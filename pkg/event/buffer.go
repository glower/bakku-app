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
	evenInCh              chan notification.Event
	timeout               time.Duration

	EvenOutCh      chan notification.Event
	BackupDoneCh   chan types.FileBackupComplete
	BackupStatusCh chan types.BackupStatus
}

// New ...
func New(ctx context.Context, eventInCh chan notification.Event) *Buffer {
	// eventCh := make(chan notification.Event)
	b := &Buffer{
		Ctx:                   ctx,
		maxElementsInBuffer:   1000,
		maxElementsInProgress: 5,
		timeout:               5 * time.Second,
		evenInCh:              eventInCh,
		EvenOutCh:             make(chan notification.Event),
		BackupDoneCh:          make(chan types.FileBackupComplete),
		BackupStatusCh:        make(chan types.BackupStatus),
	}
	go b.processEvents()
	return b
}

func (b *Buffer) processEvents() {
	for {
		select {
		case <-b.Ctx.Done():
			return
		case e := <-b.evenInCh:
			b.addEvent(e.AbsolutePath, e)
		case <-time.After(b.timeout):
			go b.sendAllBack()
		}
	}
}

func (b *Buffer) sendAllBack() {
	for _, e := range events {
		if inProgress >= b.maxElementsInProgress {
			fmt.Printf(">>> %d/%d files are progress, wait ... \n", inProgress, len(events))
			for {
				select {
				case <-b.Ctx.Done():
					return
				case <-b.BackupDoneCh:
					atomic.AddInt32(&inProgress, -1)
					fmt.Printf(">>> continue ... \n")
					return
				}
			}
		} else {
			b.EvenOutCh <- e
			removeEvent(e.AbsolutePath)
			atomic.AddInt32(&inProgress, 1)
		}
		b.BackupStatusCh <- types.BackupStatus{
			FilesInProgress: int(inProgress),
			TotalFiles:      len(events),
			Status:          "uploading",
		}
	}
	b.BackupStatusCh <- types.BackupStatus{
		FilesInProgress: 0,
		TotalFiles:      0,
		Status:          "waiting",
	}
}

// addEvent adds an event to the internal cache
func (b *Buffer) addEvent(path string, e notification.Event) {
	eventsM.Lock()
	defer eventsM.Unlock()
	events[path] = e

	b.BackupStatusCh <- types.BackupStatus{
		FilesInProgress: 0,
		TotalFiles:      len(events),
		Status:          "preparing",
	}
}

func removeEvent(path string) {
	eventsM.Lock()
	defer eventsM.Unlock()
	delete(events, path)
}
