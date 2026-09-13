package main

import (
	"log"

	"github.com/nechititudorr/GoSocial/internal/db"
	"github.com/nechititudorr/GoSocial/internal/env"
	"github.com/nechititudorr/GoSocial/internal/store"
)

const version = "0.0.1"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
	}

	conn, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB CONNECTION ESTABLISHED!")

	storage := store.NewStorage(conn)

	app := &application{
		config: cfg,
		store:  storage,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))
}
