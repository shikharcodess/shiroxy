package proxy

import (
	"net/http"
	"testing"
	"time"
)

func TestConnectionPoolStats_RecordRequestCompletion(t *testing.T) {
	s := NewConnectionPoolStats()
	if s.TotalRequests != 0 {
		t.Fatalf("expected 0 requests")
	}

	s.RecordRequestCompletion(100 * time.Millisecond)
	if s.TotalRequests != 1 {
		t.Fatalf("expected 1 request")
	}

	s.RecordRequestCompletion(200 * time.Millisecond)
	if s.TotalRequests != 2 {
		t.Fatalf("expected 2 requests")
	}
	if s.AverageRequestDuration == 0 {
		t.Fatalf("average should be non-zero")
	}

	// Ensure GetStats returns a copy
	copy := s.GetStats()
	copy.TotalRequests = 9999
	if s.TotalRequests == 9999 {
		t.Fatalf("GetStats returned reference, expected copy")
	}

	// Sanity test for HTTP2ConnectionTracer - ensure no panic when adding trace
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req2 := HTTP2ConnectionTracer(s, req)
	if req2 == nil {
		t.Fatalf("expected traced request")
	}
}
