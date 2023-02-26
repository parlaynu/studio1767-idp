package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"s1767.xyz/test/api"
	"s1767.xyz/test/internal/config"
	"s1767.xyz/test/internal/service"
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

	// create the service and api
	svc := service.New()
	handler, err := api.New(cfg, svc)
	if err != nil {
		log.Fatal(err)
	}

	// run the server
	cafile, certfile, keyfile := cfg.Https.CaCertFile, cfg.Https.CertFile, cfg.Https.KeyFile
	if strings.HasPrefix(cfg.Listener, "http://") {
		RunHTTP(cfg.Listener, handler)
	} else {
		RunHTTPS(cfg.Listener, handler, cafile, certfile, keyfile)
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
	fmt.Printf("Server listening at %s\n", listener)
	log.Fatal(srv.ListenAndServe())
}

func RunHTTPS(listener string, handler http.Handler, cafile, certfile, keyfile string) {
	// create a certificate pool with the client ca
	cacert, err := os.ReadFile(cafile)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cacert)

	// Create the TLS Config with the CA pool and enable Client certificate validation

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		MinVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
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

	fmt.Printf("Server listening at %s\n", listener)
	log.Fatal(srv.ListenAndServeTLS(certfile, keyfile))
}