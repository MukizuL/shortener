package pg_storage

import (
	"context"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStorage struct {
	conn *pgxpool.Pool
}

func (P *PGStorage) CreateShortURL(ctx context.Context, fullURL string) (string, error) {
	ID := helpers.RandomString(6)
	_, err := P.conn.Exec(ctx, `INSERT INTO urls (short_url, full_url) VALUES ($1, $2)`, ID, fullURL)
	if err != nil {
		return "", err
	}

	return "http://localhost:8080/" + ID, nil
}

func (P *PGStorage) GetLongURL(ctx context.Context, ID string) (string, error) {
	var result string
	err := P.conn.QueryRow(ctx, `SELECT full_url FROM urls WHERE short_url = $1`, ID).Scan(&result)
	if err != nil {
		return "", err
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

func New(ctx context.Context, dsn string) (*PGStorage, error) {
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &PGStorage{conn: dbpool}, nil
}
