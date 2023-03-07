package api

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/parlaynu/studio1767-oidc-idp/internal/config"
	"github.com/parlaynu/studio1767-oidc-idp/internal/middleware/clientauth"
	"github.com/parlaynu/studio1767-oidc-idp/internal/middleware/mtls"
	"github.com/parlaynu/studio1767-oidc-idp/internal/middleware/trace"
)

type Service interface {
	AuthnStart(w http.ResponseWriter, r *http.Request)
	AuthnVerify(w http.ResponseWriter, r *http.Request)

	OIDCConfiguration(w http.ResponseWriter, r *http.Request)

	Keys(w http.ResponseWriter, r *http.Request)
	Tokens(w http.ResponseWriter, r *http.Request)
}

func New(cfg *config.Config, service Service) (http.Handler, http.Handler) {

	// frontend router
	f := chi.NewRouter()

	f.Use(trace.New)
	f.Use(mtls.New)

	f.Get("/auth", service.AuthnStart)
	f.Post("/auth", service.AuthnVerify)

	// backend routes
	b := chi.NewRouter()

	b.Use(trace.New)
	b.Use(clientauth.New(cfg.Clients))
	b.Get("/.well-known/openid-configuration", service.OIDCConfiguration)
	b.Get("/keys", service.Keys)
	b.Post("/token", service.Tokens)

	return f, b
}
