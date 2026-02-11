// Package postgres 提供数据库迁移功能
package postgres

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// Migration 迁移记录
type Migration struct {
	Version   int64
	Name      string
	Applied   bool
	AppliedAt time.Time
}

// Migrator 数据库迁移器
type Migrator struct {
	client         *Client
	migrationsPath string // 迁移文件路径
}

// NewMigrator 创建迁移器
func NewMigrator(client *Client) *Migrator {
	return &Migrator{
		client:         client,
		migrationsPath: "scripts/migrations",
	}
}

// Up 执行所有待执行的迁移
func (m *Migrator) Up(ctx context.Context) error {
	// 1. 确保迁移历史表存在
	if err := m.ensureSchemaMigrationsTable(ctx); err != nil {
		return fmt.Errorf("创建迁移历史表失败: %w", err)
	}

	// 2. 获取已应用的迁移版本
	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("获取已应用迁移失败: %w", err)
	}

	// 3. 读取迁移文件
	migrations, err := m.readMigrationFiles()
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %w", err)
	}

	// 4. 执行未应用的迁移
	appliedCount := 0
	for _, migration := range migrations {
		if _, applied := appliedVersions[migration.Version]; applied {
			continue // 已应用，跳过
		}

		// 读取并执行迁移 SQL
		if err := m.applyMigration(ctx, &migration); err != nil {
			return fmt.Errorf("应用迁移 %d 失败: %w", migration.Version, err)
		}

		appliedCount++
		fmt.Printf("✓ 迁移 %d 成功应用: %s\n", migration.Version, migration.Name)
	}

	if appliedCount == 0 {
		fmt.Println("数据库已是最新版本，无需迁移")
	} else {
		fmt.Printf("成功应用 %d 个迁移\n", appliedCount)
	}

	return nil
}

// Down 回滚最后一次迁移
func (m *Migrator) Down(ctx context.Context) error {
	// 1. 确保迁移历史表存在
	if err := m.ensureSchemaMigrationsTable(ctx); err != nil {
		return fmt.Errorf("创建迁移历史表失败: %w", err)
	}

	// 2. 获取已应用的迁移版本
	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("获取已应用迁移失败: %w", err)
	}

	if len(appliedVersions) == 0 {
		fmt.Println("没有可回滚的迁移")
		return nil
	}

	// 3. 获取最新的迁移版本
	var latestVersion int64
	for version := range appliedVersions {
		if version > latestVersion {
			latestVersion = version
		}
	}

	// 4. 读取迁移文件
	migrations, err := m.readMigrationFiles()
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %w", err)
	}

	// 5. 找到对应的迁移并执行 down SQL
	var targetMigration *Migration
	for i := range migrations {
		if migrations[i].Version == latestVersion {
			targetMigration = &migrations[i]
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("找不到版本 %d 的迁移文件", latestVersion)
	}

	// 执行回滚
	if err := m.rollbackMigration(ctx, targetMigration); err != nil {
		return fmt.Errorf("回滚迁移 %d 失败: %w", latestVersion, err)
	}

	fmt.Printf("✓ 成功回滚迁移 %d: %s\n", latestVersion, targetMigration.Name)
	return nil
}

// Status 查看迁移状态
func (m *Migrator) Status(ctx context.Context) ([]Migration, error) {
	// 1. 确保迁移历史表存在
	if err := m.ensureSchemaMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("创建迁移历史表失败: %w", err)
	}

	// 2. 获取已应用的迁移
	appliedVersions, err := m.getAppliedVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取已应用迁移失败: %w", err)
	}

	// 3. 读取迁移文件
	migrations, err := m.readMigrationFiles()
	if err != nil {
		return nil, fmt.Errorf("读取迁移文件失败: %w", err)
	}

	// 4. 合并状态
	for i := range migrations {
		if appliedAt, applied := appliedVersions[migrations[i].Version]; applied {
			migrations[i].Applied = true
			migrations[i].AppliedAt = appliedAt
		}
	}

	return migrations, nil
}

// ensureSchemaMigrationsTable 确保迁移历史表存在
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

// getAppliedVersions 获取已应用的迁移版本
func (m *Migrator) getAppliedVersions(ctx context.Context) (map[int64]time.Time, error) {
	query := `SELECT version, applied_at FROM schema_migrations ORDER BY version`
	rows, err := m.client.Pool().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[int64]time.Time)
	for rows.Next() {
		var version int64
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		versions[version] = appliedAt
	}

	return versions, rows.Err()
}

// readMigrationFiles 读取迁移文件
func (m *Migrator) readMigrationFiles() ([]Migration, error) {
	// 读取迁移文件目录
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

		// 解析版本号和名称
		// 格式: {version}_{description}.up.sql
		// 例如: 000001_initial_schema.up.sql
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

	// 按版本号排序
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// applyMigration 应用单个迁移
func (m *Migrator) applyMigration(ctx context.Context, migration *Migration) error {
	// 读取 up 文件
	upFileName := fmt.Sprintf("%s/%06d_%s.up.sql", m.migrationsPath, migration.Version, migration.Name)
	upSQL, err := os.ReadFile(upFileName)
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %w", err)
	}

	// 执行迁移 SQL（使用事务）
	pool := m.client.Pool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback(ctx)

	// 执行 up SQL
	if _, err := tx.Exec(ctx, string(upSQL)); err != nil {
		return fmt.Errorf("执行迁移 SQL 失败: %w", err)
	}

	// 记录迁移历史
	insertQuery := `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`
	if _, err := tx.Exec(ctx, insertQuery, migration.Version, time.Now()); err != nil {
		return fmt.Errorf("记录迁移历史失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// rollbackMigration 回滚单个迁移
func (m *Migrator) rollbackMigration(ctx context.Context, migration *Migration) error {
	// 读取 down 文件
	downFileName := fmt.Sprintf("%s/%06d_%s.down.sql", m.migrationsPath, migration.Version, migration.Name)
	downSQL, err := os.ReadFile(downFileName)
	if err != nil {
		return fmt.Errorf("读取回滚文件失败: %w", err)
	}

	// 执行回滚 SQL（使用事务）
	pool := m.client.Pool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback(ctx)

	// 执行 down SQL
	if _, err := tx.Exec(ctx, string(downSQL)); err != nil {
		return fmt.Errorf("执行回滚 SQL 失败: %w", err)
	}

	// 删除迁移历史
	deleteQuery := `DELETE FROM schema_migrations WHERE version = $1`
	if _, err := tx.Exec(ctx, deleteQuery, migration.Version); err != nil {
		return fmt.Errorf("删除迁移历史失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetCurrentVersion 获取当前数据库版本
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

// GetLatestVersion 获取最新迁移文件版本
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

// IsUpToDate 检查数据库是否为最新版本
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
