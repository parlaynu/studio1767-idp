package service

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/parlaynu/studio1767-idp/api"
	"github.com/parlaynu/studio1767-idp/internal/config"
	"github.com/parlaynu/studio1767-idp/internal/endpoint/authbasic"
	"github.com/parlaynu/studio1767-idp/internal/endpoint/authcommon"
	"github.com/parlaynu/studio1767-idp/internal/endpoint/authmtls"
	"github.com/parlaynu/studio1767-idp/internal/endpoint/keys"
	"github.com/parlaynu/studio1767-idp/internal/endpoint/oidconfig"
	"github.com/parlaynu/studio1767-idp/internal/endpoint/token"
	"github.com/parlaynu/studio1767-idp/internal/middleware/mtls"
	"github.com/parlaynu/studio1767-idp/internal/storage/clientstore"
	"github.com/parlaynu/studio1767-idp/internal/storage/keystore"
	"github.com/parlaynu/studio1767-idp/internal/storage/tokenstore"
	"github.com/parlaynu/studio1767-idp/internal/storage/userdb"
	"github.com/parlaynu/studio1767-idp/internal/storage/userdbldap"
	"github.com/parlaynu/studio1767-idp/internal/storage/userdbyaml"
)

func New(cfg *config.Config) (api.Service, error) {

	var err error

	// create the stores
	var udb userdb.UserDb
	switch cfg.UserDb.Type {
	case "ldap":
		ldapServer, ldapPort := cfg.UserDb.LdapServer, cfg.UserDb.LdapPort
		sBase, sDn, sPw := cfg.UserDb.SearchBase, cfg.UserDb.SearchDn, cfg.UserDb.SearchPw
		udb, err = userdbldap.NewUserDb(ldapServer, ldapPort, sBase, sDn, sPw, cfg.Https.CaCertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create user db: %w", err)
		}
	case "yaml":
		dbpath := cfg.UserDb.Path
		if !filepath.IsAbs(dbpath) {
			dbpath = filepath.Join(filepath.Dir(cfg.ConfigFile), cfg.UserDb.Path)
		}
		udb, err = userdbyaml.NewUserDb(dbpath)
		if err != nil {
			return nil, fmt.Errorf("failed to create user db: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown userdb type: %s", cfg.UserDb.Type)
	}

	cstore := clientstore.New(cfg)
	kstore, err := keystore.New(5)
	if err != nil {
		return nil, fmt.Errorf("failed to create keystore: %w", err)
	}
	tstore := tokenstore.New(cfg.IssuerURL, kstore)

	// create the endpoint handlers
	cauth := authcommon.New(cstore, tstore)
	bauth, err := authbasic.New(cauth, udb, cfg.ContentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create basic auth handler: %w", err)
	}
	mauth := authmtls.New(cauth, udb)
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
