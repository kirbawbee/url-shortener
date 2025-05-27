package delete

import (
	"errors"

	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
)

//go:generate mockery --name URLDeleter --output ./moks --filename URLDeleter.go --case underscore
type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.delete.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias пустой")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("alias не может быть пустым"))

			return
		}

		err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("не найден"))

			return
		}
		if err != nil {
			log.Error("ошибка удаления url", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("внутренняя ошибка"))

			return
		}
		log.Info("url удален", slog.String("alias", alias))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}
