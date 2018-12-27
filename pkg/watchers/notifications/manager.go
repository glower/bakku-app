package notifications

import (
	"log"
	"sync"
	"time"

	"github.com/glower/bakku-app/pkg/types"
)

var notificationsMutex sync.Mutex
var notificationsChans = make(map[string]chan bool)

func RegisterFileNotification(waitChan chan bool, path string) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	notificationsChans[path] = waitChan
}

func UnregisterFileNotification(path string) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	delete(notificationsChans, path)
}

func lookupForFileNotification(path string) (chan bool, bool) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	data, ok := notificationsChans[path]
	return data, ok
}

// FileNotificationWaiter will send fileData to the chan stored in CallbackData after 5 seconds if no signal is
// received on waitChan.
func FileNotificationWaiter(waitChan chan bool, callbackChan chan types.FileChangeNotification, fileData *types.FileChangeNotification) {
	for {
		select {
		case <-waitChan:
		case <-time.After(time.Duration(5 * time.Second)):
			callbackChan <- *fileData
			UnregisterFileNotification(fileData.AbsolutePath)
			close(waitChan)
			return
		case <-time.After(time.Duration(60 * time.Second)):
			log.Printf("[ERROR] FileNotificationWaiter(): exit after 60 sec of waiting for [%s]", fileData.AbsolutePath)
			UnregisterFileNotification(fileData.AbsolutePath)
			close(waitChan)
			return
		}
	}
}
