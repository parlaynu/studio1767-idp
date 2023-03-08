package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"

	"github.com/parlaynu/studio1767-oidc-idp-test/internal/config"
)

func NewAuthMiddleware(cfg *config.Config) (func(http.Handler) http.Handler, error) {
	// create the objects
	provider, err := newProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	se := newStateEncoder(cfg.Service.StateKey)
	cc := newCookieCutter(cfg.Service.CookieHashKey, cfg.Service.CookieEncKey, provider)

	return func(next http.Handler) http.Handler {
		am := authMware{
			next:     next,
			provider: provider,
			se:       se,
			cc:       cc,
		}
		return &am
	}, nil
}

type authMware struct {
	next     http.Handler
	provider *provider
	se       *stateEncoder
	cc       *cookieCutter
}

func (am *authMware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// handle known paths that don't require authentication
	if r.URL.Path == "/auth/logout" {
		am.handleLogout(w, r)
		return
	}
	if r.URL.Path == "/auth/callback" {
		am.handleCallback(w, r)
		return
	}

	// check for the auth cookie
	idToken, err := am.cc.getCookie(r)
	if err == nil && idToken != nil {
		am.handleRequest(w, r, idToken)
		return
	}

	// force a login
	am.handleLogin(w, r)
}

func (am *authMware) handleRequest(w http.ResponseWriter, r *http.Request, token *jwt.Token) {
	// get the claims from the token
	claims := token.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	// create a new context with the claims embedded
	ctx := context.WithValue(r.Context(), "claims", email)

	// call the next handler
	am.next.ServeHTTP(w, r.WithContext(ctx))
}

func (am *authMware) handleLogin(w http.ResponseWriter, r *http.Request) {
	// create the encoded state
	sd, err := am.se.encode(r.RequestURI)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Error("auth: failed to encode request")
		return
	}

	// redirect to the auth server
	url := am.provider.authCodeURL(sd.Encoded, sd.Nonce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (am *authMware) handleLogout(w http.ResponseWriter, r *http.Request) {
	am.cc.clearCookie(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (am *authMware) handleCallback(w http.ResponseWriter, r *http.Request) {
	// check for required paramaters
	required := []string{
		"code",
		"state",
	}
	if checkParameters(r, required) == false {
		w.WriteHeader(http.StatusBadRequest)
		log.Error("auth: missing parameters in request")
		return
	}

	code := r.FormValue("code")
	state := r.FormValue("state")

	// exchange the code for the raw tokens
	idToken, err := am.provider.exchange(r.Context(), code)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Errorf("auth: failed to exchange code for token: %v", err)
		return
	}

	// get the claims
	claims := idToken.Claims.(jwt.MapClaims)

	// decode the state
	nonce := claims["nonce"].(string)
	sd, err := am.se.decode(state, nonce)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Errorf("auth: failed to decode request: %v", err)
		return
	}

	// encode the raw id token as a cookie
	err = am.cc.addCookie(w, idToken.Raw)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Errorf("auth: failed to add cookie: %v", err)
		return
	}

	// redirect to the landing page
	http.Redirect(w, r, sd.RedirectURL, http.StatusTemporaryRedirect)
}

func checkParameters(r *http.Request, params []string) bool {
	// make sure the parameters are all present
	for _, p := range params {
		v := r.FormValue(p)
		if v == "" {
			return false
		}
	}

	return true
}
