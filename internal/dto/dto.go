package dto

// Request represents a URL shortening request.
type Request struct {
	FullURL string `json:"url"`
}

// Response represents a successful URL shortening response.
type Response struct {
	Result string `json:"result"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Err string `json:"error"`
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
