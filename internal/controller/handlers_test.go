package controller

import (
	"bytes"
	"context"
	"encoding/json"
	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	mockstorage "github.com/MukizuL/shortener/internal/storage/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApplication_CreateShortURL(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		shortURL    string
	}

	tests := []struct {
		name        string
		body        string
		mockStorage func(m *mockstorage.MockRepo)
		want        want
	}{
		{
			name: "Correct working",
			body: "https://www.youtube.com",
			mockStorage: func(m *mockstorage.MockRepo) {
				m.EXPECT().
					CreateShortURL(gomock.Any(), gomock.Any(), "http://localhost:8080/", "https://www.youtube.com").
					Return("http://localhost:8080/qxDvSD", nil)
			},
			want: want{
				contentType: "text/plain",
				statusCode:  201,
				shortURL:    "http://localhost:8080/qxDvSD",
			},
		},
		{
			name: "Duplicate URL",
			body: "https://www.youtube.com",
			mockStorage: func(m *mockstorage.MockRepo) {
				m.EXPECT().
					CreateShortURL(gomock.Any(), gomock.Any(), "http://localhost:8080/", "https://www.youtube.com").
					Return("", errs.ErrDuplicate)
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  409,
				shortURL:    "Conflict\n",
			},
		},
		{
			name: "Incorrect URL #1",
			body: "https://",
			mockStorage: func(m *mockstorage.MockRepo) {

			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				shortURL:    "Bad Request\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mockstorage.NewMockRepo(ctrl)
			if tt.mockStorage != nil {
				tt.mockStorage(mockRepo)
			}

			c := &Controller{
				storage: mockRepo,
			}

			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			r.Host = "localhost:8080"
			r = r.Clone(context.WithValue(r.Context(), contextI.UserIDContextKey, "1"))

			w := httptest.NewRecorder()
			c.CreateShortURL(w, r)

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

	tests := []struct {
		name      string
		query     string
		mockSetup func(m *mockstorage.MockRepo)
		want      want
	}{
		{
			name:  "Correct URL",
			query: "qxDvSD",
			mockSetup: func(m *mockstorage.MockRepo) {
				m.EXPECT().
					GetLongURL(gomock.Any(), "qxDvSD").
					Return("https://www.youtube.com", nil)
			},
			want: want{
				statusCode: 307,
				fullURL:    "https://www.youtube.com",
			},
		},
		{
			name:  "Not present URL",
			query: "qxDvSg",
			mockSetup: func(m *mockstorage.MockRepo) {
				m.EXPECT().
					GetLongURL(gomock.Any(), "qxDvSg").
					Return("", errs.ErrNotFound)
			},
			want: want{
				statusCode: 404,
				fullURL:    "",
			},
		},
		{
			name:  "Empty URL",
			query: "",
			mockSetup: func(m *mockstorage.MockRepo) {

			},
			want: want{
				statusCode: 400,
				fullURL:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mockstorage.NewMockRepo(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			app := &Controller{
				storage: mockRepo,
			}

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

	tests := []struct {
		name        string
		body        string
		mockStorage func(m *mockstorage.MockRepo)
		want        want
	}{
		{
			name: "Correct working",
			body: "https://www.youtube.com",
			mockStorage: func(m *mockstorage.MockRepo) {
				m.EXPECT().
					CreateShortURL(gomock.Any(), gomock.Any(), "http://localhost:8080/", "https://www.youtube.com").
					Return("http://localhost:8080/qxDvSD", nil)
			},
			want: want{
				contentType: "application/json",
				statusCode:  201,
				response:    dto.Response{Result: "http://localhost:8080/qxDvSD"},
			},
		},
		{
			name: "Duplicate URL",
			body: "https://www.youtube.com",
			mockStorage: func(m *mockstorage.MockRepo) {
				m.EXPECT().
					CreateShortURL(gomock.Any(), gomock.Any(), "http://localhost:8080/", "https://www.youtube.com").
					Return("http://localhost:8080/qxDvSD", errs.ErrDuplicate)
			},
			want: want{
				contentType: "application/json",
				statusCode:  409,
				response:    dto.Response{Result: "http://localhost:8080/qxDvSD"},
			},
		},
		{
			name: "Incorrect URL #1",
			body: "https://",
			mockStorage: func(m *mockstorage.MockRepo) {

			},
			want: want{
				contentType: "application/json",
				statusCode:  400,
				response:    dto.ErrorResponse{Err: "Bad Request"},
			},
		},
		{
			name: "Incorrect URL #2",
			body: "www.something.ru",
			mockStorage: func(m *mockstorage.MockRepo) {

			},
			want: want{
				contentType: "application/json",
				statusCode:  422,
				response:    dto.ErrorResponse{Err: "Unprocessable Entity"},
			},
		},
		{
			name: "Empty URL",
			body: "",
			mockStorage: func(m *mockstorage.MockRepo) {

			},
			want: want{
				contentType: "application/json",
				statusCode:  422,
				response:    dto.ErrorResponse{Err: "Unprocessable Entity"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mockstorage.NewMockRepo(ctrl)
			if tt.mockStorage != nil {
				tt.mockStorage(mockRepo)
			}

			c := &Controller{
				storage: mockRepo,
			}

			data, err := json.Marshal(&dto.Request{FullURL: tt.body})
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(data))
			r.Header.Set("Content-Type", "application/json")
			r.Host = "localhost:8080"
			r = r.Clone(context.WithValue(r.Context(), contextI.UserIDContextKey, "1"))

			w := httptest.NewRecorder()
			c.CreateShortURLJSON(w, r)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			switch tt.want.statusCode {
			case http.StatusCreated, http.StatusConflict:
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
