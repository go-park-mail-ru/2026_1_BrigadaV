package models

import "time"

type Message struct {
	ID        uint64    `json:"id"`
	TicketID  uint64    `json:"ticket_id"`
	SenderID  uint64    `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
