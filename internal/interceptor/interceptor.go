package interceptor

import (
	"context"
	"errors"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/MukizuL/shortener/internal/config"
	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/errs"
	jwtService "github.com/MukizuL/shortener/internal/jwt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type TokenPair struct {
	AccessToken string
	UserID      string
}

type Service struct {
	jwtService jwtService.JWTServiceInterface
	cfg        *config.Config
	logger     *zap.Logger
}

func newService(jwtService jwtService.JWTServiceInterface, cfg *config.Config, logger *zap.Logger) *Service {
	return &Service{
		jwtService: jwtService,
		cfg:        cfg,
		logger:     logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(newService)
}

func (s Service) Logger(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	duration := time.Since(start)

	s.logger.Info("GRPC request", zap.String("method", info.FullMethod), zap.Duration("time", duration))
	return resp, err
}

func (s Service) Auth(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	routes := []string{
		"/shortener.Shortener/CreateGRPC",
		"/shortener.Shortener/CreateBatchGRPC",
		"/shortener.Shortener/GetUserURLsGRPC",
		"/shortener.Shortener/DeleteGRPC",
	}

	if !slices.Contains(routes, info.FullMethod) {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	var token, userID string
	var err error

	tokens := md.Get("Access-token")

	if len(tokens) == 0 {
		token, userID, err = s.jwtService.CreateOrValidateToken("")
	} else {
		token, userID, err = s.jwtService.ValidateToken(tokens[0])
	}
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrNotAuthorized), errors.Is(err, errs.ErrUnexpectedSigningMethod):
			return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
		case errors.Is(err, errs.ErrSigningToken):
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		case errors.Is(err, errs.ErrRefreshingToken):
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	data := TokenPair{
		AccessToken: token,
		UserID:      userID,
	}

	newCtx := context.WithValue(ctx, contextI.UserIDContextKey, data)

	return handler(newCtx, req)
}

func (s Service) IsTrustedCIDR(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if info.FullMethod != "/shortener.Shortener/GetStatsGRPC" {
		return handler(ctx, req)
	}

	if s.cfg.TrustedCIDR == "" {
		return nil, status.Errorf(codes.PermissionDenied, "ip is not in trusted subnet")
	}

	pr, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "unable to get peer address")
	}

	host, _, err := net.SplitHostPort(pr.Addr.String())
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "invalid peer address format")
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, status.Error(codes.PermissionDenied, "unable to parse IP")
	}

	_, subnet, err := net.ParseCIDR(s.cfg.TrustedCIDR)
	if err != nil {
		return nil, status.Error(codes.Internal, http.StatusText(http.StatusInternalServerError))
	}

	if !subnet.Contains(ip) {
		s.logger.Warn("IP is not trusted", zap.String("ip", ip.String()))
		return nil, status.Errorf(codes.PermissionDenied, "ip is not in trusted subnet")
	}

	return handler(ctx, req)
}

func (s Service) Recovery(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("panic recovered",
				zap.Any("panic", r),
				zap.String("method", info.FullMethod),
				zap.Stack("stack"),
			)
			err = status.Error(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}
