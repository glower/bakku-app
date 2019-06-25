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
	inProgress int32 // int64?
	done       int32
)

// Buffer TODO: rename me to Buffer!
type Buffer struct {
	Ctx context.Context

	maxElementsInBuffer   int32
	maxElementsInProgress int32
	evenInCh              chan notification.Event
	timeout               time.Duration

	EvenOutCh        chan notification.Event
	BackupCompleteCh chan types.BackupComplete
	BackupStatusCh   chan types.BackupStatus
}

// NewBuffer ...
func NewBuffer(ctx context.Context, res types.GlobalResources) *Buffer {
	fmt.Println("event.NewBuffer(): starting event buffer")
	b := &Buffer{
		Ctx:                   ctx,
		maxElementsInBuffer:   1000,
		maxElementsInProgress: 5,
		timeout:               1 * time.Second,
		evenInCh:              res.FileWatcher.EventCh,
		EvenOutCh:             make(chan notification.Event),
		BackupStatusCh:        make(chan types.BackupStatus),
		BackupCompleteCh:      res.BackupCompleteCh,
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

// TODO: we can have multiple BackupDone events for the same file (different backup provider!)
// so the `done` counter needs to be fixed somehow!
func (b *Buffer) sendAllBack() {
	for _, e := range events {
		if inProgress >= b.maxElementsInProgress {
			// fmt.Printf(">>> %d/%d files are in progress, wait ... \n", inProgress, len(events))
			for {
				select {
				case <-b.Ctx.Done():
					return
				case <-b.BackupCompleteCh:
					atomic.AddInt32(&inProgress, -1)
					atomic.AddInt32(&done, 1)
					return
				}
			}
		} else {
			b.EvenOutCh <- e
			removeEvent(e.AbsolutePath)
			atomic.AddInt32(&inProgress, 1)
		}

		b.BackupStatusCh <- types.BackupStatus{
			FilesDone:       int(done),
			FilesInProgress: int(inProgress),
			TotalFiles:      len(events),
			Status:          "uploading",
		}
	}

	b.BackupStatusCh <- types.BackupStatus{
		FilesDone:       0,
		FilesInProgress: 0,
		TotalFiles:      0,
		Status:          "waiting",
	}
}

// addEvent adds an event to the internal cache
func (b *Buffer) addEvent(path string, e notification.Event) {
	// fmt.Printf("event.Buffer.addEent(): file path [%s]\n", path)
	eventsM.Lock()
	defer eventsM.Unlock()
	events[path] = e
}

func removeEvent(path string) {
	eventsM.Lock()
	defer eventsM.Unlock()
	delete(events, path)
}
