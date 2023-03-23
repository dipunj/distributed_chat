package state

import (
	"time"
)

type Message struct {
	Id                int
	Group_id          string
	Sender_id         string
	Sender_name       string
	Content_type      string
	Content           string
	Reply_to_id       int
	Client_created_at time.Time
	Server_created_at time.Time
}
