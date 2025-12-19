<p align="center">
  <a href="https://github.com/jcouture/nv">
    <img src="https://user-images.githubusercontent.com/5007/120239413-3ba5c000-c22c-11eb-8008-052bc5f8e7b8.png" alt="nv" />
  </a>
</p>

[![Release](https://img.shields.io/github/release/jcouture/nv.svg?style=for-the-badge)](https://github.com/jcouture/nv/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/jcouture/nv)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcouture/nv?style=for-the-badge)](https://https://goreportcard.com/badge/github.com/jcouture/nv)

`nv` is a lightweight utility to load context specific environment variables from `.env` files before executing a command.

This repo ships two binaries:

- `nvx` is the current, feature-rich CLI.
- `nv` is the legacy v2-compatible CLI with the original UX.

Both tools build an explicit environment and then run your command. By default, `nvx` keeps `$PATH`, `$HOME`, and `$USER`, while legacy `nv` keeps only `$PATH`.

**Warning (legacy nv)**

If you are using a version manager such as [asdf](https://asdf-vm.com), both variables `$HOME` and `$USER` could be required. Please see the [Troubleshooting](#troubleshooting) section for more information.

## Why?

Why use `nv` when there are many [other](https://github.com/motdotla/dotenv) [tools](https://github.com/bkeepers/dotenv) that do pretty much the same thing automatically?

The difference is that `nv` _feeds_ an explicit environment to the process it starts, while those other tools _fetch_ an environment (based on some filename convention) after the process is started.

`nv` is also not language-specific nor framework-specific — it just _feeds_ some environment into the command it’s given to run.

## Installation

### macOS

`nv` is available via [Homebrew](#homebrew) and as a downloadable binary from the [releases page](https://github.com/jcouture/nv/releases).

#### Homebrew

| Install                       | Upgrade           |
| ----------------------------- | ----------------- |
| `brew install jcouture/nv/nv` | `brew upgrade nv` |

### Linux

`nv` is available as downloadable binaries from the [releases page](https://github.com/jcouture/nv/releases).

### Windows

`nv` is available as a downloadable binary from the [releases page](https://github.com/jcouture/nv/releases).

### Build from source

Alternatively, you can build it from source.

1. Verify you have Go 1.25+ installed

```sh
~> go version
```

If `Go` is not installed, follow the instructions on the [Go website](https://golang.org/doc/install)

2. Clone this repository

```sh
~> git clone https://github.com/jcouture/nv.git
~> cd nv
```

3. Install build dependencies (using [mise-en-place](https://mise.jdx.dev/))

```sh
~> mise install
```

4. Build

```sh
~> make build
```

While the development version is a good way to take a peek at `nv`’s latest features before they get released, be aware that it may contains bugs. Officially released versions will generally be more stable.

## Which binary should I use?

- Use `nvx` for new projects and richer features.
- Use `nv` if you want the v2 UX and behavior.

## nvx manual

Run a command with one or more `.env` files:

```sh
~> nvx run -e .env -- rails server -p 2808
~> nvx run -e .env -e .env.local -- npm start
```

Cascading mode (loads `.env`, `.env.local`, `.env.<env>`, `.env.<env>.local`):

```sh
~> nvx run --cascade --env=production -- ./deploy.sh
```

Inline overrides and strict interpolation:

```sh
~> nvx run -e .env -o PORT=4200 --strict -- ./myapp
```

Preview or export the compiled environment (masked by default):

```sh
~> nvx run -e .env --dry-run -- ./myapp
~> nvx export -e .env --format=json
```

Unredact or add masking rules:

```sh
~> nvx export -e .env --unredacted
~> nvx export -e .env --mask-pattern "(?i)token|secret"
```

Validate against a schema (default: `.env.example`):

```sh
~> nvx validate -e .env
~> nvx validate -e .env --schema .env.production.example --schema-strict
```

Schema example:

```
DATABASE_URL=postgres://localhost
OPTIONAL=
# REQUIRED: API_KEY
```

By default, `nvx` preserves `$PATH`, `$HOME`, and `$USER`. Use `-p` to adjust what is kept.

## nv (legacy v2) usage

Run a command with one or more `.env` files:

```sh
~> nv .env rails server -p 2808
~> nv .env,.env.dev rails server -p 2808
```

## Supported `.env` syntax

- `KEY=value` assignments with optional `export` prefix.
- Full-line comments beginning with `#` and inline comments outside quotes.
- Unquoted values, single-quoted literals, and double-quoted values with escapes (`\\`, `\n`, `\r`, `\t`, `\"`).
- Multiline values inside single or double quotes.
- Variable interpolation in unquoted and double-quoted values using `$VAR` and `${VAR}` (earlier definitions win, then existing environment).
- `PATH` expansions keep references to the incoming `PATH` when `$PATH`/`${PATH}` is present.

## Global variables (nv legacy)

For machine-wide overrides, create `~/.nv` (loaded last). `nvx` *does not* read `~/.nv`; use `-e` or `-o` instead.

## Troubleshooting (legacy nv)

If you rely on shims or TTY tools, add these to `~/.nv`:

```
HOME=<your home directory>
USER=<your username>
TERM=xterm-color
```

## License

`nv` is released under the MIT license. See [LICENSE](./LICENSE) for details.

The `nv` leaf logo is based on [this icon](https://thenounproject.com/term/leaf/1904973/) by [Nick Bluth](https://thenounproject.com/nickbluth/), from the Noun Project. Used under a [Creative Commons BY 3.0](http://creativecommons.org/licenses/by/3.0/) license.
