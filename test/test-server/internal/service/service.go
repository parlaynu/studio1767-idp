package service

import (
	"fmt"
	"net/http"

	"github.com/parlaynu/studio1767-idp-test/api"
)

func New() api.Service {
	return &service{}
}

type service struct{}

func (s *service) Hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello, %s\n", r.URL.Path)
}
