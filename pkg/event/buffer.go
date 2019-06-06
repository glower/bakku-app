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

// Buffer TODO: rename me to Buffer!
type Buffer struct {
	Ctx context.Context

	maxElementsInBuffer   int32
	maxElementsInProgress int32

	timeout      time.Duration
	EvenOutCh    chan notification.Event
	evenInCh     chan notification.Event
	BackupDoneCh chan types.FileBackupComplete
}

// New ...
func New(ctx context.Context, eventInCh chan notification.Event) *Buffer {
	// eventCh := make(chan notification.Event)
	c := &Buffer{
		Ctx:                   ctx,
		maxElementsInBuffer:   1000,
		maxElementsInProgress: 5,
		timeout:               5 * time.Second,
		evenInCh:              eventInCh,
		EvenOutCh:             make(chan notification.Event),
		BackupDoneCh:          make(chan types.FileBackupComplete),
	}
	go c.processEvents()
	return c
}

func (c *Buffer) processEvents() {
	for {
		select {
		case <-c.Ctx.Done():
			return
		// case <-c.FileBackupCompleteChan:
		// 	atomic.AddInt32(&inProgress, -1)
		case e := <-c.evenInCh:
			addEvent(e.AbsolutePath, e)
		case <-time.After(c.timeout):
			go c.sendAllBack()
		}
	}
}

func (c *Buffer) sendAllBack() {
	for _, e := range events {
		if inProgress >= c.maxElementsInProgress {
			fmt.Printf(">>> %d/%d files are progress, wait ... \n", inProgress, len(events))
			for {
				select {
				case <-c.Ctx.Done():
					return
				case <-c.BackupDoneCh:
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
