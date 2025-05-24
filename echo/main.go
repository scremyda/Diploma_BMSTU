package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
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

func main() {
	certFile := "/app/certs/localhost.crt"
	keyFile := "/app/certs/localhost.key"

	var atomicHolder atomic.Value

	loadCert := func() {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Printf("Failed to load cert: %v", err)
			return
		}
		atomicHolder.Store(&cert)
		log.Println("Certificate loaded/updated")
	}

	loadCert()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("fsnotify.NewWatcher error: %v", err)
	}
	defer watcher.Close()

	for _, f := range []string{certFile, keyFile} {
		if err := watcher.Add(f); err != nil {
			log.Fatalf("watcher.Add(%s) error: %v", f, err)
		}
	}

	go func() {
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					loadCert()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("fsnotify error: %v", err)
			}
		}
	}()

	// TLS config that uses GetCertificate to always fetch the latest cert
	tlsConfig := &tls.Config{
		GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			certPtr := atomicHolder.Load().(*tls.Certificate)
			return certPtr, nil
		},
	}

	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
	}

	http.HandleFunc("/", echoHandler)

	log.Println("Starting echo server at https://localhost:8443")
	// Empty strings because certs are provided via GetCertificate
	if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
