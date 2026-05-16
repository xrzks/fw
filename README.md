# fw

A simple file watcher CLI that watches directories or files for changes and executes commands.

## Features

- Watch directories or files
- Execute one or more commands on file changes
- Configurable debounce delay to handle rapid-fire events
- Shows which file and operation triggered command execution
- Simple, language-agnostic CLI - works with any project

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

Commands execute in the order specified. If a command fails, execution continues with the next command. The tool outputs "change detected: <filename> (<operation>)" when changes trigger command execution.

### Configure debounce delay

Debounce delay determines how long (in milliseconds) to wait after the last file change before running commands. This prevents running commands multiple times for rapid successive changes (e.g., during quick edits).

Set custom debounce delay (default: 500ms):

```bash
fw -d 2000 -c "npm run build"
```

Use a longer debounce for slow builds to accumulate rapid changes into a single execution:

```bash
fw ./src -c "make" -d 5000
```

## Roadmap

Planned features:

- Environment variable expansion (`$FILE`, `$EVENT`)
- File extension filtering
- Ignore patterns (`.gitignore` style)
- Graceful shutdown (SIGINT/SIGTERM)
- Dont continue with next command if the previous one fails flag, end on failure (`-e true`)

## Philosophy

**Simple over complex** - fw aims to fill the gap between:

- Complex tools like `watchexec` with many features
- Write-your-own shell scripts

**Flexible** - Language-agnostic through shell commands, works for any project.

## License

MIT
