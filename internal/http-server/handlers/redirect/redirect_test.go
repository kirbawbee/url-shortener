package redirect_test

import (
	"errors"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/redirect"
	mocks "url-shortener/internal/http-server/handlers/redirect/moks"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func testSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError string
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://www.google.com",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMocks := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != "" {
				urlGetterMocks.On("GetURL", tc.alias).
					Return(tc.url, errors.New(tc.mockError)).Once()

			}

			r := chi.NewRouter()
			r.Get("alias", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMocks))

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			require.Equal(t, tc.url, redirectedToURL)
		})
	}
}
