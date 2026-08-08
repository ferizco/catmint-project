package hashutil_test

import (
	"path/filepath"
	"testing"

	"catmint/hashutil"
)

func TestNormalizeResultsForBase(t *testing.T) {
	dir := t.TempDir()
	file := createTestFileAt(t, dir, "a.txt", "data A")

	normalized := hashutil.NormalizeResultsForBase([]hashutil.HashResult{
		{FilePath: file, HashType: "SHA256", Hash: "aaa"},
	}, dir)

	if len(normalized) != 1 || normalized[0].FilePath != "a.txt" {
		t.Fatalf("path tidak ternormalisasi relatif ke base: %+v", normalized)
	}
}

func TestExcludeResultsForBase(t *testing.T) {
	dir := t.TempDir()
	file := createTestFileAt(t, dir, "a.txt", "data A")
	refPath := createTestFileAt(t, dir, "hash.json", "reference")

	filtered := hashutil.ExcludeResultsForBase([]hashutil.HashResult{
		{FilePath: file, HashType: "SHA256", Hash: "aaa"},
		{FilePath: refPath, HashType: "SHA256", Hash: "ref"},
	}, dir, []string{refPath})

	if len(filtered) != 1 || filepath.Base(filtered[0].FilePath) != "a.txt" {
		t.Fatalf("exclude reference tidak sesuai: %+v", filtered)
	}
}
