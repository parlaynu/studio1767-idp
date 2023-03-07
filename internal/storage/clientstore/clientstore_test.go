package clientstore_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/parlaynu/studio1767-oidc-idp/internal/config"
	"github.com/parlaynu/studio1767-oidc-idp/internal/storage/clientstore"
)

func TestClientStore(t *testing.T) {
	cfg := createConfig()
	require.NotNil(t, cfg)

	cs := clientstore.New(cfg)

	for _, cfgcl := range cfg.Clients {
		cl := cs.Get(cfgcl.Id)
		require.Equal(t, cfgcl.Secret, cl.Secret)
		require.Equal(t, cfgcl.RedirectURLs[0], cl.RedirectURLs[0])
	}
}

func createConfig() *config.Config {
	clients := []config.ClientConfig{}
	nclients := 5
	for i := 0; i < nclients; i++ {
		client := config.ClientConfig{
			Id:           fmt.Sprintf("Id%d", i),
			Secret:       fmt.Sprintf("Secret%d", i),
			RedirectURLs: []string{fmt.Sprintf("http://server%d.example.com", i)},
		}
		clients = append(clients, client)
	}

	cfg := config.Config{
		Clients: clients,
	}

	return &cfg
}
