package models

// Urls data type to store urls.
type Urls struct {
	UserID      string `json:"user_id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
