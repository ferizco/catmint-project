package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
)

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
