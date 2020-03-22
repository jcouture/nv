# nv

`nv` is a lightweight utility to load context specific environment variables from either a single or multiple `.env` files before executing a command or command line program, along with its parameters.

As of version 2, the environment is cleared-out before loading context specific variables, except for `$PATH`.

## Why?

Why use `nv` when there are many [other](https://github.com/motdotla/dotenv) [tools](https://github.com/bkeepers/dotenv) that do pretty much the same thing automatically?

The difference is that `nv` _feeds_ an explicit environment to the process it starts, while those other tools _fetch_ an environment (based on some filename convention) after the process is started.

`nv` is also not language-specific nor framework-specific — it just _feeds_ some environment into the command it’s given to run.

## Install

You can install `nv` as a [Homebrew](https://brew.sh) package on macOS:

```
~> brew install jcouture/nv/nv
```

Or you can build it from source:

```
~> mkdir nv
~> cd nv
~> set -x GOPATH $PWD
~> go get -u github.com/jcouture/nv
```

## Usage example

You create a `.env` file as follows:

```
PORT=4200
SECRET_KEY_BASE=3b4476c0f6793b575050a1241438c32de8cbd3b7dec67910369657e1c4c41785
# Comments are supported
DATABASE_URL=postgres://dbuser:@localhost:5432/playground_dev?pool=10
```

You are ready to use `nv` to load your context specific environment variables.

```
~> nv .env rails server -p 2808
```

It is possible to load multiple `.env` files by separating each filenames with a comma. If a variable exists in more than one file, its value will simply be overriden as parsing goes.

```
~> nv .env,.env.dev rails server -p 2808
```

## Global variables

You might need to have global environment variables, overriding context specific ones. Create a file named `~/.nv` at the root of your home directory. It has the same format, and will be loaded last.

## Troubleshooting

If after running a command with `nv`, such as

```
nv .env rails --version
```
you get the following error

```
unknown command: rails. Perhaps you have to reshim?
```

try adding the following to your `~/.nv` file.

```
HOME=/Users/<<your username>>
USER=<<your username>>
```

## License

`nv` is released under the MIT license. See [LICENSE](./LICENSE) for details.
