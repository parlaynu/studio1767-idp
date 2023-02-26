package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	mrand "math/rand"

	"github.com/google/uuid"
)

type KeyStore interface {
	GetPublicKeys() map[string]*rsa.PublicKey
	GetPrivateKey() (string, *rsa.PrivateKey)
}

func New(nkeys int) (KeyStore, error) {

	ks := keyStore{
		nkeys: nkeys,
		keys:  make(map[string]*rsa.PrivateKey),
	}
	err := ks.createKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to create keys: %w", err)
	}

	return &ks, nil
}

type keyStore struct {
	nkeys int
	keys  map[string]*rsa.PrivateKey
	kids  []string
}

func (ks *keyStore) GetPublicKeys() map[string]*rsa.PublicKey {

	pubkeys := make(map[string]*rsa.PublicKey)
	for kid, key := range ks.keys {
		pubkeys[kid] = key.Public().(*rsa.PublicKey)
	}

	return pubkeys
}

func (ks *keyStore) GetPrivateKey() (string, *rsa.PrivateKey) {

	// choose a key at random
	kid := ks.kids[mrand.Intn(ks.nkeys)]
	key := ks.keys[kid]

	return kid, key
}

func (ks *keyStore) createKeys() error {

	// create the encryption keys
	keys := make(map[string]*rsa.PrivateKey)
	kids := make([]string, ks.nkeys)

	for i := 0; i < ks.nkeys; i++ {
		kid := uuid.New().String()

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}

		keys[kid] = key
		kids[i] = kid
	}

	// success, so store in the keystore
	ks.keys = keys
	ks.kids = kids

	return nil
}
