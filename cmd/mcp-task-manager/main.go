package main

import (
	"log"

	"github.com/gpayer/mcp-task-manager/internal/config"
	"github.com/gpayer/mcp-task-manager/internal/storage"
	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/gpayer/mcp-task-manager/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	tasksDir := cfg.TasksDir()
	mdStorage := storage.NewMarkdownStorage(tasksDir)
	index := storage.NewIndex(tasksDir, mdStorage)

	// Initialize task service
	svc := task.NewService(mdStorage, index, cfg.TaskTypes)
	if err := svc.Initialize(); err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"mcp-task-manager",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	// Register tools
	tools.Register(s, svc, cfg.TaskTypes)

	// Start server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
