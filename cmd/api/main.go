package main

import (
	"time"

	"github.com/nechititudorr/GoSocial/internal/db"
	"github.com/nechititudorr/GoSocial/internal/env"
	"github.com/nechititudorr/GoSocial/internal/mailer"
	"github.com/nechititudorr/GoSocial/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			GoSocial API
//	@description	This is a sample server celler server.
//	@version		1.0.0

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1

// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiUrl: env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			expiry: time.Hour * 24 * 2,
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	conn, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Info("DB CONNECTION ESTABLISHED!")

	storage := store.NewStorage(conn)

	app := &application{
		config: cfg,
		store:  storage,
		logger: logger,
		mailer: mailer.NoopClient{},
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
