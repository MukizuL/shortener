package app

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type application struct {
	storage    map[string]string
	createdURL map[string]struct{}
	m          sync.Mutex
	seededRand *rand.Rand
}

func Run() {
	app := &application{
		storage:    make(map[string]string),
		createdURL: make(map[string]struct{}),

		seededRand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", app.CreateShortURL)
	mux.HandleFunc("GET /{id}", app.GetFullURL)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
