package mw

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/errs"
	mockjwt "github.com/MukizuL/shortener/internal/jwt/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestApplication_Authorization(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*mockjwt.MockJWTServiceInterface)
		cookiePresent bool
		cookieValue   string
		wantStatus    int
		wantSetCookie bool
	}{
		{
			name: "success with existing valid token",
			mockSetup: func(m *mockjwt.MockJWTServiceInterface) {
				m.EXPECT().CreateOrValidateToken("valid-token").Return("new-token", "user1", nil)
			},
			cookiePresent: true,
			cookieValue:   "valid-token",
			wantStatus:    http.StatusOK,
			wantSetCookie: true,
		},
		{
			name: "success with new token creation",
			mockSetup: func(m *mockjwt.MockJWTServiceInterface) {
				m.EXPECT().CreateOrValidateToken("").Return("new-token", "user2", nil)
			},
			cookiePresent: false,
			wantStatus:    http.StatusOK,
			wantSetCookie: true,
		},
		{
			name: "error validating token",
			mockSetup: func(m *mockjwt.MockJWTServiceInterface) {
				m.EXPECT().CreateOrValidateToken("invalid-token").Return("", "", errs.ErrNotAuthorized)
			},
			cookiePresent: true,
			cookieValue:   "invalid-token",
			wantStatus:    http.StatusUnauthorized,
			wantSetCookie: false,
		},
		{
			name: "error creating token",
			mockSetup: func(m *mockjwt.MockJWTServiceInterface) {
				m.EXPECT().CreateOrValidateToken("").Return("", "", errors.New("error"))
			},
			cookiePresent: false,
			wantStatus:    http.StatusInternalServerError,
			wantSetCookie: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockJWT := mockjwt.NewMockJWTServiceInterface(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockJWT)
			}

			logger, err := zap.NewDevelopment()
			assert.NoError(t, err)

			s := &MiddlewareService{
				jwtService: mockJWT,
				logger:     logger,
			}

			r := httptest.NewRequest("GET", "/", nil)
			if tt.cookiePresent {
				r.AddCookie(&http.Cookie{
					Name:  "Access-token",
					Value: tt.cookieValue,
				})
			}

			w := httptest.NewRecorder()

			// Create a simple handler to verify context was set
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID, ok := r.Context().Value(contextI.UserIDContextKey).(string)
				if !ok && tt.wantStatus == http.StatusOK {
					t.Error("user ID not set in context")
				}
				if tt.wantStatus == http.StatusOK {
					if strings.Contains(tt.name, "existing") && userID != "user1" {
						t.Errorf("got user ID %s, want user1", userID)
					}
					if strings.Contains(tt.name, "new token") && userID != "user2" {
						t.Errorf("got user ID %s, want user2", userID)
					}
				}
			})

			handler := s.Authorization(nextHandler)
			handler.ServeHTTP(w, r)

			if status := w.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantStatus)
			}

			result := w.Result()

			if tt.wantSetCookie {
				cookies := result.Cookies()
				if len(cookies) == 0 {
					t.Error("expected cookie to be set, but none found")
				}
			}

			err = result.Body.Close()
			assert.NoError(t, err)
		})
	}
}
