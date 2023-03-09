package keystore_test

import (
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/parlaynu/studio1767-idp/internal/storage/keystore"
)

func TestKeyStore(t *testing.T) {
	ks, err := keystore.New(5)
	require.NoError(t, err)

	pubkeys := ks.GetPublicKeys()
	for i := 0; i < 10; i++ {
		kid, key := ks.GetPrivateKey()
		pubkey := key.Public().(*rsa.PublicKey)

		require.Equal(t, pubkeys[kid], pubkey)
	}
}
