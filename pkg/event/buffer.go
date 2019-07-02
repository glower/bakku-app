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
	eventsM       sync.RWMutex
	events        = make(map[string]notification.Event)
	inProgress    int32 // int64?
	done          int32
	maxInProgress = 5

	// 0, 1, 1, 2, 3, 5, 8, 13, 21, 34
	throttlingRates = []time.Duration{
		1 * 60 * time.Second,  // 1 min
		2 * 60 * time.Second,  // 2 mins
		3 * 60 * time.Second,  // 3 mins
		5 * 60 * time.Second,  // 5 mins
		8 * 60 * time.Second,  // 8 mins
		13 * 60 * time.Second, // 13 mins
		21 * 60 * time.Second, // 21 mins
		34 * 60 * time.Second, // 34 mins
	}
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

	errorsRate  *ratecounter.RateCounter
	successRate *ratecounter.RateCounter
}

// NewBuffer ...
func NewBuffer(ctx context.Context, res *types.GlobalResources) *Buffer {
	b := &Buffer{
		Ctx:                   ctx,
		maxElementsInBuffer:   1000,
		maxElementsInProgress: int32(maxInProgress),
		timeout:               11 * time.Second,
		EvenOutCh:             make(chan notification.Event, 3),
		BackupStatusCh:        make(chan types.BackupStatus),
		r:                     res,
		errorsRate:            ratecounter.NewRateCounter(60 * time.Second),
		successRate:           ratecounter.NewRateCounter(60 * time.Second),
	}
	go b.processEvents()
	return b
}

func (b *Buffer) processEvents() {
	throttlingOffset := 0
	checkErrorRate := time.Tick(5 * time.Second)
	sendBufferTicker := time.Tick(b.timeout)
	quit := make(chan bool)

	for {
		select {
		case <-b.Ctx.Done():
			return
		case e := <-b.r.FileWatcher.EventCh:
			b.setStatus("scanning")
			b.addEvent(e.AbsolutePath, e)
		case c := <-b.r.BackupCompleteCh:
			if c.Success {
				atomic.AddInt32(&inProgress, -1)
				atomic.AddInt32(&done, 1)
				b.successRate.Incr(1)
				removeEvent(c.FilePath)
				b.setStatus("uploading")
			} else {
				fmt.Printf("[ERROR] buffer: error uploading %s\n", c.FilePath)
				b.errorsRate.Incr(1)
			}
		case <-checkErrorRate:
			if b.errorsRate.Rate() == 0 && b.successRate.Rate() > 0 && throttlingOffset > 0 {
				throttlingOffset = 0
				newTimeout := throttlingRates[throttlingOffset]
				b.successRate = ratecounter.NewRateCounter(newTimeout)
				b.errorsRate = ratecounter.NewRateCounter(newTimeout)
				sendBufferTicker = time.Tick(b.timeout)

			} else if b.errorsRate.Rate() > 0 && inProgress > 0 {
				for i := 0; i < len(b.EvenOutCh); i++ {
					<-b.EvenOutCh
				}
				quit <- true

				if throttlingOffset < len(throttlingRates)-1 {
					throttlingOffset++
				}
				newTimeout := throttlingRates[throttlingOffset]
				sendBufferTicker = time.Tick(newTimeout)
				b.setStatus(fmt.Sprintf("paused for %s", newTimeout))
				b.successRate = ratecounter.NewRateCounter(2 * newTimeout)
				b.errorsRate = ratecounter.NewRateCounter(2 * newTimeout)
			}
		case <-sendBufferTicker:
			if len(events) != 0 && inProgress == 0 {
				go b.send(quit)
			}
		}
	}
}

// TODO: we can have multiple BackupDone events for the same file (different backup provider!)
// so the `done` counter needs to be fixed somehow!
func (b *Buffer) send(quit chan bool) {
	for _, e := range events {
		select {
		case <-quit:
			fmt.Println("[INFO] buffer: stopped sending")
			atomic.StoreInt32(&inProgress, 0)
			return
		default:
			b.EvenOutCh <- e
			fmt.Printf("[INFO] buffer: send to backup: %s\n", e.AbsolutePath)
			atomic.AddInt32(&inProgress, 1)
		}
	}
	b.setStatus("waiting")
}

func (b *Buffer) setStatus(status string) {
	b.BackupStatusCh <- types.BackupStatus{
		FilesDone:       int(done),
		FilesInProgress: int(inProgress),
		TotalFiles:      len(events),
		Status:          status,
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
