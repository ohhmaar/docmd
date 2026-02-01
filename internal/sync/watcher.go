package sync

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type WatchConfig struct {
	DebounceMs int
	OnChange   func(filePath string) error
}

func WatchFile(filePath string, config WatchConfig) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	debounceDuration := time.Duration(config.DebounceMs) * time.Millisecond
	var debounceTimer *time.Timer

	errChan := make(chan error, 1)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				eventPath, _ := filepath.Abs(event.Name)

				if eventPath == absPath {
					if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Chmod) {
						if debounceTimer != nil {
							debounceTimer.Stop()
						}
						debounceTimer = time.AfterFunc(debounceDuration, func() {
							if err := config.OnChange(absPath); err != nil {
								timestamp := time.Now().Format("15:04:05")
								fmt.Printf("[%s] Error: %v\n", timestamp, err)
							}
						})
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				errChan <- err
				return
			}
		}
	}()

	dir := filepath.Dir(absPath)
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	watcher.Add(absPath)

	select {
	case err := <-errChan:
		return err
	}
}

func WatchFiles(filePaths []string, config WatchConfig) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	watchedFiles := make(map[string]bool)
	watchedDirs := make(map[string]bool)

	for _, fp := range filePaths {
		absPath, err := filepath.Abs(fp)
		if err != nil {
			continue
		}
		watchedFiles[absPath] = true
		watchedDirs[filepath.Dir(absPath)] = true
	}

	debounceDuration := time.Duration(config.DebounceMs) * time.Millisecond
	debounceTimers := make(map[string]*time.Timer)

	errChan := make(chan error, 1)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				eventPath, _ := filepath.Abs(event.Name)

				if watchedFiles[eventPath] {
					if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Chmod) {
						if timer, exists := debounceTimers[eventPath]; exists && timer != nil {
							timer.Stop()
						}

						debounceTimers[eventPath] = time.AfterFunc(debounceDuration, func() {
							if err := config.OnChange(eventPath); err != nil {
								timestamp := time.Now().Format("15:04:05")
								fmt.Printf("[%s] Error: %v\n", timestamp, err)
							}
						})
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				errChan <- err
				return
			}
		}
	}()

	for dir := range watchedDirs {
		if err := watcher.Add(dir); err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", dir, err)
		}
	}

	for fp := range watchedFiles {
		watcher.Add(fp)
	}

	select {
	case err := <-errChan:
		return err
	}
}
