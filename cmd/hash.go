package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"catmint/hashutil"
	"catmint/internal"
	"catmint/output"
)

func runHash(args []string) {
	fs := flag.NewFlagSet("hash", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	filePath := fs.String("file", "", "Path of the file to generate hash")
	dirPath := fs.String("dir", "", "Path of the directory to hash files recursively")
	alg := fs.String("alg", "sha256", "Hash algorithm: sha256, sha512, sha1, md5, sha3-256 ")
	// Backward-compatible alias for algorithm flag (old: -hash sha256)
	algAlias := fs.String("hash", "", "Alias for -alg (legacy)")
	outputFile := fs.String("o", "", "Output file (supports .txt, .json, .csv)")

	for _, a := range args {
		if a == "-h" || a == "--help" {
			internal.PrintCommandUsage("hash", fs, version, `
Examples:
  catmint hash -file test.txt -alg sha256 -o hash.txt
  catmint hash -dir ./myfolder -alg sha512 -o hash.json
`)
			return
		}
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'catmint hash --help' for usage.")
		os.Exit(1)
	}

	hashType := strings.TrimSpace(*alg)
	if strings.TrimSpace(*algAlias) != "" {
		hashType = strings.TrimSpace(*algAlias)
	}

	if *filePath == "" && *dirPath == "" {
		fmt.Fprintln(os.Stderr, "Error: please provide -file or -dir")
		fmt.Fprintln(os.Stderr, "Run 'catmint hash --help' for usage.")
		os.Exit(1)
	}
	if *filePath != "" && *dirPath != "" {
		fmt.Fprintln(os.Stderr, "Error: use only one of -file or -dir")
		os.Exit(1)
	}

	outputFormat, err := detectOutputFormat(*outputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if _, err := hashutil.GetHasher(hashType); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var results []hashutil.HashResult
	hadError := false
	usedStreamingOutput := false

	if *filePath != "" {
		result, err := hashutil.GenerateFileHash(*filePath, hashType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			hadError = true
		} else {
			results = append(results, result)
		}
	}

	if *dirPath != "" {
		if *outputFile != "" {
			dirResults, err := hashutil.GenerateDirHash(*dirPath, hashType, nil, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				hadError = true
			} else {
				results = append(results, dirResults...)
			}
		} else {
			successCount := 0
			errorCount := 0
			usedStreamingOutput = true

			_, err := hashutil.GenerateDirHash(*dirPath, hashType,
				func(res hashutil.HashResult) {
					fmt.Printf("%s hash of file %s: %s\n", res.HashType, res.FilePath, res.Hash)
					successCount++
				},
				func(path string, err error) {
					fmt.Fprintf(os.Stderr, "Gagal hash file %s: %v\n", path, err)
					errorCount++
				},
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error saat menjelajah direktori: %v\n", err)
				hadError = true
			}
			fmt.Printf("\nSummary: %d success, %d failed\n", successCount, errorCount)
		}
	}

	if hadError {
		fmt.Fprintln(os.Stderr, "Run 'catmint hash --help' for usage.")
		os.Exit(1)
	}

	if len(results) == 0 && !usedStreamingOutput {
		fmt.Fprintln(os.Stderr, "Error: no results produced")
		os.Exit(1)
	}

	if *outputFile != "" {
		if err := output.SaveResultsToFile(results, *outputFile, outputFormat); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else if !usedStreamingOutput {
		for _, result := range results {
			fmt.Printf("%s hash of file %s: %s\n", result.HashType, result.FilePath, result.Hash)
		}
	}
}
