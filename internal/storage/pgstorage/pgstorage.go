package pgstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (s *PGStorage) BatchCreateShortURL(ctx context.Context, userID, urlBase string, data []dto.BatchRequest) ([]dto.BatchResponse, error) {
	const batchSize = 2
	const numCols = 3

	result := make([]dto.BatchResponse, 0, len(data))

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		s.logger.Error("pgstorage:BatchCreateShortURL Transaction start", zap.Error(err))
		return nil, errs.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	batches := helpers.SplitIntoBatches(data, batchSize)

	for _, batch := range batches {
		numRows := len(batch)
		args := make([]interface{}, 0, numRows*numCols)
		for _, item := range batch {
			ID := helpers.RandomString(6)
			args = append(args, userID, ID, item.OriginalURL)

			result = append(result, dto.BatchResponse{CorrelationID: item.CorrelationID, ShortURL: urlBase + ID})
		}

		valuesPart := helpers.BuildValuePlaceholders(numCols, numRows)

		query := fmt.Sprintf("INSERT INTO urls (user_id, short_url, full_url) VALUES %s", valuesPart)

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case pgerrcode.UniqueViolation:
					return nil, errs.ErrDuplicate
				}
			}

			s.logger.Error("pgstorage:BatchCreateShortURL other pg error", zap.Error(pgErr))
			return nil, errs.ErrInternalServerError
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error("pgstorage:BatchCreateShortURL ", zap.Error(err))
		return nil, errs.ErrInternalServerError
	}

	return result, nil
}

func (s *PGStorage) CreateShortURL(ctx context.Context, userID, urlBase, fullURL string) (string, error) {
	ID := helpers.RandomString(6)
	err := s.conn.QueryRow(ctx, `INSERT INTO urls (user_id, short_url, full_url)
										VALUES ($1, $2, $3)
										ON CONFLICT(full_url)
										DO UPDATE SET full_url = urls.full_url
										RETURNING short_url`, userID, ID, fullURL).Scan(&ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return s.CreateShortURL(ctx, userID, urlBase, fullURL)
			}
		}

		s.logger.Error("pgstorage:CreateShortURL ", zap.Error(err))
		return "", errs.ErrInternalServerError
	}

	return urlBase + ID, nil
}

func (s *PGStorage) GetLongURL(ctx context.Context, ID string) (string, error) {
	var result string
	var deleted bool
	err := s.conn.QueryRow(ctx, `SELECT full_url, deleted_flag FROM urls WHERE short_url = $1`, ID).Scan(&result, &deleted)
	if err != nil {
		s.logger.Error("pgstorage:GetLongURL ", zap.Error(err))
		return "", errs.ErrInternalServerError
	}

	if result == "" {
		return "", errs.ErrURLNotFound
	}

	if deleted {
		return "", errs.ErrGone
	}

	return result, nil
}

func (s *PGStorage) GetUserURLs(ctx context.Context, userID string) ([]dto.URLPair, error) {
	var result []dto.URLPair
	rows, err := s.conn.Query(ctx, "SELECT short_url, full_url FROM urls WHERE user_id = $1 AND deleted_flag = FALSE", userID)
	if err != nil {
		s.logger.Error("pgstorage:GetUserURLs ", zap.Error(err))
		return nil, errs.ErrInternalServerError
	}
	defer rows.Close()

	var shortURL, fullURL string
	for rows.Next() {
		err = rows.Scan(&shortURL, &fullURL)
		if err != nil {
			s.logger.Error("pgstorage:GetUserURLs Error in row", zap.Error(err))
			continue
		}

		data := dto.URLPair{
			ShortURL:    shortURL,
			OriginalURL: fullURL,
		}

		result = append(result, data)
	}

	if rows.Err() != nil {
		s.logger.Error("pgstorage:GetUserURLs Error in rows", zap.Error(err))
		return nil, rows.Err()
	}

	return result, nil
}

func (s *PGStorage) DeleteURLs(ctx context.Context, userID string, urls []string) error {
	query := "UPDATE urls SET deleted_flag = TRUE WHERE user_id = $1 AND short_url ANY($2)"

	result, err := s.conn.Exec(ctx, query, userID, pq.Array(urls))
	if err != nil {
		s.logger.Error("pgstorage:DeleteURLs ", zap.Error(err))
		return errs.ErrInternalServerError
	}

	if result.RowsAffected() == 0 {
		return errs.ErrUserMismatch
	}

	return nil
}

func (s *PGStorage) OffloadStorage(ctx context.Context, filepath string) error {
	return nil
}

func (s *PGStorage) Ping(ctx context.Context) error {
	return s.conn.Ping(ctx)
}

func (s *PGStorage) Close() {
	s.conn.Close()
}
