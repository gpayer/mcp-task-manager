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

	return 0
}
