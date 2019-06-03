package event

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/file-watcher/notification"
)

var (
	eventsM    sync.RWMutex
	events     = make(map[string]notification.Event)
	inProgress int32
)

// Cache ...
type Cache struct {
	Ctx context.Context

	MaxElementsInCache    int32
	MaxElementsInProgress int32
	// var ops uint64
	Timeout                time.Duration
	EvenInCh               chan notification.Event
	EvenOutCh              chan notification.Event
	FileBackupCompleteChan chan types.FileBackupComplete
}

// New ...
func New(ctx context.Context, eventInCh chan notification.Event, fileBackupCompleteChan chan types.FileBackupComplete) *Cache {
	// eventCh := make(chan notification.Event)
	c := &Cache{
		Ctx:                    ctx,
		MaxElementsInCache:     1000,
		MaxElementsInProgress:  5,
		Timeout:                5 * time.Second,
		EvenInCh:               eventInCh,
		EvenOutCh:              make(chan notification.Event),
		FileBackupCompleteChan: fileBackupCompleteChan,
	}
	go c.processEvents()
	return c
}

func (c *Cache) processEvents() {
	for {
		select {
		case <-c.Ctx.Done():
			return
		// case <-c.FileBackupCompleteChan:
		// 	atomic.AddInt32(&inProgress, -1)
		case e := <-c.EvenInCh:
			addEvent(e.AbsolutePath, e)
		case <-time.After(c.Timeout):
			go c.sendAllBack()
		}
	}
}

func (c *Cache) sendAllBack() {
	for _, e := range events {
		if inProgress >= c.MaxElementsInProgress {
			fmt.Printf(">>> %d/%d files are progress, wait ... \n", inProgress, len(events))
			for {
				select {
				case <-c.Ctx.Done():
					return
				case <-c.FileBackupCompleteChan:
					atomic.AddInt32(&inProgress, -1)
					fmt.Printf(">>> continue ... \n")
					return
				}
			}
		} else {
			c.EvenOutCh <- e
			removeEvent(e.AbsolutePath)
			atomic.AddInt32(&inProgress, 1)
		}
	}
}

// addEvent adds an event to the internal cache
func addEvent(path string, e notification.Event) {
	eventsM.Lock()
	defer eventsM.Unlock()
	events[path] = e
}

func removeEvent(path string) {
	eventsM.Lock()
	defer eventsM.Unlock()
	delete(events, path)
}
