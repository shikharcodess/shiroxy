package implementations

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

// func init() {
// 	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config(InsecureSk)
// }

func TrafficLabReverseProxy() {
	demoURL, err := url.Parse("http://127.0.0.1")
	if err != nil {
		log.Fatal(err)
	}

	// proxy := httputil.NewSingleHostReverseProxy(demoURL)

	proxy := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.Host = demoURL.Host
		req.URL.Host = demoURL.Host
		req.URL.Scheme = demoURL.Scheme
		req.RequestURI = ""

		s, _, _ := net.SplitHostPort(req.RemoteAddr)
		req.Header.Set("X-Forwarded-For", s)

		http2.ConfigureTransport(http.DefaultTransport.(*http.Transport))

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Println(rw, err)
			return
		}

		for key, values := range response.Header {
			for _, value := range values {
				rw.Header().Set(key, value)
			}
		}

		done := make(chan bool)
		go func() {
			for {
				select {
				case <-time.Tick(10 * time.Millisecond):
					rw.(http.Flusher).Flush()
				case <-done:
					return
				}
			}
		}()

		trailerKeys := []string{}
		for key := range response.Trailer {
			trailerKeys = append(trailerKeys, key)
		}

		rw.Header().Set("Trailer", strings.Join(trailerKeys, ","))

		rw.WriteHeader(response.StatusCode)
		io.Copy(rw, response.Body)

		for key, values := range response.Trailer {
			for _, value := range values {
				rw.Header().Set(key, value)
			}
		}

		close(done)
	})

	http.ListenAndServe(":8080", proxy)
}
