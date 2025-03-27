package models

type DataObject struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
