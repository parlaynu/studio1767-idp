package clientauth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/parlaynu/studio1767-idp/internal/config"
	"github.com/parlaynu/studio1767-idp/internal/middleware/clientauth"
)

func TestClientAuth(t *testing.T) {
	cfg := createConfig()
	mware := clientauth.New(cfg)(&handler{})

	// test with auth
	{
		url := fmt.Sprintf("http://127.0.0.1/blah?client_id=%s&client_secret=%s", cfg[0].Id, cfg[0].Secret)
		request, err := http.NewRequest("GET", url, strings.NewReader("hello world"))
		require.NoError(t, err)
		rec := httptest.NewRecorder()

		mware.ServeHTTP(rec, request)
		response := rec.Result()
		require.Equal(t, http.StatusOK, response.StatusCode)
	}
	// test with bad auth
	{
		url := fmt.Sprintf("http://127.0.0.1/blah?client_id=%s&client_secret=%s", cfg[0].Id, cfg[1].Secret)
		request, err := http.NewRequest("GET", url, strings.NewReader("hello world"))
		require.NoError(t, err)
		rec := httptest.NewRecorder()

		mware.ServeHTTP(rec, request)
		response := rec.Result()
		require.Equal(t, http.StatusUnauthorized, response.StatusCode)
	}
	// test without auth
	{
		request, err := http.NewRequest("GET", "http://127.0.0.1/blah", strings.NewReader("hello world"))
		require.NoError(t, err)
		rec := httptest.NewRecorder()

		mware.ServeHTTP(rec, request)
		response := rec.Result()
		require.Equal(t, http.StatusUnauthorized, response.StatusCode)
	}

}

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World"))
}

func createConfig() []*config.ClientConfig {
	clients := []*config.ClientConfig{}
	nclients := 5
	for i := 0; i < nclients; i++ {
		client := config.ClientConfig{
			Id:           fmt.Sprintf("Id%d", i),
			Secret:       fmt.Sprintf("Secret%d", i),
			RedirectURLs: []string{fmt.Sprintf("http://server%d.example.com", i)},
		}
		clients = append(clients, &client)
	}

	return clients
}
