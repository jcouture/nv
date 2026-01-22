<p align="center">
  <a href="https://github.com/jcouture/nv">
    <img src="https://user-images.githubusercontent.com/5007/120239413-3ba5c000-c22c-11eb-8008-052bc5f8e7b8.png" alt="nv" />
  </a>
</p>

[![Release](https://img.shields.io/github/release/jcouture/nv.svg?style=for-the-badge)](https://github.com/jcouture/nv/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/jcouture/nv)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcouture/nv?style=for-the-badge)](https://https://goreportcard.com/badge/github.com/jcouture/nv)

`nv` is a tiny tool that loads environment variables from `.env` files and runs your command with them. Use it when you want predictable, explicit env handling for a single command or script.

This repo ships two CLIs:

- `nvx` is the current, feature-rich CLI. Use this for new projects.
- `nv` is the legacy v2-compatible CLI with the original UX.

Both tools build a clean environment and then run your command. By default, `nvx` keeps `$PATH`, `$HOME`, and `$USER`, while legacy `nv` keeps only `$PATH`.

**Note (legacy nv)**

If you are using a version manager such as [asdf](https://asdf-vm.com), both `$HOME` and `$USER` could be required. See [Troubleshooting](#troubleshooting-legacy-nv).

## Quick start

### 1) Install

macOS (Homebrew):
```sh
brew install jcouture/nv/nv
```

Linux/Windows/macOS (binary):
- Download the latest release from https://github.com/jcouture/nv/releases

### 2) Run a command with `.env`

```sh
nvx run -e .env -- ./myapp
```

That is it. `nvx` loads variables from `.env` and runs your command with them.

## Which binary should I use?

- Use `nvx` for new projects and richer features.
- Use `nv` if you want the v2 UX and behavior.

## Common tasks (nvx)

### Load multiple env files

```sh
nvx run -e .env -e .env.local -- ./myapp
```

### Cascade env files automatically

Loads `.env`, `.env.local`, `.env.<env>`, `.env.<env>.local`:

```sh
nvx run --cascade --env=production -- ./deploy.sh
```

### Override values inline

```sh
nvx run -e .env -o PORT=4200 -- ./myapp
```

### Preview what will be set

```sh
nvx run -e .env --dry-run -- ./myapp
```

### Export for another tool

```sh
nvx export -e .env --format=json
```

### Validate against a schema

Default schema is `.env.example`:

```sh
nvx validate -e .env
```

Schema example:
```
DATABASE_URL=postgres://localhost
OPTIONAL=
# REQUIRED: API_KEY
```

### Print the current environment

```sh
nvx print --sort
```

## Configuration (nvx)

Configuration is optional. When you do want it, `nvx` stores it in `~/.config/nv/config.toml`.

### Quick start

```sh
nvx config init    # Create config with defaults
nvx config show    # View your current config
nvx config edit    # Edit config in $EDITOR
```

### Global variables

Global variables apply to every `nvx` command.

```sh
nvx config globals list
nvx config globals set AWS_REGION us-east-1
nvx config globals unset AWS_REGION
```

### Priority rules

When the same variable is defined multiple times, the default order is:

1. Command-line arguments
2. CLI flags
3. `.env.local` (when cascading)
4. `.env`
5. `[globals.env]`

You can flip the globals priority:

```sh
nvx config set globals.priority "last"
```

What each layer means:

- Command-line arguments: `KEY=value` prefixes on the command you run after `--`, applied by the launched process (highest priority).
- CLI flags: `-o/--override KEY=value` passed to `nvx`, applied after files and globals.
- Cascading files: in cascade mode, loaded in order `.env`, `.env.local`, `.env.<env>`, `.env.<env>.local` (missing files are skipped).
- Standard files: any `--env-file/-e` files are loaded in the order provided; with `--auto-local` defaults, `.env` is followed by optional `.env.local`.
- Globals: `[globals.env]` from config; merged before files when `globals.priority=first` (default) or after files when set to `last`.

## Legacy `nv` usage (v2)

```sh
nv .env rails server -p 2808
nv .env,.env.dev rails server -p 2808
```

## `.env` format supported

- `KEY=value` assignments with optional `export` prefix
- Full-line comments with `#` and inline comments outside quotes
- Single-quoted and double-quoted values (with escapes)
- Multiline values inside quotes
- Variable interpolation in unquoted and double-quoted values (`$VAR`, `${VAR}`)
- `PATH` expansions keep references to incoming `PATH`

## Troubleshooting (nvx)

### Legacy `~/.nv` detected

`nvx` can migrate your legacy `~/.nv` file into the config file:

```sh
nvx config migrate
```

What happens during migration:
1. Your `~/.nv` file is parsed
2. Variables are imported into `[globals]`
3. A backup is created at `~/.config/nv/nv.backup`
4. The original `~/.nv` file is removed

### Permission denied accessing config

```sh
chmod 644 ~/.config/nv/config.toml
```

### Config file corrupted

```sh
nvx config validate
nvx config reset
```

### Restore from backup

```sh
cp ~/.config/nv/nv.backup ~/.config/nv/config.toml
```

## Troubleshooting (legacy nv)

If you rely on shims or TTY tools, add these to `~/.nv`:

```
HOME=<your home directory>
USER=<your username>
TERM=xterm-color
```

## Build from source

If you want to build the latest development version:

1. Verify you have Go 1.25+ installed

```sh
go version
```

2. Clone this repository

```sh
git clone https://github.com/jcouture/nv.git
cd nv
```

3. Install build dependencies (using [mise-en-place](https://mise.jdx.dev/))

```sh
mise install
```

4. Build

```sh
make build
```

## License

`nv` is released under the MIT license. See [LICENSE](./LICENSE) for details.

The `nv` leaf logo is based on [this icon](https://thenounproject.com/term/leaf/1904973/) by [Nick Bluth](https://thenounproject.com/nickbluth/), from the Noun Project. Used under a [Creative Commons BY 3.0](http://creativecommons.org/licenses/by/3.0/) license.
