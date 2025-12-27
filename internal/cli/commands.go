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

// cmdGet handles the get command
func cmdGet(stdout, stderr io.Writer, jsonOutput bool, id int) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	t, err := svc.Get(id)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if jsonOutput {
		if err := FormatJSON(stdout, t); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskDetail(t))
	}

	return 0
}

// cmdNext handles the next command
func cmdNext(stdout, stderr io.Writer, jsonOutput bool) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	t := svc.GetNextTask()
	if t == nil {
		if jsonOutput {
			FormatJSON(stdout, map[string]string{"message": "No tasks available"})
		} else {
			fmt.Fprintln(stdout, "No tasks available.")
		}
		return 0
	}

	if jsonOutput {
		if err := FormatJSON(stdout, t); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskDetail(t))
	}

	return 0
}

// cmdCreate handles the create command
func cmdCreate(stdout, stderr io.Writer, jsonOutput bool, title, priority, taskType, description string) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	t, err := svc.Create(title, description, task.Priority(priority), taskType, nil)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if jsonOutput {
		if err := FormatJSON(stdout, t); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskDetail(t))
	}

	return 0
}

// cmdUpdate handles the update command
func cmdUpdate(stdout, stderr io.Writer, jsonOutput bool, id int, title, status, priority, taskType, description string) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	var titlePtr, descPtr, typePtr *string
	var statusPtr *task.Status
	var priorityPtr *task.Priority

	if title != "" {
		titlePtr = &title
	}
	if description != "" {
		descPtr = &description
	}
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

	t, err := svc.Update(id, titlePtr, descPtr, statusPtr, priorityPtr, typePtr)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if jsonOutput {
		if err := FormatJSON(stdout, t); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskDetail(t))
	}

	return 0
}

// cmdDelete handles the delete command
func cmdDelete(stdout, stderr io.Writer, jsonOutput bool, id int, force bool) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if err := svc.Delete(id, force); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	msg := fmt.Sprintf("Task #%d deleted.", id)
	if jsonOutput {
		if err := FormatJSONMessage(stdout, msg, id); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprintln(stdout, msg)
	}

	return 0
}

// cmdStart handles the start command
func cmdStart(stdout, stderr io.Writer, jsonOutput bool, id int) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if _, err := svc.StartTask(id); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	msg := fmt.Sprintf("Task #%d started.", id)
	if jsonOutput {
		if err := FormatJSONMessage(stdout, msg, id); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprintln(stdout, msg)
	}

	return 0
}

// cmdComplete handles the complete command
func cmdComplete(stdout, stderr io.Writer, jsonOutput bool, id int) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if _, err := svc.CompleteTask(id); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	msg := fmt.Sprintf("Task #%d completed.", id)
	if jsonOutput {
		if err := FormatJSONMessage(stdout, msg, id); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprintln(stdout, msg)
	}

	return 0
}
