package hashutil_test

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"catmint/hashutil"
)

func TestVerifyAllFromJSON(t *testing.T) {
	tmp := t.TempDir()
	createTestFileAt(t, tmp, "a.txt", "data A")
	createTestFileAt(t, tmp, "b.txt", "data B")

	results, err := hashutil.GenerateDirHash(tmp, "sha256", nil, nil)
	if err != nil {
		t.Fatalf("GenerateDirHash gagal: %v", err)
	}

	refPath := filepath.Join(tmp, "hash.json")
	saveResultsForVerifyTest(t, results, refPath, "json")

	reference, err := hashutil.LoadHashReference(refPath)
	if err != nil {
		t.Fatalf("Gagal load JSON: %v", err)
	}

	report := hashutil.CompareResults(results, reference)
	if report.HasFailures() {
		t.Fatalf("report seharusnya sukses: %+v", report)
	}
}

func TestVerifyAllFromCSV(t *testing.T) {
	tmp := t.TempDir()
	createTestFileAt(t, tmp, "a.txt", "data A")
	createTestFileAt(t, tmp, "b.txt", "data B")

	results, err := hashutil.GenerateDirHash(tmp, "sha256", nil, nil)
	if err != nil {
		t.Fatalf("GenerateDirHash gagal: %v", err)
	}

	refPath := filepath.Join(tmp, "hash.csv")
	saveResultsForVerifyTest(t, results, refPath, "csv")

	reference, err := hashutil.LoadHashReference(refPath)
	if err != nil {
		t.Fatalf("Gagal load CSV: %v", err)
	}

	report := hashutil.CompareResults(results, reference)
	if report.HasFailures() {
		t.Fatalf("report seharusnya sukses: %+v", report)
	}
}

func TestVerifyAllFromTXT(t *testing.T) {
	tmp := t.TempDir()
	createTestFileAt(t, tmp, "a.txt", "data A")
	createTestFileAt(t, tmp, "b.txt", "data B")

	results, err := hashutil.GenerateDirHash(tmp, "sha256", nil, nil)
	if err != nil {
		t.Fatalf("GenerateDirHash gagal: %v", err)
	}

	refPath := filepath.Join(tmp, "hash.txt")
	saveResultsForVerifyTest(t, results, refPath, "txt")

	reference, err := hashutil.LoadHashReference(refPath)
	if err != nil {
		t.Fatalf("Gagal load TXT: %v", err)
	}

	report := hashutil.CompareResults(results, reference)
	if report.HasFailures() {
		t.Fatalf("report seharusnya sukses: %+v", report)
	}
}

func TestLoadHashReferenceUnsupportedExtension(t *testing.T) {
	tmp := t.TempDir()
	refPath := createTestFileAt(t, tmp, "hash.xml", "<hash></hash>")

	_, err := hashutil.LoadHashReference(refPath)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "tidak didukung") {
		t.Fatalf("seharusnya error format tidak didukung, dapat: %v", err)
	}
}

func TestLoadHashReferenceInvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	refPath := createTestFileAt(t, tmp, "hash.json", "{invalid json")

	_, err := hashutil.LoadHashReference(refPath)
	if err == nil {
		t.Fatal("seharusnya error untuk JSON invalid")
	}
}

func TestLoadHashReferenceEmptyCSV(t *testing.T) {
	tmp := t.TempDir()
	refPath := createTestFileAt(t, tmp, "hash.csv", "File Path,Hash Type,Hash\n")

	_, err := hashutil.LoadHashReference(refPath)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "tidak berisi data") {
		t.Fatalf("seharusnya error CSV kosong, dapat: %v", err)
	}
}

func saveResultsForVerifyTest(t *testing.T, results []hashutil.HashResult, path, format string) {
	t.Helper()

	switch format {
	case "json":
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			t.Fatalf("MarshalIndent JSON gagal: %v", err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("WriteFile JSON gagal: %v", err)
		}

	case "csv":
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("Create CSV gagal: %v", err)
		}
		defer func() {
			_ = f.Close()
		}()

		w := csv.NewWriter(f)
		if err := w.Write([]string{"File Path", "Hash Type", "Hash"}); err != nil {
			t.Fatalf("Write header CSV gagal: %v", err)
		}
		for _, r := range results {
			if err := w.Write([]string{r.FilePath, r.HashType, r.Hash}); err != nil {
				t.Fatalf("Write row CSV gagal: %v", err)
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			t.Fatalf("Flush CSV gagal: %v", err)
		}

	case "txt":
		var lines []string
		for _, r := range results {
			lines = append(lines, fmt.Sprintf("%s hash of file %s: %s", r.HashType, r.FilePath, r.Hash))
		}
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644); err != nil {
			t.Fatalf("WriteFile TXT gagal: %v", err)
		}

	default:
		t.Fatalf("format tidak didukung: %s", format)
	}
}
