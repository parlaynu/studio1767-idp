package token

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/parlaynu/studio1767-oidc-idp/internal/endpoint/utils"
	"github.com/parlaynu/studio1767-oidc-idp/internal/storage/clientstore"
	"github.com/parlaynu/studio1767-oidc-idp/internal/storage/tokenstore"
)

func New(cs clientstore.ClientStore, ts tokenstore.TokenStore) http.Handler {
	h := tokenHandler{
		clStore: cs,
		tkStore: ts,
	}
	return &h
}

type tokenHandler struct {
	clStore clientstore.ClientStore
	tkStore tokenstore.TokenStore
}

func (th *tokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check for required paramaters
	required := []string{
		"grant_type",
		"client_id",
		"client_secret",
		"redirect_uri",
		"code",
	}
	if utils.CheckParameters(r, required) == false {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("token: missing one or more required parameters")
		return
	}

	// verify the grant type
	grant := r.FormValue("grant_type")
	if grant != "authorization_code" {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("token: unsupport grant type: %s", grant)
		return
	}

	// get the client id
	clientID := r.FormValue("client_id")
	client := th.clStore.Get(clientID)
	if client == nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("token: client id not in store: %s", clientID)
		return
	}

	// verify the client secret
	clientSecret := r.FormValue("client_secret")
	if clientSecret != client.Secret {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("token: client secret mismatch for client %s", clientID)
		return
	}

	// verify the redirect url
	redirectURL := r.FormValue("redirect_uri")

	found := false
	for _, rURL := range client.RedirectURLs {
		if redirectURL == rURL {
			found = true
			break
		}
	}
	if found == false {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("token: invalid redirect url")
		return
	}

	// get the code and token
	code := r.FormValue("code")

	token, err := th.tkStore.Get(clientID, code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("token: no token for client: %v", err)
		return
	}

	// all is good... return the token
	w.Header().Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	var response string
	for k, v := range token {
		if response != "" {
			response += "&"
		}
		response += k
		response += "="
		response += v
	}
	fmt.Fprint(w, response)
}
