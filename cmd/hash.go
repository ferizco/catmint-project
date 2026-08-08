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

	// shared variables (use StringVar so we don't deal with pointers)
	var (
		filePath   string
		dirPath    string
		alg        string
		outputFile string
		excludes   stringListFlag
		relative   bool
	)

	// file flags
	fs.StringVar(&filePath, "file", "", "Path of the file to generate hash")
	fs.StringVar(&filePath, "f", "", "Alias for -file")

	// dir flags
	fs.StringVar(&dirPath, "dir", "", "Path of the directory to hash files recursively")
	fs.StringVar(&dirPath, "d", "", "Alias for -dir")

	// algorithm flags
	fs.StringVar(&alg, "alg", "sha256", "Hash algorithm: sha256, sha512, sha1, md5, sha3-256, blake3")
	fs.StringVar(&alg, "a", "sha256", "Alias for -alg")

	// output flags
	fs.StringVar(&outputFile, "o", "", "Output file (supports .txt, .json, .csv)")
	fs.Var(&excludes, "exclude", "Path to exclude when hashing a directory (repeatable, comma-separated values supported)")
	fs.BoolVar(&relative, "relative", false, "Store directory hash file paths relative to the target directory")

	// Help for this command
	for _, a := range args {
		if a == "-h" || a == "--help" {
			internal.PrintCommandUsage("hash", fs, version, `
Examples:
  catmint hash -f test.txt -a sha256
  catmint hash -f test.txt -alg sha256 -o hash.txt
  catmint hash -d ./myfolder -alg sha512 -o hash.json
  catmint hash -d ./myfolder -alg sha512 -o hash.json --relative
  catmint hash -d ./myfolder --exclude ./myfolder/tmp --exclude ./myfolder/hash.json
`)
			return
		}
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'catmint hash --help' for usage.")
		os.Exit(1)
	}

	hashType := strings.TrimSpace(alg)

	if filePath == "" && dirPath == "" {
		fmt.Fprintln(os.Stderr, "Error: please provide -file/-f or -dir/-d")
		fmt.Fprintln(os.Stderr, "Run 'catmint hash --help' for usage.")
		os.Exit(1)
	}
	if filePath != "" && dirPath != "" {
		fmt.Fprintln(os.Stderr, "Error: use only one of -file/-f or -dir/-d")
		os.Exit(1)
	}

	outputFormat, err := detectOutputFormat(outputFile)
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

	// File mode
	if filePath != "" {
		result, err := hashutil.GenerateFileHash(filePath, hashType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			hadError = true
		} else {
			results = append(results, result)
		}
	}

	// Dir mode
	if dirPath != "" {
		if outputFile != "" {
			dirResults, err := hashutil.GenerateDirHashWithExcludes(dirPath, hashType, []string(excludes), nil, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				hadError = true
			} else {
				if relative {
					dirResults = hashutil.NormalizeResultsForBase(dirResults, dirPath)
				}
				results = append(results, dirResults...)
			}
		} else {
			successCount := 0
			errorCount := 0
			usedStreamingOutput = true

			_, err := hashutil.GenerateDirHashWithExcludes(dirPath, hashType, []string(excludes),
				func(res hashutil.HashResult) {
					if relative {
						res = hashutil.NormalizeResultForBase(res, dirPath)
					}
					fmt.Printf("%s hash of file %s: %s\n", res.HashType, res.FilePath, res.Hash)
					successCount++
				},
				func(path string, err error) {
					fmt.Fprintf(os.Stderr, "Failed to hash file %s: %v\n", path, err)
					errorCount++
				},
			)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to walk directory: %v\n", err)
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

	if outputFile != "" {
		if err := output.SaveResultsToFile(results, outputFile, outputFormat); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Saved %d hash result(s) to %s\n", len(results), outputFile)
	} else if !usedStreamingOutput {
		for _, result := range results {
			fmt.Printf("%s hash of file %s: %s\n", result.HashType, result.FilePath, result.Hash)
		}
	}
}
