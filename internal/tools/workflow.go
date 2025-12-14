package tools

import (
	"context"

	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerWorkflowTools(s *server.MCPServer, svc *task.Service) {
	// get_next_task
	nextTool := mcp.NewTool("get_next_task",
		mcp.WithDescription("Get the highest priority todo task for an agent to work on"),
	)
	s.AddTool(nextTool, getNextTaskHandler(svc))

	// start_task
	startTool := mcp.NewTool("start_task",
		mcp.WithDescription("Move a task from todo to in_progress"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID to start"),
		),
	)
	s.AddTool(startTool, startTaskHandler(svc))

	// complete_task
	completeTool := mcp.NewTool("complete_task",
		mcp.WithDescription("Move a task from in_progress to done"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID to complete"),
		),
	)
	s.AddTool(completeTool, completeTaskHandler(svc))
}

func getNextTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		t := svc.GetNextTask()
		if t == nil {
			return mcp.NewToolResultText("No tasks available"), nil
		}
		return taskResult(t)
	}
}

func startTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetInt("id", 0)

		t, err := svc.StartTask(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}

func completeTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetInt("id", 0)

		t, err := svc.CompleteTask(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}
