package dto

import "time"

type PhotoResponse struct {
	ID        uint64    `json:"id"`
	FilePath  string    `json:"file_path"`
	CreatedAt time.Time `json:"created_at"`
}
