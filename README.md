# fw

A file watcher CLI that watches directories or files for changes and executes commands. Built with Go, powered by [fsnotify](https://github.com/fsnotify/fsnotify) and [urfave/cli](https://github.com/urfave/cli).

## Features

- Watch files or directories (recursive)
- Execute one or more commands on file changes
- Built-in 500ms debounce to coalesce rapid-fire events
- File extension filtering
- Automatic watching of newly created subdirectories
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

### Debug mode

Enable debug logging to see internal event processing details:

```bash
fw -D -c "npm run build"
```

## Installation

```bash
go install github.com/xrzks/fw@latest
```

## Roadmap

- Local config file (`.fw.json`)
- Ignore patterns (`.gitignore` style)
- Environment variable expansion (`$FILE`, `$EVENT`)
- Stop on first command failure flag (`-e`)
