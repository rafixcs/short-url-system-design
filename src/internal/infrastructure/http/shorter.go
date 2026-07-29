package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rafixcs/shorter-url-design-system/src/internal/domain"
	"github.com/rafixcs/shorter-url-design-system/src/pkg/utils"
)

type ShorterHandler struct {
	service domain.ShorterUrlService
}

func NewShorterHandler(shorterService domain.ShorterUrlService) *ShorterHandler {
	return &ShorterHandler{service: shorterService}
}

func (h *ShorterHandler) CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	var reqBody CreateShortUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("[CreateShortURL]: Failed to decode request body - err: %v", err)
		http.Error(w, "bad body request format", http.StatusBadRequest)
		return
	}

	model, err := h.service.CreateShortUrl(r.Context(), reqBody.UserID, reqBody.LongUrl)
	if err != nil {
		log.Printf("[CreateShortURL]: Failed to create short URL: %v", err)
		http.Error(w, "failed to create short URL", http.StatusInternalServerError)
		return
	}

	resp := CreateShorURLResponse{
		UserID:   model.UserID,
		LongURL:  model.LongURL,
		ShortURL: model.ShortURL,
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

func (h *ShorterHandler) RedirectToSourceUrl(w http.ResponseWriter, r *http.Request) {
	linkID := chi.URLParam(r, "shortURL")

	longURL, err := h.service.GetLongUrl(r.Context(), linkID)
	if err != nil {
		log.Printf("[RedirectToSourceURL]: Failed to get long url: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}
