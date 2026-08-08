package hashutil_test

import (
	"path/filepath"
	"testing"

	"catmint/hashutil"
)

func TestCompareResultsDetectsFailures(t *testing.T) {
	actual := []hashutil.HashResult{
		{FilePath: "a.txt", HashType: "SHA256", Hash: "aaa"},
		{FilePath: "b.txt", HashType: "SHA256", Hash: "bbb"},
		{FilePath: "extra.txt", HashType: "SHA256", Hash: "extra"},
	}
	reference := []hashutil.HashResult{
		{FilePath: "a.txt", HashType: "SHA256", Hash: "aaa"},
		{FilePath: "b.txt", HashType: "SHA256", Hash: "changed"},
		{FilePath: "missing.txt", HashType: "SHA256", Hash: "missing"},
	}

	report := hashutil.CompareResults(actual, reference)
	if !report.HasFailures() {
		t.Fatal("seharusnya report gagal")
	}
	if report.MatchCount != 1 || report.MismatchCount != 1 || report.MissingReferenceCount != 1 || report.MissingActualCount != 1 {
		t.Fatalf("summary tidak sesuai: %+v", report)
	}
}

func TestCompareResultsSuccess(t *testing.T) {
	actual := []hashutil.HashResult{
		{FilePath: "a.txt", HashType: "SHA256", Hash: "aaa"},
		{FilePath: "nested/b.txt", HashType: "SHA256", Hash: "bbb"},
	}
	reference := []hashutil.HashResult{
		{FilePath: filepath.Join("a.txt"), HashType: "SHA256", Hash: "AAA"},
		{FilePath: filepath.Join("nested", "b.txt"), HashType: "SHA256", Hash: "BBB"},
	}

	report := hashutil.CompareResults(actual, reference)
	if report.HasFailures() {
		t.Fatalf("report seharusnya sukses: %+v", report)
	}
	if report.MatchCount != 2 {
		t.Fatalf("match count tidak sesuai: %+v", report)
	}
}
