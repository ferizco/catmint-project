package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"catmint/hashutil"
	"catmint/internal"
)

func runVerify(args []string) {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	filePath := fs.String("file", "", "Path of the file to verify")
	dirPath := fs.String("dir", "", "Path of the directory to verify recursively")

	// expected hash for single file verify
	expectedHash := fs.String("hash", "", "Expected hash to verify against the file")

	// reference file for directory verify against reference (replaces old -verify-all)
	refPath := fs.String("ref", "", "Path to file containing reference hashes for verification (.txt, .json, .csv)")

	alg := fs.String("alg", "sha256", "Hash algorithm: sha256, sha512, sha1, md5, sha3-256 ")
	// Backward-compatible alias for algorithm flag (legacy)
	algAlias := fs.String("hash-type", "", "Alias for -alg (legacy)")

	// Help for this command
	for _, a := range args {
		if a == "-h" || a == "--help" {
			internal.PrintCommandUsage("verify", fs, version, `
Modes:
  1) Single file verify:
     catmint verify -file <path> -hash <EXPECTED_HASH> [-alg sha256]

  2) Directory verify against reference:
     catmint verify -dir <path> -ref <hashes.json|csv|txt> [-alg sha256]

Examples:
  catmint verify -file test.txt -hash <HASH>
  catmint verify -dir ./myfolder -ref hash.json
`)
			return
		}
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'catmint verify --help' for usage.")
		os.Exit(1)
	}

	hashType := strings.TrimSpace(*alg)
	if strings.TrimSpace(*algAlias) != "" {
		hashType = strings.TrimSpace(*algAlias)
	}

	if _, err := hashutil.GetHasher(hashType); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate mode selection
	if *filePath == "" && *dirPath == "" {
		fmt.Fprintln(os.Stderr, "Error: please provide -file or -dir")
		fmt.Fprintln(os.Stderr, "Run 'catmint verify --help' for usage.")
		os.Exit(1)
	}
	if *filePath != "" && *dirPath != "" {
		fmt.Fprintln(os.Stderr, "Error: use only one of -file or -dir")
		os.Exit(1)
	}

	// Mode 1: Single file verify
	if *filePath != "" {
		if strings.TrimSpace(*expectedHash) == "" {
			fmt.Fprintln(os.Stderr, "Error: -hash (expected hash) is required when using -file")
			os.Exit(1)
		}
		if err := hashutil.VerifyFileHash(*filePath, hashType, *expectedHash); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("File %s: hash matches!\n", *filePath)
		return
	}

	// Mode 2: Directory verify against reference file
	if *dirPath != "" {
		if strings.TrimSpace(*refPath) == "" {
			fmt.Fprintln(os.Stderr, "Error: -ref is required when using -dir")
			os.Exit(1)
		}

		reference, err := hashutil.LoadHashReference(*refPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Gagal memuat referensi hash: %v\n", err)
			os.Exit(1)
		}

		actual, err := hashutil.GenerateDirHash(*dirPath, hashType, nil, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Gagal hashing direktori: %v\n", err)
			os.Exit(1)
		}

		hashutil.CompareResults(actual, reference)
		return
	}
}
