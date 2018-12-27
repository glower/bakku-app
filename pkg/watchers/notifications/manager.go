/*
Package notifications fixes the problem of multiple file change notifications for the same file from the OS API.
With FileNotificationWaiter(chanA, chanB, data) you can send data to the chanB if nothing was send to the chanA for 5 seconds

The flow:
- For each file create a channel and store it with RegisterFileNotification()
- Call FileNotificationWaiter() as a go routin with the created channel and other needed data
- On the next file change notification check if the channel for this file exists, if so send true to the channel
- If nothing was send on the channel, FileNotificationWaiter() will send the data to the provided channel after 5 seconds*/
package notifications

import (
	"log"
	"sync"
	"time"

	"github.com/glower/bakku-app/pkg/types"
)

var notificationsMutex sync.Mutex
var notificationsChans = make(map[string]chan bool)

// RegisterFileNotification channel for a given file path, use this channel for with FileNotificationWaiter() function
func RegisterFileNotification(waitChan chan bool, path string) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	notificationsChans[path] = waitChan
}

// UnregisterFileNotification channel for a given file path
func UnregisterFileNotification(path string) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	delete(notificationsChans, path)
}

// LookupForFileNotification returns a channel for a given file path
func LookupForFileNotification(path string) (chan bool, bool) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	data, ok := notificationsChans[path]
	return data, ok
}

// FileNotificationWaiter will send fileData to the chan stored in CallbackData after 5 seconds if no signal is
// received on waitChan.
func FileNotificationWaiter(waitChan chan bool, callbackChan chan types.FileChangeNotification, fileData *types.FileChangeNotification) {
	cnt := 0
	for {
		select {
		case <-waitChan:
			cnt++
			if cnt > 100 {
				log.Printf("[ERROR] FileNotificationWaiter(): exit after 100 times of notification for [%s]", fileData.AbsolutePath)
				UnregisterFileNotification(fileData.AbsolutePath)
				close(waitChan)
				return
			}
		case <-time.After(time.Duration(5 * time.Second)):
			callbackChan <- *fileData
			UnregisterFileNotification(fileData.AbsolutePath)
			close(waitChan)
			return
		}
	}
}
