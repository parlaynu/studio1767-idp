package clientstore

import (
	"github.com/parlaynu/studio1767-oidc-idp/internal/config"
)

type Client struct {
	Id           string
	Secret       string
	RedirectURLs []string
}

type ClientStore interface {
	Get(id string) *Client
}

func New(cfg *config.Config) ClientStore {
	cs := clientStore{
		clients: make(map[string]*Client),
	}

	for _, client := range cfg.Clients {
		cl := Client{
			Id:           client.Id,
			Secret:       client.Secret,
			RedirectURLs: client.RedirectURLs,
		}
		cs.clients[client.Id] = &cl
	}

	return &cs
}

type clientStore struct {
	clients map[string]*Client
}

func (cs *clientStore) Get(id string) *Client {
	return cs.clients[id]
}
