<p align="center">
  <a href="https://github.com/jcouture/nv">
    <img src="https://user-images.githubusercontent.com/5007/120239413-3ba5c000-c22c-11eb-8008-052bc5f8e7b8.png" alt="nv" />
  </a>
</p>

[![Release](https://img.shields.io/github/release/jcouture/nv.svg?style=for-the-badge)](https://github.com/jcouture/nv/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/jcouture/nv)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcouture/nv?style=for-the-badge)](https://https://goreportcard.com/badge/github.com/jcouture/nv)

`nv` runs any command with predictable environment variables from your `.env` files. Think “one-shot dotenv runner” for scripts, deploys, and local apps.

## Quick start

**Install**

- macOS (Homebrew):
  ```sh
  brew install jcouture/nv/nv
  ```
- Linux/Windows/macOS (binary): download the latest release from https://github.com/jcouture/nv/releases

**Run a command with env vars**

```sh
nv run -e .env -- ./myapp
```

`nv` loads the file(s) you point at and starts your command with those variables. Nothing is left running after the command exits.

## Everyday moves

- **Cascade the usual dotenv chain**: `.env`, `.env.local`, `.env.<env>`, `.env.<env>.local`
  ```sh
  nv run --cascade --env=production -- ./deploy.sh
  ```
  When you pass `-e/--env-file`, cascading turns off (with a warning) so only the files you listed are used.

- **Point at specific files (multiple allowed)**
  ```sh
  nv run -e .env -e .env.local -- ./myapp
  ```

- **Override inline for a one-off**
  ```sh
  nv run -e .env -o PORT=4200 -- ./myapp
  ```

- **Preview without running**
  ```sh
  nv run -e .env --dry-run -- ./myapp
  ```

- **Export for another tool**
  ```sh
  nv export -e .env --format=json
  ```

- **Validate before you run** (defaults to `.env.example` as schema)
  ```sh
  nv validate -e .env
  ```
  Validation checks that your real `.env` has every required key and that required ones are non-empty. The schema file is just a list of keys with optional example values; it is not loaded at runtime.
  Schema example:
  ```
  DATABASE_URL=postgres://localhost   # example value is ignored
  OPTIONAL=                           # empty value means "can be blank or missing"
  # REQUIRED: API_KEY                 # tag required keys with this comment
  ```

- **Use a custom schema file**
  ```sh
  nv validate -e .env --schema=.env.staging.example
  ```

- **See what is currently set**
  ```sh
  nv print --sort
  ```

## Configuration (optional)

Config lives at `$XDG_CONFIG_HOME/nv/config.toml` (defaults to `~/.config/nv` on Linux, `~/Library/Application Support/nv` on macOS). Run `nv config path` to see the exact location. If you still have a legacy `~/.nv` file from v2, `nv config migrate` will import it and back up the original alongside the config file.

```sh
nv config init    # Create config with defaults
nv config show    # View your current config
nv config edit    # Edit config in $EDITOR
```

### Globals

Globals apply to every `nv` command.

```sh
nv config globals list
nv config globals set AWS_REGION us-east-1
nv config globals unset AWS_REGION
```

### Priority at a glance

Default order (highest first):
1. `KEY=value` prefixes on the command you run after `--`
2. `-o/--override KEY=value`
3. `.env.local` (when cascading)
4. `.env`
5. `[globals.env]`

Flip globals to load last if you prefer:

```sh
nv config set globals.priority "last"
```

## `.env` syntax supported

- `KEY=value` (with optional `export`)
- `#` comments on their own line or inline (outside quotes)
- Single- and double-quoted values with escapes, multiline inside quotes
- Variable interpolation in unquoted and double-quoted values (`$VAR`, `${VAR}`)
- `PATH` expansions preserve the incoming `PATH`

## Troubleshooting

- **Permission denied on config**
  ```sh
  chmod 644 "$(nv config path)"
  ```
- **Config seems corrupted**
  ```sh
  nv config validate
  nv config reset
  ```
- **Restore backup**
  ```sh
  cp "$(dirname "$(nv config path)")/nv.backup" "$(nv config path)"
  ```

## Build from source (latest dev)

1) Verify Go 1.25+
```sh
go version
```
2) Clone
```sh
git clone https://github.com/jcouture/nv.git
cd nv
```
3) Install build deps (via [mise-en-place](https://mise.jdx.dev/))
```sh
mise install
```
4) Build
```sh
make build
```

## License

`nv` is released under the MIT license. See [LICENSE](./LICENSE) for details.

The `nv` leaf logo is based on [this icon](https://thenounproject.com/term/leaf/1904973/) by [Nick Bluth](https://thenounproject.com/nickbluth/), from the Noun Project. Used under a [Creative Commons BY 3.0](http://creativecommons.org/licenses/by/3.0/) license.
