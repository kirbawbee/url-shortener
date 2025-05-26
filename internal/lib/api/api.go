package api

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidStatusCode = errors.New("неверный статус кода")
	ErrNotFound          = errors.New("ответ не найден")
)

func GetRedirect(url string) (string, error) {
	const op = "api.GetRedirect"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("%s: request failed: %w", op, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusFound, http.StatusMovedPermanently:
		location := resp.Header.Get("Location")
		if location == "" {
			return "", fmt.Errorf("%s: empty Location header", op)
		}
		return location, nil
	case http.StatusNotFound:
		return "", fmt.Errorf("%s: %w", op, ErrNotFound)
	default:
		return "", fmt.Errorf("%s: %w (status %d)", op, ErrInvalidStatusCode, resp.StatusCode)
	}
}
