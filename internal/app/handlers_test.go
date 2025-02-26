package app

import (
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type storageMock struct {
	createdURL map[string]struct{}
}

func (m *storageMock) Create(fullURL string) (string, error) {
	if _, exist := m.createdURL[string(fullURL)]; exist {
		return "", errs.ErrDuplicate
	}
	m.createdURL[string(fullURL)] = struct{}{}

	return "http://localhost:8080/qxDvSD", nil
}

func (m *storageMock) Get(ID string) (string, error) {
	if ID == "qxDvSD" {
		return "https://www.youtube.com", nil
	}

	return "", errs.ErrNotFound
}

func TestApplication_CreateShortURL(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		shortURL    string
	}

	app := &application{
		storage: &storageMock{make(map[string]struct{})},
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "Correct working",
			body: "https://www.youtube.com",
			want: want{
				contentType: "text/plain",
				statusCode:  201,
				shortURL:    "http://localhost:8080/qxDvSD",
			},
		},
		{
			name: "Duplicate URL",
			body: "https://www.youtube.com",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  409,
				shortURL:    "Conflict\n",
			},
		},
		{
			name: "Incorrect URL #1",
			body: "https://",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				shortURL:    "Bad Request\n",
			},
		},
		{
			name: "Incorrect URL #2",
			body: "www.something.ru",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  422,
				shortURL:    "Unprocessable Entity\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			app.CreateShortURL(w, r)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			urlResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.shortURL, string(urlResult))
		})
	}
}
