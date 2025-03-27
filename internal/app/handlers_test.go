package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/go-chi/chi/v5"
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

func (m *storageMock) CreateShortURL(fullURL string) (string, error) {
	if _, exist := m.createdURL[string(fullURL)]; exist {
		return "", errs.ErrDuplicate
	}
	m.createdURL[string(fullURL)] = struct{}{}

	return "http://localhost:8080/qxDvSD", nil
}

func (m *storageMock) GetLongURL(ID string) (string, error) {
	if ID == "qxDvSD" {
		return "https://www.youtube.com", nil
	}

	return "", errs.ErrNotFound
}

func (m *storageMock) OffloadStorage(filepath string) error {
	return nil
}

func TestApplication_CreateShortURL(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		shortURL    string
	}

	app := &Application{
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
		{
			name: "Empty URL",
			body: "",
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

func TestApplication_GetFullURL(t *testing.T) {
	type want struct {
		statusCode int
		fullURL    string
	}

	app := &Application{
		storage: &storageMock{make(map[string]struct{})},
	}

	tests := []struct {
		name  string
		query string
		want  want
	}{
		{
			name:  "Correct URL",
			query: "qxDvSD",
			want: want{
				statusCode: 307,
				fullURL:    "https://www.youtube.com",
			},
		},
		{
			name:  "Not present URL",
			query: "qxDvSg",
			want: want{
				statusCode: 404,
				fullURL:    "",
			},
		},
		{
			name:  "Empty URL",
			query: "",
			want: want{
				statusCode: 400,
				fullURL:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.query)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			app.GetFullURL(w, r)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.fullURL, result.Header.Get("Location"))

			_, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestApplication_CreateShortURLJSON(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    interface{}
	}

	app := &Application{
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
				contentType: "application/json",
				statusCode:  201,
				response:    dto.Response{Result: "http://localhost:8080/qxDvSD"},
			},
		},
		{
			name: "Duplicate URL",
			body: "https://www.youtube.com",
			want: want{
				contentType: "application/json",
				statusCode:  409,
				response:    dto.ErrorResponse{Err: "Conflict"},
			},
		},
		{
			name: "Incorrect URL #1",
			body: "https://",
			want: want{
				contentType: "application/json",
				statusCode:  400,
				response:    dto.ErrorResponse{Err: "Bad Request"},
			},
		},
		{
			name: "Incorrect URL #2",
			body: "www.something.ru",
			want: want{
				contentType: "application/json",
				statusCode:  422,
				response:    dto.ErrorResponse{Err: "Unprocessable Entity"},
			},
		},
		{
			name: "Empty URL",
			body: "",
			want: want{
				contentType: "application/json",
				statusCode:  422,
				response:    dto.ErrorResponse{Err: "Unprocessable Entity"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(&dto.Request{Url: tt.body})
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(data))
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			app.CreateShortURLJSON(w, r)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			switch tt.want.statusCode {
			case http.StatusCreated:
				var resp dto.Response
				err = json.Unmarshal(body, &resp)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response.(dto.Response).Result, resp.Result)
			default:
				var errResp dto.ErrorResponse
				err = json.Unmarshal(body, &errResp)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response.(dto.ErrorResponse).Err, errResp.Err)
			}
		})
	}
}
