package auth

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/securecookie"
)

type cookieCutter struct {
	name     string
	sc       *securecookie.SecureCookie
	provider *provider
}

var (
	ErrCookieNotFound = fmt.Errorf("cookie not found")
)

func newCookieCutter(hKey, eKey []byte, p *provider) *cookieCutter {
	sc := securecookie.New(hKey, eKey)

	cc := cookieCutter{
		name:     "auth-id",
		sc:       sc,
		provider: p,
	}

	return &cc
}

func (cc *cookieCutter) addCookie(w http.ResponseWriter, rawToken string) error {

	encoded, err := cc.sc.Encode(cc.name, rawToken)
	if err != nil {
		return fmt.Errorf("auth/cookie: failed to encode cookie: %w", err)
	}

	// add the auth cookie
	idCookie := http.Cookie{
		Name:     cc.name,
		Value:    encoded,
		Path:     "/",
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &idCookie)

	return nil
}

func (cc *cookieCutter) getCookie(r *http.Request) (*jwt.Token, error) {
	// get the cookie
	cookie, err := r.Cookie(cc.name)
	if cookie == nil {
		return nil, ErrCookieNotFound
	}

	// get the cookie value
	var value string
	err = cc.sc.Decode(cc.name, cookie.Value, &value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cookie: %w", err)
	}

	// parse the token
	return cc.provider.parseToken(value)
}

func (cc *cookieCutter) clearCookie(w http.ResponseWriter) {
	idCookie := http.Cookie{
		Name:     cc.name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &idCookie)
}
