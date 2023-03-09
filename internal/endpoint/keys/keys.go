package keys

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/parlaynu/studio1767-idp/internal/storage/keystore"
)

func New(ks keystore.KeyStore) http.Handler {
	h := keysHandler{
		kStore: ks,
	}
	return &h
}

type keysHandler struct {
	kStore keystore.KeyStore
}

func (kh *keysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	keys, err := kh.getPublicKeys()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("keys: failed to get public keys: %v", err)
		return
	}

	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, keys)
}

type pubkey struct {
	Use string `json:"use"`
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (kh *keysHandler) getPublicKeys() (string, error) {

	// get the public keys and convert into the Oauth2 format
	keys := kh.kStore.GetPublicKeys()

	var pubkeys []*pubkey
	for kid, key := range keys {
		N := base64.RawURLEncoding.EncodeToString(key.N.Bytes())

		ebig := big.NewInt(int64(key.E))
		E := base64.RawURLEncoding.EncodeToString(ebig.Bytes())

		pk := pubkey{
			Use: "sig",
			Kty: "RSA",
			Kid: kid,
			Alg: "RS256",
			N:   N,
			E:   E,
		}

		pubkeys = append(pubkeys, &pk)
	}

	// serialize to the format we need
	pubmap := make(map[string][]*pubkey)
	pubmap["keys"] = pubkeys

	jdata, err := json.MarshalIndent(&pubmap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("keys: failed to marshal key data")
	}

	return string(jdata), nil
}
