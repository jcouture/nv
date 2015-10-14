# nv

nv is a lightweight utility to load context specific environment variables from a `.env` file before executing a command or command line program.

## Build the Source

```
$ mkdir nv
$ cd nv
$ export GOPATH=`pwd`
$ go get -u github.com/jcouture/nv
```

## Usage Example

You create a `.env` file as follows:

```
PORT=4200
SECRET_KEY_BASE=3b4476c0f6793b575050a1241438c32de8cbd3b7dec67910369657e1c4c41785
DATABASE_URL=postgres://dbuser:@localhost:5432/playground_dev?pool=10
```

You are ready to use `nv` to load your context specific environment variables

```
$ nv .env rails server
```

## License

nv is released under the MIT license. See LICENSE for details.
