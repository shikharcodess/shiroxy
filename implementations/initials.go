package implementations

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"net/http/httputil"
)

func RProxy() {
	// Define the reverse proxy for domain1
	domainProxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "13.235.77.167:8000", // The actual address where domain1's server is running
	})

	// domainProxy := httputil.ReverseProxy{
	// 	Rewrite: func(r *httputil.ProxyRequest) {
	// 		r.SetURL(target)
	// 		r.Out.Host = r.In.Host // if desired
	// 	},
	// }

	// Define a handler that dispatches requests to the correct reverse proxy
	handler := func(w http.ResponseWriter, r *http.Request) {
		domainProxy.ServeHTTP(w, r)
	}

	// Create a server
	server := &http.Server{
		Addr:    ":443", // Listen on the standard HTTPS port
		Handler: http.HandlerFunc(handler),
		TLSConfig: &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				var cert tls.Certificate
				var err error

				var serverName string = strings.TrimSpace(info.ServerName)
				// Dynamically load the appropriate certificate based on the domain
				// Note: In production, you should cache the certificates to avoid loading them for each request
				if serverName == "shikharcode.in" {
					cert, err = tls.LoadX509KeyPair("/home/shikharcode/Main/public_project/shiroxy/live/shikharcode.in/fullchain.pem", "/home/shikharcode/Main/public_project/shiroxy/live/shikharcode.in/privkey.pem")
					if err != nil {
						fmt.Println("err: ", err)
					}
				}
				// else if info.ServerName == "domain2.com" {
				// 	cert, err = tls.LoadX509KeyPair("path/to/domain2.crt", "path/to/domain2.key")
				// } else {
				// 	return nil, fmt.Errorf("no certificate found for domain: %s", info.ServerName)
				// }

				if err != nil {
					return nil, err
				}

				return &cert, nil
			},
		},
	}

	// Start the server with TLS
	fmt.Println("Server running on port 443")
	log.Fatal(server.ListenAndServeTLS("", ""))
}
