// Package postgres provides database migration functionality
package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Migration represents a migration record
type Migration struct {
	Version   int64
	Name      string
	Applied   bool
	AppliedAt time.Time
}

// Migrator handles database migrations
type Migrator struct {
	client         *Client
	migrationsPath string
}

// NewMigrator creates a new migrator
func NewMigrator(client *Client) *Migrator {
	// Resolve migrations path relative to executable
	// Try multiple paths in order:
	// 1. Relative to working directory (for development)
	// 2. Relative to executable (for production builds)
	// 3. Multiple candidate paths for different scenarios

	// Candidate paths to try
	candidatePaths := []string{
		// When running from packages/server directory
		"internal/store/postgres/migrations",
		// When running from packages/server/bin directory
		"../internal/store/postgres/migrations",
		// When running from project root
		"packages/server/internal/store/postgres/migrations",
		// When running from packages/bin directory
		"../server/internal/store/postgres/migrations",
	}

	// Try each candidate path
	for _, candidatePath := range candidatePaths {
		if _, err := os.Stat(candidatePath); err == nil {
			return &Migrator{
				client:         client,
				migrationsPath: candidatePath,
			}
		}
	}

	// Try path relative to executable as last resort
	exePath, err := os.Executable()
	if err == nil {
		binDir := filepath.Dir(exePath)
		// Try relative to executable
		for _, relPath := range []string{"../internal/store/postgres/migrations", "internal/store/postgres/migrations"} {
			candidatePath := filepath.Join(binDir, relPath)
			if _, err := os.Stat(candidatePath); err == nil {
				return &Migrator{
					client:         client,
					migrationsPath: candidatePath,
				}
			}
		}
	}

	// Fallback to relative path
	return &Migrator{
		client:         client,
		migrationsPath: "internal/store/postgres/migrations",
	}
}

// NewMigratorWithPath creates a new migrator with a custom migrations path
func NewMigratorWithPath(client *Client, path string) *Migrator {
	return &Migrator{
		client:         client,
		migrationsPath: path,
	}
}

