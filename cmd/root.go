package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"catmint/hashutil"
	"catmint/internal"
	"catmint/output"
)

const version = "1.0.0"

func Execute() {
	// No args -> show root usage
	if len(os.Args) < 2 {
		printRootUsage()
		os.Exit(0)
	}

	// Support common help/version flags at root level
	switch os.Args[1] {
	case "-h", "--help", "help":
		printRootUsage()
		return
	case "-version", "--version", "version":
		fmt.Printf("catmint version: %s\n", version)
		return
	}

	// Legacy mode fallback (if user still uses old flags without command)
	// Example: catmint -file test.txt -hash sha256
	if strings.HasPrefix(os.Args[1], "-") {
		runLegacy(os.Args[1:])
		return
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
		os.Exit(1)
	}
}

func printRootUsage() {
	// Root usage uses existing internal.PrintUsage so it compiles without changing other files.
	fs := flag.NewFlagSet("catmint", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	internal.PrintUsage(fs, version)

}

// runLegacy keeps your old behaviour so existing users/scripts won't break.
// You can remove this later once you fully migrate to subcommands.
func runLegacy(args []string) {
	fs := flag.NewFlagSet("catmint", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	filePath := fs.String("file", "", "Path of the file to generate hash")
	dirPath := fs.String("dir", "", "Path of the directory to hash files recursively")
	hashType := fs.String("hash", "sha256", "Hash type: sha256, sha512, sha1, md5, sha3-256")
	outputFile := fs.String("o", "", "Output file (supports .txt, .json, .csv)")
	verifyHash := fs.String("verify", "", "Hash to verify against the file")
	verifyAllPath := fs.String("verify-all", "", "Path to file containing reference hashes for verification")
	showVersion := fs.Bool("version", false, "Show the version of the application")

	for _, arg := range args {
		if arg == "-help" || arg == "--help" || arg == "-h" {
			internal.PrintUsage(fs, version)
			return
		}
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'catmint --help' for usage.")
		os.Exit(1)
	}

	if *showVersion {
		fmt.Printf("catmint version: %s\n", version)
		return
	}

	outputFormat, err := detectOutputFormat(*outputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	if _, err := hashutil.GetHasher(*hashType); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *verifyHash != "" && *filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: -verify must be used with -file")
		os.Exit(1)
	}

	if *filePath != "" && *verifyHash != "" {
		if err := hashutil.VerifyFileHash(*filePath, *hashType, *verifyHash); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		} else {
			fmt.Printf("File %s: hash matches!\n", *filePath)
		}
		return
	}

	var results []hashutil.HashResult
	hadError := false

	if *filePath != "" {
		result, err := hashutil.GenerateFileHash(*filePath, *hashType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			hadError = true
		} else {
			results = append(results, result)
		}
	}

	usedStreamingOutput := false
	if *dirPath != "" && *verifyAllPath == "" {
		if *outputFile != "" {
			dirResults, err := hashutil.GenerateDirHash(*dirPath, *hashType, nil, nil)
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

			_, err := hashutil.GenerateDirHash(*dirPath, *hashType,
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

	if *dirPath != "" && *verifyAllPath != "" {
		reference, err := hashutil.LoadHashReference(*verifyAllPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Gagal memuat referensi hash: %v\n", err)
			os.Exit(1)
		}

		actual, err := hashutil.GenerateDirHash(*dirPath, *hashType, nil, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Gagal hashing direktori: %v\n", err)
			os.Exit(1)
		}

		hashutil.CompareResults(actual, reference)
		return
	}

	if hadError {
		fmt.Fprintln(os.Stderr, "Run 'catmint --help' for usage.")
		return
	}

	if len(results) == 0 && !usedStreamingOutput {
		fmt.Fprintln(os.Stderr, "Error: Please provide a file path (-file) or directory path (-dir).")
		fmt.Fprintln(os.Stderr, "Run 'catmint --help' for usage.")
		return
	}

	if *outputFile != "" {
		if err := output.SaveResultsToFile(results, *outputFile, outputFormat); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	} else {
		for _, result := range results {
			fmt.Printf("%s hash of file %s: %s\n", result.HashType, result.FilePath, result.Hash)
		}
	}
}

func detectOutputFormat(outputFile string) (string, error) {
	if strings.TrimSpace(outputFile) == "" {
		return "txt", nil
	}
	ext := strings.ToLower(filepath.Ext(outputFile))
	switch ext {
	case ".json", ".csv", ".txt":
		return strings.TrimPrefix(ext, "."), nil
	default:
		return "", fmt.Errorf("Error: Output format not supported. Please use .txt, .json, or .csv.")
	}
}
