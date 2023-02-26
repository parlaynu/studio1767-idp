package service

import (
	"fmt"
	"net/http"
	"path/filepath"

	"s1767.xyz/idp/api"
	"s1767.xyz/idp/internal/config"
	"s1767.xyz/idp/internal/endpoint/authbasic"
	"s1767.xyz/idp/internal/endpoint/authcommon"
	"s1767.xyz/idp/internal/endpoint/authmtls"
	"s1767.xyz/idp/internal/endpoint/keys"
	"s1767.xyz/idp/internal/endpoint/oidconfig"
	"s1767.xyz/idp/internal/endpoint/token"
	"s1767.xyz/idp/internal/middleware/mtls"
	"s1767.xyz/idp/internal/storage/clientstore"
	"s1767.xyz/idp/internal/storage/keystore"
	"s1767.xyz/idp/internal/storage/tokenstore"
	"s1767.xyz/idp/internal/storage/userdbyaml"
)

func New(cfg *config.Config) (api.Service, error) {

	// create the stores
	dbpath := cfg.UserDb.Path
	if !filepath.IsAbs(dbpath) {
		dbpath = filepath.Join(filepath.Dir(cfg.ConfigFile), cfg.UserDb.Path)
	}
	userdb, err := userdbyaml.NewUserDb(dbpath)
	if err != nil {
		return nil, fmt.Errorf("failed to create user db: %w", err)
	}
	cstore := clientstore.New(cfg)
	kstore, err := keystore.New(5)
	if err != nil {
		return nil, fmt.Errorf("failed to create keystore: %w", err)
	}
	tstore := tokenstore.New(cfg.IssuerURL, kstore)

	// create the endpoint handlers
	cauth := authcommon.New(cstore, tstore, userdb)
	bauth, err := authbasic.New(cauth, userdb, cfg.ContentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create basic auth handler: %w", err)
	}
	mauth := authmtls.New(cauth)
	oconfig, err := oidconfig.New(cfg.IssuerURL, cfg.AuthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create oidc config handler: %w", err)
	}
	khandler := keys.New(kstore)
	thandler := token.New(cstore, tstore)

	// create the service
	svc := service{
		authBasic: bauth,
		authMtls:  mauth,
		oidConfig: oconfig,
		keys:      khandler,
		token:     thandler,
	}
	return &svc, nil
}

type service struct {
	authBasic http.Handler
	authMtls  http.Handler
	oidConfig http.Handler
	keys      http.Handler
	token     http.Handler
}

func (s *service) OIDCConfiguration(w http.ResponseWriter, r *http.Request) {
	s.oidConfig.ServeHTTP(w, r)
}

func (s *service) Keys(w http.ResponseWriter, r *http.Request) {
	s.keys.ServeHTTP(w, r)
}

func (s *service) Tokens(w http.ResponseWriter, r *http.Request) {
	s.token.ServeHTTP(w, r)
}

func (s *service) AuthnStart(w http.ResponseWriter, r *http.Request) {
	if r.Context().Value(mtls.MTLSKey{}) != nil {
		s.authMtls.ServeHTTP(w, r)
	} else {
		s.authBasic.ServeHTTP(w, r)
	}
}

func (s *service) AuthnVerify(w http.ResponseWriter, r *http.Request) {
	s.authBasic.ServeHTTP(w, r)
}
