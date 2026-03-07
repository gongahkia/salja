package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mark3labs/mcp-go/server"

	"github.com/gongahkia/salja/internal/config"
	"github.com/gongahkia/salja/internal/logging"
	mcpserver "github.com/gongahkia/salja/internal/mcp"
)

var version = "dev"

func main() {
	logPath := filepath.Join(config.DataDir(), "salja.log")
	if err := logging.Init(logPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to init log: %v\n", err)
	}
	defer logging.Shutdown()
	logging.Default().Info("system", "salja-mcp started")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv := mcpserver.NewServer(version)
	stdio := server.NewStdioServer(srv)

	if err := stdio.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
