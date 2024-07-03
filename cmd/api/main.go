package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	//we will generate this automatically at build time, but for now we will just store the version number as hardcoded global const
	version = "1.0.0"
)

// this will hold all the config settings for our app,
// port is the port the app will run at, and env is the operating environment (dev,staging,production)
type config struct {
	port int
	env  string
}

// this will hold the dependencies for our http handlers, helpers and middleware.
type application struct {
	config config
	logger *slog.Logger
}

func main() {
	//instance of config struct
	var cfg config
	//read the value of the config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "environment (development|staging|production)")
	flag.Parse()

	//init a new slog that writes log entries to stdout stream
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	//declare instance of app strcut containg the config struct and logger
	app := &application{
		config: cfg,
		logger: logger,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)
	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
