package main

import (
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type DebouncedWatcher struct {
	mu       sync.Mutex
	pending  map[string]struct{}
	timer    *time.Timer
	debounce time.Duration
	callback func(files []string)
}

func NewDebouncedWatcher(debounce time.Duration, callback func(files []string)) *DebouncedWatcher {
	return &DebouncedWatcher{
		pending:  make(map[string]struct{}),
		debounce: debounce,
		callback: callback,
	}
}

func (dw *DebouncedWatcher) AddEvent(path string) {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	dw.pending[path] = struct{}{}

	if dw.timer != nil {
		dw.timer.Stop()
	}

	dw.timer = time.AfterFunc(dw.debounce, func() {
		dw.mu.Lock()
		files := make([]string, 0, len(dw.pending))
		for f := range dw.pending {
			files = append(files, f)
		}
		dw.pending = make(map[string]struct{})
		dw.mu.Unlock()

		dw.callback(files)
	})
}

func WatchDirectory(dir string, debounce time.Duration, callback func(files []string)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	dw := NewDebouncedWatcher(debounce, callback)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op.Has(fsnotify.Write) ||
					event.Op.Has(fsnotify.Create) ||
					event.Op.Has(fsnotify.Remove) {
					dw.AddEvent(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("watcher error:", err)
			}
		}
	}()

	if err := watcher.Add(dir); err != nil {
		log.Fatal(err)
	}
}
