package authcommon

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"s1767.xyz/idp/internal/endpoint/utils"
	"s1767.xyz/idp/internal/storage/clientstore"
	"s1767.xyz/idp/internal/storage/tokenstore"
	"s1767.xyz/idp/internal/storage/userdb"
)

type Authenticator interface {
	Authenticate(w http.ResponseWriter, r *http.Request, user *userdb.User)
}

func New(cs clientstore.ClientStore, ts tokenstore.TokenStore) Authenticator {
	ah := &authenticator{
		cstore: cs,
		tstore: ts,
	}
	return ah
}

type authenticator struct {
	cstore clientstore.ClientStore
	tstore tokenstore.TokenStore
}

func (au *authenticator) Authenticate(w http.ResponseWriter, r *http.Request, user *userdb.User) {
	// check for required paramaters
	required := []string{
		"client_id",
		"scope",
		"redirect_uri",
		"nonce",
		"state",
		"response_type",
	}
	if utils.CheckParameters(r, required) == false {
		w.WriteHeader(http.StatusBadRequest)
		log.Error("authcommon: missing parameters in request")
		return
	}

	// extract form parameters
	client_id := r.FormValue("client_id")
	redirect_url := r.FormValue("redirect_uri")
	nonce := r.FormValue("nonce")
	state := r.FormValue("state")
	response_type := r.FormValue("response_type")

	scopes := make(map[string]bool)
	for _, s := range strings.Split(r.FormValue("scope"), " ") {
		scopes[s] = true
	}

	// verify the client and redirect URL
	client := au.cstore.Get(client_id)
	if client == nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("authcommon: no client with id %s", client_id)
		return
	}
	found := false
	for _, url := range client.RedirectURLs {
		if redirect_url == url {
			found = true
			break
		}
	}
	if found == false {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("no matching redirect url for client %s:%s", client_id, redirect_url)
		return
	}

	// create the token
	ti := tokenstore.TokenInfo{
		User:         user,
		ClientID:     client_id,
		Scopes:       scopes,
		RedirectURL:  redirect_url,
		Nonce:        nonce,
		State:        state,
		ResponseType: response_type,
	}

	token, err := au.tstore.NewToken(&ti)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("failed to create token: %v", err)
		return
	}

	code := au.tstore.Put(ti.ClientID, token)

	url := ti.RedirectURL
	url += "?state=" + ti.State
	url += "&code=" + code

	http.Redirect(w, r, url, http.StatusSeeOther)
}
