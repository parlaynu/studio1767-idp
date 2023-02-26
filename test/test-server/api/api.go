package api

// authentication - who are you
// authorizing - who can call the functions
// routing - paths to functions
// validating - parameters and responses
// delegation - pass on to implementations

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"

	"s1767.xyz/test/internal/auth"
	"s1767.xyz/test/internal/config"
	"s1767.xyz/test/internal/trace"
)

type Service interface {
	Hello(w http.ResponseWriter, r *http.Request)
}

func New(cfg *config.Config, service Service) (http.Handler, error) {

	// create the top router
	r := chi.NewRouter()

	// auth middleware handles login, logout, etc. workflows
	r.Use(trace.New)

	mw, err := auth.NewAuthMiddleware(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed creating auth middleware: %w", err)
	}
	r.Use(mw)

	r.Get("/", service.Hello)

	return r, nil
}
