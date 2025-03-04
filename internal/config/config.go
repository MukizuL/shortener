package config

import (
	"errors"
	"flag"
	"github.com/caarlos0/env/v11"
	"log"
	"os"
	"strconv"
	"strings"
)

var ErrMalformedFlags = errors.New("error parsing flags")

type Params struct {
	Addr string `env:"SERVER_ADDRESS"`
	Base string `env:"BASE_URL"`
}

// GetParams fetches parameters, firstly from env variables, secondly from flags
func GetParams() *Params {
	result := &Params{}

	err := env.Parse(result)
	if err != nil {
		log.Fatal("Error parsing env variables")
	}

	if result.Addr == "" {
		flag.StringVar(&result.Addr, "a", "localhost:8080", "Sets server address.")
	}

	if result.Base == "" {
		flag.StringVar(&result.Base, "b", "", "Sets server URL base. Example: string1/string2")
	}

	flag.Parse()

	err = checkParams(result)
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	return result
}

func checkParams(data *Params) error {
	if data.Addr != "" {
		addr := strings.Split(data.Addr, ":")
		if len(addr) != 2 {
			return ErrMalformedFlags
		}

		_, err := strconv.Atoi(addr[1])
		if err != nil {
			return ErrMalformedFlags
		}
	}

	if data.Base != "" {
		if string(data.Base[0]) == "/" || string(data.Base[len(data.Base)-1]) == "/" {
			return ErrMalformedFlags
		}
	}

	return nil
}
