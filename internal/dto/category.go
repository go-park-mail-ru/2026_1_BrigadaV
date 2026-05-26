//go:generate easyjson -all

package dto

//easyjson:json
type CategoryRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	ApplicableTypes []string `json:"applicable_types"`
}

//easyjson:json
type CategoryResponse struct {
	ID              uint64   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	ApplicableTypes []string `json:"applicable_types"`
}
