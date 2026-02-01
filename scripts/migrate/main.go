// Package main 数据库迁移工具
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func main() {
	// 解析命令行参数
	action := flag.String("action", "up", "迁移动作: up 或 down")
	dsn := flag.String("dsn", "", "数据库连接字符串 (默认从环境变量POSTGRESQL_URL读取)")
	migrationsDir := flag.String("dir", "./migrations", "迁移文件目录")
	flag.Parse()

	// 获取数据库连接字符串
	dataSourceName := *dsn
	if dataSourceName == "" {
		dataSourceName = os.Getenv("POSTGRESQL_URL")
	}
	if dataSourceName == "" {
		log.Fatal("请提供数据库连接字符串: -dsn 或设置环境变量POSTGRESQL_URL")
	}

	// 连接数据库
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 验证连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	log.Println("数据库连接成功")

	// 执行迁移
	if *action == "up" {
		if err := runMigrations(db, *migrationsDir, "up"); err != nil {
			log.Fatalf("迁移失败: %v", err)
		}
		log.Println("迁移完成")
	} else if *action == "down" {
		if err := runMigrations(db, *migrationsDir, "down"); err != nil {
			log.Fatalf("回滚失败: %v", err)
		}
		log.Println("回滚完成")
	} else {
		log.Fatalf("未知的动作: %s (支持: up, down)", *action)
	}
}

// runMigrations 执行迁移
func runMigrations(db *sql.DB, dir, action string) error {
	// 查找迁移文件
	files, err := filepath.Glob(filepath.Join(dir, fmt.Sprintf("*.%s.sql", action)))
	if err != nil {
		return fmt.Errorf("查找迁移文件失败: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("未找到迁移文件")
	}

	log.Printf("找到 %d 个迁移文件", len(files))

	// 执行迁移文件
	for _, file := range files {
		log.Printf("执行迁移: %s", filepath.Base(file))

		if err := executeSQLFile(db, file); err != nil {
			return fmt.Errorf("执行 %s 失败: %w", file, err)
		}
	}

	return nil
}

// executeSQLFile 执行SQL文件
func executeSQLFile(db *sql.DB, filename string) error {
	// 读取SQL文件
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 执行SQL
	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("执行SQL失败: %w", err)
	}

	return nil
}
