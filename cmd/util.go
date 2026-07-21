package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
)

type stringListFlag []string

func (f *stringListFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *stringListFlag) Set(value string) error {
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			*f = append(*f, item)
		}
	}
	return nil
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
