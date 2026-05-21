# fw

A file watcher CLI that watches directories or files for changes and executes commands. Built with Go, powered by
[fsnotify](https://github.com/fsnotify/fsnotify) and [urfave/cli](https://github.com/urfave/cli).

## Features

- Watch files or directories (recursive)
- Execute one or more commands on file changes
- Built-in 500ms debounce to coalesce rapid-fire events
- File extension filtering
- Ignore patterns with glob support (`.gitignore` auto-loaded)
- Automatic watching of newly created subdirectories
- TOML config file support (`fw.toml`)
- Debug logging mode
- Graceful shutdown via SIGINT/SIGTERM

## Usage

### Basic usage

Watch the current directory:

```bash
fw
```

Watch a specific directory:

```bash
fw ./src
```

Watch a specific file:

```bash
fw ./config.yaml
```

### Execute commands on changes

Run a command when files change:

```bash
fw -c "npm run build"
```

Watch a directory and run tests on changes:

```bash
fw ./src -c "go test ./..."
```

Run multiple commands in sequence:

```bash
fw -c "npm run build" -c "npm run test"
```

Commands execute in the order specified. If a command fails, execution continues with the next command.

### Extension filtering

Filter events to specific file extensions:

```bash
fw -c "go test ./..." -e .go
```

Watch multiple extensions:

```bash
fw -c "npm run build" -e .js -e .ts -e .css
```

Extensions can be specified with or without a leading dot (`go` and `.go` both work). Matching is case-insensitive.

### Ignore patterns

Ignore files or directories matching glob patterns:

```bash
fw -c "go test ./..." -i "*_test.go" -i vendor
```

Patterns prefixed with `!` act as exceptions (un-ignore previously matched paths). If a `.gitignore` file exists in the
current directory, its entries are automatically loaded.

### Debug mode

Enable debug logging to see internal event processing details:

```bash
fw -D -c "npm run build"
```

### Config file

Create a `fw.toml` (or `.fw.toml`) file in your project directory:

```toml
path = "./src"
commands = ["npm run build", "npm test"]
extensions = [".js", ".ts"]
ignore = ["*_test.go", "vendor", ".git"]
debug = false
```

All fields are optional. `commands`, `extensions`, and `ignore` are merged between the config file and CLI flags; `path`
and `debug` are overridden by CLI flags. Auto-detected in the current directory; use `-C` to specify a custom path:

```bash
fw -C ./config/fw.toml
```

## Installation

```bash
go install github.com/xrzks/fw@latest
```

## Roadmap

- Environment variable expansion (`$FILE`, `$EVENT`)
- Stop on first command failure flag
