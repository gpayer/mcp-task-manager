package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerManagementTools(s *server.MCPServer, svc *task.Service, validTypes []string) {
	// create_task
	createTool := mcp.NewTool("create_task",
		mcp.WithDescription("Create a new task"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Task title"),
		),
		mcp.WithString("description",
			mcp.Description("Task description (markdown supported)"),
		),
		mcp.WithString("priority",
			mcp.Required(),
			mcp.Description("Task priority"),
			mcp.Enum("critical", "high", "medium", "low"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Task type"),
			mcp.Enum(validTypes...),
		),
		mcp.WithNumber("parent_id",
			mcp.Description("Parent task ID (creates a subtask)"),
		),
	)
	s.AddTool(createTool, createTaskHandler(svc))

	// get_task
	getTool := mcp.NewTool("get_task",
		mcp.WithDescription("Get a task by ID"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
	)
	s.AddTool(getTool, getTaskHandler(svc))

	// update_task
	updateTool := mcp.NewTool("update_task",
		mcp.WithDescription("Update an existing task"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
		mcp.WithString("title",
			mcp.Description("New title"),
		),
		mcp.WithString("description",
			mcp.Description("New description"),
		),
		mcp.WithString("status",
			mcp.Description("New status"),
			mcp.Enum("todo", "in_progress", "done"),
		),
		mcp.WithString("priority",
			mcp.Description("New priority"),
			mcp.Enum("critical", "high", "medium", "low"),
		),
		mcp.WithString("type",
			mcp.Description("New task type"),
			mcp.Enum(validTypes...),
		),
	)
	s.AddTool(updateTool, updateTaskHandler(svc))

	// delete_task
	deleteTool := mcp.NewTool("delete_task",
		mcp.WithDescription("Delete a task"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
		mcp.WithBoolean("delete_subtasks",
			mcp.Description("If true, also delete all subtasks (required if task has subtasks)"),
		),
	)
	s.AddTool(deleteTool, deleteTaskHandler(svc))

	// list_tasks
	listTool := mcp.NewTool("list_tasks",
		mcp.WithDescription("List tasks with optional filters"),
		mcp.WithString("status",
			mcp.Description("Filter by status"),
			mcp.Enum("todo", "in_progress", "done"),
		),
		mcp.WithString("priority",
			mcp.Description("Filter by priority"),
			mcp.Enum("critical", "high", "medium", "low"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by task type"),
			mcp.Enum(validTypes...),
		),
	)
	s.AddTool(listTool, listTasksHandler(svc))
}

func createTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title := req.GetString("title", "")
		description := req.GetString("description", "")
		priority := task.Priority(req.GetString("priority", ""))
		taskType := req.GetString("type", "")

		var parentID *int
		args := req.GetArguments()
		if _, ok := args["parent_id"]; ok {
			id := req.GetInt("parent_id", 0)
			parentID = &id
		}

		t, err := svc.Create(title, description, priority, taskType, parentID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}

func getTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetInt("id", 0)

		t, err := svc.Get(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}

func updateTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetInt("id", 0)

		var title, description, taskType *string
		var status *task.Status
		var priority *task.Priority

		args := req.GetArguments()
		if _, ok := args["title"]; ok {
			v := req.GetString("title", "")
			title = &v
		}
		if _, ok := args["description"]; ok {
			v := req.GetString("description", "")
			description = &v
		}
		if _, ok := args["status"]; ok {
			s := task.Status(req.GetString("status", ""))
			status = &s
		}
		if _, ok := args["priority"]; ok {
			p := task.Priority(req.GetString("priority", ""))
			priority = &p
		}
		if _, ok := args["type"]; ok {
			v := req.GetString("type", "")
			taskType = &v
		}

		t, err := svc.Update(id, title, description, status, priority, taskType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}

func deleteTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetInt("id", 0)
		deleteSubtasks := req.GetBool("delete_subtasks", false)

		if err := svc.Delete(id, deleteSubtasks); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Task %d deleted", id)), nil
	}
}

func listTasksHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var status *task.Status
		var priority *task.Priority
		var taskType *string

		args := req.GetArguments()
		if _, ok := args["status"]; ok {
			s := task.Status(req.GetString("status", ""))
			status = &s
		}
		if _, ok := args["priority"]; ok {
			p := task.Priority(req.GetString("priority", ""))
			priority = &p
		}
		if _, ok := args["type"]; ok {
			v := req.GetString("type", "")
			taskType = &v
		}

		tasks := svc.List(status, priority, taskType, nil)

		if len(tasks) == 0 {
			return mcp.NewToolResultText("No tasks found"), nil
		}

		data, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

func taskResult(t *task.Task) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
