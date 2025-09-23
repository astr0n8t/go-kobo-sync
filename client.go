package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Hardcoded path lookup for security purposes
// The Kobo doesnt have these CA certs in its local store so these either
// have to exist or you have to allow an unsecure connection which is just
// bad practice
const caCertPath = "/mnt/onboard/.adds/go-kobo-sync/ca-certs"

func loadCaCertsFromPath(dir string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var block *pem.Block
			rest := data
			for {
				block, rest = pem.Decode(rest)
				if block == nil {
					break
				}
				if block.Type == "CERTIFICATE" {
					cert, err := x509.ParseCertificate(block.Bytes)
					if err != nil {
						return err
					}
					pool.AddCert(cert)
				}
			}

			return nil
		})

	if err != nil {
		return nil, err
	}
	return pool, nil
}

func GetClient() *http.Client {

	caCertPool, err := loadCaCertsFromPath(caCertPath)

	if err != nil {
		log.Fatalf("Unable to load CA Certificates. Error [%s]", err)
	}

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	return client
}
