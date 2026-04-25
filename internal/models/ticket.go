package models

import "time"

type TicketCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TicketStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Ticket struct {
	ID         uint64    `json:"id"`
	UserID     uint64    `json:"user_id"`
	CategoryID int       `json:"category_id"`
	StatusID   int       `json:"status_id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
