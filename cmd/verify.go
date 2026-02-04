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

	// shared variables (use StringVar so we don't deal with pointers)
	var (
		filePath     string
		dirPath      string
		expectedHash string
		refPath      string
		alg          string
	)

	// file flags
	fs.StringVar(&filePath, "file", "", "Path of the file to verify")
	fs.StringVar(&filePath, "f", "", "Alias for -file")

	// dir flags
	fs.StringVar(&dirPath, "dir", "", "Path of the directory to verify recursively")
	fs.StringVar(&dirPath, "d", "", "Alias for -dir")

	// expected hash for single file verify
	fs.StringVar(&expectedHash, "hash", "", "Expected hash to verify against the file")

	// reference file for directory verify against reference
	fs.StringVar(&refPath, "ref", "", "Path to file containing reference hashes (.txt, .json, .csv)")

	// algorithm flags
	fs.StringVar(&alg, "alg", "sha256", "Hash algorithm: sha256, sha512, sha1, md5, sha3-256")
	fs.StringVar(&alg, "a", "sha256", "Alias for -alg")

	// Help for this command
	for _, a := range args {
		if a == "-h" || a == "--help" {
			internal.PrintCommandUsage("verify", fs, version, `
Modes:
  1) Single file verify:
     catmint verify -f <path> -hash <EXPECTED_HASH> [-a sha256]

  2) Directory verify against reference file:
     catmint verify -d <path> -ref <hashes.json|csv|txt> [-a sha256]

Examples:
  catmint verify -f test.txt -hash <HASH>
  catmint verify -d ./myfolder -ref hash.json
`)
			return
		}
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'catmint verify --help' for usage.")
		os.Exit(1)
	}

	hashType := strings.TrimSpace(alg)

	// Validate algo early
	if _, err := hashutil.GetHasher(hashType); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate mode selection
	if filePath == "" && dirPath == "" {
		fmt.Fprintln(os.Stderr, "Error: please provide -file/-f or -dir/-d")
		fmt.Fprintln(os.Stderr, "Run 'catmint verify --help' for usage.")
		os.Exit(1)
	}
	if filePath != "" && dirPath != "" {
		fmt.Fprintln(os.Stderr, "Error: use only one of -file/-f or -dir/-d")
		os.Exit(1)
	}

	// Mode 1: Single file verify
	if filePath != "" {
		if strings.TrimSpace(expectedHash) == "" {
			fmt.Fprintln(os.Stderr, "Error: -hash (expected hash) is required when using -file/-f")
			os.Exit(1)
		}
		if err := hashutil.VerifyFileHash(filePath, hashType, expectedHash); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("File %s: hash matches!\n", filePath)
		return
	}

	// Mode 2: Directory verify against reference file
	if dirPath != "" {
		if strings.TrimSpace(refPath) == "" {
			fmt.Fprintln(os.Stderr, "Error: -ref is required when using -dir/-d")
			os.Exit(1)
		}

		reference, err := hashutil.LoadHashReference(refPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Gagal memuat referensi hash: %v\n", err)
			os.Exit(1)
		}

		actual, err := hashutil.GenerateDirHash(dirPath, hashType, nil, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Gagal hashing direktori: %v\n", err)
			os.Exit(1)
		}

		hashutil.CompareResults(actual, reference)
		return
	}
}
