package internal

import (
	"flag"
	"fmt"
	"os"
)

const banner = `               __          .__        __    .__                  .__     
  ____ _____ _╱  │_  _____ │__│ _____╱  │_  │  │__ _____    _____│  │__  
_╱ ___╲╲__  ╲╲   __╲╱     ╲│  │╱    ╲   __╲ │  │  ╲╲__  ╲  ╱  ___╱  │  ╲ 
╲  ╲___ ╱ __ ╲│  │ │  Y Y  ╲  │   │  ╲  │   │   Y  ╲╱ __ ╲_╲___ ╲│   Y  ╲
 ╲___  >____  ╱__│ │__│_│  ╱__│___│  ╱__│   │___│  (____  ╱____  >___│  ╱
     ╲╱     ╲╱           ╲╱        ╲╱            ╲╱     ╲╱     ╲╱     ╲╱ 
`

// PrintUsage prints ROOT help (catmint with no args / catmint help).
func PrintUsage(fs *flag.FlagSet, version string) {
	fmt.Print(banner)
	fmt.Fprintf(os.Stderr, `simple, fast & ergonomic hash utility
Version: %s
Author: Ferizco

Usage:
  catmint <command> [options]

Commands:
  hash        Generate hash for a file or directory
  verify      Verify file hash OR verify directory against reference file
  version     Show the version of the application
  help        Show this help message

Global flags:
  -h, --help  Show help
  --version   Show version

Use "catmint <command> --help" to see available options.

Examples:
  catmint hash -file test.txt -alg sha256 -o hash.txt
  catmint hash -dir ./myfolder -alg sha512 -o hash.json
  catmint verify -file test.txt -hash <EXPECTED_HASH> -alg sha256
  catmint verify -dir ./myfolder -ref hash.json -alg sha256
`, version)

	// Root help doesn't need fs.PrintDefaults() because root has no flags in command mode.
	// We keep fs parameter for backward compatibility.
	_ = fs
}

// PrintCommandUsage prints help for a specific subcommand and includes fs.PrintDefaults().
func PrintCommandUsage(command string, fs *flag.FlagSet, version string, extra string) {
	fmt.Print(banner)
	fmt.Fprintf(os.Stderr, `simple, fast & ergonomic hash utility
Version: %s

Usage:
  catmint %s [options]

Options:
`, version, command)

	// Important: this prints the actual flags defined on the command FlagSet
	fs.SetOutput(os.Stderr)
	fs.PrintDefaults()

	if extra != "" {
		// extra should typically start with a newline for spacing.
		fmt.Fprint(os.Stderr, extra)
	}
}
