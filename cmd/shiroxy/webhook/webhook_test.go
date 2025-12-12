package webhook

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
)

func TestWebhook_FireSendsPayload(t *testing.T) {
	var received string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		received = string(b)
		w.WriteHeader(200)
	}))
	defer server.Close()

	logg, _ := logger.StartLogger(nil)
	wg := &sync.WaitGroup{}
	config := models.Webhook{Enable: true, Url: server.URL, Events: []string{"test-event"}}
	wh, err := StartWebhookHandler(config, logg, wg, "")
	if err != nil {
		t.Fatalf("start webhook handler: %v", err)
	}

	// Fire event
	wh.Fire("test-event", map[string]string{"domain": "example.com"})

	// wait briefly for goroutine to fire webhook
	wg.Add(1)
	go func() {
		defer wg.Done()
		// allow the single goroutine in StartWebhookHandler to run
	}()
	wg.Wait()

	if !strings.Contains(received, "test-event") {
		t.Fatalf("expected event in payload, got: %s", received)
	}
}
