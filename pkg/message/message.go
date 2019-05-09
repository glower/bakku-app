package message

import "time"

// Message ...
type Message struct {
	Message string
	Type    string
	Source  string
	Time    time.Time
}

func FormatMessage(t, msg, from string) Message {
	return Message{
		Message: msg,
		Type:    t,
		Time:    time.Now(),
		Source:  from,
	}
}
