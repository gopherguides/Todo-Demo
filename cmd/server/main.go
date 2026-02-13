package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"syscall"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/labstack/echo/v4"

	"todo-demo/internal/config"
	"todo-demo/internal/database"
	"todo-demo/internal/handler"
	mw "todo-demo/internal/middleware"
)

const maxPortAttempts = 20

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
	defer func() { _ = db.Close() }()

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

	e.Use(echo.WrapMiddleware(chimw.Logger))

	ln, port, err := findOpenPort(cfg.Port, maxPortAttempts)
	if err != nil {
		slog.Error("no available port found", "start", cfg.Port, "attempts", maxPortAttempts, "error", err)
		os.Exit(1)
	}

	slog.Info("starting server", "port", port, "url", fmt.Sprintf("http://localhost:%d", port))
	e.Listener = ln
	if err := e.Start(""); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func findOpenPort(startPort, maxAttempts int) (net.Listener, int, error) {
	for i := range maxAttempts {
		port := startPort + i
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			return ln, port, nil
		}
		if !isAddrInUse(err) {
			return nil, 0, fmt.Errorf("listening on %s: %w", addr, err)
		}
		slog.Debug("port in use, trying next", "port", port)
	}
	return nil, 0, fmt.Errorf("ports %d-%d all in use", startPort, startPort+maxAttempts-1)
}

func isAddrInUse(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var syscallErr *os.SyscallError
		if errors.As(opErr.Err, &syscallErr) {
			return errors.Is(syscallErr.Err, syscall.EADDRINUSE)
		}
	}
	return false
}
