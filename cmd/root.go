package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"catmint/internal"
)

const version = "1.0.0"

func Execute() {
	// No args -> show root usage
	if len(os.Args) < 2 {
		printRootUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		printRootUsage()
		return
	case "-version", "--version", "version":
		fmt.Printf("catmint version: %s\n", version)
		return
	}

	// Optional: if user forgets command and starts with a flag, show help (no legacy)
	if strings.HasPrefix(os.Args[1], "-") {
		fmt.Fprint(os.Stderr, "Error: missing command. Use 'catmint <command> [options]'.\n")
		printRootUsage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "hash":
		runHash(args)
	case "verify":
		runVerify(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printRootUsage()
		os.Exit(2)
	}
}

func printRootUsage() {
	internal.PrintUsage(flag.NewFlagSet("catmint", flag.ContinueOnError), version)
}