// Up executes all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	// 1. Ensure schema_migrations table exists
	if err := m.ensureSchemaMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 2. Get applied migration versions
	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 3. Read migration files
	migrations, err := m.readMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// 4. Execute pending migrations
	appliedCount := 0
	for _, migration := range migrations {
		fmt.Printf("DEBUG: Checking migration %d (name: %s), appliedVersions: %v\n", migration.Version, migration.Name, appliedVersions)
		if _, applied := appliedVersions[migration.Version]; applied {
			fmt.Printf("DEBUG: Skipping migration %d (already applied)\n", migration.Version)
			continue // Already applied, skip
		}

		// Read and execute migration SQL
		if err := m.applyMigration(ctx, &migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		appliedCount++
		fmt.Printf("✓ Migration %d applied: %s\n", migration.Version, migration.Name)
	}

	if appliedCount == 0 {
		fmt.Println("Database is up to date, no migrations needed")
	} else {
		fmt.Printf("Successfully applied %d migrations\n", appliedCount)
	}

	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down(ctx context.Context) error {
	// 1. Ensure schema_migrations table exists
	if err := m.ensureSchemaMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 2. Get applied migration versions
	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(appliedVersions) == 0 {
		fmt.Println("No migrations to roll back")
		return nil
	}

	// 3. Get the latest migration version
	var latestVersion int64
	for version := range appliedVersions {
		if version > latestVersion {
			latestVersion = version
		}
	}

	// 4. Read migration files
	migrations, err := m.readMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// 5. Find the target migration and execute down SQL
	var targetMigration *Migration
	for i := range migrations {
		if migrations[i].Version == latestVersion {
			targetMigration = &migrations[i]
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration file for version %d not found", latestVersion)
	}

	// Execute rollback
	if err := m.rollbackMigration(ctx, targetMigration); err != nil {
		return fmt.Errorf("failed to roll back migration %d: %w", latestVersion, err)
	}

	fmt.Printf("✓ Successfully rolled back migration %d: %s\n", latestVersion, targetMigration.Name)
	return nil
}

// Status returns migration status
func (m *Migrator) Status(ctx context.Context) ([]Migration, error) {
	// 1. Ensure schema_migrations table exists
	if err := m.ensureSchemaMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 2. Get applied migrations
	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 3. Read migration files
	migrations, err := m.readMigrationFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to read migration files: %w", err)
	}

	// 4. Merge status
	for i := range migrations {
		if appliedAt, applied := appliedVersions[migrations[i].Version]; applied {
			migrations[i].Applied = true
			migrations[i].AppliedAt = appliedAt
		}
	}

	return migrations, nil
}

// ensureSchemaMigrationsTable ensures the schema_migrations table exists
func (m *Migrator) ensureSchemaMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL
		);
	`
	_, err := m.client.Pool().Exec(ctx, query)
	return err
}

// getAppliedVersions gets applied migration versions
func (m *Migrator) getAppliedVersions(ctx context.Context) (map[int64]time.Time, error) {
	query := `SELECT version, applied_at FROM schema_migrations ORDER BY version`
	rows, err := m.client.Pool().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[int64]time.Time)
	for rows.Next() {
		var versionStr string
		var appliedAt time.Time
		if err := rows.Scan(&versionStr, &appliedAt); err != nil {
			return nil, err
		}
		var version int64
		if _, err := fmt.Sscanf(versionStr, "%d", &version); err != nil {
			return nil, fmt.Errorf("failed to parse version %q: %w", versionStr, err)
		}
		versions[version] = appliedAt
	}

	return versions, rows.Err()
}

// readMigrationFiles reads migration files
func (m *Migrator) readMigrationFiles() ([]Migration, error) {
	// Read migration directory
	entries, err := os.ReadDir(m.migrationsPath)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		// Parse version and name
		// Format: {version}_{description}.up.sql
		// Example: 000001_initial_schema.up.sql
		baseName := strings.TrimSuffix(name, ".up.sql")
		parts := strings.SplitN(baseName, "_", 2)
		if len(parts) < 2 {
			continue
		}

		var version int64
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			continue
		}

		description := parts[1]
		migrations = append(migrations, Migration{
			Version: version,
			Name:    description,
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(ctx context.Context, migration *Migration) error {
	// Read up file
	upFileName := fmt.Sprintf("%s/%06d_%s.up.sql", m.migrationsPath, migration.Version, migration.Name)
	fmt.Printf("DEBUG: Attempting to read migration file: %s\n", upFileName)
	upSQL, err := os.ReadFile(upFileName)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration SQL (with transaction)
	pool := m.client.Pool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Execute up SQL
	if _, err := tx.Exec(ctx, string(upSQL)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration history
	insertQuery := `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`
	if _, err := tx.Exec(ctx, insertQuery, migration.Version, time.Now()); err != nil {
		return fmt.Errorf("failed to record migration history: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(ctx context.Context, migration *Migration) error {
	// Read down file
	downFileName := fmt.Sprintf("%s/%06d_%s.down.sql", m.migrationsPath, migration.Version, migration.Name)
	downSQL, err := os.ReadFile(downFileName)
	if err != nil {
		return fmt.Errorf("failed to read rollback file: %w", err)
	}

	// Execute rollback SQL (with transaction)
	pool := m.client.Pool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Execute down SQL
	if _, err := tx.Exec(ctx, string(downSQL)); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	// Delete migration history
	deleteQuery := `DELETE FROM schema_migrations WHERE version = $1`
	if _, err := tx.Exec(ctx, deleteQuery, migration.Version); err != nil {
		return fmt.Errorf("failed to delete migration history: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetCurrentVersion returns the current database version
func (m *Migrator) GetCurrentVersion(ctx context.Context) (int64, error) {
	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return 0, err
	}

	if len(appliedVersions) == 0 {
		return 0, nil
	}

	var latestVersion int64
	for version := range appliedVersions {
		if version > latestVersion {
			latestVersion = version
		}
	}

	return latestVersion, nil
}

// GetLatestVersion returns the latest migration file version
func (m *Migrator) GetLatestVersion(ctx context.Context) (int64, error) {
	migrations, err := m.readMigrationFiles()
	if err != nil {
		return 0, err
	}

	if len(migrations) == 0 {
		return 0, nil
	}

	return migrations[len(migrations)-1].Version, nil
}

// IsUpToDate checks if database is up to date
func (m *Migrator) IsUpToDate(ctx context.Context) (bool, error) {
	current, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return false, err
	}

	latest, err := m.GetLatestVersion(ctx)
	if err != nil {
		return false, err
	}

	return current == latest, nil
}

// RunMigrations is a convenience function to run all pending migrations
func RunMigrations(ctx context.Context, client *Client) error {
	migrator := NewMigrator(client)
	return migrator.Up(ctx)
}
