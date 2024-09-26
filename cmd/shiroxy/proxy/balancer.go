package proxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/webhook"
	"shiroxy/pkg/models"
	"shiroxy/public"
	"sync"
	"time"
)

type TagRoutingDetails struct {
	Current         int                // Round-robin index
	ConnectionCount map[*Server]int    // Least-connection count
	StickySessions  map[string]*Server // Sticky session mapping
}

type Frontends struct {
	handlerFunc http.HandlerFunc
}

type BackendServers struct {
	Servers []*Server
}

type ServerByTags struct {
	Servers map[string]*BackendServers
}

type Server struct {
	Id                            string   `json:"id"`
	URL                           *url.URL `json:"url"`
	HealthCheckUrl                *url.URL `json:"health_check_url"`
	Alive                         bool     `json:"alive"`
	Shiroxy                       *Shiroxy `json:"-"`
	FireWebhookOnFirstHealthCheck bool     `json:"fire_webhook_on_first_health_check"`
	Tags                          []string `json:"-"`
}

type LoadBalancer struct {
	Ready               bool
	MaxRetry            int
	RetryCounter        int
	configuration       *models.Config
	Servers             *BackendServers
	ServerByTag         *ServerByTags
	Frontends           map[string]*Frontends
	Mutex               sync.Mutex
	RoutingDetailsByTag map[string]*TagRoutingDetails
	HealthChecker       *HealthChecker
	DomainStorage       *domains.Storage
	TagCache            *TagCache
	TagTrie             *TrieNode // Tag indexing
}

// Initialize a new LoadBalancer with caching and indexing
func NewLoadBalancer(configuration *models.Config, servers *BackendServers, webhookHandler *webhook.WebhookHandler, domainStorage *domains.Storage, wg *sync.WaitGroup) *LoadBalancer {
	healthChecker := NewHealthChecker(servers, webhookHandler, time.Second*time.Duration(configuration.Backend.HealthCheckTriggerDuration), wg)
	healthChecker.StartHealthCheck()

	lb := LoadBalancer{
		Ready:         true,
		MaxRetry:      3,
		RetryCounter:  0,
		configuration: configuration,
		Servers:       servers,
		ServerByTag: &ServerByTags{
			Servers: make(map[string]*BackendServers),
		},
		RoutingDetailsByTag: make(map[string]*TagRoutingDetails),
		HealthChecker:       healthChecker,
		DomainStorage:       domainStorage,
		TagCache:            NewTagCache(100), // Cache capacity of 100
		TagTrie:             NewTrieNode(),
		Frontends:           make(map[string]*Frontends),
	}
	// For fallback domains which does not contain any tag for routing.
	lb.RoutingDetailsByTag[""] = &TagRoutingDetails{
		Current:         0,
		ConnectionCount: map[*Server]int{},
		StickySessions:  map[string]*Server{},
	}
	lb.ExtractTags()

	// go func() {
	// 	var checks []bool = []bool{}
	// 	for _, server := range lb.Servers.Servers {
	// 		checks = append(checks, healthChecker.CheckHealth(server))
	// 	}

	// 	fmt.Println("checks: ", checks)

	// 	allTrue := allTrue(checks)

	// 	if configuration.Backend.NoServerAction == "strict" && allTrue {
	// 		lb.Ready = true
	// 	} else if configuration.Backend.NoServerAction == "strict" && !allTrue {
	// 		lb.Ready = false
	// 	} else {
	// 		lb.Ready = true
	// 	}
	// }()
	return &lb
}

func (lb *LoadBalancer) ExtractTags() {
	serverByTags := map[string]*BackendServers{}
	for _, server := range lb.Servers.Servers {
		if len(server.Tags) > 0 {
			for _, tag := range server.Tags {

				if serverByTags[tag] != nil {
					serverByTags[tag].Servers = append(serverByTags[tag].Servers, server)
				} else {
					serverByTags[tag] = &BackendServers{}
					serverByTags[tag].Servers = []*Server{server}
					// serverByTags[tag].Servers = append(serverByTags[tag].Servers, server)

					lb.RoutingDetailsByTag[tag] = &TagRoutingDetails{
						Current:         0,
						ConnectionCount: map[*Server]int{},
						StickySessions:  map[string]*Server{},
					}
				}
			}
		}
	}
}

func (lb *LoadBalancer) getNextServerRoundRobin(tag string, servers []*Server) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	var backendServers *BackendServers
	var routingDetails *TagRoutingDetails

	if tag != "" {
		backendServers = lb.ServerByTag.Servers[tag]
		routingDetails = lb.RoutingDetailsByTag[tag]
		if backendServers == nil || routingDetails == nil {
			return nil // Handle no servers for the tag
		}
	} else {
		backendServers = lb.Servers
		routingDetails = lb.RoutingDetailsByTag[""]
	}

	fmt.Println("backendServers: ", backendServers)
	fmt.Println("routingDetails: ", routingDetails)

	server := backendServers.Servers[routingDetails.Current]
	routingDetails.Current = (routingDetails.Current + 1) % len(backendServers.Servers)

	return server
}

