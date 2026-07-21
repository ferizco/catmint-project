package output_test

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"catmint/hashutil"
	"catmint/output"
)

var dummyResults = []hashutil.HashResult{
	{FilePath: filepath.Join("tmp", "test1.txt"), HashType: "SHA256", Hash: "abc123"},
	{FilePath: filepath.Join("tmp", "test2.txt"), HashType: "SHA256", Hash: "def456"},
}

func TestSaveResultsToFileJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "output.json")

	if err := output.SaveResultsToFile(dummyResults, path, "json"); err != nil {
		t.Fatalf("gagal menyimpan file JSON: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("gagal membaca output JSON: %v", err)
	}

	var parsed []hashutil.HashResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("file JSON tidak valid: %v", err)
	}

	if len(parsed) != len(dummyResults) {
		t.Errorf("jumlah hasil tidak cocok, dapat %d, ingin %d", len(parsed), len(dummyResults))
	}
}

func TestSaveResultsToFileCSV(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "output.csv")

	if err := output.SaveResultsToFile(dummyResults, path, "csv"); err != nil {
		t.Fatalf("gagal menyimpan file CSV: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("gagal membaca output CSV: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV tidak valid: %v", err)
	}

	if len(rows) != len(dummyResults)+1 {
		t.Errorf("jumlah baris CSV tidak sesuai: dapat %d, ingin %d", len(rows), len(dummyResults)+1)
	}
	if len(rows[0]) < 3 || rows[0][0] != "File Path" {
		t.Errorf("header CSV tidak sesuai: %v", rows[0])
	}
}

func TestSaveResultsToFileTXT(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "output.txt")

	if err := output.SaveResultsToFile(dummyResults, path, "txt"); err != nil {
		t.Fatalf("gagal menyimpan file TXT: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("gagal membaca output TXT: %v", err)
	}

	content := string(data)
	for _, result := range dummyResults {
		expected := fmt.Sprintf("%s hash of file %s: %s", result.HashType, result.FilePath, result.Hash)
		if !strings.Contains(content, expected) {
			t.Errorf("konten TXT tidak mengandung baris: %s", expected)
		}
	}
}

func TestSaveResultsToFileUnsupportedFormat(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "output.xml")

	err := output.SaveResultsToFile(dummyResults, path, "xml")
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unsupported") {
		t.Fatalf("seharusnya error unsupported format, dapat: %v", err)
	}
}
