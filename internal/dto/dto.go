package dto

type Request struct {
	Url string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Err string `json:"error"`
}
