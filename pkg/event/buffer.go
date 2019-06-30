package event

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/file-watcher/notification"
	metrics "github.com/rcrowley/go-metrics"
)

var (
	eventsM    sync.RWMutex
	events     = make(map[string]notification.Event)
	inProgress int32 // int64?
	done       int32
	// numberOfErrors
	maxInProgress = 5
)

// Buffer TODO: rename me to Buffer!
type Buffer struct {
	Ctx context.Context
	r   *types.GlobalResources

	maxElementsInBuffer   int32
	maxElementsInProgress int32
	evenInCh              chan notification.Event
	timeout               time.Duration

	EvenOutCh        chan notification.Event
	BackupCompleteCh chan types.BackupComplete
	BackupStatusCh   chan types.BackupStatus

	errorsCount metrics.EWMA
}

// NewBuffer ...
func NewBuffer(ctx context.Context, res *types.GlobalResources) *Buffer {

	// a := metrics.NewEWMA1()

	b := &Buffer{
		Ctx:                   ctx,
		maxElementsInBuffer:   1000,
		maxElementsInProgress: int32(maxInProgress),
		timeout:               5 * time.Second,
		EvenOutCh:             make(chan notification.Event),
		BackupStatusCh:        make(chan types.BackupStatus),
		r:                     res,
		errorsCount:           metrics.NewEWMA1(),
	}
	go b.processEvents()
	return b
}

func (b *Buffer) processEvents() {

	checkErrorRate := time.Tick(5 * time.Second)

	for {
		select {
		case <-b.Ctx.Done():
			return
		case e := <-b.r.FileWatcher.EventCh:
			// TODO: add something like:
			// Status: "scanning",
			b.addEvent(e.AbsolutePath, e)
		case c := <-b.r.BackupCompleteCh:
			if c.Success {
				atomic.AddInt32(&inProgress, -1)
				atomic.AddInt32(&done, 1)
				removeEvent(c.FilePath)
				b.BackupStatusCh <- types.BackupStatus{
					FilesDone:       int(done),
					FilesInProgress: int(inProgress),
					TotalFiles:      len(events),
					Status:          "uploading",
				}
			} else {
				// TODO: count errors, slow down, ...
				fmt.Printf("!!!!!! buffer: error uploading %s\n", c.FilePath)
				b.errorsCount.Update(1)
			}
		case <-checkErrorRate:
			fmt.Printf("!!!!!! buffer: errors per min: %v\n", b.errorsCount.Rate())
			b.errorsCount.Tick()
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
