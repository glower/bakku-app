package backup

import (
	"context"

	"github.com/glower/bakku-app/pkg/types"
)

// StorageManager ...
type StorageManager struct {
	ctx context.Context

	FileChangeNotificationChannel chan *types.FileChangeNotification
	FileBackupProgressChannel     chan *Progress
	FileBackupCompleteChannel     chan *types.FileBackupComplete
	// SSEServer *sse.Server
}
