package pgstorage

import (
	"context"
	"errors"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PGStorage struct {
	conn   *pgxpool.Pool
	logger *zap.Logger
}

func (P *PGStorage) BatchCreateShortURL(ctx context.Context, data []dto.BatchRequest) ([]dto.BatchResponse, error) {
	result := make([]dto.BatchResponse, 0, len(data))

	tx, err := P.conn.Begin(ctx)
	if err != nil {
		P.logger.Error("pgstorage:BatchCreateShortURL Transaction start", zap.Error(err))
		return nil, errs.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	for _, v := range data {
		ID := helpers.RandomString(6)
		_, err = tx.Exec(ctx, `INSERT INTO urls (short_url, full_url) VALUES ($1, $2)`, ID, v.OriginalURL)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case pgerrcode.UniqueViolation:
					return nil, errs.ErrDuplicate
				}
			}

			P.logger.Error("pgstorage:BatchCreateShortURL other pg error", zap.Error(pgErr))
			return nil, errs.ErrInternalServerError
		}

		result = append(result, dto.BatchResponse{CorrelationID: v.CorrelationID, ShortURL: "http://localhost:8080/" + ID})
	}

	err = tx.Commit(ctx)
	if err != nil {
		P.logger.Error("pgstorage:BatchCreateShortURL ", zap.Error(err))
		return nil, errs.ErrInternalServerError
	}

	return result, nil
}

func (P *PGStorage) CreateShortURL(ctx context.Context, fullURL string) (string, error) {
	ID := helpers.RandomString(6)
	_, err := P.conn.Exec(ctx, `INSERT INTO urls (short_url, full_url) VALUES ($1, $2)`, ID, fullURL)
	if err != nil { // What a horrific abomination
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				switch pgErr.ConstraintName {
				case "urls_full_url_key":
					err = P.conn.QueryRow(ctx, `SELECT short_url FROM urls WHERE full_url = $1`, fullURL).Scan(&ID)
					if err != nil {
						P.logger.Error("pgstorage:BatchCreateShortURL ", zap.Error(err))
						return "", errs.ErrInternalServerError
					}

					return "http://localhost:8080/" + ID, errs.ErrDuplicate
				case "urls_short_url_key":
					return P.CreateShortURL(ctx, fullURL)
				}
			}
		}

		P.logger.Error("pgstorage:BatchCreateShortURL ", zap.Error(err))
		return "", errs.ErrInternalServerError
	}

	return "http://localhost:8080/" + ID, nil
}

func (P *PGStorage) GetLongURL(ctx context.Context, ID string) (string, error) {
	var result string
	err := P.conn.QueryRow(ctx, `SELECT full_url FROM urls WHERE short_url = $1`, ID).Scan(&result)
	if err != nil {
		P.logger.Error("pgstorage:BatchCreateShortURL ", zap.Error(err))
		return "", errs.ErrInternalServerError
	}

	if result == "" {
		return "", errs.ErrNotFound
	}

	return result, nil
}

func (P *PGStorage) OffloadStorage(ctx context.Context, filepath string) error {
	return nil
}

func (P *PGStorage) Ping(ctx context.Context) error {
	return P.conn.Ping(ctx)
}

func (P *PGStorage) Close() {
	P.conn.Close()
}

func New(ctx context.Context, dsn string, logger *zap.Logger) (*PGStorage, error) {
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &PGStorage{conn: dbpool, logger: logger}, nil
}
