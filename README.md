<p align="center">
  <a href="https://github.com/jcouture/nv">
    <img src="https://user-images.githubusercontent.com/5007/120239413-3ba5c000-c22c-11eb-8008-052bc5f8e7b8.png" alt="nv" />
  </a>
</p>

[![Release](https://img.shields.io/github/release/jcouture/nv.svg?style=for-the-badge)](https://github.com/jcouture/nv/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/jcouture/nv)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcouture/nv?style=for-the-badge)](https://https://goreportcard.com/badge/github.com/jcouture/nv)

`nv` is a tiny tool that loads environment variables from `.env` files and runs your command with them. Use it when you want predictable, explicit env handling for a single command or script. v3 is the next iteration and replaces the old nv/nvx split; if you want to stick to v2, pin an older release via Homebrew.

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
nv run -e .env -- ./myapp
```

That is it. `nv` loads variables from `.env` and runs your command with them.

## Common tasks

### Load multiple env files

```sh
nv run -e .env -e .env.local -- ./myapp
```

Note: when you pass one or more `-e/--env-file`, cascading is automatically disabled (with a warning) so only the explicit files are used. Use `--cascade` without `-e` when you want the cascade chain (`.env`, `.env.local`, `.env.<env>`, `.env.<env>.local`).

### Cascade env files automatically

Loads `.env`, `.env.local`, `.env.<env>`, `.env.<env>.local`:

```sh
nv run --cascade --env=production -- ./deploy.sh
```

### Override values inline

```sh
nv run -e .env -o PORT=4200 -- ./myapp
```

### Preview what will be set

```sh
nv run -e .env --dry-run -- ./myapp
```

### Export for another tool

```sh
nv export -e .env --format=json
```

### Validate against a schema

Default schema is `.env.example`:

```sh
nv validate -e .env
```

Schema example:
```
DATABASE_URL=postgres://localhost
OPTIONAL=
# REQUIRED: API_KEY
```

### Print the current environment

```sh
nv print --sort
```

## Configuration

Configuration is optional. When you do want it, `nv` stores it in `~/.config/nv/config.toml`.

If you still have a legacy `~/.nv` globals file from v2, `nv config migrate` will import it into the config and back up the original to `~/.config/nv/nv.backup`.

### Quick start

```sh
nv config init    # Create config with defaults
nv config show    # View your current config
nv config edit    # Edit config in $EDITOR
```

### Global variables

Global variables apply to every `nv` command.

```sh
nv config globals list
nv config globals set AWS_REGION us-east-1
nv config globals unset AWS_REGION
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
nv config set globals.priority "last"
```

What each layer means:

- Command-line arguments: `KEY=value` prefixes on the command you run after `--`, applied by the launched process (highest priority).
- CLI flags: `-o/--override KEY=value` passed to `nv`, applied after files and globals.
- Cascading files: in cascade mode, loaded in order `.env`, `.env.local`, `.env.<env>`, `.env.<env>.local` (missing files are skipped).
- Standard files: any `--env-file/-e` files are loaded in the order provided; with `--auto-local` defaults, `.env` is followed by optional `.env.local`.
- Globals: `[globals.env]` from config; merged before files when `globals.priority=first` (default) or after files when set to `last`.

## `.env` format supported

- `KEY=value` assignments with optional `export` prefix
- Full-line comments with `#` and inline comments outside quotes
- Single-quoted and double-quoted values (with escapes)
- Multiline values inside quotes
- Variable interpolation in unquoted and double-quoted values (`$VAR`, `${VAR}`)
- `PATH` expansions keep references to incoming `PATH`

## Troubleshooting

### Permission denied accessing config

```sh
chmod 644 ~/.config/nv/config.toml
```

### Config file corrupted

```sh
nv config validate
nv config reset
```

### Restore from backup

```sh
cp ~/.config/nv/nv.backup ~/.config/nv/config.toml
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
