package tokenstore_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/parlaynu/studio1767-oidc-idp/internal/storage/keystore"
	"github.com/parlaynu/studio1767-oidc-idp/internal/storage/tokenstore"
	"github.com/parlaynu/studio1767-oidc-idp/internal/storage/userdb"
)

func TestTokenStore(t *testing.T) {
	ks, err := keystore.New(5)
	require.NoError(t, err)

	ts := tokenstore.New("https://issuer.example,com", ks)

	type TokenInfo struct {
		User         *userdb.User
		ClientID     string
		Scopes       map[string]bool
		RedirectURL  string
		State        string
		Nonce        string
		ResponseType string
	}

	user := userdb.User{
		UserId:     1001,
		GroupId:    1001,
		UserName:   "tokenuser",
		Password:   "tokenpass",
		FullName:   "token user",
		GivenName:  "token",
		FamilyName: "user",
		Email:      "token@example.com",
		Groups:     []string{"tokengroup1"},
	}

	scopes := make(map[string]bool)
	scopes["openid"] = true
	scopes["profile"] = true
	scopes["email"] = true

	ti := tokenstore.TokenInfo{
		User:         &user,
		ClientID:     "clientid",
		Scopes:       scopes,
		RedirectURL:  "https://redirect.example.com",
		State:        "kasdfkajsf;",
		Nonce:        "kajfd;ajsf",
		ResponseType: "code",
	}

	token, err := ts.NewToken(&ti)
	require.NoError(t, err)

	code := ts.Put(ti.ClientID, token)
	token2, err := ts.Get(ti.ClientID, code)
	require.NoError(t, err)
	require.Equal(t, token, token2)
}
