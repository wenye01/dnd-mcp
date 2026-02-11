// Package persistence 提供数据持久化和迁移功能
package persistence

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migration 迁移记录
type Migration struct {
	Version   string
	Name      string
	AppliedAt bool
}

// Migrator 数据库迁移器
type Migrator struct {
	db *pgx.Conn
}

// NewMigrator 创建迁移器
func NewMigrator(db *pgx.Conn) *Migrator {
	return &Migrator{db: db}
}

// Initialize 初始化迁移表
func (m *Migrator) Initialize(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := m.db.Exec(ctx, query)
	return err
}

// GetAppliedMigrations 获取已应用的迁移
func (m *Migrator) GetAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	query := `SELECT version FROM schema_migrations;`
	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// GetPendingMigrations 获取待应用的迁移
func (m *Migrator) GetPendingMigrations(ctx context.Context) ([]Migration, error) {
	// 获取已应用的迁移
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取已应用迁移失败: %w", err)
	}

	// 读取所有迁移文件
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("读取迁移文件失败: %w", err)
	}

	var pending []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), ".sql")
		if !applied[version] {
			pending = append(pending, Migration{
				Version:   version,
				Name:      entry.Name(),
				AppliedAt: false,
			})
		}
	}

	// 按版本号排序
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	return pending, nil
}

// GetStatus 获取迁移状态
func (m *Migrator) GetStatus(ctx context.Context) ([]Migration, error) {
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	// 读取所有迁移文件
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), ".sql")
		migrations = append(migrations, Migration{
			Version:   version,
			Name:      entry.Name(),
			AppliedAt: applied[version],
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// Up 应用所有待应用的迁移
func (m *Migrator) Up(ctx context.Context) error {
	pending, err := m.GetPendingMigrations(ctx)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		log.Println("✓ 所有迁移都已应用，无需执行")
		return nil
	}

	log.Printf("发现 %d 个待应用的迁移", len(pending))

	for _, migration := range pending {
		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("应用迁移 %s 失败: %w", migration.Version, err)
		}
		log.Printf("✓ 迁移 %s 应用成功", migration.Version)
	}

	return nil
}

// ApplyVersion 应用指定版本的迁移
func (m *Migrator) ApplyVersion(ctx context.Context, version string) error {
	// 读取迁移文件
	filename := version + ".sql"
	content, err := migrationFS.ReadFile("migrations/" + filename)
	if err != nil {
		return fmt.Errorf("读取迁移文件失败: %w", err)
	}

	// 在事务中执行迁移
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback(ctx)

	// 执行迁移 SQL
	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("执行迁移 SQL 失败: %w", err)
	}

	// 记录迁移
	recordQuery := `INSERT INTO schema_migrations (version) VALUES ($1);`
	if _, err := tx.Exec(ctx, recordQuery, version); err != nil {
		return fmt.Errorf("记录迁移失败: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	log.Printf("✓ 迁移 %s 应用成功", version)
	return nil
}

// applyMigration 应用单个迁移
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	return m.ApplyVersion(ctx, migration.Version)
}
