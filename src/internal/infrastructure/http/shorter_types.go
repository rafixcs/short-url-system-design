package http

type CreateShortUrlRequest struct {
	UserID  string `json:"user_id"`
	LongUrl string `json:"long_url"`
}

type CreateShorURLResponse struct {
	UserID   string `json:"user_id"`
	LongURL  string `json:"long_url"`
	ShortURL string `json:"shor_url"`
}
