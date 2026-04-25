package dto

type CreateTicketRequest struct {
	CategoryID int    `json:"category_id"`
	Title      string `json:"title"`
	Body       string `json:"body"`
}

type TicketResponse struct {
	ID         uint64 `json:"id"`
	UserID     uint64 `json:"user_id,omitempty"`
	CategoryID int    `json:"category_id"`
	StatusID   int    `json:"status_id"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type UpdateStatusRequest struct {
	TicketID    uint64 `json:"ticket_id"`
	NewStatusID int    `json:"new_status_id"`
}

type StatisticsResponse struct {
	Total      int         `json:"total"`
	ByStatus   map[int]int `json:"by_status"`
	ByCategory map[int]int `json:"by_category"`
}
