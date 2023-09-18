<p align="center">
  <a href="https://github.com/jcouture/nv">
    <img src="https://user-images.githubusercontent.com/5007/120239413-3ba5c000-c22c-11eb-8008-052bc5f8e7b8.png" alt="nv" />
  </a>
</p>

`nv` is a lightweight utility to load context specific environment variables from either a single or multiple `.env` files before executing a command or command line program, along with its parameters.

As of version 2, the environment is cleared-out before loading context specific variables, except for `$PATH`.

**Warning**

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

1. Verify you have Go 1.20+ installed

```sh
~> go version
```

If `Go` is not installed, follow the instructions on the [Go website](https://golang.org/doc/install)

2. Clone this repository

```sh
~> git clone https://github.com/jcouture/nv.git
~> cd nv
```

3. Build

```sh
~> go mod tidy
~> go build ./cmd/nv/
```

While the development version is a good way to take a peek at `nv`’s latest features before they get released, be aware that it may contains bugs. Officially released versions will generally be more stable.

## Usage example

Create a `.env` file as follows:

```
PORT=4200
SECRET_KEY_BASE=3b4476c0f6793b575050a1241438c32de8cbd3b7dec67910369657e1c4c41785
# Comments are supported
DATABASE_URL=postgres://dbuser:@localhost:5432/playground_dev?pool=10
```

You are then ready to use `nv` to load your context specific environment variables.

```sh
~> nv .env rails server -p 2808
```

It is possible to load multiple `.env` files by separating each filenames with a comma. If a variable exists in more than one file, its value will simply be overriden as parsing goes.

```sh
~> nv .env,.env.dev rails server -p 2808
```

## Global variables

You might need to have global environment variables, overriding context specific ones. Create a file named `~/.nv` at the root of your home directory. It has the same format, and _will be loaded last_.

## Troubleshooting

### Shims

If after executing a command with `nv`, such as:

```sh
~> nv .env rails --version
```

you get the following error:

```sh
unknown command: rails. Perhaps you have to reshim?
```

add the following to your `~/.nv` file:

```
HOME=<your home directory>
USER=<your username>
```

### Interactive TTY

If after executing a command with `nv`, such as:

```sh
~> nv .env less README.md
```

you get a similar error or warning:

```
WARNING: terminal is not fully functional
-  (press RETURN)
```

add the following to your `~/.nv` file:

```
TERM=xterm-color # or any other relevant value for `TERM`
```

## License

`nv` is released under the MIT license. See [LICENSE](./LICENSE) for details.

The `nv` leaf logo is based on [this icon](https://thenounproject.com/term/leaf/1904973/) by [Nick Bluth](https://thenounproject.com/nickbluth/), from the Noun Project. Used under a [Creative Commons BY 3.0](http://creativecommons.org/licenses/by/3.0/) license.
