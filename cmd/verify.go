package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"catmint/hashutil"
	"catmint/internal"
)

type fileVerifyReport struct {
	FilePath     string `json:"file_path"`
	Matched      bool   `json:"matched"`
	ExpectedHash string `json:"expected_hash,omitempty"`
	ActualHash   string `json:"actual_hash,omitempty"`
	HashType     string `json:"hash_type"`
	Error        string `json:"error,omitempty"`
}

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
		excludes     stringListFlag
		quiet        bool
		jsonOutput   bool
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
	fs.Var(&excludes, "exclude", "Path to exclude when verifying a directory (repeatable, comma-separated values supported)")
	fs.BoolVar(&quiet, "quiet", false, "Suppress successful verification output")
	fs.BoolVar(&jsonOutput, "json", false, "Print verification report as JSON")

	// algorithm flags
	fs.StringVar(&alg, "alg", "sha256", "Hash algorithm: sha256, sha512, sha1, md5, sha3-256, blake3")
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
  catmint verify -d ./myfolder -ref hash.json --exclude ./myfolder/tmp
  catmint verify -d ./myfolder -ref hash.json --json
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

	if quiet && jsonOutput {
		fmt.Fprintln(os.Stderr, "Error: use only one of --quiet or --json")
		os.Exit(1)
	}

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

		result, err := hashutil.GenerateFileHash(filePath, hashType)
		if err != nil {
			if jsonOutput {
				writeFileVerifyReport(fileVerifyReport{
					FilePath: filePath,
					Matched:  false,
					HashType: strings.ToUpper(hashType),
					Error:    err.Error(),
				})
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}

		matched := strings.EqualFold(result.Hash, expectedHash)
		if jsonOutput {
			writeFileVerifyReport(fileVerifyReport{
				FilePath:     filePath,
				Matched:      matched,
				ExpectedHash: strings.TrimSpace(expectedHash),
				ActualHash:   result.Hash,
				HashType:     result.HashType,
			})
		} else {
			if matched && !quiet {
				fmt.Printf("File %s: hash matches!\n", filePath)
			}
			if !matched {
				fmt.Fprintf(os.Stderr, "Error: hash does not match. Expected: %s, Got: %s\n", expectedHash, result.Hash)
			}
		}
		if !matched {
			os.Exit(1)
		}
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
			fmt.Fprintf(os.Stderr, "Failed to load hash reference: %v\n", err)
			os.Exit(1)
		}

		allExcludes := append([]string{refPath}, []string(excludes)...)

		actual, err := hashutil.GenerateDirHashWithExcludes(dirPath, hashType, allExcludes, nil, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to hash directory: %v\n", err)
			os.Exit(1)
		}

		reference = hashutil.ExcludeResultsForBase(reference, dirPath, allExcludes)
		actual = hashutil.NormalizeResultsForBase(actual, dirPath)
		reference = hashutil.NormalizeResultsForBase(reference, dirPath)

		report := hashutil.CompareResults(actual, reference)
		if jsonOutput {
			if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else if !quiet {
			hashutil.PrintCompareReport(report)
		}
		if report.HasFailures() {
			os.Exit(1)
		}
		return
	}
}

func writeFileVerifyReport(report fileVerifyReport) {
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
