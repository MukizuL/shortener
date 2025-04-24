package app

import (
	"context"
	"github.com/MukizuL/shortener/internal/dto"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//go:generate mockgen -source=app.go -destination=mocks/app.go -package=mocksapp

type Repo interface {
	CreateShortURL(ctx context.Context, urlBase, fullURL string) (string, error)
	BatchCreateShortURL(ctx context.Context, urlBase string, data []dto.BatchRequest) ([]dto.BatchResponse, error)
	GetLongURL(ctx context.Context, ID string) (string, error)
	OffloadStorage(ctx context.Context, filepath string) error
	Ping(ctx context.Context) error
}

//	func Run() error {
//		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
//		defer stop()
//
//		log, err := zap.NewDevelopment()
//		if err != nil {
//			return err
//		}
//		defer log.Sync()
//
//		params := config.GetParams()
//
//		var repository Repo
//		if params.DSN != "" {
//			err = Migrate(params.DSN)
//			if err != nil {
//				return err
//			}
//
//			db, err := pgstorage.New(ctx, params.DSN, log)
//			if err != nil {
//				return err
//			}
//			defer db.Close()
//
//			repository = db
//
//		} else {
//			repository, err = mapstorage.New(ctx, params.Filepath, log)
//			if err != nil {
//				return err
//			}
//		}
//
//		app := NewApplication(repository, log)
//
//		r := NewRouter(params.Base, app)
//
//		err = runServer(ctx, params.Addr, r)
//		if err != nil {
//			return err
//		}
//
//		err = app.storage.OffloadStorage(ctx, params.Filepath)
//		if err != nil {
//			return err
//		}
//
//		return nil
//	}
