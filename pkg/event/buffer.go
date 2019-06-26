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
		case <-b.BackupCompleteCh:
			fmt.Println("processEvents(): continue ...")
			atomic.AddInt32(&inProgress, -1)
			atomic.AddInt32(&done, 1)
			b.BackupStatusCh <- types.BackupStatus{
				FilesDone:       int(done),
				FilesInProgress: int(inProgress),
				TotalFiles:      len(events),
				Status:          "uploading",
			}
		case <-time.After(b.timeout):
			if len(events) != 0 && inProgress < b.maxElementsInProgress {
				e := getEvent()
				b.EvenOutCh <- e
				removeEvent(e.AbsolutePath)
				atomic.AddInt32(&inProgress, 1)
			}
		}
	}
}

// // TODO: we can have multiple BackupDone events for the same file (different backup provider!)
// // so the `done` counter needs to be fixed somehow!
// func (b *Buffer) sendAllBack() {
// 	total := len(events)
// 	for _, e := range events {
// 		if inProgress >= b.maxElementsInProgress {
// 			<-b.BackupCompleteCh
// 			fmt.Println("sendAllBack(): continue a ...")
// 			atomic.AddInt32(&inProgress, -1)
// 			atomic.AddInt32(&done, 1)
// 		} else {
// 		}
// 		b.BackupStatusCh <- types.BackupStatus{
// 			FilesDone:       int(done),
// 			FilesInProgress: int(inProgress),
// 			TotalFiles:      total,
// 			Status:          "uploading",
// 		}
// 	}
// 	b.BackupStatusCh <- types.BackupStatus{
// 		FilesDone:       0,
// 		FilesInProgress: 0,
// 		TotalFiles:      0,
// 		Status:          "waiting",
// 	}
// }

func getEvent() notification.Event {
	for _, e := range events {
		return e
	}
	return notification.Event{}
}

// addEvent adds an event to the internal cache
func (b *Buffer) addEvent(path string, e notification.Event) {
	eventsM.Lock()
	defer eventsM.Unlock()
	events[path] = e
}

func removeEvent(path string) {
	eventsM.Lock()
	defer eventsM.Unlock()
	delete(events, path)
}
