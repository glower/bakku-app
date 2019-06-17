package watcher

import (
	"context"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/file-watcher/notification"
)

// Setup adds a watcher for a file changes in specified directories and returns a channel for notifications
func Setup(ctx context.Context, conf *config.WatchConfig, actionFilters []notification.ActionType, fileFilters []string, options *Options) (chan notification.Event, chan notification.Error) {
	eventCh := make(chan notification.Event)
	errorCh := make(chan notification.Error)

	if options == nil {
		options = &Options{IgnoreDirectoies: true}
	}

	watcher := Create(eventCh, errorCh, actionFilters, fileFilters, options)

	for _, e := range conf.DirsToWatch {
		if e.Active {
			go watcher.StartWatching(e.Path)
		}
	}

	return eventCh, errorCh
}
