package main

import (
	"log"

	"github.com/nechititudorr/GoSocial/internal/db"
	"github.com/nechititudorr/GoSocial/internal/env"
	"github.com/nechititudorr/GoSocial/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewStorage(conn)

	db.Seed(store)
}
