package mtls

import (
	"context"
	"crypto/x509"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func New(next http.Handler) http.Handler {
	mm := mtlsMware{
		next: next,
	}
	return &mm
}

type MTLSKey struct{}

type MTLSInfo struct {
	CommonName   string
	Organization string
	Email        string
}

type mtlsMware struct {
	next http.Handler
}

func (mm *mtlsMware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.TLS != nil {
		for _, pc := range r.TLS.PeerCertificates {
			mi := mm.parseCertificate(pc)
			if mi != nil {
				ctx := context.WithValue(r.Context(), MTLSKey{}, mi)
				r = r.WithContext(ctx)
				break
			}
		}
	}

	mm.next.ServeHTTP(w, r)
}

func (mm *mtlsMware) parseCertificate(cert *x509.Certificate) *MTLSInfo {
	var mi MTLSInfo

	// extrace certificate information
	mi.CommonName = cert.Subject.CommonName
	for _, o := range cert.Subject.Organization {
		mi.Organization = o
		break
	}
	if mi.Organization == "" {
		log.Error("mtls: required field 'organization' missing")
		return nil
	}

	// process the uris
	for _, uri := range cert.URIs {
		if uri.Scheme == "email" {
			mi.Email = uri.Opaque
		}
	}
	if mi.Email == "" {
		log.Error("mtls: required uri 'email' missing")
		return nil
	}

	// we have a matching certificate
	return &mi
}
