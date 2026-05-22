[![Go Report Card](https://goreportcard.com/badge/github.com/xrzks/fw)](https://goreportcard.com/report/github.com/xrzks/fw)

# fw - file watcher

CLI tool that watches directories or files for changes and executes commands.
Built with Go, powered by [fsnotify](https://github.com/fsnotify/fsnotify) and
[urfave/cli](https://github.com/urfave/cli).

## Installation

```bash
go install github.com/xrzks/fw@latest
```

## Usage

```bash
# watch current directory and run a command on every change
fw -c "go test ./..."

# watch a specific path
fw ./src -c "npm run build"

# run multiple commands in sequence
fw -c "go build ./..." -c "go test ./..."

# only watch .go files
fw -e .go -c "go test ./..."

# ignore a directory
fw -i vendor -c "go build ./..."

# stop on first command failure
fw -f -c "go build ./..." -c "go test ./..."
```

## Options

| Flag                 | Alias | Description                                       |
| -------------------- | ----- | ------------------------------------------------- |
| `--command <cmd>`    | `-c`  | Command to run on change (repeatable)             |
| `--extension <ext>`  | `-e`  | Only watch files with this extension (repeatable) |
| `--ignore <pattern>` | `-i`  | Glob pattern to ignore (repeatable)               |
| `--config <path>`    | `-C`  | Path to config file                               |
| `--no-gitignore`     |       | Disable automatic `.gitignore` loading            |
| `--fail-fast`        | `-f`  | Stop subsequent commands on first failure         |
| `--debug`            | `-D`  | Enable debug logging                              |

## Config file

`fw` auto-detects `fw.toml` or `.fw.toml` in the current directory. CLI flags
take precedence over config file values. Commands from both sources are merged
and run in order.

```toml
path = "."
commands = ["go test ./..."]
extensions = [".go"]
ignore = ["vendor", ".git"]
fail-fast = false
no-gitignore = false
debug = false
```
