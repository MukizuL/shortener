package config

import (
	"errors"
	"flag"
	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrMalformedFlags = errors.New("error parsing flags")

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
		flag.StringVar(&result.Base, "b", "", "Sets server URL base. Example: string1/string2")
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
		os.Exit(1)
	}

	return result
}

func checkParams(cfg *Config) error {
	if cfg.Addr != "" {
		addr := strings.Split(cfg.Addr, ":")
		if len(addr) != 2 {
			return ErrMalformedFlags
		}

		_, err := strconv.Atoi(addr[1])
		if err != nil {
			return ErrMalformedFlags
		}
	}

	if cfg.Base != "" {
		if string(cfg.Base[0]) == "/" || string(cfg.Base[len(cfg.Base)-1]) == "/" {
			return ErrMalformedFlags
		}
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
