package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	mockstorage "github.com/MukizuL/shortener/internal/storage/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
					Return("http://localhost:8080/qxDvSD", errs.ErrDuplicate)
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  409,
				shortURL:    "http://localhost:8080/qxDvSD\n",
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
					Return("", errs.ErrURLNotFound)
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

func TestApplication_GetURLs(t *testing.T) {
	type want struct {
		statusCode int
		fullURL    []dto.URLPair
	}

	tests := []struct {
		name      string
		mockSetup func(m *mockstorage.MockRepo)
		user      string
		want      want
	}{
		{
			name: "Correct UserID with links",
			mockSetup: func(m *mockstorage.MockRepo) {
				m.EXPECT().GetUserURLs(gomock.Any(), "user1").Return([]dto.URLPair{
					{ShortURL: "https://link1.com", OriginalURL: "https://localhost:8080/1"},
					{ShortURL: "https://link2.com", OriginalURL: "https://localhost:8080/2"},
					{ShortURL: "https://link3.com", OriginalURL: "https://localhost:8080/3"},
					{ShortURL: "https://link4.com", OriginalURL: "https://localhost:8080/4"},
					{ShortURL: "https://link5.com", OriginalURL: "https://localhost:8080/5"},
				}, nil)
			},
			user: "user1",
			want: want{
				statusCode: 200,
				fullURL: []dto.URLPair{
					{ShortURL: "https://link1.com", OriginalURL: "https://localhost:8080/1"},
					{ShortURL: "https://link2.com", OriginalURL: "https://localhost:8080/2"},
					{ShortURL: "https://link3.com", OriginalURL: "https://localhost:8080/3"},
					{ShortURL: "https://link4.com", OriginalURL: "https://localhost:8080/4"},
					{ShortURL: "https://link5.com", OriginalURL: "https://localhost:8080/5"},
				},
			},
		},
		{
			name: "Correct UserID with no links",
			mockSetup: func(m *mockstorage.MockRepo) {
				m.EXPECT().GetUserURLs(gomock.Any(), "user1").Return([]dto.URLPair{}, nil)
			},
			user: "user1",
			want: want{
				statusCode: 204,
				fullURL:    nil,
			},
		},
		{
			name: "Error in storage",
			mockSetup: func(m *mockstorage.MockRepo) {
				m.EXPECT().GetUserURLs(gomock.Any(), "user1").Return([]dto.URLPair{}, errs.ErrInternalServerError)
			},
			user: "user1",
			want: want{
				statusCode: 500,
				fullURL:    []dto.URLPair{},
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

			r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

			r = r.Clone(context.WithValue(r.Context(), contextI.UserIDContextKey, tt.user))

			w := httptest.NewRecorder()
			app.GetURLs(w, r)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.statusCode == 200 {
				var urls []dto.URLPair
				err := json.NewDecoder(result.Body).Decode(&urls)
				assert.NoError(t, err)

				err = result.Body.Close()
				require.NoError(t, err)

				assert.Equal(t, tt.want.fullURL, urls)
			}
		})
	}
}

func TestApplication_DeleteURLs(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name      string
		mockSetup func(m *mockstorage.MockRepo)
		user      string
		links     []string
		want      want
	}{
		{
			name: "Correct UserID with links",
			mockSetup: func(m *mockstorage.MockRepo) {
				m.EXPECT().DeleteURLs(gomock.Any(), "user1", []string{
					"https://link1.com",
					"https://link2.com",
					"https://link3.com",
					"https://link4.com",
					"https://link5.com",
				}).Return(nil)
			},
			user: "user1",
			links: []string{
				"https://link1.com",
				"https://link2.com",
				"https://link3.com",
				"https://link4.com",
				"https://link5.com",
			},
			want: want{
				statusCode: 202,
			},
		},
		{
			name: "Error in storage",
			mockSetup: func(m *mockstorage.MockRepo) {
				m.EXPECT().DeleteURLs(gomock.Any(), "user1", []string{
					"https://link1.com",
					"https://link2.com",
					"https://link3.com",
					"https://link4.com",
					"https://link5.com",
				}).Return(errs.ErrInternalServerError)
			},
			user: "user1",
			links: []string{
				"https://link1.com",
				"https://link2.com",
				"https://link3.com",
				"https://link4.com",
				"https://link5.com",
			},
			want: want{
				statusCode: 500,
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

			buf := &bytes.Buffer{}
			err := json.NewEncoder(buf).Encode(tt.links)
			assert.NoError(t, err)

			r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", buf)

			r = r.Clone(context.WithValue(r.Context(), contextI.UserIDContextKey, tt.user))

			w := httptest.NewRecorder()
			app.DeleteURLs(w, r)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			err = result.Body.Close()
			assert.NoError(t, err)
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
				response:    dto.Envelope{"short_url": "http://localhost:8080/qxDvSD"},
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
				response:    dto.Envelope{"short_url": "http://localhost:8080/qxDvSD"},
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
				response:    dto.Envelope{"error": "Bad Request"},
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
				response:    dto.Envelope{"error": "Unprocessable Entity"},
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
				response:    dto.Envelope{"error": "Unprocessable Entity"},
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
				var resp dto.Envelope
				err = json.Unmarshal(body, &resp)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response.(dto.Envelope)["short_url"], resp["short_url"])
			default:
				var errResp dto.Envelope
				err = json.Unmarshal(body, &errResp)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response.(dto.Envelope)["error"], errResp["error"])
			}
		})
	}
}

