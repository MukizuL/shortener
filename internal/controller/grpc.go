package controller

import (
	"context"
	"errors"

	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/MukizuL/shortener/internal/interceptor"
	pb "github.com/MukizuL/shortener/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c Controller) CreateGRPC(
	ctx context.Context,
	in *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	var response pb.CreateShortURLResponse
	url, err := helpers.CheckURL([]byte(in.OriginalUrl))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "original url is not a url")
	}

	pair := ctx.Value(contextI.UserIDContextKey).(interceptor.TokenPair)

	shortURL, err := c.storage.CreateShortURL(ctx, pair.UserID, "", url)
	if err != nil {
		if errors.Is(err, errs.ErrDuplicate) {
			response.ShortUrl = shortURL
			return &response, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	response.ShortUrl = shortURL
	response.AccessToken = pair.AccessToken

	return &response, nil
}

func (c Controller) CreateBatchGRPC(
	ctx context.Context,
	in *pb.CreateBatchShortURLRequest) (*pb.CreateBatchShortURLResponse, error) {
	var response pb.CreateBatchShortURLResponse

	var req []dto.BatchRequest

	for _, v := range in.Batch {
		_, err := helpers.CheckURL([]byte(v.OriginalUrl))
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%s is not a url", v.OriginalUrl)
		}
		temp := dto.BatchRequest{
			CorrelationID: v.CorrelationId,
			OriginalURL:   v.OriginalUrl,
		}

		req = append(req, temp)
	}

	pair := ctx.Value(contextI.UserIDContextKey).(interceptor.TokenPair)

	resp, err := c.storage.BatchCreateShortURL(ctx, pair.UserID, "", req)
	if err != nil {
		if errors.Is(err, errs.ErrDuplicate) {
			return nil, status.Error(codes.AlreadyExists, "some urls already exist")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	var batch []*pb.BatchResponse
	for _, v := range resp {
		temp := pb.BatchResponse{
			CorrelationId: v.CorrelationID,
			ShortUrl:      v.ShortURL,
		}

		batch = append(batch, &temp)
	}

	response.Batch = batch
	response.AccessToken = pair.AccessToken

	return &response, nil
}

func (c Controller) GetOriginalURLGRPC(
	ctx context.Context,
	in *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	var response pb.GetOriginalURLResponse

	fullURL, err := c.storage.GetLongURL(ctx, in.ShortUrl)
	if err != nil {
		if errors.Is(err, errs.ErrURLNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		if errors.Is(err, errs.ErrGone) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	response.OriginalUrl = fullURL

	return &response, nil
}

func (c Controller) GetUserURLsGRPC(
	ctx context.Context,
	in *pb.GetUserURLRequest) (*pb.GetUserURLResponse, error) {
	var response pb.GetUserURLResponse

	var (
		pair interceptor.TokenPair
		ok   bool
	)

	if pair, ok = ctx.Value(contextI.UserIDContextKey).(interceptor.TokenPair); !ok {
		return nil, status.Error(codes.FailedPrecondition, "user id not found in context")
	}

	data, err := c.storage.GetUserURLs(ctx, pair.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pairs []*pb.URLPair
	for _, v := range data {
		temp := pb.URLPair{
			OriginalUrl: v.OriginalURL,
			ShortUrl:    v.ShortURL,
		}

		pairs = append(pairs, &temp)
	}

	response.Pairs = pairs
	response.AccessToken = pair.AccessToken

	return &response, nil
}

func (c Controller) DeleteGRPC(
	ctx context.Context,
	in *pb.DeleteShortURLRequest) (*pb.DeleteShortURLResponse, error) {
	var response pb.DeleteShortURLResponse

	pair := ctx.Value(contextI.UserIDContextKey).(interceptor.TokenPair)

	err := c.storage.DeleteURLs(ctx, pair.UserID, in.ShortUrls)
	if err != nil {
		if errors.Is(err, errs.ErrUserMismatch) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	response.AccessToken = pair.AccessToken

	return &response, nil
}

func (c Controller) GetStatsGRPC(
	ctx context.Context,
	in *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	var response pb.GetStatsResponse

	urls, users, err := c.storage.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response.Urls = int32(urls)
	response.Users = int32(users)

	return &response, nil
}