// Least connection server selection per tag
func (lb *LoadBalancer) getLeastConnectionServer(tag string, servers []*Server) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	var backendServers *BackendServers
	var routingDetails *TagRoutingDetails

	if tag != "" {
		backendServers = lb.ServerByTag.Servers[tag]
		routingDetails = lb.RoutingDetailsByTag[tag]
		if backendServers == nil || routingDetails == nil {
			return nil
		}
	} else {
		backendServers = lb.Servers
		routingDetails = lb.RoutingDetailsByTag[""]
	}

	var leastConnServer *Server
	minConn := int(^uint(0) >> 1)
	for _, server := range backendServers.Servers {
		if routingDetails.ConnectionCount[server] < minConn {
			minConn = routingDetails.ConnectionCount[server]
			leastConnServer = server
		}
	}
	routingDetails.ConnectionCount[leastConnServer]++
	return leastConnServer
}

// Sticky session-based server selection per tag
func (lb *LoadBalancer) getStickySessionServer(clientIP string, tag string, servers []*Server) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	var backendServers *BackendServers
	var routingDetails *TagRoutingDetails

	if tag != "" {
		backendServers = lb.ServerByTag.Servers[tag]
		routingDetails = lb.RoutingDetailsByTag[tag]
		if backendServers == nil || routingDetails == nil {
			return nil
		}
	} else {
		backendServers = lb.Servers
		routingDetails = lb.RoutingDetailsByTag[""]
	}

	if server, exists := routingDetails.StickySessions[clientIP]; exists {
		return server
	}

	server := lb.getNextServerRoundRobin(tag, servers)
	routingDetails.StickySessions[clientIP] = server
	return server
}

