package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStorage struct {
	conn *pgxpool.Pool
}

func (P *PGStorage) CreateShortURL(fullURL string) (string, error) {
	return "", nil
}

func (P *PGStorage) GetLongURL(ID string) (string, error) {
	return "", nil
}

func (P *PGStorage) OffloadStorage(filepath string) error {
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
