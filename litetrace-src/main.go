// Package main provides the entry point for the litetrace application.
// Litetrace is a lightweight ftrace tool written in pure Go for Linux kernel tracing.
//
// The application requires root privileges to access the ftrace subsystem via tracefs.
// It provides graceful shutdown handling to ensure kernel state is restored on exit.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/M410550/lite-tracer-mygo/cmd"
	"github.com/M410550/lite-tracer-mygo/internal/ftrace"
)

// Global variables for graceful shutdown handling
var (
	cancel context.CancelFunc
	// engine is the global ftrace engine instance, used for emergency shutdown
	engine *ftrace.Engine
)

// SetupCloseHandler sets up a signal handler for graceful shutdown.
// It captures SIGINT (Ctrl+C) and SIGTERM signals to perform cleanup
// before exiting the application.
//
// The handler performs the following actions on signal reception:
// 1. Cancels the global context to stop ongoing operations
// 2. Calls SafeShutdown on the engine to restore kernel state
// 3. Prints status messages and exits cleanly
func SetupCloseHandler() {
	// Create a buffered channel to receive OS signals
	// Buffer size 2 prevents signal loss if multiple signals arrive quickly
	c := make(chan os.Signal, 2)
	// Register for SIGINT (Ctrl+C) and SIGTERM signals
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// Start a goroutine to handle the signals
	go func() {
		// Wait for a signal
		<-c
		fmt.Println("\n[!] Emergency Signal detected (Ctrl+C). Safely shutting down ftrace...")
		// Cancel the global context to stop ongoing operations
		if cancel != nil {
			cancel()
		}
		// Call SafeShutdown to restore kernel state
		if engine != nil {
			engine.SafeShutdown()
		}
		fmt.Println("[+] Kernel state restored to safety. Exiting.")
		os.Exit(0)
	}()
}

// main is the entry point of the application.
// It performs the following steps:
// 1. Checks for root privileges
// 2. Sets up the graceful shutdown handler
// 3. Executes the command
func main() {
	// Check if running as root
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "🚨 Fatal: Root privileges required")
		os.Exit(1)
	}
	// Setup the signal handler for graceful shutdown
	SetupCloseHandler()
	// Execute the command
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "🚨 Fatal: %v\n", err)
		os.Exit(1)
	}
}
