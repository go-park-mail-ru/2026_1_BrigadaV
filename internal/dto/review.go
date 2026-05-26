//go:generate easyjson -all review.go

package dto

//easyjson:json
type CreateReviewRequest struct {
	PlaceID   uint64  `json:"place_id"`
	Title     *string `json:"title,omitempty"`
	Rating    int16   `json:"rating"`
	Content   string  `json:"content"`
	VisitDate *string `json:"visit_date,omitempty"`
}
