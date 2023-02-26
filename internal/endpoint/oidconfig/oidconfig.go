package oidconfig

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func New(issuerURL, authURL string) (http.Handler, error) {

	h := configHandler{
		Issuer:        issuerURL,
		AuthEndpoint:  authURL,
		TokenEndpoint: issuerURL + "/token",
		TokenEndpointAuthSupported: []string{
			"client_secret_basic",
		},
		JwksURI: issuerURL + "/keys",
		ScopesSupported: []string{
			"openid",
			"email",
			"profile",
		},
		ClaimsSupported: []string{
			"aud",
			"email",
			"email_verified",
			"exp",
			"iat",
			"iss",
			"locale",
			"name",
			"sub",
		},
		GrantTypesSupported: []string{
			"authorization_code",
		},
		ResponseTypesSupported: []string{
			"code",
		},
		IdTokenSigningAlgsSupported: []string{
			"RS256",
		},
		SubjectTypesSupported: []string{
			"public",
		},
	}

	jdata, err := json.MarshalIndent(&h, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("oidconfig: failed to marhshal configuration")
	}

	h.Serialized = string(jdata)

	return &h, nil
}

type configHandler struct {
	Issuer                      string   `json:"issuer"`
	AuthEndpoint                string   `json:"authorization_endpoint"`
	JwksURI                     string   `json:"jwks_uri"`
	TokenEndpoint               string   `json:"token_endpoint"`
	ClaimsSupported             []string `json:"claims_supported"`
	GrantTypesSupported         []string `json:"grant_types_supported"`
	IdTokenSigningAlgsSupported []string `json:"id_token_signing_alg_values_supported"`
	ResponseTypesSupported      []string `json:"response_types_supported"`
	ScopesSupported             []string `json:"scopes_supported"`
	SubjectTypesSupported       []string `json:"subject_types_supported"`
	TokenEndpointAuthSupported  []string `json:"token_endpoint_auth_methods_supported"`
	Serialized                  string   `json:"serialized,omitempty"`
}

func (ch *configHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, ch.Serialized)
}
