package tokenstore

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"

	"github.com/parlaynu/studio1767-oidc-idp/internal/storage/userdb"
)

type TokenInfo struct {
	User         *userdb.User
	ClientID     string
	Scopes       map[string]bool
	RedirectURL  string
	State        string
	Nonce        string
	ResponseType string
}

func (ts *tokenStore) NewToken(ti *TokenInfo) (Token, error) {

	// get the times
	now := time.Now()
	duration := time.Hour * 24
	exp := now.Add(duration)
	now = now.Add(time.Second * -5)

	// an event id
	event_id := uuid.New().String()

	// the oatoken
	token := make(Token)
	token["token_type"] = "Bearer"
	token["expires_in"] = strconv.Itoa(int(duration.Seconds()) - 1)

	// create the access token
	atoken, err := ts.accessToken(ti, now, exp, event_id)
	if err != nil {
		return nil, err
	}
	token["access_token"] = atoken

	// create the idtoken
	if ti.Scopes["openid"] {
		idtoken, err := ts.openidToken(ti, now, exp, event_id, atoken)
		if err != nil {
			return nil, err
		}

		token["id_token"] = idtoken
	}

	return token, nil
}

func (ts *tokenStore) accessToken(ti *TokenInfo, now, exp time.Time, event_id string) (string, error) {

	// create the claims
	claims := make(jwt.MapClaims)

	subject := base64.RawURLEncoding.EncodeToString([]byte(ti.User.Email))

	claims["token_use"] = "access"
	claims["event_id"] = event_id
	claims["iss"] = ts.issuer
	claims["sub"] = subject
	claims["aud"] = ti.ClientID
	claims["exp"] = exp.Unix()
	claims["iat"] = now.Unix()

	scopes := make([]string, 0, len(ti.Scopes))
	for scope := range ti.Scopes {
		scopes = append(scopes, scope)
	}
	claims["scope"] = strings.Join(scopes, " ")

	if ti.Scopes["profile"] {
		claims["name"] = ti.User.FullName
		claims["given_name"] = ti.User.GivenName
		claims["family_name"] = ti.User.FamilyName
		claims["username"] = ti.User.Name
		claims["groups"] = ti.User.Groups
	}

	if ti.Scopes["email"] {
		claims["email"] = ti.User.Email
		claims["email_verified"] = true
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	kid, key := ts.kstore.GetPrivateKey()
	token.Header["kid"] = kid

	ss, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return ss, nil
}

func (ts *tokenStore) openidToken(ti *TokenInfo, now, exp time.Time, event_id, atoken string) (string, error) {
	// create the claims
	claims := make(jwt.MapClaims)

	subject := base64.RawURLEncoding.EncodeToString([]byte(ti.User.Email))

	claims["token_use"] = "id"
	claims["event_id"] = event_id
	claims["iss"] = ts.issuer
	claims["sub"] = subject
	claims["aud"] = ti.ClientID
	claims["exp"] = exp.Unix()
	claims["iat"] = now.Unix()

	if ti.Nonce != "" {
		claims["nonce"] = ti.Nonce
	}

	athash := sha256.Sum256([]byte(atoken))
	claims["at_hash"] = base64.RawURLEncoding.EncodeToString(athash[:16])

	claims["given_name"] = ti.User.GivenName
	claims["family_name"] = ti.User.FamilyName
	claims["username"] = ti.User.Name
	claims["email"] = ti.User.Email
	claims["email_verified"] = true

	claims["groups"] = ti.User.Groups

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	kid, key := ts.kstore.GetPrivateKey()
	token.Header["kid"] = kid

	ss, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign openid token: %w", err)
	}

	return ss, nil
}
