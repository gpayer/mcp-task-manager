package tools

import (
	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/mark3labs/mcp-go/server"
)

// Register registers all MCP tools with the server
func Register(s *server.MCPServer, svc *task.Service, validTypes []string) {
	registerManagementTools(s, svc, validTypes)
	// registerWorkflowTools will be added in Task 8
	// registerWorkflowTools(s, svc)
}

// Temporary stub for registerWorkflowTools (will be implemented in Task 8)
func registerWorkflowTools(s *server.MCPServer, svc *task.Service) {
	// TODO: Implement workflow tools
}
