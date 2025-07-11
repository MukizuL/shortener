package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MukizuL/shortener/docs"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"
)

var ErrMalformedFlags = errors.New("error parsing flags")
var ErrMalformedAddr = errors.New("address of wrong format")
var ErrMalformedBase = errors.New("base should be an url")

// Config holds all application configuration.
type Config struct {
	Addr           string `env:"SERVER_ADDRESS" json:"server_address"`
	Base           string `env:"BASE_URL" json:"base_url"`
	Config         string `env:"CONFIG" json:"config"`
	Filepath       string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DSN            string `env:"DATABASE_DSN" json:"database_dsn"`
	MasterPassword string `env:"MASTER_PASSWORD" json:"master_password"`
	HTTPS          bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	Cert           string `env:"CERT_PATH" json:"cert_path"`
	PK             string `env:"PK_PATH" json:"pk_path"`
	Debug          bool   `env:"DEBUG" json:"debug"`
}

// newConfig fetches parameters, firstly from env variables, secondly from flags, then from file.
func newConfig() (*Config, error) {
	resultCfg := &Config{}

	envCfg, err := envConfig()
	if err != nil {
		flag.Usage()
		return nil, fmt.Errorf("error loading config from env: %w", err)
	}

	flagCfg, err := flagConfig()
	if err != nil {
		flag.Usage()
		return nil, fmt.Errorf("error loading config from flag: %w", err)
	}

	if envCfg.Config != "" || flagCfg.Config != "" {
		fileCfg, err := fileConfig("")
		if err != nil {
			return nil, fmt.Errorf("error loading config from file: %w", err)
		}
		mergeConfig(resultCfg, fileCfg)
	}

	mergeConfig(resultCfg, flagCfg)

	mergeConfig(resultCfg, envCfg)

	fmt.Printf("%+v\n", resultCfg)
	err = checkParams(resultCfg)
	if err != nil {
		flag.Usage()
		return nil, err
	}

	return resultCfg, nil
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

	if cfg.HTTPS {
		if cfg.Cert == "" {
			return errs.ErrNoCert
		}

		if cfg.PK == "" {
			return errs.ErrNoPK
		}

		err := checkFiles(cfg.Cert, cfg.PK)
		if err != nil {
			return err
		}
	}

	//if cfg.MasterPassword == "" {
	//	return fmt.Errorf("missing private key")
	//}

	setSwagger(cfg)

	return nil
}

func checkFiles(cert, pk string) error {
	if _, err := os.Stat(cert); errors.Is(err, os.ErrNotExist) {
		return errs.ErrNoCert
	}

	if _, err := os.Stat(pk); errors.Is(err, os.ErrNotExist) {
		return errs.ErrNoPK
	}

	return nil
}

func setSwagger(cfg *Config) {
	docs.SwaggerInfo.Host = cfg.Addr
	docs.SwaggerInfo.BasePath = cfg.Base
}

func envConfig() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func flagConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "a", "0.0.0.0:8080", "Sets server address.")

	flag.StringVar(&cfg.Base, "b", "", "Sets server URL base. Example: http(s)://address:port/your/base")

	flag.StringVar(&cfg.Filepath, "r", "./storage.json", "Sets server storage file path.")

	flag.StringVar(&cfg.Config, "c", "", "Sets server config file name.")

	flag.StringVar(&cfg.DSN, "d", "", "Sets server DSN.")

	flag.BoolVar(&cfg.HTTPS, "s", false, "Turns on HTTPS. Requires cert and pk to be set.")

	flag.StringVar(&cfg.Cert, "cert", "", "Sets certificate file path.")

	flag.StringVar(&cfg.PK, "pk", "", "Sets private key file path.")

	flag.BoolVar(&cfg.Debug, "debug", false, "Sets server debug mode.")

	flag.Parse()

	return cfg, nil
}

func fileConfig(name string) (*Config, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func mergeConfig(dst, src *Config) {
	if src == nil {
		return
	}

	if src.Addr != "" {
		dst.Addr = src.Addr
	}
	if src.Base != "" {
		dst.Base = src.Base
	}
	if src.Config != "" {
		dst.Config = src.Config
	}
	if src.Filepath != "" {
		dst.Filepath = src.Filepath
	}
	if src.DSN != "" {
		dst.DSN = src.DSN
	}
	if src.MasterPassword != "" {
		dst.MasterPassword = src.MasterPassword
	}
	if src.Cert != "" {
		dst.Cert = src.Cert
	}
	if src.PK != "" {
		dst.PK = src.PK
	}
	// Booleans: only overwrite if true to preserve priority
	if src.HTTPS {
		dst.HTTPS = true
	}
	if src.Debug {
		dst.Debug = true
	}
}

func Provide() fx.Option {
	return fx.Provide(newConfig)
}
