package dto

type SendMessageRequest struct {
	TicketID uint64 `json:"ticket_id"`
	Content  string `json:"content"`
}

type MessageResponse struct {
	ID        uint64 `json:"id"`
	TicketID  uint64 `json:"ticket_id"`
	SenderID  uint64 `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
