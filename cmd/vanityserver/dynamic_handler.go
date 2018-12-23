package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type dynamicHandler struct {
	*fsnotify.Watcher

	h         http.Handler
	unhealthy bool
	mu        sync.RWMutex
}

func newDynamicHandler(file string, generator func() (http.Handler, error)) *dynamicHandler {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create fsnotify.Watcher: %v", err)
	}

	if err := w.Add(file); err != nil {
		log.Fatalf("Could not watch file %q: %v", file, err)
	}

	h, err := generator()
	if err != nil {
		log.Fatalf("Failed generating initial handler: %v", err)
	}

	dh := &dynamicHandler{Watcher: w, h: h}
	updateHandler := func(h http.Handler, err error) error {
		dh.mu.Lock()
		defer dh.mu.Unlock()

		if err != nil {
			dh.unhealthy = true
			return err
		}
		dh.unhealthy = false

		if h == nil {
			panic("nil handler returned from generator")
		}
		dh.h = h
		return nil
	}

	go func() {
		log.Printf("Watching for changes to %q", file)
		for {
			select {
			case evt, ok := <-w.Events:
				if !ok {
					return
				}
				if evt.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename) == 0 {
					continue
				}
				if err := updateHandler(generator()); err != nil {
					log.Printf("Error switching to new handler: %v", err)
				} else {
					log.Printf("Updated to new handler based on %q", file)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Printf("Error in fsnotify.Watcher: %v", err)
			}
		}
	}()

	return dh
}

func (dh *dynamicHandler) IsHealthy() bool {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	return !dh.unhealthy
}

func (dh *dynamicHandler) getHandler() http.Handler {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	return dh.h
}

func (dh *dynamicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dh.getHandler().ServeHTTP(w, r)
}
