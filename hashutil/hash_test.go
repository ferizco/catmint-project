package hashutil_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"catmint/hashutil"
)

func createTestFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("gagal membuat file: %v", err)
	}
	return filePath
}

func createTestFileAt(t *testing.T, dir, name, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatalf("gagal membuat direktori test: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("gagal membuat file di direktori: %v", err)
	}
	return filePath
}

func TestGenerateFileHash(t *testing.T) {
	file := createTestFile(t, "test.txt", "hello world")

	result, err := hashutil.GenerateFileHash(file, "sha256")
	if err != nil {
		t.Fatalf("GenerateFileHash gagal: %v", err)
	}
	if result.Hash == "" {
		t.Error("hash tidak boleh kosong")
	}
	if result.FilePath != file {
		t.Errorf("path tidak sesuai: dapat %s, ingin %s", result.FilePath, file)
	}
}

func TestGenerateDirHash(t *testing.T) {
	dir := t.TempDir()
	createTestFileAt(t, dir, "a.txt", "data A")
	createTestFileAt(t, dir, "b.txt", "data B")

	var collected []hashutil.HashResult

	results, err := hashutil.GenerateDirHash(dir, "sha1",
		func(res hashutil.HashResult) {
			collected = append(collected, res)
		},
		func(path string, err error) {
			t.Errorf("hash gagal untuk file %s: %v", path, err)
		},
	)

	if err != nil {
		t.Fatalf("GenerateDirHash gagal: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("seharusnya ada 2 hash, dapat %d", len(results))
	}
	if len(collected) != 2 {
		t.Errorf("callback onResult hanya dipanggil %d kali, seharusnya 2", len(collected))
	}
}

func TestGenerateDirHashEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	results, err := hashutil.GenerateDirHash(dir, "sha256", nil, nil)
	if err != nil {
		t.Fatalf("GenerateDirHash gagal: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("direktori kosong seharusnya menghasilkan 0 hash, dapat %d", len(results))
	}
}

func TestGenerateDirHashWithExcludes(t *testing.T) {
	dir := t.TempDir()
	createTestFileAt(t, dir, "a.txt", "data A")
	refPath := createTestFileAt(t, dir, "hash.json", "reference")

	results, err := hashutil.GenerateDirHashWithExcludes(dir, "sha256", []string{refPath}, nil, nil)
	if err != nil {
		t.Fatalf("GenerateDirHashWithExcludes gagal: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("seharusnya hanya 1 file yang di-hash, dapat %d", len(results))
	}
	if filepath.Base(results[0].FilePath) != "a.txt" {
		t.Fatalf("file exclude ikut ter-hash: %s", results[0].FilePath)
	}
}

func TestGenerateDirHashWithRelativeDirectoryExclude(t *testing.T) {
	dir := t.TempDir()
	createTestFileAt(t, dir, "a.txt", "data A")
	createTestFileAt(t, dir, filepath.Join("tmp", "ignored.txt"), "ignore me")

	results, err := hashutil.GenerateDirHashWithExcludes(dir, "sha256", []string{"tmp"}, nil, nil)
	if err != nil {
		t.Fatalf("GenerateDirHashWithExcludes gagal: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("seharusnya hanya 1 file yang di-hash, dapat %d", len(results))
	}
	if filepath.Base(results[0].FilePath) != "a.txt" {
		t.Fatalf("exclude directory tidak sesuai: %s", results[0].FilePath)
	}
}

func TestVerifyFileHash(t *testing.T) {
	file := createTestFile(t, "verify.txt", "verifikasi data")

	result, err := hashutil.GenerateFileHash(file, "sha256")
	if err != nil {
		t.Fatalf("GenerateFileHash gagal: %v", err)
	}

	if err := hashutil.VerifyFileHash(file, "sha256", result.Hash); err != nil {
		t.Errorf("hash seharusnya cocok, tapi gagal: %v", err)
	}

	if err := hashutil.VerifyFileHash(file, "sha256", "1234567890abcdef"); err == nil {
		t.Error("seharusnya gagal jika hash tidak cocok")
	}
}

func TestGetHasher(t *testing.T) {
	valid := []string{"sha256", "sha512", "sha1", "md5", "sha3-256", "blake3"}
	for _, algo := range valid {
		_, err := hashutil.GetHasher(algo)
		if err != nil {
			t.Errorf("seharusnya mendukung %s, tapi error: %v", algo, err)
		}
	}

	_, err := hashutil.GetHasher("unsupported")
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unsupported") {
		t.Error("seharusnya error untuk hash tidak didukung")
	}
}

func TestGenerateFileHashBlake3(t *testing.T) {
	file := createTestFile(t, "blake3.txt", "hello blake3")

	result, err := hashutil.GenerateFileHash(file, "blake3")
	if err != nil {
		t.Fatalf("GenerateFileHash blake3 gagal: %v", err)
	}
	if result.HashType != "BLAKE3" || result.Hash == "" {
		t.Fatalf("hasil blake3 tidak sesuai: %+v", result)
	}
}

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
