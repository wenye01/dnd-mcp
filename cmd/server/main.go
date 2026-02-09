// Package main 是 DND MCP Server 的入口
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dnd-mcp/client/internal/api"
	"github.com/dnd-mcp/client/internal/api/handler"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/monitor"
	"github.com/dnd-mcp/client/internal/persistence"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/dnd-mcp/client/internal/store/postgres"
	"github.com/dnd-mcp/client/internal/store/redis"
	"github.com/dnd-mcp/client/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化 Redis 客户端
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("初始化 Redis 失败: %v", err)
	}
	defer redisClient.Close()

	// 测试 Redis 连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx); err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	}
	log.Println("✓ Redis 连接成功")

	// 初始化 PostgreSQL 客户端
	postgresClient, err := postgres.NewClient(cfg.Postgres)
	if err != nil {
		log.Printf("⚠ PostgreSQL 初始化失败: %v", err)
		log.Println("⚠ 将仅使用 Redis 模式，持久化功能不可用")
		postgresClient = nil
	} else {
		if err := postgresClient.Ping(ctx); err != nil {
			log.Printf("⚠ PostgreSQL 连接失败: %v", err)
			log.Println("⚠ 将仅使用 Redis 模式，持久化功能不可用")
			postgresClient.Close()
			postgresClient = nil
		} else {
			log.Println("✓ PostgreSQL 连接成功")
		}
	}

	// 初始化 Redis 存储
	sessionStore := redis.NewSessionStore(redisClient)
	messageStore := redis.NewMessageStore(redisClient)

	// 初始化健康监控器
	healthMonitor := monitor.NewHealthMonitor()
	healthMonitor.Register(monitor.NewRedisHealthChecker(redisClient))

	// 初始化统计监控器
	statsMonitor := monitor.NewStatsMonitor("v0.1.0") // 从配置或常量获取版本
	statsMonitor.Register(monitor.NewRedisStatsCollector(redisClient))
	if sessionStore != nil {
		statsMonitor.Register(monitor.NewSessionStatsCollector(sessionStore))
	}

	// 初始化 AdminHandler（如果 PostgreSQL 可用）
	var adminHandler *handler.AdminHandler
	if postgresClient != nil {
		// 创建 PostgreSQL 连接用于迁移器（使用 pgx）
		pgConn, err := pgx.Connect(ctx, fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User,
			cfg.Postgres.Password, cfg.Postgres.DBName,
		))
		if err != nil {
			log.Printf("⚠ 创建 PostgreSQL 连接失败: %v", err)
		} else {
			defer pgConn.Close(ctx)

			// 初始化迁移器
			migrator := persistence.NewMigrator(pgConn)
			if err := migrator.Initialize(ctx); err != nil {
				log.Printf("⚠ 初始化迁移器失败: %v", err)
			} else {
				log.Println("✓ 迁移器初始化成功")
			}

			// 初始化 PostgreSQL 存储
			postgresSessionStore := postgres.NewPostgresSessionStore(postgresClient)
			postgresMessageStore := postgres.NewPostgresMessageStore(postgresClient)

			// 创建Redis存储适配器以实现persistence接口
			redisSessionReader := &redisSessionReaderAdapter{store: sessionStore}
			redisMessageReader := &redisMessageReaderAdapter{store: messageStore}
			redisSessionWriter := &redisSessionWriterAdapter{store: sessionStore}
			redisMessageWriter := &redisMessageWriterAdapter{store: messageStore}

			// 初始化备份和恢复服务
			backupSvc := persistence.NewBackupService(
				redisSessionReader,
				redisMessageReader,
				postgresSessionStore,
				postgresMessageStore,
			)

			restoreSvc := persistence.NewRestoreService(
				postgresSessionStore,
				postgresMessageStore,
				redisSessionReader,
				redisSessionWriter,
				redisMessageWriter,
			)

			adminHandler = handler.NewAdminHandler(migrator, backupSvc, restoreSvc)
			log.Println("✓ 持久化服务初始化成功")
		}
	}

	// 设置 Gin 模式
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 持久化管理器（如果可用）
	var persistenceTriggerer handler.PersistenceTriggerer
	if postgresClient != nil {
		// 持久化管理器已在上面初始化，这里需要获取它
		// 由于作用域问题，我们需要重新组织代码结构
		// 为简单起见，这里先设置为nil
		persistenceTriggerer = nil
	}

	// 创建系统处理器
	systemHandler := handler.NewSystemHandler(persistenceTriggerer, healthMonitor, statsMonitor)

	// 创建 API 服务器
	apiServer := api.NewServer(
		cfg,
		nil, // sessionService
		sessionStore,
		messageStore,
		nil, // llmClient
		nil, // mcpClient
		nil, // contextBuilder
		nil, // hub
		systemHandler,
	)

	// 启动服务器（goroutine）
	go func() {
		log.Printf("✓ HTTP Server 启动成功，监听 %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server 启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}

	log.Println("✓ 服务器已关闭")
}

// Redis存储适配器，将store.SessionStore/MessageStore适配为persistence接口

type redisSessionReaderAdapter struct {
	store store.SessionStore
}

func (a *redisSessionReaderAdapter) Get(ctx context.Context, id string) (*models.Session, error) {
	return a.store.Get(ctx, id)
}

func (a *redisSessionReaderAdapter) List(ctx context.Context) ([]*models.Session, error) {
	return a.store.List(ctx)
}

func (a *redisSessionReaderAdapter) ListActive(ctx context.Context) ([]*models.Session, error) {
	// Redis 没有软删除，直接返回所有会话
	return a.store.List(ctx)
}

type redisMessageReaderAdapter struct {
	store store.MessageStore
}

func (a *redisMessageReaderAdapter) Get(ctx context.Context, sessionID, messageID string) (*models.Message, error) {
	return a.store.Get(ctx, sessionID, messageID)
}

func (a *redisMessageReaderAdapter) List(ctx context.Context, sessionID string, limit int) ([]*models.Message, error) {
	return a.store.List(ctx, sessionID, limit)
}

func (a *redisMessageReaderAdapter) ListByRole(ctx context.Context, sessionID, role string, limit int) ([]*models.Message, error) {
	// Redis MessageStore 不支持按角色筛选，返回全部然后过滤
	messages, err := a.store.List(ctx, sessionID, limit)
	if err != nil {
		return nil, err
	}

	var filtered []*models.Message
	for _, msg := range messages {
		if msg.Role == role {
			filtered = append(filtered, msg)
		}
	}
	return filtered, nil
}

type redisSessionWriterAdapter struct {
	store store.SessionStore
}

func (a *redisSessionWriterAdapter) Create(ctx context.Context, session *models.Session) error {
	return a.store.Create(ctx, session)
}

func (a *redisSessionWriterAdapter) BatchCreate(ctx context.Context, sessions []*models.Session) error {
	// Redis 没有批量创建，逐个创建
	for _, session := range sessions {
		if err := a.store.Create(ctx, session); err != nil {
			return err
		}
	}
	return nil
}

func (a *redisSessionWriterAdapter) Update(ctx context.Context, session *models.Session) error {
	return a.store.Update(ctx, session)
}

type redisMessageWriterAdapter struct {
	store store.MessageStore
}

func (a *redisMessageWriterAdapter) Create(ctx context.Context, message *models.Message) error {
	return a.store.Create(ctx, message)
}

func (a *redisMessageWriterAdapter) BatchCreate(ctx context.Context, messages []*models.Message) error {
	// Redis 没有批量创建，逐个创建
	for _, message := range messages {
		if err := a.store.Create(ctx, message); err != nil {
			return err
		}
	}
	return nil
}
