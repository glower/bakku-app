package event

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/file-watcher/notification"
	"github.com/paulbellamy/ratecounter"
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

	errorsRate *ratecounter.RateCounter
}

// NewBuffer ...
func NewBuffer(ctx context.Context, res *types.GlobalResources) *Buffer {

	// a := metrics.NewEWMA1()

	b := &Buffer{
		Ctx:                   ctx,
		maxElementsInBuffer:   1000,
		maxElementsInProgress: int32(maxInProgress),
		timeout:               11 * time.Second,
		EvenOutCh:             make(chan notification.Event, 3),
		BackupStatusCh:        make(chan types.BackupStatus),
		r:                     res,
		errorsRate:            ratecounter.NewRateCounter(60 * time.Second),
	}
	go b.processEvents()
	return b
}

func (b *Buffer) processEvents() {

	checkErrorRate := time.Tick(5 * time.Second)
	sendBufferTicker := time.Tick(b.timeout)
	quit := make(chan bool)
	newTimeout := b.timeout

	for {
		select {
		case <-b.Ctx.Done():
			fmt.Printf("buffer: DONE, WTF?!!!\n")
			return
		case e := <-b.r.FileWatcher.EventCh:
			// TODO: add something like:
			// Status: "scanning",
			// fmt.Printf("buffer: add %s to the buffer\n", e.AbsolutePath)
			b.addEvent(e.AbsolutePath, e)
		case c := <-b.r.BackupCompleteCh:
			fmt.Printf("buffer: BackupCompleteCh\n")
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
				fmt.Printf("!!!!!!!!!!!!! buffer: error uploading %s\n", c.FilePath)
				b.errorsRate.Incr(1)
			}
		case <-checkErrorRate:
			// fmt.Printf("buffer: checkErrorRate\n")
			if b.errorsRate.Rate() == 0 && newTimeout != b.timeout {
				fmt.Println("buffer: BACK TO NORMAL: error rate is 0 !!!!!!!!!!!!!!!")
				newTimeout = b.timeout
				sendBufferTicker = time.Tick(b.timeout)
			} else if b.errorsRate.Rate() > 0 && inProgress > 0 {
				fmt.Printf("!!!!!!!!!!! buffer: error rate is %d/min with %d/%d events in the queue inProgress=%d\n", b.errorsRate.Rate(), len(b.EvenOutCh), cap(b.EvenOutCh), inProgress)
				for i := 0; i < len(b.EvenOutCh); i++ {
					<-b.EvenOutCh
				}
				// fmt.Printf("clean of the channel done: %d/%d inProgress=%d\n", len(b.EvenOutCh), cap(b.EvenOutCh), inProgress)
				quit <- true
				newTimeout = time.Duration(newTimeout * 2)
				sendBufferTicker = time.Tick(newTimeout)
			}
			// fallthrough
		case tick := <-sendBufferTicker:
			fmt.Printf("buffer: send file events to backup afer %s\n", tick)
			if len(events) != 0 && inProgress == 0 {
				go b.send(quit)
			}
			// default:
			// 	// fmt.Println("buffer: ...")
		}
	}
}

// TODO: we can have multiple BackupDone events for the same file (different backup provider!)
// so the `done` counter needs to be fixed somehow!
func (b *Buffer) send(quit chan bool) {
	for _, e := range events {
		select {
		case <-quit:
			fmt.Println("!!!!!!!!!!!!!!!!!! buffer: stop sending !!!!!!!!!!!!!!!!!!!!!!!!")
			atomic.StoreInt32(&inProgress, 0)
			return
		default:
			b.EvenOutCh <- e
			fmt.Printf(">>> send to backup: %s\n", e.AbsolutePath)
			atomic.AddInt32(&inProgress, 1)
			b.BackupStatusCh <- types.BackupStatus{
				FilesDone:       int(done),
				FilesInProgress: int(inProgress),
				TotalFiles:      len(events),
				Status:          "uploading",
			}
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
