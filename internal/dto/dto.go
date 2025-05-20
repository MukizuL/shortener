package dto

type Request struct {
	FullURL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Err string `json:"error"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URLPair struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
