package save

import (
	"errors"
	"log/slog"
	"net/http"
	resp "urlShorter/internal/lib/api/response"
	"urlShorter/internal/lib/logger/sl"
	"urlShorter/internal/lib/random"
	"urlShorter/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

// Константу можно перенести в конфиг
const aliasLength = 4

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	const op = "handlers.url.save.New"
	return func(w http.ResponseWriter, r *http.Request) {
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("ошибка декодирования тела запроса", sl.Err(err))
			render.JSON(w, r, resp.Error("ошибка декодирования запроса"))
			return
		}

		log.Info("тело запроса декодировано", slog.Any("запрос:", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("неверный запрос", sl.Err(err))
			validateErr := err.(validator.ValidationErrors)
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
			// Посмотреть как можно реализовать исправление совпадающих alias
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExist) {
			log.Info("url уже существует", slog.String("url", req.URL))
			render.JSON(w, r, resp.Error("url уже существует"))
			return
		}
		if err != nil {
			log.Error("ошибка добавления url", sl.Err(err))
			render.JSON(w, r, resp.Error("ошибка добавления url"))
			return
		}

		log.Info("url добавлен", slog.Int64("id", id))
		responseOK(w, r, alias)

	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