func TestApplication_BatchCreateShortURLJSON(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    interface{}
	}

	tests := []struct {
		name        string
		body        []dto.BatchRequest
		mockStorage func(m *mockstorage.MockRepo)
		want        want
	}{
		{
			name: "Correct working",
			body: []dto.BatchRequest{
				{
					CorrelationID: "1",
					OriginalURL:   "https://www.youtube1.com",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "https://www.youtube2.com",
				},
				{
					CorrelationID: "3",
					OriginalURL:   "https://www.youtube3.com",
				},
			},
			mockStorage: func(m *mockstorage.MockRepo) {
				m.EXPECT().BatchCreateShortURL(gomock.Any(), "user1", "http://localhost:8080/", gomock.Any()).Return([]dto.BatchResponse{
					{
						CorrelationID: "1",
						ShortURL:      "http://localhost:8080/qxDvSD",
					},
					{
						CorrelationID: "2",
						ShortURL:      "http://localhost:8080/qxDvSS",
					},
					{
						CorrelationID: "3",
						ShortURL:      "http://localhost:8080/qxDvSB",
					},
				}, nil)
			},
			want: want{
				contentType: "application/json",
				statusCode:  201,
				response: []dto.BatchResponse{
					{
						CorrelationID: "1",
						ShortURL:      "http://localhost:8080/qxDvSD",
					},
					{
						CorrelationID: "2",
						ShortURL:      "http://localhost:8080/qxDvSS",
					},
					{
						CorrelationID: "3",
						ShortURL:      "http://localhost:8080/qxDvSB",
					},
				},
			},
		},
		{
			name: "Error in storage",
			body: []dto.BatchRequest{
				{
					CorrelationID: "1",
					OriginalURL:   "https://www.youtube1.com",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "https://www.youtube2.com",
				},
				{
					CorrelationID: "3",
					OriginalURL:   "https://www.youtube3.com",
				},
			},
			mockStorage: func(m *mockstorage.MockRepo) {
				m.EXPECT().
					BatchCreateShortURL(gomock.Any(), "user1", "http://localhost:8080/", gomock.Any()).
					Return([]dto.BatchResponse{}, errs.ErrInternalServerError)
			},
			want: want{
				contentType: "application/json",
				statusCode:  500,
				response:    dto.Envelope{"error": "Internal Server Error"},
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

			buf := &bytes.Buffer{}

			err := json.NewEncoder(buf).Encode(tt.body)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", buf)
			r.Header.Set("Content-Type", "application/json")
			r.Host = "localhost:8080"
			r = r.Clone(context.WithValue(r.Context(), contextI.UserIDContextKey, "user1"))

			w := httptest.NewRecorder()
			c.BatchCreateShortURLJSON(w, r)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			fmt.Println(result.StatusCode)

			switch tt.want.statusCode {
			case http.StatusCreated:
				var resp []dto.BatchResponse
				err = json.NewDecoder(result.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response.([]dto.BatchResponse), resp)
			case http.StatusInternalServerError:
				var errResp dto.Envelope
				err = json.NewDecoder(result.Body).Decode(&errResp)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response.(dto.Envelope)["error"], errResp["error"])
			}

			err = result.Body.Close()
			assert.NoError(t, err)
		})
	}
}
