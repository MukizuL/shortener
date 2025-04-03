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
