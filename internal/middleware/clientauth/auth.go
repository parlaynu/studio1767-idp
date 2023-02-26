package clientauth

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"s1767.xyz/idp/internal/config"
)

func New(clients []config.ClientConfig) func(http.Handler) http.Handler {
	cmap := make(map[string]config.ClientConfig)
	for _, client := range clients {
		cmap[client.Id] = client
	}

	return func(next http.Handler) http.Handler {
		cauth := clientAuth{
			clients: cmap,
			next:    next,
		}
		return &cauth
	}
}

type clientAuth struct {
	clients map[string]config.ClientConfig
	next    http.Handler
}

func (cauth *clientAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// validate the form values for client and secret
	id := r.FormValue("client_id")
	secret := r.FormValue("client_secret")

	if len(id) == 0 || len(secret) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		log.Error("clientauth: id or secret zero length")
		return
	}

	ccfg, exists := cauth.clients[id]
	if exists == false || secret != ccfg.Secret {
		w.WriteHeader(http.StatusUnauthorized)
		if exists == false {
			log.Errorf("clientauth: clientid not found (%s)", id)
		} else {
			log.Errorf("clientauth: client secret does not match (%s:%s)", id, secret)
		}
		return
	}

	cauth.next.ServeHTTP(w, r)
}
