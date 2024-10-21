// Author: @ShikharY10
// Docs Author: ChatGPT

// Package proxy implements load balancing and reverse proxy functionalities
// including tag-based routing, sticky sessions, round-robin, and least connection server selections.
package proxy

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"shiroxy/cmd/shiroxy/domains" // Custom package for domain metadata handling.
	"shiroxy/cmd/shiroxy/webhook" // Custom package for webhook handling.
	"shiroxy/pkg/models"          // Custom package for configuration models.
	"shiroxy/public"              // Custom package for public constants and assets.
	"sync"
	"time"
)

// TagRoutingDetails maintains routing details per tag, supporting round-robin,
// least connection, and sticky session routing algorithms.
type TagRoutingDetails struct {
	Current         int                // Index for round-robin routing.
	ConnectionCount map[*Server]int    // Map of servers to their current connection counts for least-connection routing.
	StickySessions  map[string]*Server // Map of client IPs to servers for sticky session management.
}

// Frontends holds an HTTP handler function for serving incoming requests.
type Frontends struct {
	handlerFunc http.HandlerFunc // Function to handle HTTP requests.
}

// BackendServers contains a list of servers for load balancing.
type BackendServers struct {
	Servers []*Server // Slice of server instances.
}

// ServerByTags maps tags to their respective backend servers, supporting tag-based routing.
type ServerByTags struct {
	Servers map[string]*BackendServers // Map of tag names to backend server lists.
}

// Server represents a backend server with associated metadata and status.
type Server struct {
	Id                            string   `json:"id"`                                 // Unique identifier for the server.
	URL                           *url.URL `json:"url"`                                // URL of the server.
	HealthCheckUrl                string   `json:"health_check_url"`                   // URL used for health checks.
	Alive                         bool     `json:"alive"`                              // Indicates if the server is healthy.
	Shiroxy                       *Shiroxy `json:"-"`                                  // Shiroxy reverse proxy instance for the server.
	FireWebhookOnFirstHealthCheck bool     `json:"fire_webhook_on_first_health_check"` // Flag to trigger webhook on first successful health check.
	Tags                          []string `json:"-"`                                  // Tags for routing purposes.
	Lock                          *sync.RWMutex
}

// LoadBalancer implements the main load-balancing logic, supporting various routing mechanisms.
type LoadBalancer struct {
	Ready               bool
	MaxRetry            int
	RetryCounter        int
	configuration       *models.Config
	Servers             *BackendServers
	ServerByTag         *ServerByTags
	Frontends           map[string]*Frontends
	Mutex               sync.RWMutex
	RoutingDetailsByTag map[string]*TagRoutingDetails
	HealthChecker       *HealthChecker
	DomainStorage       *domains.Storage
	TagCache            *TagCache
	TagTrie             *TrieNode // Trie structure for tag indexing.
}

// NewLoadBalancer initializes a LoadBalancer with health checking, tag indexing, and caching mechanisms.
// Parameters:
//   - configuration: *models.Config, configuration settings for the load balancer.
//   - servers: *BackendServers, list of servers to load balance across.
//   - webhookHandler: *webhook.WebhookHandler, handles webhook actions on events.
//   - domainStorage: *domains.Storage, contains domain metadata.
//   - wg: *sync.WaitGroup, for synchronization with goroutines.
//
// Returns:
//   - *LoadBalancer: a new LoadBalancer instance.
func NewLoadBalancer(configuration *models.Config, servers *BackendServers, webhookHandler *webhook.WebhookHandler, domainStorage *domains.Storage, wg *sync.WaitGroup) *LoadBalancer {
	// Initialize health checker and start health checks on the servers.
	healthChecker := NewHealthChecker(servers, webhookHandler, time.Second*time.Duration(configuration.Backend.HealthCheckTriggerDuration), wg)
	healthChecker.StartHealthCheck()

	// Create the LoadBalancer instance.
	lb := LoadBalancer{
		Ready:         true,
		MaxRetry:      3, // Maximum retry count for failed server requests.
		RetryCounter:  0,
		configuration: configuration,
		Servers:       servers,
		ServerByTag: &ServerByTags{
			Servers: make(map[string]*BackendServers),
		},
		RoutingDetailsByTag: make(map[string]*TagRoutingDetails),
		HealthChecker:       healthChecker,
		DomainStorage:       domainStorage,
		TagCache:            NewTagCache(100), // Initialize a cache with a capacity of 100 entries.
		TagTrie:             NewTrieNode(),    // Initialize a trie for tag-based routing.
		Frontends:           make(map[string]*Frontends),
	}

	// Add a default routing entry for requests without specific tags.
	lb.RoutingDetailsByTag[""] = &TagRoutingDetails{
		Current:         0,
		ConnectionCount: map[*Server]int{},
		StickySessions:  map[string]*Server{},
	}

	// Extract and index tags for routing.
	lb.ExtractTags()

	return &lb
}

