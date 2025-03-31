package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	fmt.Fprintf(w, "Received: %s", body)
}

func certificateReloader(certFile, keyFile string) func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var mu sync.Mutex
	var currentCert *tls.Certificate

	loadCert := func() error {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
		mu.Lock()
		currentCert = &cert
		mu.Unlock()
		log.Println("Certificate reloaded successfully")
		return nil
	}

	if err := loadCert(); err != nil {
		log.Fatalf("Error loading initial certificate: %v", err)
	}

	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		mu.Lock()
		defer mu.Unlock()
		return currentCert, nil
	}
}

func main() {
	time.Sleep(60 * time.Second)
	certFile := "/app/certs/example.com/cert.pem"
	keyFile := "/app/certs/example.com/key.pem"

	tlsConfig := &tls.Config{
		GetCertificate: certificateReloader(certFile, keyFile),
	}

	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
	}

	http.HandleFunc("/", echoHandler)

	log.Println("Starting echo server at https://localhost:8443")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
