package mw

import (
	"errors"
	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/errs"
	mockjwt "github.com/MukizuL/shortener/internal/jwt/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
				m.EXPECT().ValidateToken("valid-token").Return("new-token", "user1", nil)
			},
			cookiePresent: true,
			cookieValue:   "valid-token",
			wantStatus:    http.StatusOK,
			wantSetCookie: true,
		},
		{
			name: "success with new token creation",
			mockSetup: func(m *mockjwt.MockJWTServiceInterface) {
				m.EXPECT().CreateToken().Return("new-token", "user2", nil)
			},
			cookiePresent: false,
			wantStatus:    http.StatusOK,
			wantSetCookie: true,
		},
		{
			name: "error validating token",
			mockSetup: func(m *mockjwt.MockJWTServiceInterface) {
				m.EXPECT().ValidateToken("invalid-token").Return("", "", errs.ErrNotAuthorized)
			},
			cookiePresent: true,
			cookieValue:   "invalid-token",
			wantStatus:    http.StatusInternalServerError,
			wantSetCookie: false,
		},
		{
			name: "error creating token",
			mockSetup: func(m *mockjwt.MockJWTServiceInterface) {
				m.EXPECT().CreateToken().Return("", "", errors.New("error"))
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

			req := httptest.NewRequest("GET", "/", nil)
			if tt.cookiePresent {
				req.AddCookie(&http.Cookie{
					Name:  "Access-token",
					Value: tt.cookieValue,
				})
			}

			rr := httptest.NewRecorder()

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
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantStatus)
			}

			if tt.wantSetCookie {
				cookies := rr.Result().Cookies()
				if len(cookies) == 0 {
					t.Error("expected cookie to be set, but none found")
				}
			}
		})
	}
}
