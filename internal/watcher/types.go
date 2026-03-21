package watcher

// Config holds the watcher configuration
type Config struct {
	Path     string // Path to watch
	Command  string // Command to execute on changes
	Debounce int    // Debounce delay in milliseconds
}

// Event represents a filesystem event
type Event struct {
	Path string
	Op   string
}
