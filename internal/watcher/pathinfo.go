package watcher

import (
	"fmt"
	"os"

	"github.com/xrzks/fw/internal/logger"
)

type PathInfo struct {
	Path     string
	PathType string
	logger   *logger.Logger
}

func NewPathInfo(path string, logger *logger.Logger) (*PathInfo, error) {
	logger.Debug("Creating path info for: %s", path)
	info, err := os.Stat(path)
	if err != nil {
		logger.Error("Failed to stat path: %v", err)
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	pathType := "directory"
	if info.Mode().IsRegular() {
		pathType = "file"
	}

	logger.Debug("Path type: %s", pathType)
	return &PathInfo{
		Path:     path,
		PathType: pathType,
		logger:   logger,
	}, nil
}

func (p *PathInfo) PrintStatus(debounceMs int) {
	p.logger.Debug("Printing status with debounce: %dms", debounceMs)
	p.logger.Info("Watching %s: %s (debounce: %dms)", p.PathType, p.Path, debounceMs)
}
