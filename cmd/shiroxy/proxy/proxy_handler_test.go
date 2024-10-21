package proxy_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/cmd/shiroxy/webhook"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
)

func TestStartShiroxyHandler(t *testing.T) {
	fmt.Println("TestStartShiroxyHandler")

	var wg sync.WaitGroup

	// Mock configuration
	config := &models.Config{
		Frontend: models.Frontend{
			Mode: "http",
			Bind: []models.FrontendBind{
				{
					Host:   "localhost",
					Port:   "8080",
					Target: "multiple",
				},
			},
		},
		Backend: models.Backend{
			Servers: []models.BackendServer{
				{
					Id:        "server1",
					Host:      "localhost",
					Port:      "9090",
					Tags:      "tag1,tag2",
					HealthUrl: "/health",
				},
			},
		},
	}

	// Mock storage
	storage := &domains.Storage{
		DomainMetadata: map[string]*domains.DomainMetadata{
			"localhost": {
				Status:          "active",
				DnsChallengeKey: "challenge_key",
			},
		},
	}

	// Mock webhook handler
	webhookHandler := &webhook.WebhookHandler{}

	// Mock logger
	logHandler := &logger.Logger{}

	// Start the proxy handler
	lb, err := proxy.StartShiroxyHandler(config, storage, webhookHandler, logHandler, &wg)

	// Test LoadBalancer creation
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lb == nil {
		t.Fatalf("expected load balancer to be non-nil")
	}
}

func TestCreateMultipleTargetServer(t *testing.T) {
	fmt.Println("TestCreateMultipleTargetServer")

	// Mock frontend bind data
	bindData := &models.FrontendBind{
		Host:   "localhost",
		Port:   "8080",
		Secure: false,
	}

	// Mock storage
	storage := &domains.Storage{
		DomainMetadata: map[string]*domains.DomainMetadata{
			"localhost": {
				Status:          "active",
				DnsChallengeKey: "challenge_key",
			},
		},
	}

	// Mock handler function
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create server
	server, secure, err := proxy.CreateMultipleTargetServer(bindData, storage, handlerFunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server == nil {
		t.Fatalf("expected server to be non-nil")
	}
	if secure {
		t.Fatalf("expected server to be insecure (no TLS)")
	}
}

func TestCreateSingleTargetServer(t *testing.T) {
	fmt.Println("TestCreateSingleTargetServer")

	// Mock frontend bind data
	bindData := &models.FrontendBind{
		Host:   "localhost",
		Port:   "8080",
		Secure: true,
		SecureSetting: models.FrontendSecuritySetting{
			SingleTargetMode: "certandkey",
			CertAndKey: models.FrontendSecuritySettingCertAndKey{
				Cert:   "testcert.pem",
				Key:    "testkey.pem",
				Domain: "localhost",
			},
		},
	}

	// Mock storage
	storage := &domains.Storage{
		DomainMetadata: map[string]*domains.DomainMetadata{
			"localhost": {
				Status: "active",
			},
		},
	}

	// Mock handler function
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create server
	server, secure, err := proxy.CreateSingleTargetServer(bindData, storage, handlerFunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server == nil {
		t.Fatalf("expected server to be non-nil")
	}
	if !secure {
		t.Fatalf("expected server to be secure (with TLS)")
	}
}

func TestResolveSecurityPolicy(t *testing.T) {
	fmt.Println("TestResolveSecurityPolicy")

	tests := []struct {
		policy   string
		expected tls.ClientAuthType
	}{
		{"none", tls.NoClientCert},
		{"optional", tls.RequestClientCert},
		{"required", tls.RequireAndVerifyClientCert},
		{"invalid", tls.NoClientCert}, // Default to NoClientCert for invalid policies
	}

	for _, test := range tests {
		result := proxy.ResolveSecurityPolicy(test.policy)
		if result != test.expected {
			t.Errorf("expected %v, got %v for policy %s", test.expected, result, test.policy)
		}
	}
}

func TestLoadErrorPageHtmlContent(t *testing.T) {
	fmt.Println("TestLoadErrorPageHtmlContent")

	// Mock error response configuration
	config := &models.ErrorRespons{
		ErrorPageButtonName: "Retry",
		ErrorPageButtonUrl:  "https://retry.com",
	}

	// Example HTML content with placeholders
	htmlContent := "<html><body><a href='{{button_url}}'>{{button_name}}</a></body></html>"

	// Load error page content
	result := proxy.LoadErrorPageHtmlContent(htmlContent, config)

	// Check if placeholders were replaced correctly
	expected := "<html><body><a href='https://retry.com'>Retry</a></body></html>"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestListenAndServe(t *testing.T) {
	fmt.Println("TestListenAndServe")

	var wg sync.WaitGroup

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Parse the URL of the test server
	parsedURL, _ := url.Parse(server.URL)

	// Mock logger
	logHandler := &logger.Logger{}

	// Create a new HTTP server
	httpServer := &http.Server{
		Addr:    parsedURL.Host,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	}

	// Start the server and wait for it to start
	proxy.ListenAndServe(httpServer, false, logHandler, &wg)
	wg.Wait()

	// Test if the server starts correctly
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code 200, got %v", resp.StatusCode)
	}
}
