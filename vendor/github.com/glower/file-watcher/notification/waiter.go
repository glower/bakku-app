/*
Package notifications fixes the problem of multiple file change notifications for the same file from the OS API.
With FileNotificationWaiter(chanA, chanB, data) you can send data to the chanB if nothing was send to the chanA for 5 seconds

The flow:
- For each file create a channel and store it with RegisterFileNotification()
- Call FileNotificationWaiter() as a go routin with the created channel and other needed data
- On the next file change notification check if the channel for this file exists, if so send true to the channel
- If nothing was send on the channel, FileNotificationWaiter() will send the data to the provided channel after 5 seconds*/
package notification

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

// Waiter ...
type Waiter struct {
	EventCh  chan Event
	ErrorCh  chan Error
	Timeout  time.Duration
	MaxCount int
}

var notificationsMutex sync.Mutex
var notificationsChans = make(map[string]chan bool)

// RegisterFileNotification channel for a given file path, use this channel for with FileNotificationWaiter() function
func (w *Waiter) RegisterFileNotification(path string) {
	waitChan := make(chan bool)
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	notificationsChans[path] = waitChan
}

// UnregisterFileNotification channel for a given file path
func (w *Waiter) UnregisterFileNotification(path string) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	delete(notificationsChans, path)
}

// LookupForFileNotification returns a channel for a given file path
func (w *Waiter) LookupForFileNotification(path string) (chan bool, bool) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	data, ok := notificationsChans[path]
	return data, ok
}

// Wait will send fileData to the chan stored in CallbackData after 5 seconds if no signal is
// received on waitChan.
// TODO: this can be done better with a general type of channel and any data
func (w *Waiter) Wait(fileData *Event) {
	waitChan, exists := w.LookupForFileNotification(fileData.AbsolutePath)
	if !exists {
		w.ErrorCh <- Error{
			Stack:   string(debug.Stack()),
			Message: fmt.Sprintf("NotificationWaiter.Wait(): no notification if registered for the path %s", fileData.AbsolutePath),
		}
		return
	}
	cnt := 0
	for {
		select {
		case <-waitChan:
			cnt++
			if cnt == w.MaxCount {
				w.ErrorCh <- Error{
					Stack:   string(debug.Stack()),
					Message: fmt.Sprintf("exit after %d times of notification for [%s]", w.MaxCount, fileData.AbsolutePath),
				}
				w.UnregisterFileNotification(fileData.AbsolutePath)
				close(waitChan)
				return
			}
		case <-time.After(w.Timeout):
			w.EventCh <- *fileData
			w.UnregisterFileNotification(fileData.AbsolutePath)
			close(waitChan)
			return
		}
	}
}