// Serve HTTP requests with tag-based routing and fallback mechanisms
func (lb *LoadBalancer) serveHTTP(w http.ResponseWriter, r *ShiroxyRequest) {
	if lb.Ready {
		var server *Server
		clientIP := r.Request.RemoteAddr

		fmt.Println("r.Request.Host: ", r.Request.Host)

		ip := net.ParseIP(r.Request.Host)
		if ip == nil {
			domainName, _, err := net.SplitHostPort(r.Request.Host)
			if err != nil {
				http.Error(w, "Invalid host format", http.StatusBadRequest)
				return
			}

			domainData := lb.DomainStorage.DomainMetadata[domainName]
			if domainData == nil {
				http.Error(w, "Domain not found", http.StatusNotFound)
				return
			}

			// Extract tags and apply tag rule
			tags := domainData.Metadata["tags"]
			if tags == "" {
				if lb.configuration.Backend.Tagrule == "strict" {
					http.Error(w, "No tag found and strict tag rule is enabled", http.StatusServiceUnavailable)
					return
					// Use fallback server selection method
				} else {
					server = lb.selectServerBasedOnRule(clientIP, "")
				}
			} else {
				server = lb.selectServerBasedOnRule(clientIP, tags)
			}
		} else {
			server = lb.selectServerBasedOnRule(clientIP, "")
		}

		// If server is found, forward the request
		if server != nil {
			err := server.Shiroxy.ServeHTTP(w, r)
			if err != nil {
				server.Alive = false
				if r.RetryCount <= lb.MaxRetry {
					r.RetryCount++
					lb.serveHTTP(w, r)
				} else {
					// server.Shiroxy.DefaultErrorHandler(w, r.Request, fmt.Errorf("dial up failed, host : %s", r.Request.RemoteAddr))
					shiroxyNotReadyResponse := loadErrorPageHtmlContent(public.DOMAIN_NOT_FOUND_ERROR, &models.ErrorRespons{
						ErrorPageButtonName: "Shiroxy",
						ErrorPageButtonUrl:  "",
					})
					w.Header().Add("Content-Type", "text/html")
					w.WriteHeader(400)
					_, err := w.Write([]byte(shiroxyNotReadyResponse))
					if err != nil {
						log.Printf("failed to write response: %v", err)
					}
				}
			}
		} else {
			http.Error(w, "No available servers for the tag", http.StatusServiceUnavailable)
		}
	} else {
		shiroxyNotReadyResponse := loadErrorPageHtmlContent(public.SHIROXY_NOT_READY, &models.ErrorRespons{
			ErrorPageButtonName: "Shiroxy",
			ErrorPageButtonUrl:  "",
		})
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(400)
		_, err := w.Write([]byte(shiroxyNotReadyResponse))
		if err != nil {
			log.Printf("failed to write response: %v", err)
		}
		// http.Error(w, domainNotFoundErrorResponse, http.StatusServiceUnavailable)
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shiroxyRequest := &ShiroxyRequest{
		RetryCount: 3,
		Request:    r,
	}

	lb.serveHTTP(w, shiroxyRequest)
}

// Updating server tags with indexing and caching
func (lb *LoadBalancer) updateServerTags() {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	lb.ServerByTag.Servers = make(map[string]*BackendServers)
	lb.RoutingDetailsByTag = make(map[string]*TagRoutingDetails)

	// Reset the tag cache and indexing
	lb.TagCache = NewTagCache(100)
	lb.TagTrie = NewTrieNode()

	for _, server := range lb.Servers.Servers {
		if server.Alive {
			for _, tag := range server.Tags {
				if _, exists := lb.ServerByTag.Servers[tag]; !exists {
					lb.ServerByTag.Servers[tag] = &BackendServers{Servers: []*Server{}}
					lb.RoutingDetailsByTag[tag] = &TagRoutingDetails{
						ConnectionCount: make(map[*Server]int),
						StickySessions:  make(map[string]*Server),
					}
				}
				lb.ServerByTag.Servers[tag].Servers = append(lb.ServerByTag.Servers[tag].Servers, server)
				lb.TagCache.Set(tag, lb.ServerByTag.Servers[tag])   // Add to cache
				lb.TagTrie.Insert(tag, lb.ServerByTag.Servers[tag]) // Add to trie
			}
		}
	}
}

func (lb *LoadBalancer) selectServerBasedOnRule(clientIP, tag string) *Server {
	// Check the cache first
	if cachedServers, found := lb.TagCache.Get(tag); found {
		return lb.selectServerFromList(clientIP, cachedServers, tag)
	}

	// Search in the trie for a matching tag
	if servers, found := lb.TagTrie.Search(tag); found {
		lb.TagCache.Set(tag, servers) // Cache the found servers
		return lb.selectServerFromList(clientIP, servers, tag)
	}

	// If not found in cache or trie, fallback to global list without tags
	return lb.selectServerFromList(clientIP, lb.Servers, "")
}

func (lb *LoadBalancer) selectServerFromList(clientIP string, servers *BackendServers, tag string) *Server {
	switch lb.configuration.Backend.Balance {
	case "round-robin":
		return lb.getNextServerRoundRobin(tag, servers.Servers)
	case "least-count":
		return lb.getLeastConnectionServer(tag, servers.Servers)
	case "sticky-session":
		return lb.getStickySessionServer(clientIP, tag, servers.Servers)
	default:
		return lb.getNextServerRoundRobin(tag, servers.Servers)
	}
}

// Caching mechanism for frequently used tags
func NewTagCache(capacity int) *TagCache {
	return &TagCache{
		cache:    make(map[string]*BackendServers),
		capacity: capacity,
	}
}

type TagCache struct {
	cache    map[string]*BackendServers
	capacity int
	keys     []string // For LRU eviction
}

func (tc *TagCache) Get(tag string) (*BackendServers, bool) {
	if servers, found := tc.cache[tag]; found {
		tc.moveToEnd(tag)
		return servers, true
	}
	return nil, false
}

func (tc *TagCache) Set(tag string, servers *BackendServers) {
	if len(tc.cache) >= tc.capacity {
		oldestKey := tc.keys[0]
		tc.keys = tc.keys[1:]
		delete(tc.cache, oldestKey)
	}
	tc.cache[tag] = servers
	tc.keys = append(tc.keys, tag)
}

func (tc *TagCache) moveToEnd(tag string) {
	for i, key := range tc.keys {
		if key == tag {
			tc.keys = append(tc.keys[:i], tc.keys[i+1:]...)
			tc.keys = append(tc.keys, tag)
			break
		}
	}
}

type TrieNode struct {
	Children map[rune]*TrieNode
	Servers  *BackendServers
	End      bool
}

// Trie-based tag indexing for efficient tag lookups
func NewTrieNode() *TrieNode {
	return &TrieNode{
		Children: make(map[rune]*TrieNode),
	}
}

func (node *TrieNode) Insert(tag string, servers *BackendServers) {
	currentNode := node
	for _, ch := range tag {
		if _, found := currentNode.Children[ch]; !found {
			currentNode.Children[ch] = NewTrieNode()
		}
		currentNode = currentNode.Children[ch]
	}
	currentNode.Servers = servers
	currentNode.End = true
}

func (node *TrieNode) Search(tag string) (*BackendServers, bool) {
	currentNode := node
	for _, ch := range tag {
		if _, found := currentNode.Children[ch]; !found {
			return nil, false
		}
		currentNode = currentNode.Children[ch]
	}
	if currentNode.End {
		return currentNode.Servers, true
	}
	return nil, false
}

func allTrue(arr []bool) bool {
	for _, v := range arr {
		if !v {
			return false
		}
	}
	return true
}
