package message

import "time"

type Message struct {
	Message string    `json:"message"`
	Type    string    `json:"type"`
	Source  string    `json:"source"`
	Time    time.Time `json:"time"`
}

func FormatMessage(t, msg, from string) Message {
	return Message{
		Message: msg,
		Type:    t,
		Time:    time.Now(),
		Source:  from,
	}
}
