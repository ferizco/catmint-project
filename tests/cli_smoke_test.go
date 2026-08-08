package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type smokeResult struct {
	name       string
	args       []string
	wantCode   int
	wantOutput string
}

const smokeVersion = "v1.2.0-smoke"

func TestCLISmoke(t *testing.T) {
	root := projectRoot(t)
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "catmint")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	build := exec.Command("go", "build", "-buildvcs=false", "-ldflags", "-X catmint/cmd.version="+smokeVersion, "-o", bin, ".")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build catmint binary: %v\n%s", err, output)
	}

	sample := filepath.Join(tmp, "sample")
	mustWriteFile(t, filepath.Join(sample, "a.txt"), "alpha")
	mustWriteFile(t, filepath.Join(sample, "b.txt"), "bravo")
	mustWriteFile(t, filepath.Join(sample, "tmp", "ignored.txt"), "ignored")

	hashJSON := filepath.Join(sample, "hash.json")
	hashCSV := filepath.Join(sample, "hash.csv")
	hashTXT := filepath.Join(sample, "hash.txt")
	hashRelativeJSON := filepath.Join(sample, "hash-relative.json")

	runSmokeCases(t, bin, []smokeResult{
		{name: "root help with no args", wantCode: 0, wantOutput: "Usage:"},
		{name: "help flag", args: []string{"--help"}, wantCode: 0, wantOutput: "Commands:"},
		{name: "help command", args: []string{"help"}, wantCode: 0, wantOutput: "show-update"},
		{name: "version command", args: []string{"version"}, wantCode: 0, wantOutput: smokeVersion},
		{name: "version flag", args: []string{"--version"}, wantCode: 0, wantOutput: smokeVersion},
		{name: "hash help", args: []string{"hash", "--help"}, wantCode: 0, wantOutput: "--exclude"},
		{name: "verify help", args: []string{"verify", "--help"}, wantCode: 0, wantOutput: "--json"},
		{name: "hash file sha256", args: []string{"hash", "-f", filepath.Join(sample, "a.txt"), "-a", "sha256"}, wantCode: 0, wantOutput: "SHA256 hash of file"},
		{name: "hash file blake3", args: []string{"hash", "-f", filepath.Join(sample, "a.txt"), "-a", "blake3"}, wantCode: 0, wantOutput: "BLAKE3 hash of file"},
		{name: "hash dir json with exclude", args: []string{"hash", "-d", sample, "-o", hashJSON, "--exclude", "tmp"}, wantCode: 0, wantOutput: "Saved 2 hash result(s)"},
		{name: "hash dir csv with comma exclude", args: []string{"hash", "-d", sample, "-o", hashCSV, "--exclude", "tmp,hash.json"}, wantCode: 0, wantOutput: "Saved 2 hash result(s)"},
		{name: "hash dir txt with repeated exclude", args: []string{"hash", "-d", sample, "-o", hashTXT, "--exclude", "tmp", "--exclude", "hash.json", "--exclude", "hash.csv"}, wantCode: 0, wantOutput: "Saved 2 hash result(s)"},
		{name: "hash dir json relative", args: []string{"hash", "-d", sample, "-o", hashRelativeJSON, "--exclude", "tmp,hash.json,hash.csv,hash.txt", "--relative"}, wantCode: 0, wantOutput: "Saved 2 hash result(s)"},
	})
	assertFileContains(t, hashRelativeJSON, `"file_path": "a.txt"`)
	assertFileDoesNotContain(t, hashRelativeJSON, sample)

	hashOutput, code := runCommand(t, bin, "hash", "-f", filepath.Join(sample, "a.txt"), "-a", "sha256")
	if code != 0 {
		t.Fatalf("hash file for expected hash failed with exit %d:\n%s", code, hashOutput)
	}
	expectedHash := strings.TrimSpace(hashOutput[strings.LastIndex(hashOutput, ": ")+2:])

	runSmokeCases(t, bin, []smokeResult{
		{name: "verify file success", args: []string{"verify", "-f", filepath.Join(sample, "a.txt"), "-hash", expectedHash, "-a", "sha256"}, wantCode: 0, wantOutput: "hash matches"},
		{name: "verify file success json", args: []string{"verify", "-f", filepath.Join(sample, "a.txt"), "-hash", expectedHash, "-a", "sha256", "--json"}, wantCode: 0, wantOutput: `"matched":true`},
		{name: "verify file success quiet", args: []string{"verify", "-f", filepath.Join(sample, "a.txt"), "-hash", expectedHash, "-a", "sha256", "--quiet"}, wantCode: 0},
		{name: "verify file mismatch fails", args: []string{"verify", "-f", filepath.Join(sample, "a.txt"), "-hash", "deadbeef", "-a", "sha256"}, wantCode: 1, wantOutput: "hash does not match"},
		{name: "verify file mismatch json fails", args: []string{"verify", "-f", filepath.Join(sample, "a.txt"), "-hash", "deadbeef", "-a", "sha256", "--json"}, wantCode: 1, wantOutput: `"matched":false`},
		{name: "verify dir json success", args: []string{"verify", "-d", sample, "-ref", hashJSON, "--exclude", "tmp,hash.csv,hash.txt,hash-relative.json", "--json"}, wantCode: 0, wantOutput: `"mismatch_count":0`},
		{name: "verify dir quiet success", args: []string{"verify", "-d", sample, "-ref", hashJSON, "--exclude", "tmp,hash.csv,hash.txt,hash-relative.json", "--quiet"}, wantCode: 0},
	})

	mustWriteFile(t, filepath.Join(sample, "b.txt"), "changed")
	runSmokeCases(t, bin, []smokeResult{
		{name: "verify dir mismatch fails", args: []string{"verify", "-d", sample, "-ref", hashJSON, "--exclude", "tmp,hash.csv,hash.txt,hash-relative.json"}, wantCode: 1, wantOutput: "Mismatch:"},
	})

	if err := os.Remove(filepath.Join(sample, "b.txt")); err != nil {
		t.Fatalf("remove b.txt: %v", err)
	}
	runSmokeCases(t, bin, []smokeResult{
		{name: "verify dir missing actual fails", args: []string{"verify", "-d", sample, "-ref", hashJSON, "--exclude", "tmp,hash.csv,hash.txt,hash-relative.json"}, wantCode: 1, wantOutput: "Missing from actual:"},
		{name: "invalid algorithm fails", args: []string{"hash", "-f", filepath.Join(sample, "a.txt"), "-a", "invalid"}, wantCode: 1, wantOutput: "unsupported hash type"},
		{name: "invalid output extension fails", args: []string{"hash", "-d", sample, "-o", filepath.Join(sample, "hash.xml")}, wantCode: 1, wantOutput: "Output format not supported"},
		{name: "missing command fails usage", args: []string{"-a", "sha256"}, wantCode: 2, wantOutput: "missing command"},
	})
}

