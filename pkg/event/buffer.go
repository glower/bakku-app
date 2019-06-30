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
	eventsM       sync.RWMutex
	events        = make(map[string]notification.Event)
	inProgress    int32 // int64?
	done          int32
	maxInProgress = 5
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

	r *types.GlobalResources
}

// NewBuffer ...
func NewBuffer(ctx context.Context, res *types.GlobalResources) *Buffer {
	b := &Buffer{
		Ctx:                   ctx,
		maxElementsInBuffer:   1000,
		maxElementsInProgress: int32(maxInProgress),
		timeout:               1 * time.Second,
		EvenOutCh:             make(chan notification.Event),
		BackupStatusCh:        make(chan types.BackupStatus),
		r:                     res,
	}
	go b.processEvents()
	return b
}

func (b *Buffer) processEvents() {
	for {
		select {
		case <-b.Ctx.Done():
			return
		case e := <-b.r.FileWatcher.EventCh:
			b.addEvent(e.AbsolutePath, e)
		case c := <-b.r.BackupCompleteCh:
			atomic.AddInt32(&inProgress, -1)
			atomic.AddInt32(&done, 1)
			fmt.Printf("<<<<< buffer.processEvents(): BackupComplete: %s, inProgress=%d, total done=%d\n", c.FilePath, inProgress, done)
			b.BackupStatusCh <- types.BackupStatus{
				FilesDone:       int(done),
				FilesInProgress: int(inProgress),
				TotalFiles:      len(events),
				Status:          "uploading",
			}
		case <-time.After(b.timeout):
			if len(events) != 0 && inProgress == 0 {
				go b.send()
			}
		}
	}
}

// TODO: we can have multiple BackupDone events for the same file (different backup provider!)
// so the `done` counter needs to be fixed somehow!
func (b *Buffer) send() {
	for _, e := range events {
		b.EvenOutCh <- e
		fmt.Printf("<<<<< buffer.send(): BackupComplete: %s, inProgress=%d, total done=%d\n", e.AbsolutePath, inProgress, done)
		removeEvent(e.AbsolutePath)
		atomic.AddInt32(&inProgress, 1)
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
	eventsM.Lock()
	defer eventsM.Unlock()
	events[path] = e
}

func removeEvent(path string) {
	eventsM.Lock()
	defer eventsM.Unlock()
	delete(events, path)
}
