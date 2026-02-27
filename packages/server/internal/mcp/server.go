// Package mcp provides MCP (Model Context Protocol) server implementation
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dnd-mcp/server/pkg/config"
	"github.com/gin-gonic/gin"
)

// Server represents an MCP server
type Server struct {
	registry    *Registry
	cfg         *config.Config
	httpServer  *http.Server
	mu          sync.RWMutex
	initialized bool
}

// ServerInfo contains server metadata
var serverInfo = ServerInfo{
	Name:    "dnd-mcp-server",
	Version: "1.0.0",
}

// NewServer creates a new MCP server
func NewServer(cfg *config.Config) *Server {
	return &Server{
		registry: NewRegistry(),
		cfg:      cfg,
	}
}

// Registry returns the tool registry for registering tools
func (s *Server) Registry() *Registry {
	return s.registry
}

// Handler returns the HTTP handler for testing purposes
// This allows tests to use the real routing configuration
func (s *Server) Handler() http.Handler {
	// Set Gin mode
	if s.cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// CORS middleware
	if s.cfg.HTTP.EnableCORS {
		router.Use(corsMiddleware())
	}

	// Register routes
	s.registerRoutes(router)

	return router
}

// RegisterTool is a convenience method to register a tool
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) error {
	return s.registry.Register(tool, handler)
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()

	// Set Gin mode
	if s.cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// CORS middleware
	if s.cfg.HTTP.EnableCORS {
		router.Use(corsMiddleware())
	}

	// Register routes
	s.registerRoutes(router)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.HTTP.Host, s.cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(s.cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.cfg.HTTP.WriteTimeout) * time.Second,
	}

	s.mu.Unlock()

	// Start server
	fmt.Printf("MCP Server starting on %s\n", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.HTTP.ShutdownTimeout)*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

// registerRoutes registers HTTP routes
func (s *Server) registerRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", s.handleHealth)

	// MCP endpoints
	router.POST("/mcp/initialize", s.handleInitialize)
	router.GET("/mcp/tools", s.handleListTools)
	router.POST("/mcp/tools/call", s.handleCallTool)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": serverInfo.Version,
	})
}

// handleInitialize handles MCP initialize requests
func (s *Server) handleInitialize(c *gin.Context) {
	var req InitializeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()

	c.JSON(http.StatusOK, InitializeResponse{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: serverInfo,
	})
}

// handleListTools handles tool listing requests
func (s *Server) handleListTools(c *gin.Context) {
	tools := s.registry.List()
	c.JSON(http.StatusOK, ListToolsResponse{
		Tools: tools,
	})
}

// handleCallTool handles tool call requests
func (s *Server) handleCallTool(c *gin.Context) {
	var req CallToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert arguments to JSON
	argsJSON, err := json.Marshal(req.Arguments)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to marshal arguments"})
		return
	}

	// Create tool request
	toolReq := ToolRequest{
		ToolName:  req.Name,
		Arguments: argsJSON,
	}

	// Call the tool
	resp := s.registry.Call(c.Request.Context(), toolReq)

	// Return response
	c.JSON(http.StatusOK, CallToolResponse{
		Content: resp.Content,
		IsError: resp.IsError,
	})
}

// corsMiddleware returns a CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
