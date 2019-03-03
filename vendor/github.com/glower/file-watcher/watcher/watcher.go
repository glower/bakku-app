package watcher

import (
	"context"
	"log"

	"github.com/glower/file-watcher/notification"
)

// Setup adds a watcher for a file changes in specified directories and returns a channel for notifications
func Setup(ctx context.Context, dirsToWatch []string, actionFilters []notification.ActionType, fileFilters []string, options *Options) (chan notification.Event, chan notification.Error) {
	log.Printf("watchers.SetupFSWatchers(): for %v\n", dirsToWatch)

	eventCh := make(chan notification.Event)
	errorCh := make(chan notification.Error)

	if options == nil {
		options = &Options{IgnoreDirectoies: true}
	}

	watcher := Create(eventCh, errorCh, actionFilters, fileFilters, options)

	for _, dir := range dirsToWatch {
		go watcher.StartWatching(dir)
	}

	return eventCh, errorCh
}
