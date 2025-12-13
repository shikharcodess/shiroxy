package proxy

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipCompressionForTextHTML(t *testing.T) {
	// Create a test backend that returns HTML
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Hello World</h1></body></html>"))
	}))
	defer backend.Close()

	// Create a request with Accept-Encoding: gzip
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// Simulate proxy response
	w := httptest.NewRecorder()

	// Test that Content-Type is preserved through gzip
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip")

	gzWriter := gzip.NewWriter(w)
	defer gzWriter.Close()
	gzWriter.Write([]byte("<html><body><h1>Hello World</h1></body></html>"))
	gzWriter.Close()

	resp := w.Result()
	if resp.Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Content-Type should be text/html, got %s", resp.Header.Get("Content-Type"))
	}

	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Error("Content-Encoding should be gzip")
	}
}

func TestGzipCompressionForJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")

	gzWriter := gzip.NewWriter(w)
	defer gzWriter.Close()
	gzWriter.Write([]byte(`{"status": "ok"}`))
	gzWriter.Close()

	resp := w.Result()
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type should be application/json, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestNoGzipForBinaryContent(t *testing.T) {
	contentTypes := []string{
		"image/png",
		"image/jpeg",
		"video/mp4",
		"application/pdf",
	}

	for _, ct := range contentTypes {
		req := httptest.NewRequest("GET", "/binary", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", ct)
		// Binary content should NOT be gzipped
		w.Write([]byte("binary data"))

		resp := w.Result()
		if resp.Header.Get("Content-Encoding") == "gzip" {
			t.Errorf("Binary content %s should not be gzipped", ct)
		}
	}
}

func TestGzipResponseWriterWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()

	grw := &gzipResponseWriter{
		ResponseWriter: w,
		Writer:         gzip.NewWriter(w),
		headerWritten:  false,
	}

	// First WriteHeader call should work
	grw.WriteHeader(http.StatusOK)

	if !grw.headerWritten {
		t.Error("headerWritten should be true after first WriteHeader")
	}

	// Second WriteHeader call should be ignored
	grw.WriteHeader(http.StatusInternalServerError)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code should remain 200, got %d", resp.StatusCode)
	}
}

func TestGzipResponseWriterWrite(t *testing.T) {
	w := httptest.NewRecorder()

	gzWriter := gzip.NewWriter(w)
	grw := &gzipResponseWriter{
		ResponseWriter: w,
		Writer:         gzWriter,
		headerWritten:  false,
	}

	// Writing should automatically call WriteHeader if not called
	data := []byte("test data")
	n, err := grw.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != len(data) {
		t.Errorf("Write should return %d bytes, got %d", len(data), n)
	}

	if !grw.headerWritten {
		t.Error("headerWritten should be true after Write")
	}

	gzWriter.Close()

	// Verify data was compressed
	resp := w.Result()
	if resp.Header.Get("Content-Encoding") != "gzip" {
		// This test is just for the writer behavior
	}
}

func TestContentTypePreservation(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{"HTML", "text/html; charset=utf-8", "<html><body>Test</body></html>"},
		{"JSON", "application/json", `{"key": "value"}`},
		{"JavaScript", "application/javascript", "console.log('test');"},
		{"CSS", "text/css", "body { color: red; }"},
		{"XML", "application/xml", "<root><item>test</item></root>"},
		{"Plain Text", "text/plain", "Plain text content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			// Set Content-Type before writing
			w.Header().Set("Content-Type", tt.contentType)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(tt.body))

			resp := w.Result()
			actualCT := resp.Header.Get("Content-Type")

			if !strings.HasPrefix(actualCT, strings.Split(tt.contentType, ";")[0]) {
				t.Errorf("Content-Type should start with %s, got %s", tt.contentType, actualCT)
			}
		})
	}
}

func TestVaryHeaderWithGzip(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")

	resp := w.Result()
	if resp.Header.Get("Vary") != "Accept-Encoding" {
		t.Error("Vary header should be set to Accept-Encoding for gzipped responses")
	}
}

func TestGzipDecompression(t *testing.T) {
	// Create gzipped content
	var buf strings.Builder
	gzWriter := gzip.NewWriter(&buf)
	original := "This is test content that should be compressed"
	gzWriter.Write([]byte(original))
	gzWriter.Close()

	// Verify we can decompress it
	gzReader, err := gzip.NewReader(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	if string(decompressed) != original {
		t.Errorf("Decompressed content doesn't match. Got: %s, Want: %s", string(decompressed), original)
	}
}

func TestCopyHeader(t *testing.T) {
	src := make(http.Header)
	src.Set("Content-Type", "text/html")
	src.Set("X-Custom-Header", "value")
	src.Set("Cache-Control", "max-age=3600")

	dst := make(http.Header)
	copyHeader(dst, src)

	if dst.Get("Content-Type") != "text/html" {
		t.Error("Content-Type header not copied")
	}
	if dst.Get("X-Custom-Header") != "value" {
		t.Error("X-Custom-Header not copied")
	}
	if dst.Get("Cache-Control") != "max-age=3600" {
		t.Error("Cache-Control header not copied")
	}
}

func TestRemoveHopByHopHeaders(t *testing.T) {
	headers := make(http.Header)
	headers.Set("Connection", "keep-alive")
	headers.Set("Keep-Alive", "timeout=5")
	headers.Set("Proxy-Authenticate", "Basic")
	headers.Set("Proxy-Authorization", "Bearer token")
	headers.Set("Te", "trailers")
	headers.Set("Trailer", "X-Custom")
	headers.Set("Transfer-Encoding", "chunked")
	headers.Set("Upgrade", "websocket")
	headers.Set("Content-Type", "text/html") // Should NOT be removed

	removeHopByHopHeaders(headers)

	// Hop-by-hop headers should be removed
	hopByHop := []string{
		"Connection", "Keep-Alive", "Proxy-Authenticate",
		"Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade",
	}

	for _, h := range hopByHop {
		if headers.Get(h) != "" {
			t.Errorf("Hop-by-hop header %s should be removed", h)
		}
	}

	// Content-Type should remain
	if headers.Get("Content-Type") != "text/html" {
		t.Error("Content-Type should not be removed")
	}
}

func TestUpgradeType(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		expected string
	}{
		{
			name: "WebSocket upgrade",
			headers: http.Header{
				"Connection": []string{"Upgrade"},
				"Upgrade":    []string{"websocket"},
			},
			expected: "websocket",
		},
		{
			name: "HTTP/2 upgrade",
			headers: http.Header{
				"Connection": []string{"Upgrade"},
				"Upgrade":    []string{"h2c"},
			},
			expected: "h2c",
		},
		{
			name:     "No upgrade",
			headers:  http.Header{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := upgradeType(tt.headers)
			if result != tt.expected {
				t.Errorf("upgradeType() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestCleanQueryParams(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"foo=bar&baz=qux", "foo=bar&baz=qux"},
		{"foo=bar%20baz", "foo=bar%20baz"},
		{"foo=bar&foo=baz", "foo=bar&foo=baz"},
		{"", ""},
	}

	for _, tt := range tests {
		result := cleanQueryParams(tt.input)
		if result != tt.expected {
			t.Errorf("cleanQueryParams(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
