package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/integrii/flaggy"
)

// Version is set at build time
var Version = "dev"

// Run executes the CLI with os.Args
func Run() {
	code := RunWithArgs(os.Args, os.Stdout, os.Stderr)
	os.Exit(code)
}

// RunWithArgs executes the CLI with given arguments (for testing)
func RunWithArgs(args []string, stdout, stderr io.Writer) int {
	// Reset flaggy for fresh parsing
	flaggy.ResetParser()

	flaggy.SetName("mcp-task-manager")
	flaggy.SetDescription("Task manager for Claude and coding agents")

	// Disable built-in version flag since we're using a version subcommand
	flaggy.DefaultParser.DisableShowVersionWithVersion()

	// Version subcommand
	versionCmd := flaggy.NewSubcommand("version")
	versionCmd.Description = "Show version information"
	flaggy.AttachSubcommand(versionCmd, 1)

	// List subcommand
	listCmd := flaggy.NewSubcommand("list")
	listCmd.Description = "List tasks with optional filters"
	var listStatus, listPriority, listType string
	var listJSON bool
	listCmd.String(&listStatus, "s", "status", "Filter by status (todo|in_progress|done)")
	listCmd.String(&listPriority, "p", "priority", "Filter by priority (critical|high|medium|low)")
	listCmd.String(&listType, "t", "type", "Filter by type")
	listCmd.Bool(&listJSON, "j", "json", "Output as JSON")
	flaggy.AttachSubcommand(listCmd, 1)

	// Get subcommand
	getCmd := flaggy.NewSubcommand("get")
	getCmd.Description = "Get task details by ID"
	var getIDStr string
	var getJSON bool
	getCmd.AddPositionalValue(&getIDStr, "id", 1, true, "Task ID")
	getCmd.Bool(&getJSON, "j", "json", "Output as JSON")
	flaggy.AttachSubcommand(getCmd, 1)

	// Next subcommand
	nextCmd := flaggy.NewSubcommand("next")
	nextCmd.Description = "Get highest priority todo task"
	var nextJSON bool
	nextCmd.Bool(&nextJSON, "j", "json", "Output as JSON")
	flaggy.AttachSubcommand(nextCmd, 1)

	// Create subcommand
	createCmd := flaggy.NewSubcommand("create")
	createCmd.Description = "Create a new task"
	var createTitle string
	var createPriority = "medium"
	var createType = "feature"
	var createDesc string
	var createJSON bool
	createCmd.AddPositionalValue(&createTitle, "title", 1, true, "Task title")
	createCmd.String(&createPriority, "p", "priority", "Priority (default: medium)")
	createCmd.String(&createType, "t", "type", "Type (default: feature)")
	createCmd.String(&createDesc, "d", "description", "Task description")
	createCmd.Bool(&createJSON, "j", "json", "Output as JSON")
	flaggy.AttachSubcommand(createCmd, 1)

	// Update subcommand
	updateCmd := flaggy.NewSubcommand("update")
	updateCmd.Description = "Update an existing task"
	var updateIDStr string
	var updateTitle, updateStatus, updatePriority, updateType, updateDesc string
	var updateJSON bool
	updateCmd.AddPositionalValue(&updateIDStr, "id", 1, true, "Task ID")
	updateCmd.String(&updateTitle, "", "title", "New title")
	updateCmd.String(&updateStatus, "s", "status", "New status")
	updateCmd.String(&updatePriority, "p", "priority", "New priority")
	updateCmd.String(&updateType, "t", "type", "New type")
	updateCmd.String(&updateDesc, "d", "description", "New description")
	updateCmd.Bool(&updateJSON, "j", "json", "Output as JSON")
	flaggy.AttachSubcommand(updateCmd, 1)

	// Delete subcommand
	deleteCmd := flaggy.NewSubcommand("delete")
	deleteCmd.Description = "Delete a task"
	var deleteIDStr string
	var deleteJSON bool
	var deleteForce bool
	deleteCmd.AddPositionalValue(&deleteIDStr, "id", 1, true, "Task ID")
	deleteCmd.Bool(&deleteJSON, "j", "json", "Output as JSON")
	deleteCmd.Bool(&deleteForce, "f", "force", "Force delete (also deletes subtasks)")
	flaggy.AttachSubcommand(deleteCmd, 1)

	// Start subcommand
	startCmd := flaggy.NewSubcommand("start")
	startCmd.Description = "Start a task (todo -> in_progress)"
	var startIDStr string
	var startJSON bool
	startCmd.AddPositionalValue(&startIDStr, "id", 1, true, "Task ID")
	startCmd.Bool(&startJSON, "j", "json", "Output as JSON")
	flaggy.AttachSubcommand(startCmd, 1)

	// Complete subcommand
	completeCmd := flaggy.NewSubcommand("complete")
	completeCmd.Description = "Complete a task (in_progress -> done)"
	var completeIDStr string
	var completeJSON bool
	completeCmd.AddPositionalValue(&completeIDStr, "id", 1, true, "Task ID")
	completeCmd.Bool(&completeJSON, "j", "json", "Output as JSON")
	flaggy.AttachSubcommand(completeCmd, 1)

	// Parse with custom args
	flaggy.ParseArgs(args[1:])

	// Handle subcommands
	if versionCmd.Used {
		fmt.Fprintf(stdout, "mcp-task-manager %s\n", Version)
		return 0
	}

	if listCmd.Used {
		return cmdList(stdout, stderr, listJSON, listStatus, listPriority, listType)
	}

	if getCmd.Used {
		getID, err := strconv.Atoi(getIDStr)
		if err != nil {
			fmt.Fprintf(stderr, "Error: invalid task ID: %v\n", err)
			return 1
		}
		return cmdGet(stdout, stderr, getJSON, getID)
	}

	if nextCmd.Used {
		return cmdNext(stdout, stderr, nextJSON)
	}

	if createCmd.Used {
		return cmdCreate(stdout, stderr, createJSON, createTitle, createPriority, createType, createDesc)
	}

	if updateCmd.Used {
		updateID, err := strconv.Atoi(updateIDStr)
		if err != nil {
			fmt.Fprintf(stderr, "Error: invalid task ID: %s\n", updateIDStr)
			return 1
		}
		return cmdUpdate(stdout, stderr, updateJSON, updateID, updateTitle, updateStatus, updatePriority, updateType, updateDesc)
	}

	if deleteCmd.Used {
		deleteID, err := strconv.Atoi(deleteIDStr)
		if err != nil {
			fmt.Fprintf(stderr, "Error: invalid task ID: %s\n", deleteIDStr)
			return 1
		}
		return cmdDelete(stdout, stderr, deleteJSON, deleteID, deleteForce)
	}

	if startCmd.Used {
		startID, err := strconv.Atoi(startIDStr)
		if err != nil {
			fmt.Fprintf(stderr, "Error: invalid task ID: %s\n", startIDStr)
			return 1
		}
		return cmdStart(stdout, stderr, startJSON, startID)
	}

	if completeCmd.Used {
		completeID, err := strconv.Atoi(completeIDStr)
		if err != nil {
			fmt.Fprintf(stderr, "Error: invalid task ID: %s\n", completeIDStr)
			return 1
		}
		return cmdComplete(stdout, stderr, completeJSON, completeID)
	}

	return 0
}
