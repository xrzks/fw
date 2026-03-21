package watcher

import (
	"fmt"
	"os"
)

type PathInfo struct {
	Path     string
	PathType string
}

func NewPathInfo(path string) (*PathInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	pathType := "directory"
	if info.Mode().IsRegular() {
		pathType = "file"
	}

	return &PathInfo{
		Path:     path,
		PathType: pathType,
	}, nil
}

func (p *PathInfo) PrintStatus() {
	fmt.Printf("Watching %s: %s\n", p.PathType, p.Path)
}
