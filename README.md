# catmint ![version](https://img.shields.io/badge/version-1.3.0-blue.svg)

catmint is a small command-line tool for generating and verifying file hashes. It is useful for integrity checks, folder checksum snapshots, and CI scripts that need a clear success or failure exit code.

## Highlights

- Hash a single file or an entire directory recursively.
- Verify a single file against an expected hash.
- Verify a directory against a previously exported reference file.
- Export hash results as TXT, CSV, or JSON.
- Exclude files or directories from recursive hashing and verification.
- Print human-readable or JSON verification reports.

## Supported Algorithms

- SHA256
- SHA512
- SHA1
- MD5
- SHA3-256
- Blake3

For security-sensitive verification, prefer SHA256, SHA512, SHA3-256, or Blake3. MD5 and SHA1 are kept for compatibility.

## Installation

Download the binary for your operating system from the [Releases](https://github.com/ferizco/catmint-project/releases) page.

Available release builds:

- Linux amd64
- Windows amd64

## Project Documents

- [Changelog](CHANGELOG.md)
- [Security Policy](SECURITY.md)
- [License](LICENSE)

After installing, confirm that catmint is available:

```sh
catmint --version
catmint --help
```

Local builds use `dev` as the default version. Release builds inject the version from the Git tag at build time.

## Docker / GHCR

catmint can also run as a container image from GitHub Container Registry:

```sh
docker run --rm ghcr.io/ferizco/catmint:v1.3.0 --version
```

Use it in a project directory by mounting the current working directory to `/work`:

```sh
docker run --rm -v "$PWD:/work" -w /work ghcr.io/ferizco/catmint:v1.3.0 hash -d ./dist -o hash.json --relative
docker run --rm -v "$PWD:/work" -w /work ghcr.io/ferizco/catmint:v1.3.0 verify -d ./dist -ref hash.json --json
```

In CI/CD pipelines, use the `verify` command exit code as the gate. A mismatch or missing file exits with code `1`, which should fail the pipeline.

Example GitHub Actions step:

```yaml
- name: Verify release artifacts
  run: |
    docker run --rm -v "$PWD:/work" -w /work ghcr.io/ferizco/catmint:v1.3.0 verify -d ./dist -ref hash.json --json
```

## Quick Start

Hash a single file:

```sh
catmint hash -f test.txt -a sha256
```

Hash a directory and save the result as JSON:

```sh
catmint hash -d ./myfolder -a sha256 -o hash.json
```

Hash a directory using portable relative paths:

```sh
catmint hash -d ./myfolder -a sha256 -o hash.json --relative
```

Verify a single file:

```sh
catmint verify -f test.txt -hash <EXPECTED_HASH> -a sha256
```

Verify a directory against a reference file:

```sh
catmint verify -d ./myfolder -ref hash.json -a sha256
```

## Commands

### `hash`

Generate hashes for a file or directory.

```sh
catmint hash -f <file> [-a algorithm] [-o output.txt|output.csv|output.json]
catmint hash -d <directory> [-a algorithm] [-o output.txt|output.csv|output.json] [--exclude path] [--relative]
```

Useful options:

- `-f`, `-file` - File to hash.
- `-d`, `-dir` - Directory to hash recursively.
- `-a`, `-alg` - Hash algorithm. Default: `sha256`.
- `-o` - Output file. Supports `.txt`, `.csv`, and `.json`.
- `--exclude` - Exclude a file or directory while hashing a directory. Can be repeated and also accepts comma-separated values.
- `--relative` - Store directory hash file paths relative to the target directory.

Examples:

```sh
catmint hash -d ./myfolder -o hash.json --exclude tmp --exclude hash.json
catmint hash -d ./myfolder -o hash.json --exclude tmp,dist
catmint hash -d ./myfolder -o hash.json --relative
```

### `verify`

Verify a file or directory.

```sh
catmint verify -f <file> -hash <expected_hash> [-a algorithm] [--quiet] [--json]
catmint verify -d <directory> -ref <reference.json|reference.csv|reference.txt> [-a algorithm] [--exclude path] [--quiet] [--json]
```

Useful options:

- `-f`, `-file` - File to verify.
- `-d`, `-dir` - Directory to verify recursively.
- `-hash` - Expected hash for single-file verification.
- `-ref` - Reference file for directory verification. Supports `.txt`, `.csv`, and `.json`.
- `-a`, `-alg` - Hash algorithm. Default: `sha256`.
- `--exclude` - Exclude a file or directory from directory verification. Can be repeated and also accepts comma-separated values.
- `--quiet` - Suppress successful verification output. Errors are still printed to stderr.
- `--json` - Print the verification result as JSON.

Directory verification normalizes paths relative to the target directory. The reference file passed with `-ref` is ignored automatically, so storing `hash.json` inside the verified folder will not cause a false failure.

For single-file verification, `--json` prints a structured result for both success and mismatch cases. Mismatch still exits with code `1`, but the output remains machine-readable.

Examples:

```sh
catmint verify -d ./myfolder -ref hash.json --exclude tmp
catmint verify -d ./myfolder -ref hash.json --exclude tmp,dist --json
catmint verify -d ./myfolder -ref hash.json --quiet
```

### `show-update`

Check whether a newer GitHub release is available.

```sh
catmint show-update
```

## Exit Codes

- `0` - Command completed successfully.
- `1` - Hashing, verification, parsing, or output writing failed.
- `2` - Command usage error, such as a missing or unknown command.

For directory verification, exit code `1` is returned when catmint finds mismatches, files missing from the reference, or reference entries missing from the actual directory.

## Reference Files

Directory verification uses reference files generated by catmint:

```sh
catmint hash -d ./myfolder -o hash.json
catmint verify -d ./myfolder -ref hash.json
```

Supported reference formats:

- TXT
- CSV
- JSON

## Testing

Run all unit and smoke tests:

```sh
go test ./...
```

Run tests with verbose output:

```sh
go test -v ./...
```

Run only the CLI smoke tests:

```sh
go test -v ./tests
```

The CLI smoke tests build a temporary Catmint binary and run the main command flows end-to-end, including `hash`, `verify`, `--exclude`, `--quiet`, `--json`, and expected failure cases.

The `show-update` smoke test is disabled by default because it requires network access to the GitHub API. Enable it explicitly when needed:

PowerShell:

```powershell
$env:CATMINT_SMOKE_NETWORK = "1"
go test -v ./tests
```

Shell:

```sh
CATMINT_SMOKE_NETWORK=1 go test -v ./tests
```

Run static analysis:

```sh
go vet ./...
```

Build locally:

```sh
go build .
```

Build locally with an explicit version:

```sh
go build -ldflags "-X catmint/cmd.version=v1.3.0" .
```

## Security Notes

catmint helps verify file integrity, but it does not prove that a file is trustworthy by itself. Always compare hashes from trusted sources and keep catmint updated.
