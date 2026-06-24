# Changelog

All notable changes to nv are documented here.

---

## v3.0.0 - 2026-06-24

### Breaking Changes

- **Simplified configuration**: removed `paths`, `warn_on_missing`, `general.auto_validate`, and `validation.allow_extra` settings from config.
- **Legacy `~/.nv` globals replaced**: the old flat global env file is superseded by per-user TOML config, with migration support for existing users.

### New Features

- **Lexer-based env parsing**: env parsing now supports interpolation, quoting, comments, and `PATH` preservation rules.
- **TOML-based configuration**: added per-user TOML config for defaults, validation settings, and global environment variables.
- **Config command suite**: added `nv config` subcommands for init, show, path, validate, edit, reset, migrate, get, set, and global variable management.
- **Schema validation**: validate environment variables against schemas with required-key enforcement and optional extra-key detection in strict mode.
- **Sensitive value masking**: automatically mask values that look like secrets, with `--unredacted` and custom `--mask-pattern` controls.
- **JSON export**: export resolved environment variables as JSON for integration with other tools.
- **`print` command**: added `nv print` to inspect the current OS environment with optional fuzzy filtering and sorting.
- **Verbose tracing**: `nv run --verbose` now shows which env files were loaded and in what order.
- **Expanded version output**: `nv version` now reports Go version, commit, compiler, and platform, with text and JSON output.

### Architecture

- **Internal package restructuring**: reorganized the codebase into discrete internal packages: `cli`, `config`, `exec`, `exporter`, `loader`, `parser`, `sys`, and `validator`.

### Changes

- **Per-user globals**: global environment variables now live in config instead of a flat file.
- **Cascade behavior**: cascading env resolution and auto-local loading are now explicit parts of the v3 environment pipeline.

### Bug Fixes

- **TTY foreground process group handling**: fixed `SIGTTOU` when setting the foreground process group and properly restore the child process group in TTY sessions.
- **Explicit env file handling**: when env files are explicitly specified, cascading file discovery is correctly disabled.
- **Parser stability**: fixed empty identifier handling and related parse edge cases.
- **Validation and env error handling**: fixed error handling for environment clearing and schema scanning.

### Documentation

- Refreshed README with updated usage guide, env layer priorities, build dependencies, config behavior, and schema validation examples.

### Tests

- Added coverage for config loader, tracer, runner (table-driven), secret-key detection, validator scenarios, and JSON exporter error paths.

### Maintenance

- Bumped Go from `1.20` to `1.26.4`.
- Added `gosec` security scanner to pre-commit checks.
- Refreshed copyright headers to 2026.
- Improved Makefile with parameterized names, `uninstall` and `print-version` targets.

---

## v2.2.1 - 2023-10-25

### Changes

- **Removed UPX compression** — stopped compressing binaries with UPX to reduce false positives from antivirus software.

---

## v2.2.0 - 2023-10-25

### Changes

- **Replaced `env` module with internal implementation** — replaced external `env` module with internal functionality for tighter control.
- **Switched to `just` command runner** — replaced `make` with `just` for build recipes.
- **Reorganized code** — restructured code layout for better test coverage.
- **Renamed main file** — renamed entrypoint to `nv.go`.

### Bug Fixes

- **GoReleaser homebrew config** — fixed homebrew tap configuration and enabled binary shrinking.
- **GoReleaser deprecated flag** — replaced deprecated GoReleaser configuration flag.

### Tests

- Improved test coverage for the parser and main package.

### Maintenance

- Updated to Go `1.20.9`.
- Integrated CodeQL for code analysis.
- Set up Dependabot for version updates.
- Added GitHub Actions workflow to publish releases.
- Enabled binary signing for releases.
- Updated copyright notice.

---

## v2.1.1 - 2022-01-10

### Changes

- **Isolated environment code** — extracted environment-related code into its own module.
- **Reproducible builds** — adopted the Reproducible Builds philosophy for deterministic binary output.
- **ARM64 Windows binary** — added ARM64 binary target for Windows.

### Bug Fixes

- **Version command exit code** — `nv version` no longer exits with a non-zero code.

### Documentation

- Improved usage description.
- Added non-interactive TTY warning to troubleshooting section.

### Maintenance

- Added a basic Makefile.
- Added copyright notice in `build.go`.
- Updated copyright notice.

---

## v2.1.0 - 2021-05-31

### New Features

- **Build metadata display** — `nv` now tracks and displays the build date and version at runtime.

### Changes

- **Broader binary targets** — GoReleaser now produces binaries for more platforms, including ARM on Windows.
- **Refreshed project organization** — cleaned up code layout and project structure.

### Documentation

- Updated and enhanced README with build-from-source instructions, logo copyright, and general improvements.

### Maintenance

- Updated copyright notice.

---

## v2.0.0 - 2018-11-19

### New Features

- **Global environment file** — added support for a global `~/.env` file loaded before project-specific files.
- **Multiple environment files** — `nv` now accepts multiple `-f` flags to load several env files in order.
- **Comment support** — environment files now support `#`-prefixed comment lines.
- **Environment clearing** — the environment is cleared before loading context-specific variables, ensuring a clean slate.

### Changes

- **GoReleaser integration** — integrated GoReleaser for streamlined cross-platform release management.

### Maintenance

- Added dependency on `github.com/mitchellh/go-homedir`.
- Updated copyright notice.

---

## v1.0.0 - 2017-11-30

### Bug Fixes

- **Line ending handling** — fixed parsing to handle all line ending styles (LF, CRLF, CR).

---

## v0.0.3 - 2015-10-20

### Bug Fixes

- **File readability check** — fixed an issue where the `.env` file was incorrectly reported as unreadable.

---

## v0.0.2 - 2015-10-20

### New Features

- **File existence validation** — `nv` now checks if the `.env` file exists and is readable before attempting to parse it.

### Documentation

- Added install instructions to the README.

---

## v0.0.1 - 2015-10-14

Initial version.
