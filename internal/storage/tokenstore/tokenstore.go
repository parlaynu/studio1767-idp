package tokenstore

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/parlaynu/studio1767-oidc-idp/internal/storage/keystore"
)

type Token map[string]string

type TokenStore interface {
	Put(clientID string, tk Token) string
	Get(clientID string, code string) (Token, error)

	NewToken(ti *TokenInfo) (Token, error)
}

func New(issuerURL string, ks keystore.KeyStore) TokenStore {

	// create the structure
	ts := tokenStore{
		issuer: issuerURL,
		tokens: make(map[string]*storeData),
		kstore: ks,
	}

	return &ts
}

type storeData struct {
	stamp    time.Time
	clientID string
	token    Token
}

type tokenStore struct {
	mutex  sync.Mutex
	issuer string
	tokens map[string]*storeData
	kstore keystore.KeyStore
}

func (ts *tokenStore) Put(clientID string, tk Token) string {

	code := uuid.New().String()

	td := storeData{
		stamp:    time.Now(),
		clientID: clientID,
		token:    tk,
	}

	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.tokens[code] = &td

	return code
}

func (ts *tokenStore) Get(cliendID, code string) (Token, error) {

	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	td := ts.tokens[code]
	if td == nil {
		return nil, errors.New("tokenstore: token not found")
	}

	if td.clientID != cliendID {
		return nil, errors.New("tokenstore: incorrect client id provided")
	}

	delete(ts.tokens, code)

	return td.token, nil
}
