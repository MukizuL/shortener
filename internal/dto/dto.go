package dto

type ResponseWrapper map[string]interface{}

// Request represents a URL shortening request.
type Request struct {
	FullURL string `json:"url"`
}

// BatchRequest represents a batch URL shortening request item.
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse represents a batch URL shortening response item.
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// URLPair represents a pair of original and shortened URLs
type URLPair struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Stats provides info on how many shortURLs and users in the system.
type Stats struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}
