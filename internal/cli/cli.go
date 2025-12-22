package cli

import (
	"fmt"
	"io"
	"os"

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
	// Enable panic instead of exit for testing
	flaggy.PanicInsteadOfExit = true

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

	return 0
}
