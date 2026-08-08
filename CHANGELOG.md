# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project follows [Semantic Versioning](https://semver.org/).

## [Draft v1.3]

- Added smoke test package
- Added user experience improvements
- Refactored hashutil package structure
- Updated documentation and release notes workflow
- Switched CLI version handling to build-time injection

## [v1.2.0] 2026-07-21

### Added

- Added structured reports for folder hash comparison, including match count, mismatch count, files missing from the reference, and reference files missing from the actual folder.
- Added path exclusion support for folder hashing, used to ignore the reference file passed with `-ref`.
- Added the `--exclude` CLI option for recursive `hash` and `verify` commands.
- Added the `--quiet` and `--json` CLI options for `verify`.
- Added tests for reference file exclusion, compare reports, path normalization, and Blake3 support.
- Added tests for empty directories, successful compare results, unsupported output formats, and invalid or unsupported hash reference files.
- Added tests for repeatable and comma-separated CLI exclude values.
- Moved package-specific tests into `hashutil/`, `output/`, and `cmd/` test files.

### Changed

- Directory verification now normalizes paths relative to the target directory, making comparisons more stable across absolute and relative paths.
- Directory verification now ignores the `-ref` file entry, preventing the reference file itself from causing verification failures.
- Folder hashing now supports excluding both files and directory subtrees.
- Compare report calculation is now separated from human-readable report printing.
- JSON, CSV, and TXT output writing now propagates write, flush, and close errors more reliably.
- The `verify` command help text now lists `blake3` as a supported algorithm.
- CLI version and documentation have been aligned with `v1.2.0`.
- README, Security Policy, and Release Roadmap have been updated to match the v1.2.0 feature set and release target.
- README has been reorganized with clearer installation, quick start, command, exit code, and reference file sections.
- Release workflow now runs `go test ./...` and `go vet ./...` before building, signing, packaging, or publishing release assets.

### Fixed

- Directory verification now exits with code `1` when mismatches are found, files are missing from the reference, or reference files are missing from the actual folder.
- Folder comparison now detects files that exist only in the reference, not only actual files missing from the reference.
- Fixed mojibake text in README, CLI help banner, and release workflow notes.

## [v1.1.1] - 2026-03-27

### Fixed

- Fixed a bug in the v1.1.x release line.

## [v1.1.0] - 2026-03-20

### Added

- Added the `show-update` command to check for available Catmint updates.
- Added a response message when folder hash results are saved to a file.
- Added support for the `blake3` algorithm.

## [v1.0.0] - 2026-02-04

### Added

- Initial Catmint release.
