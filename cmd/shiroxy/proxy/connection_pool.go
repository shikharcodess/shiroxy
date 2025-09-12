package proxy

import (
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)

// ConnectionPoolStats stores statistics about the HTTP connection pool
type ConnectionPoolStats struct {
	ActiveConnections      int           `json:"active_connections"`
	IdleConnections        int           `json:"idle_connections"`
	ConnectionsCreated     int64         `json:"connections_created"`
	ConnectionsClosed      int64         `json:"connections_closed"`
	ConnectionReuseCount   int64         `json:"connection_reuse_count"`
	TotalRequests          int64         `json:"total_requests"`
	AverageRequestDuration time.Duration `json:"average_request_duration"`
	LastUpdated            time.Time     `json:"last_updated"`

	mu sync.RWMutex
}

// NewConnectionPoolStats creates a new connection pool stats tracker
func NewConnectionPoolStats() *ConnectionPoolStats {
	return &ConnectionPoolStats{
		LastUpdated: time.Now(),
	}
}

// HTTP2ConnectionTracer adds HTTP/2 connection tracing to a request
func HTTP2ConnectionTracer(stats *ConnectionPoolStats, req *http.Request) *http.Request {
	trace := &httptrace.ClientTrace{
		// Track when a connection is created
		ConnectStart: func(network, addr string) {
			stats.mu.Lock()
			stats.ConnectionsCreated++
			stats.ActiveConnections++
			stats.mu.Unlock()
		},

		// Track when a connection is closed
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				stats.mu.Lock()
				stats.ConnectionsClosed++
				stats.ActiveConnections--
				stats.mu.Unlock()
			}
		},

		// Track when a connection is reused (important for connection pooling)
		GotConn: func(info httptrace.GotConnInfo) {
			stats.mu.Lock()
			if info.Reused {
				stats.ConnectionReuseCount++
			}
			stats.mu.Unlock()
		},

		// Track when a connection becomes idle
		PutIdleConn: func(err error) {
			stats.mu.Lock()
			if err == nil {
				stats.ActiveConnections--
				stats.IdleConnections++
			}
			stats.mu.Unlock()
		},
	}

	return req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
}

// RecordRequestCompletion updates the total request count and calculates average duration
func (s *ConnectionPoolStats) RecordRequestCompletion(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalRequests++
	// Simple moving average for request duration
	if s.AverageRequestDuration == 0 {
		s.AverageRequestDuration = duration
	} else {
		s.AverageRequestDuration = (s.AverageRequestDuration*time.Duration(s.TotalRequests-1) + duration) / time.Duration(s.TotalRequests)
	}
	s.LastUpdated = time.Now()
}

// GetStats returns a copy of the current statistics without the mutex
func (s *ConnectionPoolStats) GetStats() ConnectionPoolStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a new instance to avoid returning the mutex
	return ConnectionPoolStats{
		ActiveConnections:      s.ActiveConnections,
		IdleConnections:        s.IdleConnections,
		ConnectionsCreated:     s.ConnectionsCreated,
		ConnectionsClosed:      s.ConnectionsClosed,
		ConnectionReuseCount:   s.ConnectionReuseCount,
		TotalRequests:          s.TotalRequests,
		AverageRequestDuration: s.AverageRequestDuration,
		LastUpdated:            s.LastUpdated,
	}
}
