package proxy_test

import (
	"net/url"
	"sync"
	"testing"

	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/cmd/shiroxy/webhook"
	"shiroxy/pkg/models"
)

func TestNewLoadBalancer(t *testing.T) {
	// Set up test data
	config := &models.Config{
		Backend: models.Backend{
			HealthCheckTriggerDuration: 5, // Set a health check duration for the test
		},
	}
	serverURL, _ := url.Parse("https://example.com/")
	servers := &proxy.BackendServers{
		Servers: []*proxy.Server{
			{
				Id:             "1",
				URL:            serverURL,
				Alive:          true,
				Tags:           []string{"tag1"},
				Lock:           &sync.RWMutex{},
				HealthCheckUrl: "/health",
			},
		},
	}
	webhookHandler := &webhook.WebhookHandler{}
	domainStorage := &domains.Storage{}
	var wg sync.WaitGroup

	// Call NewLoadBalancer
	lb := proxy.NewLoadBalancer(config, servers, webhookHandler, domainStorage, &wg)

	// Assertions
	if lb == nil {
		t.Fatalf("Expected non-nil LoadBalancer, got nil")
	}
	if lb.HealthChecker == nil {
		t.Errorf("Expected HealthChecker to be initialized, got nil")
	}
	if !lb.Ready {
		t.Errorf("Expected LoadBalancer to be ready, got not ready")
	}
	if len(lb.RoutingDetailsByTag[""].StickySessions) != 0 {
		t.Errorf("Expected default routing entry for sticky sessions, got non-empty")
	}
}

func TestGetNextServerRoundRobin(t *testing.T) {
	// Set up test data
	server1 := &proxy.Server{
		Id:    "1",
		Alive: true,
		Lock:  &sync.RWMutex{},
	}
	server2 := &proxy.Server{
		Id:    "2",
		Alive: true,
		Lock:  &sync.RWMutex{},
	}
	servers := []*proxy.Server{server1, server2}
	lb := &proxy.LoadBalancer{
		Servers: &proxy.BackendServers{
			Servers: servers,
		},
		RoutingDetailsByTag: map[string]*proxy.TagRoutingDetails{
			"": {
				Current: 0,
			},
		},
	}

	// Call getNextServerRoundRobin and assert
	selectedServer := lb.GetNextServerRoundRobin("", servers)
	if selectedServer.Id != "1" {
		t.Errorf("Expected server 1, got server %s", selectedServer.Id)
	}

	// Call again and assert round-robin behavior
	selectedServer = lb.GetNextServerRoundRobin("", servers)
	if selectedServer.Id != "2" {
		t.Errorf("Expected server 2, got server %s", selectedServer.Id)
	}

	// Test round-robin wrap-around
	selectedServer = lb.GetNextServerRoundRobin("", servers)
	if selectedServer.Id != "1" {
		t.Errorf("Expected server 1 after wrap-around, got server %s", selectedServer.Id)
	}
}

func TestGetLeastConnectionServer(t *testing.T) {
	server1 := &proxy.Server{
		Id:    "1",
		Alive: true,
		Lock:  &sync.RWMutex{},
	}
	server2 := &proxy.Server{
		Id:    "2",
		Alive: true,
		Lock:  &sync.RWMutex{},
	}
	servers := []*proxy.Server{server1, server2}

	lb := &proxy.LoadBalancer{
		Servers: &proxy.BackendServers{
			Servers: servers,
		},
		RoutingDetailsByTag: map[string]*proxy.TagRoutingDetails{
			"": {
				ConnectionCount: map[*proxy.Server]int{
					server1: 10,
					server2: 5, // Server 2 has fewer connections
				},
				Current: 0,
			},
		},
	}

	// Call getLeastConnectionServer and assert
	selectedServer := lb.GetLeastConnectionServer("", servers)
	if selectedServer.Id != "2" {
		t.Errorf("Expected server 2 with least connections, got server %s", selectedServer.Id)
	}
}

func TestGetStickySessionServer(t *testing.T) {
	server1 := &proxy.Server{
		Id:    "1",
		Alive: true,
		Lock:  &sync.RWMutex{},
	}
	server2 := &proxy.Server{
		Id:    "2",
		Alive: true,
		Lock:  &sync.RWMutex{},
	}
	servers := []*proxy.Server{server1, server2}

	lb := &proxy.LoadBalancer{
		Servers: &proxy.BackendServers{
			Servers: servers,
		},
		RoutingDetailsByTag: map[string]*proxy.TagRoutingDetails{
			"": {
				StickySessions: make(map[string]*proxy.Server),
				Current:        0,
			},
		},
	}

	clientIP := "192.168.0.1"

	// Test sticky session creation
	selectedServer := lb.GetStickySessionServer(clientIP, "", servers)
	if selectedServer == nil {
		t.Fatalf("Expected a server, got nil")
	}

	// Ensure sticky session is saved
	stickyServer := lb.RoutingDetailsByTag[""].StickySessions[clientIP]
	if stickyServer.Id != selectedServer.Id {
		t.Errorf("Expected sticky session to be saved for server %s, got %s", selectedServer.Id, stickyServer.Id)
	}

	// Test that sticky session returns the same server
	selectedServer2 := lb.GetStickySessionServer(clientIP, "", servers)
	if selectedServer.Id != selectedServer2.Id {
		t.Errorf("Expected sticky session to return the same server, got different servers: %s, %s", selectedServer.Id, selectedServer2.Id)
	}
}
