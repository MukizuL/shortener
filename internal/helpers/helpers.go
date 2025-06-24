package helpers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().Unix()))

// RandomString generates string, with variable length, of random characters.
func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// WriteJSON writes status and any object as JSON. Reports no error if Encoder fails.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteCookie prepares a cookie then sets it in ResponseWriter.
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

func SplitIntoBatches[T any](items []T, batchSize int) [][]T {
	batches := make([][]T, 0, (len(items)+batchSize-1)/batchSize)

	for batchSize < len(items) {
		batches = append(batches, items[0:batchSize:batchSize])
		items = items[batchSize:]
	}
	if len(items) > 0 {
		batches = append(batches, items)
	}

	return batches
}

func BuildValuePlaceholders(numCols, numRows int) string {
	placeholders := make([]string, 0, numRows)

	for i := 0; i < numRows; i++ {
		single := make([]string, numCols)
		for j := 0; j < numCols; j++ {
			single[j] = fmt.Sprintf("$%d", i*numCols+j+1)
		}
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(single, ", ")))
	}

	return strings.Join(placeholders, ", ")
}
