package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"shiroxy/pkg/models"
	"sync"
	"time"
)

type Server struct {
	Id                            string
	URL                           *url.URL
	HealthCheckUrl                *url.URL
	Alive                         bool
	Shiroxy                       *Shiroxy
	FireWebhookOnFirstHealthCheck bool
}

type Frontends struct {
	handlerFunc http.HandlerFunc
}

type LoadBalancer struct {
	MaxRetry        int
	RetryCounter    int
	configuration   *models.Config
	Servers         []*Server
	Frontends       map[string]*Frontends
	Current         int
	Mutex           sync.Mutex
	ConnectionCount map[*Server]int
	StickySessions  map[string]*Server
	HealthChecker   *HealthChecker
}

func NewLoadBalancer(configuration *models.Config, servers []*Server, wg *sync.WaitGroup) *LoadBalancer {
	healthChecker := NewHealthChecker(servers, time.Second*5, wg)
	healthChecker.StartHealthCheck()

	return &LoadBalancer{
		MaxRetry:        3,
		RetryCounter:    0,
		configuration:   configuration,
		Servers:         servers,
		HealthChecker:   healthChecker,
		ConnectionCount: make(map[*Server]int),
		StickySessions:  make(map[string]*Server),
		Frontends:       make(map[string]*Frontends),
	}
}

func (lb *LoadBalancer) getNextServerRoundRobin() *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()
	server := lb.Servers[lb.Current]
	if server.Alive {
		lb.Current = (lb.Current + 1) % len(lb.Servers)
		return server
	} else {
		lb.Current = (lb.Current + 1) % len(lb.Servers)
		server := lb.Servers[lb.Current]
		return server
	}
}

func (lb *LoadBalancer) getLeastConnectionServer() *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()
	var leastConnServer *Server
	minConn := int(^uint(0) >> 1) // max int value
	for _, server := range lb.Servers {
		if lb.ConnectionCount[server] < minConn {
			minConn = lb.ConnectionCount[server]
			leastConnServer = server
		}
	}
	lb.ConnectionCount[leastConnServer]++
	return leastConnServer
}

func (lb *LoadBalancer) getStickySessionServer(clientIP string) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()
	if server, exists := lb.StickySessions[clientIP]; exists {
		return server
	}
	server := lb.getNextServerRoundRobin()
	lb.StickySessions[clientIP] = server
	return server
}

func (lb *LoadBalancer) serveHTTP(w http.ResponseWriter, r *ShiroxyRequest) {

	var server *Server
	clientIP := r.Request.RemoteAddr

	switch {
	case lb.configuration.Backend.Balance == "round-robin":
		server = lb.getNextServerRoundRobin()
	case lb.configuration.Backend.Balance == "sticky-session":
		server = lb.getStickySessionServer(clientIP)
	case lb.configuration.Backend.Balance == "least-count":
		server = lb.getLeastConnectionServer()
	}

	err := server.Shiroxy.ServeHTTP(w, r)
	if err != nil {
		server.Alive = false
		if r.RetryCount <= lb.MaxRetry {
			r.RetryCount++
			lb.serveHTTP(w, r)
		} else {
			server.Shiroxy.DefaultErrorHandler(w, r.Request, fmt.Errorf("dial up failed, host : %s", r.Request.RemoteAddr))
		}
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shiroxyRequest := &ShiroxyRequest{
		RetryCount: 0,
		Request:    r,
	}

	lb.serveHTTP(w, shiroxyRequest)
}
