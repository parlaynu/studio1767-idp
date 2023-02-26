package authmtls

import (
	"net/http"

	"s1767.xyz/idp/internal/endpoint/authcommon"
	"s1767.xyz/idp/internal/middleware/mtls"
)

func New(au authcommon.Authenticator) http.Handler {
	ah := &authMtls{
		auth: au,
	}
	return ah
}

type authMtls struct {
	auth authcommon.Authenticator
}

func (am *authMtls) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(mtls.MTLSKey{})
	mi := val.(*mtls.MTLSInfo)

	am.auth.Authenticate(w, r, mi.CommonName, mi.Email)
}
