package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"shiroxy/pkg/models"
)

func TestProxy_ForwardsToBackend(t *testing.T) {
	// upstream server
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("upstream-response"))
	}))
	defer up.Close()

	targetURL, _ := url.Parse(up.URL)

	// create server entry with Shiroxy configured to point to target
	srv := &Server{Id: "u1", URL: targetURL, Alive: true, Lock: &sync.RWMutex{}}
	// configure Shiroxy director to route to target
	srv.Shiroxy = &Shiroxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
		},
		Transport:  http.DefaultTransport,
		BufferPool: NewSyncBufferPool(32 * 1024),
	}

	servers := &BackendServers{Servers: []*Server{srv}}
	lb := &LoadBalancer{
		Ready:               true,
		Servers:             servers,
		RoutingDetailsByTag: map[string]*TagRoutingDetails{"": {Current: 0}},
		configuration:       &models.Config{Backend: models.Backend{Balance: "round-robin"}},
		TagCache:            &TagCache{cache: make(map[string]*BackendServers), capacity: 10},
		TagTrie:             &TrieNode{Children: make(map[rune]*TrieNode)},
	}

	// start a frontend server that uses lb.ServeHTTP
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ensure host resolves to localhost path used by lb
		r.Host = "localhost"
		lb.ServeHTTP(w, r)
	}))
	defer frontend.Close()

	resp, err := http.Get(frontend.URL)
	if err != nil {
		t.Fatalf("http get: %v", err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(b), "upstream-response") {
		t.Fatalf("unexpected body: %s", string(b))
	}
}
