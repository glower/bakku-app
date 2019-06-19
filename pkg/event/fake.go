package event

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/file-watcher/notification"
	"github.com/glower/file-watcher/watcher"
	"github.com/google/uuid"
)

var cnt = 0

func Fake(ctx context.Context, _ *config.WatchConfig) *watcher.Watch {
	eventCh := make(chan notification.Event)
	errorCh := make(chan notification.Error)
	go createFakeEvents(ctx, eventCh)
	// return eventCh, errorCh
	w := &watcher.Watch{
		EventCh: eventCh,
		ErrorCh: errorCh,
	}
}

func createFakeEvents(ctx context.Context, e chan notification.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(getRandomTimer() * time.Second):
			max := 10 + rand.Intn(50)
			for i := 0; i < max; i++ {
				e <- getRandomEvent()
			}
		}
	}
}

func getRandomTimer() time.Duration {
	r := 10 + rand.Intn(100)
	return time.Duration(r)
}

func getRandomEvent() notification.Event {
	cnt++
	name := fmt.Sprintf("image_%04d.jpg", cnt)
	size := 100000000 + rand.Intn(200000000)
	return notification.Event{
		MimeType:           "image/jpg",
		AbsolutePath:       fmt.Sprintf("c:/files/images/%s", name),
		Action:             notification.ActionType(notification.FileAdded),
		DirectoryPath:      "c:/files/images/",
		Machine:            "testing",
		FileName:           name,
		RelativePath:       fmt.Sprintf("images/%s", name),
		Size:               int64(size),
		Timestamp:          time.Now(),
		WatchDirectoryName: "images",
		UUID:               uuid.New(),
	}
}
