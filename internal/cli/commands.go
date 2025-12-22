package cli

import (
	"fmt"
	"io"

	"github.com/gpayer/mcp-task-manager/internal/config"
	"github.com/gpayer/mcp-task-manager/internal/storage"
	"github.com/gpayer/mcp-task-manager/internal/task"
)

// initService initializes the task service
func initService() (*task.Service, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	tasksDir := cfg.TasksDir()
	mdStorage := storage.NewMarkdownStorage(tasksDir)
	index := storage.NewIndex(tasksDir, mdStorage)
	svc := task.NewService(mdStorage, index, cfg.TaskTypes)

	if err := svc.Initialize(); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize: %w", err)
	}

	return svc, cfg, nil
}

// cmdList handles the list command
func cmdList(stdout, stderr io.Writer, jsonOutput bool, status, priority, taskType string) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	var statusPtr *task.Status
	var priorityPtr *task.Priority
	var typePtr *string

	if status != "" {
		s := task.Status(status)
		statusPtr = &s
	}
	if priority != "" {
		p := task.Priority(priority)
		priorityPtr = &p
	}
	if taskType != "" {
		typePtr = &taskType
	}

	tasks := svc.List(statusPtr, priorityPtr, typePtr)

	if jsonOutput {
		// Ensure we always output a JSON array, even if empty
		if tasks == nil {
			tasks = []*task.Task{}
		}
		if err := FormatJSON(stdout, tasks); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskTable(tasks))
	}

	return 0
}
