package authmtls

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"s1767.xyz/idp/internal/endpoint/authcommon"
	"s1767.xyz/idp/internal/middleware/mtls"
	"s1767.xyz/idp/internal/storage/userdb"
)

func New(au authcommon.Authenticator, udb userdb.UserDb) http.Handler {
	ah := &authMtls{
		auth:   au,
		userdb: udb,
	}
	return ah
}

type authMtls struct {
	auth   authcommon.Authenticator
	userdb userdb.UserDb
}

func (am *authMtls) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(mtls.MTLSKey{})
	mi := val.(*mtls.MTLSInfo)

	// lookup the user - we don't have a password, but we have a verified
	//   certificate for the user, so use the lookup interface
	user, err := am.userdb.LookupUser(mi.CommonName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Errorf("authcommon: login failed: %v", err)
		return
	}

	// verify the email
	if mi.Email != user.Email {
		w.WriteHeader(http.StatusUnauthorized)
		log.Errorf("authcommon: email doesn't match: %s -> %s", user.Email, mi.Email)
		return
	}

	am.auth.Authenticate(w, r, user)
}
