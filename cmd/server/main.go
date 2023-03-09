package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/parlaynu/studio1767-idp/api"
	"github.com/parlaynu/studio1767-idp/internal/config"
	"github.com/parlaynu/studio1767-idp/internal/service"
)

func main() {
	// setup logging
	log.SetLevel(log.TraceLevel)
	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)

	// parse command line
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("Usage: %s <config-file>", filepath.Base(os.Args[0]))
	}
	cfgFile := flag.Arg(0)

	// load the configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	// create the service and API
	//   the api package defines a service interface that needs to be implemented
	//   the internals/service package implements it
	svc, err := service.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	fe, be := api.New(cfg, svc)

	// run the server
	//   if the listener scheme is http, run a http server, otherwise, https
	cafile, certfile, keyfile := cfg.Https.CaCertFile, cfg.Https.CertFile, cfg.Https.KeyFile
	if strings.HasPrefix(cfg.Listeners.Frontend, "http://") {
		go func() {
			RunHTTP(cfg.Listeners.Frontend, fe)
		}()
	} else {
		go func() {
			RunHTTPS(cfg.Listeners.Frontend, fe, cafile, certfile, keyfile)
		}()
	}
	if strings.HasPrefix(cfg.Listeners.Backend, "http://") {
		RunHTTP(cfg.Listeners.Backend, be)
	} else {
		RunHTTPS(cfg.Listeners.Backend, be, cafile, certfile, keyfile)
	}
}

func RunHTTP(listener string, handler http.Handler) {
	address := strings.TrimPrefix(listener, "http://")
	srv := &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Infof("Server listening at %s\n", listener)
	log.Fatal(srv.ListenAndServe())
}

func RunHTTPS(listener string, handler http.Handler, cafile, certfile, keyfile string) {

	// create a certificate pool with the ca certificate
	cacert, err := os.ReadFile(cafile)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cacert)

	// create the TLS config and enable client authentication
	// note: using tls 1.3 so all default ciphers etc. are secure
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequestClientCert,
	}

	address := strings.TrimPrefix(listener, "https://")
	srv := &http.Server{
		Addr:         address,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Infof("Server listening at %s\n", listener)
	log.Fatal(srv.ListenAndServeTLS(certfile, keyfile))
}
