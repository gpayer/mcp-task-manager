package tools

import (
	"fmt"
	"strings"

	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/mark3labs/mcp-go/server"
)

// Register registers all MCP tools with the server
func Register(s *server.MCPServer, svc *task.Service, validTypes []string, relationTypes []string) {
	registerManagementTools(s, svc, validTypes)
	registerWorkflowTools(s, svc)
	registerRelationTools(s, svc, relationTypes)
}

func allowedValuesDescription(label string, values []string) string {
	return fmt.Sprintf("%s Allowed values: %s.", label, strings.Join(values, ", "))
}
