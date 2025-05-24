package main

import (
	"crypto/tls"
	"html/template"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

var pageTpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <title>Добро пожаловать</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link
    href="https://cdn.jsdelivr.net/npm/bootstrap@5.4.3/dist/css/bootstrap.min.css"
    rel="stylesheet"
    integrity="sha384-..."
    crossorigin="anonymous">
  <style>
    body {
      background: linear-gradient(135deg, #667eea, #764ba2);
      color: white;
      height: 100vh;
      margin: 0;
    }
    .hero {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 100%;
      text-align: center;
    }
    h1 {
      font-size: 4rem;
      margin: 0;
    }
  </style>
</head>
<body>
  <div class="hero">
    <h1>Добро пожаловать!</h1>
  </div>
</body>
</html>
`))

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

	tlsConfig := &tls.Config{
		GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return atomicHolder.Load().(*tls.Certificate), nil
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := pageTpl.Execute(w, nil); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
		}
	})

	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
	}

	log.Println("Starting server at https://localhost:8443")
	if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