func TestCLISmokeShowUpdate(t *testing.T) {
	if os.Getenv("CATMINT_SMOKE_NETWORK") != "1" {
		t.Skip("set CATMINT_SMOKE_NETWORK=1 to run network smoke test")
	}

	root := projectRoot(t)
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "catmint")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	build := exec.Command("go", "build", "-buildvcs=false", "-ldflags", "-X catmint/cmd.version="+smokeVersion, "-o", bin, ".")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build catmint binary: %v\n%s", err, output)
	}

	output, code := runCommand(t, bin, "show-update")
	if code != 0 {
		t.Fatalf("show-update failed with exit %d:\n%s", code, output)
	}
	if !strings.Contains(output, "Current version") {
		t.Fatalf("show-update output missing current version:\n%s", output)
	}
}

func runSmokeCases(t *testing.T, bin string, cases []smokeResult) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, code := runCommand(t, bin, tc.args...)
			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d\noutput:\n%s", code, tc.wantCode, output)
			}
			if tc.wantOutput != "" && !strings.Contains(output, tc.wantOutput) {
				t.Fatalf("output does not contain %q\noutput:\n%s", tc.wantOutput, output)
			}
		})
	}
}

func runCommand(t *testing.T, bin string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return string(output), exitErr.ExitCode()
	}
	t.Fatalf("run %s %v: %v\n%s", bin, args, err, output)
	return "", -1
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	return filepath.Dir(filepath.Dir(file))
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("%s does not contain %q\ncontent:\n%s", path, want, content)
	}
}

func assertFileDoesNotContain(t *testing.T, path, unwanted string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if strings.Contains(string(content), unwanted) {
		t.Fatalf("%s contains unexpected %q\ncontent:\n%s", path, unwanted, content)
	}
}
