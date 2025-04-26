package config

import (
	"errors"
	"flag"
	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"
	"log"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrMalformedFlags = errors.New("error parsing flags")
var ErrMalformedAddr = errors.New("address of wrong format")
var ErrMalformedBase = errors.New("base should be an url")

type Config struct {
	Addr     string `env:"SERVER_ADDRESS"`
	Base     string `env:"BASE_URL"`
	Filepath string `env:"FILE_STORAGE_PATH"`
	DSN      string `env:"DATABASE_DSN"`
	Key      string `env:"PRIVATE_KEY"`
}

// NewConfig fetches parameters, firstly from env variables, secondly from flags
func NewConfig() *Config {
	result := &Config{}

	err := env.Parse(result)
	if err != nil {
		log.Fatal("Error parsing env variables")
	}

	if result.Addr == "" {
		flag.StringVar(&result.Addr, "a", "0.0.0.0:8080", "Sets server address.")
	}

	if result.Base == "" {
		flag.StringVar(&result.Base, "b", "", "Sets server URL base. Example: http(s)://address:port/*")
	}

	absPath, _ := filepath.Abs("./storage.json")

	if result.Filepath == "" {
		flag.StringVar(&result.Filepath, "r", absPath, "Sets server storage file path.")
	}

	if result.DSN == "" {
		flag.StringVar(&result.DSN, "d", "", "Sets server DSN.")
	}

	flag.Parse()

	err = checkParams(result)
	if err != nil {
		flag.Usage()
		log.Fatal(err)
	}

	return result
}

func checkParams(cfg *Config) error {
	if cfg.Addr != "" {
		addr := strings.Split(cfg.Addr, ":")
		if len(addr) != 2 {
			return ErrMalformedAddr
		}

		_, err := strconv.Atoi(addr[1])
		if err != nil {
			return ErrMalformedAddr
		}
	}

	if cfg.Base != "" {
		parsedURL, err := url.Parse(cfg.Base)
		if err != nil {
			return ErrMalformedBase
		}

		cfg.Base = strings.TrimSuffix(parsedURL.RequestURI(), "/")
	}

	if !filepath.IsAbs(cfg.Filepath) {
		temp, err := filepath.Abs(cfg.Filepath)
		if err != nil {
			return ErrMalformedFlags
		}

		cfg.Filepath = temp
	}

	//if cfg.Key == "" {
	//	return fmt.Errorf("missing private key in PRIVATE_KEY env")
	//}

	return nil
}

func Provide() fx.Option {
	return fx.Provide(NewConfig)
}
