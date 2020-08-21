package api

import (
	"github.com/jackc/pgx"
)

type Config struct {
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPassword string

	natsURL string
}

type initAPI struct {
	Db     *pgx.ConnPool
	config Config
}

type initWorker struct {
	api *initAPI
}

type QueueItem struct {
	Schedule       string `json:"scedule"`
	Description    string `json:"description"`
	Activation     int32  `json:"Activation"`
	TriggerBy      string `json:"trigger-by"`
	ActivationTime int64  `json:"activation_time"`
}

type MessageItem struct {
	Message string `json:"message"`
	Email   string `json:"email"`
	Number  string `json:"number"`
}

type QueueRequest struct {
	Event string       `json:"events"`
	Item  *MessageItem `json:"item"`
}

type QueueResponse struct {
	Id        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
}

type MessageRequest struct {
	Events string `json:"events"`
	Times  int64  `json:"times"`
	Number string `json:"number"`
}
