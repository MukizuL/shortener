package helpers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().Unix()))

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteCookie(w http.ResponseWriter, token string) {
	tokenCookie := &http.Cookie{
		Name:     "Access-token",
		Value:    token,
		Path:     "/",
		MaxAge:   876000,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}

	http.SetCookie(w, tokenCookie)
}

func FillParameters(userID string, urls []string) ([]interface{}, string) {
	query := ""

	params := make([]interface{}, 0, len(urls)+1)
	params = append(params, userID)

	for i, url := range urls {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", i+2)
		params = append(params, url)
	}

	return params, query
}
