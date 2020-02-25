package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/fsnotify.v1"
)

// watch watches for changes starting at the root directory. Changes are passed to
// the returned channel.
func watch(ctx context.Context, root string) (<-chan string, error) {
	if root == "" {
		root = "."
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("watcher: %w", err)
	}

	ch := make(chan string, 1)
	go func() {
		defer watcher.Close()
		for {
			select {
			case event := <-watcher.Events:
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					if err := watchAll(watcher, event.Name); err != nil {
						log.Print(err)
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
				case event.Op&fsnotify.Write == fsnotify.Write:
				case event.Op&fsnotify.Rename == fsnotify.Rename:
				default:
					// Ignore other events.
					continue
				}

				if !strings.HasSuffix(event.Name, ".go") {
					break
				}
				ch <- event.Name
			case err := <-watcher.Errors:
				log.Print(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	err = watchAll(watcher, root)
	return ch, err
}

// watchAll adds path and all of its child directories to the watch list.
func watchAll(watcher *fsnotify.Watcher, path string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Apply filters.
		// TODO: Make this configurable.
		if info.IsDir() {
			if filepath.Base(path) == "node_modules" {
				return filepath.SkipDir
			}
			if filepath.Base(path) == ".git" {
				return filepath.SkipDir
			}
			if err := watcher.Add(path); err != nil {
				return fmt.Errorf("watcher: %w", err)
			}
			fmt.Println("Watching:", path)
		}
		return nil
	})
}
