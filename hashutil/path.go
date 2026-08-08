package hashutil

import (
	"os"
	"path/filepath"
	"strings"
)

func NormalizeResultForBase(result HashResult, basePath string) HashResult {
	result.FilePath = normalizePathForBase(result.FilePath, basePath)
	return result
}

func NormalizeResultsForBase(results []HashResult, basePath string) []HashResult {
	normalized := make([]HashResult, 0, len(results))
	for _, result := range results {
		normalized = append(normalized, NormalizeResultForBase(result, basePath))
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
