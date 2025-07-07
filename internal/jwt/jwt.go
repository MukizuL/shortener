package jwt

import (
	"fmt"
	"time"

	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/fx"
)

//go:generate mockgen -source=jwt.go -destination=mocks/jwt.go -package=mockjwt

type JWTServiceInterface interface {
	ValidateToken(token string) (string, string, error)
	CreateToken() (string, string, error)
	RefreshToken(userID string) (string, error)
}

type JWTService struct {
	key []byte
}

func newJWTService(cfg *config.Config) JWTServiceInterface {
	return &JWTService{
		key: []byte(cfg.Key),
	}
}

func Provide() fx.Option {
	return fx.Provide(newJWTService)
}

// ValidateToken returns parsed token, userID, and an error
func (s *JWTService) ValidateToken(token string) (string, string, error) {
	var claims jwt.RegisteredClaims
	accessToken, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", errs.ErrUnexpectedSigningMethod, token.Header["alg"])
		}

		return s.key, nil
	})
	if err != nil {
		return "", "", err
	}

	if !accessToken.Valid {
		return "", "", errs.ErrNotAuthorized
	}

	newToken, err := s.RefreshToken(claims.Subject)
	if err != nil {
		return "", "", err
	}

	return newToken, claims.Subject, nil
}

// CreateToken returns a new token, user_id and an error
func (s *JWTService) CreateToken() (string, string, error) {
	userID := uuid.New().String()

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(876000 * time.Second)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	accessTokenSigned, err := accessToken.SignedString(s.key)
	if err != nil {
		return "", "", errs.ErrSigningToken
	}

	return accessTokenSigned, userID, nil
}

// RefreshToken returns a new token with same user_id and an error
func (s *JWTService) RefreshToken(userID string) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(876000 * time.Second)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	accessTokenSigned, err := accessToken.SignedString(s.key)
	if err != nil {
		return "", errs.ErrRefreshingToken
	}

	return accessTokenSigned, nil
}
