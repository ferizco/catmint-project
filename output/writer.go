package output

import (
	"catmint/hashutil"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
)

func SaveResultsToFile(results []hashutil.HashResult, outputFile, format string) (err error) {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	switch format {
	case "json":
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		return encoder.Encode(results)
	case "csv":
		writer := csv.NewWriter(file)
		if err := writer.Write([]string{"File Path", "Hash Type", "Hash"}); err != nil {
			return err
		}
		for _, r := range results {
			if err := writer.Write([]string{r.FilePath, r.HashType, r.Hash}); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	case "txt":
		for _, r := range results {
			if _, err := fmt.Fprintf(file, "%s hash of file %s: %s\n", r.HashType, r.FilePath, r.Hash); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	return nil
}