// ExtractTags processes servers to group them by their tags and initialize routing details.
func (lb *LoadBalancer) ExtractTags() {
	serverByTags := map[string]*BackendServers{}
	for _, server := range lb.Servers.Servers {
		if len(server.Tags) > 0 {
			for _, tag := range server.Tags {
				if serverByTags[tag] != nil {
					// Append server to the existing tag group.
					serverByTags[tag].Servers = append(serverByTags[tag].Servers, server)
				} else {
					// Initialize a new BackendServers instance for the tag.
					serverByTags[tag] = &BackendServers{}
					serverByTags[tag].Servers = []*Server{server}

					// Create routing details for the tag.
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

// GetNextServerRoundRobin selects the next server in a round-robin manner for the specified tag.
// Returns the selected server.
func (lb *LoadBalancer) GetNextServerRoundRobin(tag string, servers []*Server) *Server {
	var backendServers *BackendServers
	var routingDetails *TagRoutingDetails

	lb.Mutex.RLock()
	if tag != "" {
		backendServers = lb.ServerByTag.Servers[tag]
		routingDetails = lb.RoutingDetailsByTag[tag]
		if backendServers == nil || routingDetails == nil {
			lb.Mutex.RUnlock()
			return nil // No servers for the tag.
		}
	} else {
		backendServers = lb.Servers
		routingDetails = lb.RoutingDetailsByTag[""]
	}
	lb.Mutex.RUnlock()

	var server *Server
	serverlistLength := len(backendServers.Servers)

	index := routingDetails.Current % serverlistLength

	for i := 0; i < serverlistLength; i++ {
		server = backendServers.Servers[index]

		server.Lock.RLock()
		if server.Alive {
			routingDetails.Current = (index + 1) % serverlistLength
			server.Lock.RUnlock() // Ensure lock is released before returning
			return server
		}
		server.Lock.RUnlock() // Release the lock if the server is not alive

		index = (index + 1) % serverlistLength
	}

	return nil // Return nil if no alive server is found
}

// GetLeastConnectionServer selects the server with the least number of active connections for the specified tag.
// Returns the selected server.
func (lb *LoadBalancer) GetLeastConnectionServer(tag string, servers []*Server) *Server {
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
	minConn := int(^uint(0) >> 1) // Set to max int value.
	for _, server := range backendServers.Servers {
		if routingDetails.ConnectionCount[server] < minConn {
			minConn = routingDetails.ConnectionCount[server]
			leastConnServer = server
		}
	}
	routingDetails.ConnectionCount[leastConnServer]++
	return leastConnServer
}

// GetStickySessionServer returns the server associated with the client IP for sticky sessions.
// If no association exists, selects a server using round-robin and creates a new sticky session.
func (lb *LoadBalancer) GetStickySessionServer(clientIP string, tag string, servers []*Server) *Server {

	var backendServers *BackendServers
	var routingDetails *TagRoutingDetails

	lb.Mutex.RLock()
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
	lb.Mutex.RUnlock()

	if server, exists := routingDetails.StickySessions[clientIP]; exists {
		return server // Return existing session association.
	}

	// Select a server using round-robin and associate it with the client IP.
	if len(servers) > 0 {
		server := lb.GetNextServerRoundRobin(tag, servers)
		routingDetails.StickySessions[clientIP] = server
		return server
	} else {
		server := lb.GetNextServerRoundRobin(tag, backendServers.Servers)
		routingDetails.StickySessions[clientIP] = server
		return server
	}
}

// serveHTTP processes HTTP requests, performing tag-based routing and handling fallbacks.
// Parameters:
//   - w: http.ResponseWriter, for writing HTTP responses.
//   - r: *ShiroxyRequest, the incoming HTTP request with associated metadata.
func (lb *LoadBalancer) serveHTTP(w http.ResponseWriter, r *ShiroxyRequest) {
	if lb.Ready {
		var server *Server
		clientIP := r.Request.RemoteAddr

		host, _, err := net.SplitHostPort(r.Request.Host)
		if err != nil {
			// Handle the case where no port is provided (e.g., IP address without a port)
			host = r.Request.Host
			// port = "" // or assign a default port if needed
		}

		// First, check if the host is an IP address.
		ip := net.ParseIP(host)
		if ip != nil {
			// If it's an IP, proceed with selecting a server based on the IP.
			server = lb.selectServerBasedOnRule(clientIP, "")
		} else {
			// If it's not an IP, assume it's a domain name.
			domainData := lb.DomainStorage.DomainMetadata[host]
			if domainData == nil {
				http.Error(w, "Domain not found", http.StatusNotFound)
				return
			}

			// Extract tags and apply tag rules.
			tags := domainData.Metadata["tags"]
			if tags == "" && lb.configuration.Backend.Tagrule == "strict" {
				http.Error(w, "No tag found and strict tag rule is enabled", http.StatusServiceUnavailable)
				return
			}

			// Select the server based on the tags (or an empty tag if no tags are present).
			server = lb.selectServerBasedOnRule(clientIP, tags)
		}

		// fmt.Println("SERVER FOUND==========: ", server)

		// If server is found, forward the request.

		if server != nil {
			server.Lock.RLock()
			alive := server.Alive
			instanceOfServer := *server
			server.Lock.RUnlock()

			if alive {
				err := instanceOfServer.Shiroxy.ServeHTTP(w, r)
				if err != nil {

					if r.RetryCount <= lb.MaxRetry {
						r.RetryCount++
						lb.serveHTTP(w, r)
					} else {

						server.Lock.Lock()
						server.Alive = false
						server.Lock.Unlock()

						// Load error page on server failure.
						shiroxyNotReadyResponse := LoadErrorPageHtmlContent(public.DOMAIN_NOT_FOUND_ERROR, &models.ErrorRespons{
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
				return
			} else {
				// Load error page if the load balancer is not ready.
				shiroxyNotReadyResponse := LoadErrorPageHtmlContent(public.SHIROXY_NOT_READY, &models.ErrorRespons{
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
		} else {
			http.Error(w, "No available servers for the tag", http.StatusServiceUnavailable)
		}
	} else {
		// Load error page if the load balancer is not ready.
		shiroxyNotReadyResponse := LoadErrorPageHtmlContent(public.SHIROXY_NOT_READY, &models.ErrorRespons{
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

// ServeHTTP is a wrapper for serveHTTP that prepares the ShiroxyRequest.
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shiroxyRequest := &ShiroxyRequest{
		RetryCount: 3,
		Request:    r,
	}
	lb.serveHTTP(w, shiroxyRequest)
}

// updateServerTags reindexes servers based on their tags and updates the caching mechanisms.
func (lb *LoadBalancer) updateServerTags() {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	// Reset ServerByTag and RoutingDetailsByTag.
	lb.ServerByTag.Servers = make(map[string]*BackendServers)
	lb.RoutingDetailsByTag = make(map[string]*TagRoutingDetails)

	// Reset the tag cache and trie indexing.
	lb.TagCache = NewTagCache(100)
	lb.TagTrie = NewTrieNode()

	// Re-index servers by their tags.
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
				lb.TagCache.Set(tag, lb.ServerByTag.Servers[tag])   // Add to cache.
				lb.TagTrie.Insert(tag, lb.ServerByTag.Servers[tag]) // Add to trie.
			}
		}
	}
}

// selectServerBasedOnRule selects a server based on the configured load balancing method (round-robin, least connection, or sticky session).
// Parameters:
//   - clientIP: string, the client's IP address for sticky sessions.
//   - tag: string, the tag used for routing.
//
// Returns:
//   - *Server: the selected server.
func (lb *LoadBalancer) selectServerBasedOnRule(clientIP, tag string) *Server {
	// Check the cache first.
	if cachedServers, found := lb.TagCache.Get(tag); found {
		return lb.selectServerFromList(clientIP, cachedServers, tag)
	}

	// Search in the trie for a matching tag.
	if servers, found := lb.TagTrie.Search(tag); found {
		lb.TagCache.Set(tag, servers) // Cache the found servers.
		return lb.selectServerFromList(clientIP, servers, tag)
	}

	// If not found in cache or trie, fallback to global list without tags.
	return lb.selectServerFromList(clientIP, lb.Servers, "")
}

// selectServerFromList chooses a server based on the load balancing method.
// Parameters:
//   - clientIP: string, the client's IP address for sticky sessions.
//   - servers: *BackendServers, the list of servers.
//   - tag: string, the tag for routing.
//
// Returns:
//   - *Server: the selected server.
func (lb *LoadBalancer) selectServerFromList(clientIP string, servers *BackendServers, tag string) *Server {
	switch lb.configuration.Backend.Balance {
	case "round-robin":
		return lb.GetNextServerRoundRobin(tag, servers.Servers)
	case "least-count":
		return lb.GetLeastConnectionServer(tag, servers.Servers)
	case "sticky-session":
		return lb.GetStickySessionServer(clientIP, tag, servers.Servers)
	default:
		return lb.GetNextServerRoundRobin(tag, servers.Servers)
	}
}

// NewTagCache creates a new TagCache instance with the specified capacity.
// Parameters:
//   - capacity: int, the maximum number of entries in the cache.
//
// Returns:
//   - *TagCache: the initialized tag cache.
func NewTagCache(capacity int) *TagCache {
	return &TagCache{
		cache:    make(map[string]*BackendServers),
		capacity: capacity,
	}
}

// TagCache is a cache for frequently accessed tags to optimize routing.
type TagCache struct {
	cache    map[string]*BackendServers
	capacity int
	keys     []string // Keys for LRU (Least Recently Used) eviction.
}

// Get retrieves the servers associated with a tag from the cache.
// Parameters:
//   - tag: string, the tag to retrieve.
//
// Returns:
//   - *BackendServers, bool: the cached servers and whether they were found.
func (tc *TagCache) Get(tag string) (*BackendServers, bool) {
	if servers, found := tc.cache[tag]; found {
		tc.moveToEnd(tag)
		return servers, true
	}
	return nil, false
}

// Set adds a tag and its servers to the cache, evicting the oldest entry if the cache is full.
// Parameters:
//   - tag: string, the tag to cache.
//   - servers: *BackendServers, the servers to associate with the tag.
func (tc *TagCache) Set(tag string, servers *BackendServers) {
	if len(tc.cache) >= tc.capacity {
		oldestKey := tc.keys[0]
		tc.keys = tc.keys[1:]
		delete(tc.cache, oldestKey)
	}
	tc.cache[tag] = servers
	tc.keys = append(tc.keys, tag)
}

// moveToEnd moves the specified tag to the end of the keys slice to mark it as recently used.
// Parameters:
//   - tag: string, the tag to move.
func (tc *TagCache) moveToEnd(tag string) {
	for i, key := range tc.keys {
		if key == tag {
			tc.keys = append(tc.keys[:i], tc.keys[i+1:]...)
			tc.keys = append(tc.keys, tag)
			break
		}
	}
}

// TrieNode represents a node in a trie used for tag indexing.
type TrieNode struct {
	Children map[rune]*TrieNode
	Servers  *BackendServers
	End      bool // Indicates if the node represents the end of a valid tag.
}

// NewTrieNode creates and returns a new TrieNode instance.
func NewTrieNode() *TrieNode {
	return &TrieNode{
		Children: make(map[rune]*TrieNode),
	}
}

// Insert adds a tag and its servers to the trie.
// Parameters:
//   - tag: string, the tag to insert.
//   - servers: *BackendServers, the servers to associate with the tag.
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

// Search looks up a tag in the trie.
// Parameters:
//   - tag: string, the tag to search for.
//
// Returns:
//   - *BackendServers, bool: the servers associated with the tag and whether the tag was found.
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

// allTrue checks if all elements in a boolean slice are true.
// Parameters:
//   - arr: []bool, the slice of booleans.
//
// Returns:
//   - bool: true if all elements are true, false otherwise.
func allTrue(arr []bool) bool {
	for _, v := range arr {
		if !v {
			return false
		}
	}
	return true
}
