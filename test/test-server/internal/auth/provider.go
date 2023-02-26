package auth

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"

	"s1767.xyz/test/internal/config"
)

type provider struct {
	client   *http.Client
	pvConfig *providerConfig
	oaConfig *oauth2.Config
	keys     map[string]*jwKey
}

type providerConfig struct {
	Issuer              string   `json:"issuer"`
	AuthURL             string   `json:"authorization_endpoint"`
	JwksURL             string   `json:"jwks_uri"`
	TokenURL            string   `json:"token_endpoint"`
	TokenURLAuthMethods []string `json:"token_endpoint_auth_methods_supported"`
	Scopes              []string `json:"scopes_supported"`
	SubjectTypes        []string `json:"subject_types_supported"`
	ResponseTypes       []string `json:"response_types_supported"`
	GrantTypes          []string `json:"grant_types_supported"`
	Claims              []string `json:"claims_supported"`
	SigningAlgs         []string `json:"id_token_signing_alg_values_supported"`
}

type jwKey struct {
	Kid    string
	Kty    string
	Alg    string
	E      string
	N      string
	Use    string
	Public *rsa.PublicKey
}

func newProvider(cfg *config.Config) (*provider, error) {

	// build the http client
	client, err := newHttpClient(cfg.Idp.CaCertFile)
	if err != nil {
		return nil, fmt.Errorf("auth/provider: failed to create client: %w", err)
	}

	// get the config from the provider
	configURL := cfg.Idp.IssuerURL + "/.well-known/openid-configuration"

	params := url.Values{}
	params.Add("client_id", cfg.Service.Id)
	params.Add("client_secret", cfg.Service.Secret)

	configURL += "?" + params.Encode()

	pvConfig, err := loadConfig(client, configURL)
	if err != nil {
		return nil, fmt.Errorf("failed loading configuration: %w", err)
	}

	// load the keys
	keyURL := pvConfig.JwksURL + "?" + params.Encode()

	keys, err := loadKeys(client, keyURL)
	if err != nil {
		return nil, fmt.Errorf("failed loading keys: %w", err)
	}

	// create the oauth2 config
	ep := oauth2.Endpoint{
		AuthURL:   pvConfig.AuthURL,
		TokenURL:  pvConfig.TokenURL,
		AuthStyle: oauth2.AuthStyleInParams,
	}
	oaConfig := oauth2.Config{
		ClientID:     cfg.Service.Id,
		ClientSecret: cfg.Service.Secret,
		RedirectURL:  cfg.Service.RedirectURLs[0],
		Endpoint:     ep,
		Scopes:       []string{"openid", "profile", "email"}, // removed "groups" ... might be included in "profile"
	}

	// create the provider
	p := &provider{
		client:   client,
		pvConfig: pvConfig,
		oaConfig: &oaConfig,
		keys:     keys,
	}

	return p, nil
}

func (p *provider) authCodeURL(state, nonce string) string {
	opts := oauth2.SetAuthURLParam("nonce", nonce)
	return p.oaConfig.AuthCodeURL(state, opts)
}

func (p *provider) exchange(ctx context.Context, code string) (*jwt.Token, error) {
	// create a contet with the correct client to use
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client)

	// exchange the code for the oa token
	oaToken, err := p.oaConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	// get the raw id token
	rawIdToken, ok := oaToken.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("provider: missing id token in response")
	}

	// convert into id token
	return p.parseToken(rawIdToken)
}

func (p *provider) parseToken(rawToken string) (*jwt.Token, error) {
	// parse and verify the token
	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("provider: unsupported signing method")
		}

		keyid := token.Header["kid"].(string)
		key := p.getKey(keyid)
		if key == nil {
			return nil, fmt.Errorf("provider: missing key: %s", keyid)
		}
		return key.Public, nil
	})
	if err != nil {
		return nil, err
	}

	if token.Valid == false {
		return nil, fmt.Errorf("provider: token is not value")
	}

	return token, nil
}

func (p *provider) getKey(kid string) *jwKey {
	return p.keys[kid]
}

func loadConfig(client *http.Client, url string) (*providerConfig, error) {
	cresp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("provider: failed to get configuration: %w", err)
	}
	defer cresp.Body.Close()

	if cresp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider: error getting configuration: %s", cresp.Status)
	}

	cbody, err := io.ReadAll(cresp.Body)
	if err != nil {
		return nil, fmt.Errorf("provider: failed to read configuration: %w", err)
	}

	// decode the json
	var pvConfig providerConfig
	err = json.Unmarshal(cbody, &pvConfig)
	if err != nil {
		return nil, fmt.Errorf("provider: failed to unmarshal configuration: %w", err)
	}

	return &pvConfig, nil
}

func loadKeys(client *http.Client, url string) (map[string]*jwKey, error) {
	// get the keys
	kres, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("provider: failed to get keys: %w", err)
	}
	defer kres.Body.Close()

	if kres.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider: error getting keys: %s", kres.Status)
	}

	kbody, err := io.ReadAll(kres.Body)
	if err != nil {
		return nil, fmt.Errorf("provider: failed to read keys: %w", err)
	}

	// unmarshall the json
	kdata := make(map[string][]jwKey)
	err = json.Unmarshal(kbody, &kdata)
	if err != nil {
		return nil, fmt.Errorf("provider: failed to unmarshal keys: %w", err)
	}
	klist := kdata["keys"]

	// convert to map and create the public key
	keys := make(map[string]*jwKey)
	for i, _ := range klist {
		k := &klist[i]

		if k.Alg != "RS256" {
			return nil, fmt.Errorf("provider: unsupported key type: %s", k.Alg)
		}

		// create the RSA public key
		nbytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			return nil, fmt.Errorf("provider: failed to decode N: %w", err)
		}
		n := big.NewInt(0)
		n.SetBytes(nbytes)

		ebytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			return nil, fmt.Errorf("provider: failed to decode E: %w", err)
		}

		var buffer bytes.Buffer
		buffer.WriteByte(0)
		buffer.Write(ebytes)
		e := binary.BigEndian.Uint32(buffer.Bytes())

		k.Public = &rsa.PublicKey{
			N: n,
			E: int(e),
		}

		keys[k.Kid] = k
	}

	return keys, nil
}
