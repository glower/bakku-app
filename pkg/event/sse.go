package event

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/file-watcher/notification"
	"github.com/gorilla/mux"
	"github.com/r3labs/sse"
)

var streams = []string{"files", "messages", "ping", "status"}

// SSE ...
type SSE struct {
	ctx context.Context

	streams []string
	server  *sse.Server
	router  *mux.Router
}

// NewSSE ...
func NewSSE(ctx context.Context, router *mux.Router, backupProgressCh chan types.BackupProgress, res *types.GlobalResources, eventBuffer *Buffer) *SSE {
	events := sse.New()
	for _, name := range streams {
		events.CreateStream(name)
	}

	s := &SSE{
		ctx:     ctx,
		streams: streams,
		server:  events,
		router:  router,
	}

	s.addEventRoute()

	// backupProgressCh chan types.BackupProgress
	// errorCh chan notification.Error
	// messageCh chan message.Message
	// fileBackupCompleteBCh broadcast.Broadcast
	go s.processBackupStatus(eventBuffer.BackupStatusCh)
	go s.processErrors(res.FileWatcher.ErrorCh, res.MessageCh)
	go s.processProgressCallback(backupProgressCh)
	// go s.ping()

	return s
}

// StopSSE stops SSE service
func (s *SSE) StopSSE() {
	for _, name := range s.streams {
		s.server.RemoveStream(name)
	}
	s.server.Close()
}

func (s *SSE) addEventRoute() {
	s.router.Methods("GET").Path("/events").HandlerFunc(s.server.HTTPHandler)
}

func (s *SSE) processBackupStatus(status chan types.BackupStatus) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case bs := <-status:
			if bs.Status == "waiting" {
				fmt.Printf("[SSE] processBackupStatus(): [%s...]\n", bs.Status)
			} else {
				fmt.Printf("[SSE] processBackupStatus(): [%s] [%d to sync] (%d done)\n", bs.Status, bs.TotalFiles, bs.FilesDone)
			}
			stautsJSON, err := json.Marshal(bs)
			if err != nil {
				stautsJSON = []byte(fmt.Sprintf(`{"message": "%s", "type": "error"}`, err.Error()))
			}
			s.server.Publish("status", &sse.Event{
				Data: stautsJSON,
			})
		}
	}
}

func (s *SSE) processProgressCallback(backupProgressCh chan types.BackupProgress) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case progress := <-backupProgressCh:
			// TODO: don't report progress on backup of snapshot file
			// get this name from config/package
			if strings.Contains(progress.FileName, ".snapshot") {
				continue
			}
			// fmt.Printf("[SSE] ProcessProgressCallback(): [%s] [%s]\t%.2f%%\n", progress.StorageName, progress.FileName, progress.Percent)
			progressJSON, err := json.Marshal(progress)
			if err != nil {
				progressJSON = []byte(fmt.Sprintf(`{"message": "%s", "type": "error"}`, err.Error()))
			}
			// file fotification for the frontend client over the SSE
			s.server.Publish("files", &sse.Event{
				Data: progressJSON,
			})
		}
	}
}

func (s *SSE) processErrors(errorCh chan notification.Error, messageCh chan message.Message) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case err := <-errorCh:
			if err.Level == "ERROR" || err.Level == "CRITICAL" {
				s.publishEventMessage(message.Message{
					Type:    err.Level,
					Message: err.Message,
					Source:  "watcher",
				}, "messages")
			}
		case msg := <-messageCh:
			s.publishEventMessage(msg, "messages")
		}
	}
}

func (s *SSE) ping() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(60 * time.Second):
			s.publishEventMessage(message.Message{
				Message: "ping",
				Type:    "INFO",
				Source:  "main",
				Time:    time.Now(),
			}, "ping")
		}
	}
}

func (s *SSE) publishEventMessage(msg message.Message, channel string) {
	messageJSON, err := json.Marshal(msg)
	if err != nil {
		messageJSON = []byte(fmt.Sprintf(`{"message": "%s", "type": "error"}`, err.Error()))
	}
	// fmt.Printf("[SSE] [%s] %s: %s\n", msg.Type, msg.Source, msg.Message)
	s.server.Publish(channel, &sse.Event{
		Data: messageJSON,
	})
}
