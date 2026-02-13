package tools

import (
	"context"
	"fmt"

	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRelationTools(s *server.MCPServer, svc *task.Service, relationTypes []string) {
	// add_relation
	addTool := mcp.NewTool("add_relation",
		mcp.WithDescription("Add a relation between two tasks"),
		mcp.WithNumber("source",
			mcp.Required(),
			mcp.Description("Source task ID"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Relation type"),
			mcp.Enum(relationTypes...),
		),
		mcp.WithNumber("target",
			mcp.Required(),
			mcp.Description("Target task ID"),
		),
	)
	s.AddTool(addTool, addRelationHandler(svc))

	// remove_relation
	removeTool := mcp.NewTool("remove_relation",
		mcp.WithDescription("Remove a relation between two tasks"),
		mcp.WithNumber("source",
			mcp.Required(),
			mcp.Description("Source task ID"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Relation type"),
			mcp.Enum(relationTypes...),
		),
		mcp.WithNumber("target",
			mcp.Required(),
			mcp.Description("Target task ID"),
		),
	)
	s.AddTool(removeTool, removeRelationHandler(svc))
}

func addRelationHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		source := req.GetInt("source", 0)
		relationType := req.GetString("type", "")
		target := req.GetInt("target", 0)

		if err := svc.AddRelation(source, relationType, target); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Added %s relation from task %d to task %d", relationType, source, target)), nil
	}
}

func removeRelationHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		source := req.GetInt("source", 0)
		relationType := req.GetString("type", "")
		target := req.GetInt("target", 0)

		if err := svc.RemoveRelation(source, relationType, target); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Removed %s relation from task %d to task %d", relationType, source, target)), nil
	}
}
