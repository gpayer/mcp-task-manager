package main

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"mcp-task-manager",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
