// Copyright 2026 boris-korol. Licensed under Apache-2.0. See LICENSE.

package main

import (
	"fmt"
	"os"

	mcptools "github.com/BKproduct/kie-cli/internal/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"Kie Ai",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	mcptools.RegisterTools(s)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}
