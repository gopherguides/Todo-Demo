package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/config"
	"todo-demo/internal/database"
	"todo-demo/internal/handler"
	mw "todo-demo/internal/middleware"
)

func main() {
	setupLogger()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("loading config", "error", err)
		os.Exit(1)
	}

	mw.ClerkInit(cfg.ClerkSecretKey)

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("opening database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		slog.Error("running migrations", "error", err)
		os.Exit(1)
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	for _, m := range mw.Common() {
		e.Use(m)
	}

	e.Use(mw.OptionalClerkAuth())
	e.Static("/static", "static")

	h := handler.New(db, cfg)
	h.Register(e)

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("starting server", "addr", addr)
	if err := e.Start(addr); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
