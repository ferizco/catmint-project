package hashutil

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeebo/blake3"
	"golang.org/x/crypto/sha3"
)

type HashResult struct {
	FilePath string `json:"file_path"`
	HashType string `json:"hash_type"`
	Hash     string `json:"hash"`
}

type CompareReport struct {
	MatchCount            int      `json:"match_count"`
	MismatchCount         int      `json:"mismatch_count"`
	MissingReferenceCount int      `json:"missing_reference_count"`
	MissingActualCount    int      `json:"missing_actual_count"`
	MismatchFiles         []string `json:"mismatch_files,omitempty"`
	MissingReferenceFiles []string `json:"missing_reference_files,omitempty"`
	MissingActualFiles    []string `json:"missing_actual_files,omitempty"`
}

func (r CompareReport) HasFailures() bool {
	return r.MismatchCount > 0 || r.MissingReferenceCount > 0 || r.MissingActualCount > 0
}

func GetHasher(hashType string) (hash.Hash, error) {
	switch strings.ToLower(hashType) {
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "md5":
		return md5.New(), nil
	case "sha3-256":
		return sha3.New256(), nil
	case "blake3":
		return blake3.New(), nil
	default:
		return nil, fmt.Errorf("unsupported hash type: %s", hashType)
	}
}

func GenerateFileHash(filePath, hashType string) (HashResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return HashResult{}, err
	}
	defer file.Close()

	hasher, err := GetHasher(hashType)
	if err != nil {
		return HashResult{}, err
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return HashResult{}, err
	}

	hashString := hex.EncodeToString(hasher.Sum(nil))
	return HashResult{
		FilePath: filePath,
		HashType: strings.ToUpper(hashType),
		Hash:     hashString,
	}, nil
}

func GenerateDirHash(dirPath, hashType string, onResult func(HashResult), onError func(string, error)) ([]HashResult, error) {
	return GenerateDirHashWithExcludes(dirPath, hashType, nil, onResult, onError)
}

func GenerateDirHashWithExcludes(dirPath, hashType string, excludePaths []string, onResult func(HashResult), onError func(string, error)) ([]HashResult, error) {
	var results []HashResult
	excludes := normalizeExcludePaths(dirPath, excludePaths)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if onError != nil {
				onError(path, err)
			}
			return nil
		}

		if isPathExcluded(path, excludes) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		result, err := GenerateFileHash(path, hashType)
		if err != nil {
			if onError != nil {
				onError(path, err)
			}
			return nil
		}

		if onResult != nil {
			onResult(result)
		}
		results = append(results, result)
		return nil
	})
	return results, err
}

func normalizeExcludePaths(dirPath string, excludePaths []string) []string {
	var excludes []string
	for _, path := range excludePaths {
		cleanPath := filepath.Clean(strings.TrimSpace(path))
		if cleanPath == "." || cleanPath == "" {
			continue
		}

		if abs, err := filepath.Abs(cleanPath); err == nil {
			excludes = append(excludes, abs)
		}
		if !filepath.IsAbs(cleanPath) {
			if abs, err := filepath.Abs(filepath.Join(dirPath, cleanPath)); err == nil {
				excludes = append(excludes, abs)
			}
		}
	}
	return excludes
}

func isPathExcluded(path string, excludes []string) bool {
	normalizedPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}

	for _, excluded := range excludes {
		rel, err := filepath.Rel(excluded, normalizedPath)
		if err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))) {
			return true
		}
	}
	return false
}

func VerifyFileHash(filePath, hashType, expectedHash string) error {
	result, err := GenerateFileHash(filePath, hashType)
	if err != nil {
		return err
	}
	if !strings.EqualFold(result.Hash, expectedHash) {
		return fmt.Errorf("hash does not match. Expected: %s, Got: %s", expectedHash, result.Hash)
	}
	return nil
}

func NormalizeResultsForBase(results []HashResult, basePath string) []HashResult {
	normalized := make([]HashResult, 0, len(results))
	for _, result := range results {
		result.FilePath = normalizePathForBase(result.FilePath, basePath)
		normalized = append(normalized, result)
	}
	return normalized
}

func ExcludeResultsForBase(results []HashResult, basePath string, excludePaths []string) []HashResult {
	excludes := make(map[string]struct{})
	for _, path := range excludePaths {
		excludes[comparePathKey(normalizePathForBase(path, basePath))] = struct{}{}
	}

	filtered := make([]HashResult, 0, len(results))
	for _, result := range results {
		key := comparePathKey(normalizePathForBase(result.FilePath, basePath))
		if _, excluded := excludes[key]; excluded {
			continue
		}
		filtered = append(filtered, result)
	}
	return filtered
}

func normalizePathForBase(path, basePath string) string {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	baseAbs, baseErr := filepath.Abs(filepath.Clean(strings.TrimSpace(basePath)))

	if baseErr == nil {
		pathAbs, pathErr := filepath.Abs(cleanPath)
		if pathErr == nil {
			rel, relErr := filepath.Rel(baseAbs, pathAbs)
			if relErr == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return filepath.ToSlash(rel)
			}
		}
	}

	return filepath.ToSlash(cleanPath)
}

func comparePathKey(path string) string {
	key := filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	if os.PathSeparator == '\\' {
		key = strings.ToLower(key)
	}
	return key
}

// CompareResults membandingkan hash hasil saat ini dengan referensi.
func CompareResults(actual, reference []HashResult) CompareReport {
	referenceMap := make(map[string]HashResult)
	for _, ref := range reference {
		referenceMap[comparePathKey(ref.FilePath)] = ref
	}

	actualMap := make(map[string]HashResult)
	report := CompareReport{}

	for _, a := range actual {
		key := comparePathKey(a.FilePath)
		actualMap[key] = a

		ref, found := referenceMap[key]
		if !found {
			report.MissingReferenceFiles = append(report.MissingReferenceFiles, a.FilePath)
			report.MissingReferenceCount++
			continue
		}

		if strings.EqualFold(a.Hash, ref.Hash) {
			report.MatchCount++
		} else {
			report.MismatchFiles = append(report.MismatchFiles, a.FilePath)
			report.MismatchCount++
		}
	}

	for _, ref := range reference {
		key := comparePathKey(ref.FilePath)
		if _, found := actualMap[key]; !found {
			report.MissingActualFiles = append(report.MissingActualFiles, ref.FilePath)
			report.MissingActualCount++
		}
	}

	return report
}

func PrintCompareReport(report CompareReport) {
	fmt.Printf("Summary: %d match, %d mismatch, %d not found in reference, %d missing from actual\n",
		report.MatchCount,
		report.MismatchCount,
		report.MissingReferenceCount,
		report.MissingActualCount,
	)

	if report.MismatchCount > 0 {
		fmt.Println("\nMismatch:")
		for _, path := range report.MismatchFiles {
			fmt.Printf("- %s\n", path)
		}
	}

	if report.MissingReferenceCount > 0 {
		fmt.Println("\nNot found in reference:")
		for _, path := range report.MissingReferenceFiles {
			fmt.Printf("- %s\n", path)
		}
	}

	if report.MissingActualCount > 0 {
		fmt.Println("\nMissing from actual:")
		for _, path := range report.MissingActualFiles {
			fmt.Printf("- %s\n", path)
		}
	}
}
