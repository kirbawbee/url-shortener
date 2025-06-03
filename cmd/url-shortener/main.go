package main

import (
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/save"
	mwLogger "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Конфиг : cleanenv
	cfg := config.MustLoad()

	// Логер : slog
	log := setupLogger(cfg.Env)
	log.Info("начата работа url-shorter", slog.String("env", cfg.Env))
	log.Debug("дебаг-сообщения были включены")

	// Storage: sqllite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("ошибка инициализации базы данных", sl.Err(err))
		os.Exit(1)

	}

	// Роутер: chi, chi render
	router := chi.NewRouter()
	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password}))

		r.Post("/", save.New(log, storage))
		router.Delete("/{alias}", redirect.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("Запуск сервера", slog.String("address", cfg.Address))
	// Запуск сервера
	srv := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("ошибка запуска сервера")
	}

	log.Error("сервер остановлен")
}

func setupLogger(Env string) *slog.Logger {
	var log *slog.Logger
	switch Env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}
