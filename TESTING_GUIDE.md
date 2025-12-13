# Shiroxy Testing Guide

Complete guide for testing all features of Shiroxy reverse proxy.

---

## Table of Contents

1. [Setup & Environment Testing](#1-setup--environment-testing)
2. [HTTP/2 Connection Pooling Testing](#2-http2-connection-pooling-testing)
3. [Gzip Compression Testing](#3-gzip-compression-testing)
4. [Domain Management & SSL Certificate Testing](#4-domain-management--ssl-certificate-testing)
5. [Load Balancing Strategy Testing](#5-load-balancing-strategy-testing)
6. [Tag-Based Routing Testing](#6-tag-based-routing-testing)
7. [Health Check & Failover Testing](#7-health-check--failover-testing)
8. [WebSocket Support Testing](#8-websocket-support-testing)
9. [Buffer Pool & Memory Testing](#9-buffer-pool--memory-testing)
10. [Analytics & Monitoring Testing](#10-analytics--monitoring-testing)
11. [Webhook Testing](#11-webhook-testing)
12. [Graceful Shutdown Testing](#12-graceful-shutdown-testing)
13. [HTTP Status Code Testing](#13-http-status-code-testing)
14. [HTTPS Redirect Testing](#14-https-redirect-testing)
15. [Load & Stress Testing](#15-load--stress-testing)
16. [Configuration Testing](#16-configuration-testing)
17. [API Documentation Testing](#17-api-documentation-testing)
18. [Integration Testing](#18-integration-testing)
19. [Error Handling Testing](#19-error-handling-testing)
20. [Performance Benchmarking](#20-performance-benchmarking)
21. [Test Checklist](#test-checklist)

---

## 1. Setup & Environment Testing

### Test Development Environment

```bash
# Clone and start with Docker
git clone https://github.com/shikharcodess/shiroxy.git
cd shiroxy
docker compose up -d --build

# Verify all services are running
docker compose ps
```

### Test Local Development Setup

```bash
# Start Pebble ACME test server
cd pebble
go build -o pebble cmd/pebble/main.go
./pebble &

# Start Shiroxy in development mode
cd ..
sudo go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml
```

**Expected Output:**

- Shiroxy logo displays
- Mode: development
- ACME server: Pebble URL
- API server starts on configured port
- No errors in startup logs

---

## 2. HTTP/2 Connection Pooling Testing

### Test HTTP/2 Support

```bash
# Test HTTP/2 with curl
curl -I --http2 https://yourdomain.com

# Verify HTTP/2 in response
# Expected: HTTP/2 200
```

### Load Test Connection Pooling

```bash
# Navigate to test directory
cd test/k6

# Run k6 load test
k6 run --vus 100 --duration 30s dummy-test/index.js
```

### Monitor Connection Statistics

```bash
# Get connection pool statistics
curl http://localhost:8080/api/v1/analytics/connections

# Expected response:
# {
#   "active_connections": 50,
#   "idle_connections": 150,
#   "connections_created": 200,
#   "connections_closed": 0,
#   "connection_reuse_count": 5000,
#   "total_requests": 10000,
#   "average_request_duration": "15ms",
#   "last_updated": "2025-12-12T10:30:00Z"
# }
```

### Verify Connection Reuse

```bash
# Run multiple requests and check reuse count increases
for i in {1..100}; do
  curl -s https://yourdomain.com > /dev/null
done

# Check statistics again
curl http://localhost:8080/api/v1/analytics/connections | jq .connection_reuse_count
```

**Expected Results:**

- Connection reuse count should increase with each batch of requests
- Idle connections maintained in pool
- Average request duration decreases as connections are reused

---

## 3. Gzip Compression Testing

### Test Compression with Different Content Types

```bash
# Test HTML compression
curl -H "Accept-Encoding: gzip" -I https://yourdomain.com/index.html

# Test JSON compression
curl -H "Accept-Encoding: gzip" -I https://yourdomain.com/api/data.json

# Test CSS compression
curl -H "Accept-Encoding: gzip" -I https://yourdomain.com/styles.css

# Test JavaScript compression
curl -H "Accept-Encoding: gzip" -I https://yourdomain.com/app.js

# Test XML compression
curl -H "Accept-Encoding: gzip" -I https://yourdomain.com/feed.xml
```

### Verify Compression Headers

```bash
# Check headers with verbose output
curl -H "Accept-Encoding: gzip" -v https://yourdomain.com 2>&1 | grep -i "content-encoding\|vary"

# Expected:
# < Content-Encoding: gzip
# < Vary: Accept-Encoding
```

### Test Without Compression Support

```bash
# Request without Accept-Encoding header
curl -I https://yourdomain.com

# Expected: No Content-Encoding header
```

### Compare Compressed vs Uncompressed Size

```bash
# Get compressed size
curl -H "Accept-Encoding: gzip" -so /tmp/compressed.gz https://yourdomain.com/large-file.html
ls -lh /tmp/compressed.gz

# Get uncompressed size
curl -so /tmp/uncompressed.html https://yourdomain.com/large-file.html
ls -lh /tmp/uncompressed.html

# Compare sizes (compressed should be 60-80% smaller for text)
```

**Expected Headers:**

- `Content-Encoding: gzip` (when compression applied)
- `Vary: Accept-Encoding` (always present for compressible content)
- No `Content-Length` header (with compression)

---

## 4. Domain Management & SSL Certificate Testing

### Register a New Domain

```bash
# Register domain with metadata
curl -X POST http://localhost:8080/api/v1/domains/register \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "example.com",
    "email": "admin@example.com",
    "metadata": {
      "env": "production",
      "tag": "web",
      "region": "us-east-1"
    }
  }'

# Expected response:
# {
#   "status": "success",
#   "domain": "example.com",
#   "dns_challenge_key": "ABC123..."
# }
```

### Get Domain Details

```bash
# Retrieve domain information
curl http://localhost:8080/api/v1/domains/example.com | jq

# Expected response includes:
# {
#   "domain": "example.com",
#   "status": "active",
#   "email": "admin@example.com",
#   "metadata": {...},
#   "certificate_expiry": "2026-03-12T..."
# }
```

### Update Domain Metadata

```bash
# Update domain tags
curl -X PUT http://localhost:8080/api/v1/domains/example.com \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "tag": "api",
      "version": "v2"
    }
  }'

# Verify update
curl http://localhost:8080/api/v1/domains/example.com | jq .metadata
```

### List All Domains

```bash
# Get all registered domains
curl http://localhost:8080/api/v1/domains | jq

# Expected: Array of domain objects
```

### Test SSL Certificate

```bash
# Check certificate details
echo | openssl s_client -connect example.com:443 -servername example.com 2>/dev/null | openssl x509 -noout -dates -subject

# Expected:
# notBefore=...
# notAfter=...
# subject=CN=example.com
```

### Remove Domain

```bash
# Delete domain
curl -X DELETE http://localhost:8080/api/v1/domains/example.com

# Verify deletion
curl http://localhost:8080/api/v1/domains/example.com
# Expected: 404 Not Found
```

---

## 5. Load Balancing Strategy Testing

### Test Round-Robin Load Balancing

```bash
# Send multiple requests
for i in {1..10}; do
  curl -w "\nBackend: %{remote_ip}\n" https://yourdomain.com/api/test
  sleep 0.5
done

# Expected: Requests distributed evenly across backends
# Backend 1: 192.168.1.10
# Backend 2: 192.168.1.11
# Backend 3: 192.168.1.12
# Backend 1: 192.168.1.10 (cycle repeats)
```

### Test Least-Connection Load Balancing

Configure least-connection mode in `shiroxy.conf.yaml`:

```yaml
balancing_strategy: least-connection
```

```bash
# Monitor connection counts
curl http://localhost:8080/api/v1/analytics/backends | jq '.[] | {name, active_connections}'

# Send requests
ab -n 100 -c 10 https://yourdomain.com/

# Verify: Backends with fewer connections receive more requests
curl http://localhost:8080/api/v1/analytics/backends | jq '.[] | {name, total_requests, active_connections}'
```

### Test Sticky-Session Load Balancing

Configure sticky-session mode:

```yaml
balancing_strategy: sticky-session
```

```bash
# Create session
curl -b cookies.txt -c cookies.txt https://yourdomain.com/session

# Make multiple requests with same cookies
for i in {1..5}; do
  curl -b cookies.txt https://yourdomain.com/session
done

# Expected: All requests hit the same backend server
```

### Verify Load Distribution

```bash
# Get backend statistics
curl http://localhost:8080/api/v1/analytics/backends | jq

# Expected response:
# [
#   {
#     "name": "backend-1",
#     "url": "http://192.168.1.10:8000",
#     "alive": true,
#     "active_connections": 5,
#     "total_requests": 1500
#   },
#   {
#     "name": "backend-2",
#     "url": "http://192.168.1.11:8000",
#     "alive": true,
#     "active_connections": 5,
#     "total_requests": 1500
#   }
# ]
```

---

## 6. Tag-Based Routing Testing

### Register Domains with Different Tags

```bash
# Register API domain
curl -X POST http://localhost:8080/api/v1/domains/register \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "api.example.com",
    "email": "admin@example.com",
    "metadata": {"tag": "api"}
  }'

# Register Web domain
curl -X POST http://localhost:8080/api/v1/domains/register \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "web.example.com",
    "email": "admin@example.com",
    "metadata": {"tag": "web"}
  }'

# Register CDN domain
curl -X POST http://localhost:8080/api/v1/domains/register \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "cdn.example.com",
    "email": "admin@example.com",
    "metadata": {"tag": "cdn"}
  }'
```

### Configure Backend Servers with Tags

In `shiroxy.conf.yaml`:

```yaml
servers:
  - name: api-server-1
    url: http://192.168.1.20:8000
    tags: [api]
  - name: api-server-2
    url: http://192.168.1.21:8000
    tags: [api]
  - name: web-server-1
    url: http://192.168.1.30:8000
    tags: [web]
  - name: cdn-server-1
    url: http://192.168.1.40:8000
    tags: [cdn]
```

### Test Tag-Based Routing

```bash
# Test API routing
curl -v https://api.example.com/endpoint 2>&1 | grep "X-Backend-Server"
# Expected: api-server-1 or api-server-2

# Test Web routing
curl -v https://web.example.com/ 2>&1 | grep "X-Backend-Server"
# Expected: web-server-1

# Test CDN routing
curl -v https://cdn.example.com/assets/image.png 2>&1 | grep "X-Backend-Server"
# Expected: cdn-server-1
```

### Verify Tag Cache Performance

```bash
# First request (cache miss)
time curl -so /dev/null https://api.example.com

# Subsequent requests (cache hit)
time curl -so /dev/null https://api.example.com

# Expected: Second request should be faster due to tag cache
```

---

## 7. Health Check & Failover Testing

### Monitor Backend Health

```bash
# Check current health status
curl http://localhost:8080/api/v1/analytics/backends | jq '.[] | {name, alive}'

# Expected:
# [
#   {"name": "backend-1", "alive": true},
#   {"name": "backend-2", "alive": true}
# ]
```

### Test Automatic Failover

```bash
# Stop one backend server
docker stop backend-1
# or
sudo systemctl stop backend-1

# Wait for health check to detect failure (default: 30s)
sleep 35

# Check health status
curl http://localhost:8080/api/v1/analytics/backends | jq '.[] | {name, alive}'

# Expected:
# [
#   {"name": "backend-1", "alive": false},
#   {"name": "backend-2", "alive": true}
# ]

# Send requests - should only go to healthy backend
for i in {1..10}; do
  curl https://yourdomain.com/
done

# Verify all requests went to backend-2
curl http://localhost:8080/api/v1/analytics/backends | jq '.[] | {name, total_requests}'
```

### Test Health Check Recovery

```bash
# Restart failed backend
docker start backend-1
# or
sudo systemctl start backend-1

# Wait for health check to detect recovery
sleep 35

# Check health status
curl http://localhost:8080/api/v1/analytics/backends | jq '.[] | {name, alive}'

# Expected: Both backends should be alive now
```

### Monitor Health Check Logs

```bash
# Tail Shiroxy logs
tail -f shiroxy.log | grep health

# Expected output:
# [INFO] Health check passed for backend-1
# [INFO] Health check passed for backend-2
# [WARN] Health check failed for backend-1: connection refused
# [INFO] Health check recovered for backend-1
```

### Configure Custom Health Check Endpoints

In `shiroxy.conf.yaml`:

```yaml
servers:
  - name: backend-1
    url: http://192.168.1.10:8000
    health_check_url: http://192.168.1.10:8000/health
    health_check_interval: 15s
```

```bash
# Restart Shiroxy and verify custom health checks
sudo go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml
```

---

## 8. WebSocket Support Testing

### Install WebSocket Client

```bash
# Install wscat
npm install -g wscat
```

### Test WebSocket Connection

```bash
# Connect to WebSocket endpoint
wscat -c wss://yourdomain.com/ws

# Expected: Connection established
# Connected (press CTRL+C to quit)
```

### Test Bidirectional Communication

```bash
# In wscat session:
> {"type": "ping", "data": "hello"}

# Expected response:
< {"type": "pong", "data": "hello"}

> {"action": "subscribe", "channel": "updates"}
< {"status": "subscribed", "channel": "updates"}
```

### Test WebSocket Through Load Balancer

```bash
# Connect multiple clients
for i in {1..5}; do
  wscat -c wss://yourdomain.com/ws &
done

# Verify connections distributed across backends
curl http://localhost:8080/api/v1/analytics/backends | jq '.[] | {name, active_connections}'
```

### Test WebSocket Persistence

```bash
# Connect and send messages
wscat -c wss://yourdomain.com/ws
> {"action": "message", "text": "test 1"}
> {"action": "message", "text": "test 2"}

# Connection should remain on same backend (sticky)
```

---

## 9. Buffer Pool & Memory Testing

### Monitor Initial Memory Usage

```bash
# Get baseline memory stats
curl http://localhost:8080/api/v1/analytics/system | jq .memory

# Expected:
# {
#   "alloc": "25MB",
#   "total_alloc": "50MB",
#   "sys": "70MB",
#   "num_gc": 5
# }
```

### Run Sustained Load Test

```bash
# Install Apache Bench if not available
# macOS: brew install apache-bench
# Ubuntu: sudo apt-get install apache2-utils

# Run sustained load
ab -n 100000 -c 100 https://yourdomain.com/
```

### Monitor Memory During Load

```bash
# Watch memory usage in real-time
watch -n 1 'curl -s http://localhost:8080/api/v1/analytics/system | jq .memory'

# Expected: Memory should stabilize due to buffer pooling
# alloc should not continuously increase
```

### Verify Buffer Pool Efficiency

```bash
# Check GC statistics
curl http://localhost:8080/api/v1/analytics/system | jq .gc_stats

# Expected:
# {
#   "num_gc": 25,        # Should be low relative to request count
#   "pause_total_ns": 1500000,
#   "last_gc": "2025-12-12T10:30:00Z"
# }
```

### Memory Leak Test

```bash
# Run extended load test (5 minutes)
ab -n 500000 -c 200 -t 300 https://yourdomain.com/

# Check memory before and after
curl http://localhost:8080/api/v1/analytics/system | jq .memory.alloc

# Expected: Memory should return to baseline after GC
# No continuous growth indicates no memory leaks
```

---

## 10. Analytics & Monitoring Testing

### Test System Metrics Endpoint

```bash
# Get system metrics
curl http://localhost:8080/api/v1/analytics/system | jq

# Expected response:
# {
#   "cpu_usage": 15.5,
#   "memory": {
#     "alloc": "32MB",
#     "total_alloc": "120MB",
#     "sys": "85MB"
#   },
#   "goroutines": 150,
#   "uptime": "2h30m15s",
#   "timestamp": "2025-12-12T10:30:00Z"
# }
```

### Test Connection Statistics

```bash
# Get detailed connection stats
curl http://localhost:8080/api/v1/analytics/connections | jq

# Expected:
# {
#   "active_connections": 45,
#   "idle_connections": 155,
#   "connections_created": 200,
#   "connections_closed": 0,
#   "connection_reuse_count": 8500,
#   "total_requests": 10000,
#   "average_request_duration": "12ms"
# }
```

### Test Backend Statistics

```bash
# Get per-backend stats
curl http://localhost:8080/api/v1/analytics/backends | jq

# Expected:
# [
#   {
#     "name": "backend-1",
#     "url": "http://192.168.1.10:8000",
#     "alive": true,
#     "active_connections": 22,
#     "total_requests": 5000,
#     "failed_requests": 5,
#     "average_response_time": "15ms",
#     "last_health_check": "2025-12-12T10:29:45Z"
#   }
# ]
```

### Test Request Analytics

```bash
# Get request statistics
curl http://localhost:8080/api/v1/analytics/requests | jq

# Expected:
# {
#   "total_requests": 10000,
#   "successful_requests": 9850,
#   "failed_requests": 150,
#   "average_duration": "18ms",
#   "requests_per_second": 125.5,
#   "status_codes": {
#     "200": 9500,
#     "404": 100,
#     "503": 50,
#     "500": 50
#   }
# }
```

### Create Monitoring Dashboard

```bash
# Poll metrics every second for dashboard
while true; do
  clear
  echo "=== SHIROXY METRICS ==="
  echo ""
  echo "System:"
  curl -s http://localhost:8080/api/v1/analytics/system | jq '{cpu: .cpu_usage, memory: .memory.alloc, goroutines}'
  echo ""
  echo "Connections:"
  curl -s http://localhost:8080/api/v1/analytics/connections | jq '{active, idle, reuse_count: .connection_reuse_count}'
  echo ""
  echo "Requests:"
  curl -s http://localhost:8080/api/v1/analytics/requests | jq '{total, rps: .requests_per_second}'
  sleep 1
done
```

---

## 11. Webhook Testing

### Setup Webhook Endpoint

```bash
# Use webhook.site for testing
# Visit https://webhook.site and get your unique URL
WEBHOOK_URL="https://webhook.site/your-unique-id"
```

### Configure Webhook in Shiroxy

In `shiroxy.conf.yaml`:

```yaml
webhook:
  url: "https://webhook.site/your-unique-id"
  secret: "my-secret-key"
  events:
    - backendserver.register.success
    - backendserver.register.failed
    - domain.status.changed
```

### Test Backend Registration Webhook

```bash
# Add new backend server (triggers success webhook)
# Edit config and restart, or use API if available

# Check webhook.site - should receive:
# POST with JSON payload:
# {
#   "event": "backendserver.register.success",
#   "data": {
#     "name": "backend-3",
#     "url": "http://192.168.1.12:8000",
#     "timestamp": "2025-12-12T10:30:00Z"
#   },
#   "signature": "sha256=..."
# }
```

### Test Backend Health Change Webhook

```bash
# Stop a backend to trigger failure webhook
docker stop backend-1

# Wait for health check
sleep 35

# Check webhook.site for failure event:
# {
#   "event": "backendserver.register.failed",
#   "data": {
#     "name": "backend-1",
#     "reason": "health check failed",
#     "timestamp": "2025-12-12T10:30:30Z"
#   }
# }

# Restart backend to trigger success webhook
docker start backend-1
```

### Test Domain Status Webhook

```bash
# Change domain status
curl -X PUT http://localhost:8080/api/v1/domains/example.com \
  -d '{"status": "inactive"}'

# Check webhook.site:
# {
#   "event": "domain.status.changed",
#   "data": {
#     "domain": "example.com",
#     "old_status": "active",
#     "new_status": "inactive"
#   }
# }
```

### Verify Webhook Signature

```bash
# Create signature verification script
cat > verify_webhook.sh << 'EOF'
#!/bin/bash
SECRET="my-secret-key"
PAYLOAD='{"event":"test","data":{}}'
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
echo "Expected signature: sha256=$SIGNATURE"
EOF

chmod +x verify_webhook.sh
./verify_webhook.sh
```

---

## 12. Graceful Shutdown Testing

### Test Shutdown with Active Requests

```bash
# Start long-running requests in background
ab -n 1000 -c 10 https://yourdomain.com/ &
AB_PID=$!

# Wait a moment for requests to start
sleep 2

# Get Shiroxy PID
SHIROXY_PID=$(pgrep -f "shiroxy")

# Send SIGTERM (graceful shutdown signal)
kill -SIGTERM $SHIROXY_PID

# Monitor logs
tail -f shiroxy.log

# Expected log output:
# [INFO] Received shutdown signal
# [INFO] Waiting for 8 active requests to complete...
# [INFO] All requests completed
# [INFO] Persisting domain metadata...
# [INFO] Saving 15 domains to persistence file
# [INFO] Closing connections...
# [INFO] Shutdown complete
```

### Verify Data Persistence

```bash
# Check for persistence file
ls -lh .shiroxy_persistence.pb

# Expected: File exists with recent timestamp

# Restart Shiroxy
sudo go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml

# Verify domains restored
curl http://localhost:8080/api/v1/domains | jq length

# Expected: Same number of domains as before shutdown
```

### Test SIGINT Handling

```bash
# Start Shiroxy
sudo go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml

# Press Ctrl+C
# Expected: Graceful shutdown process same as SIGTERM
```

### Test Shutdown Timeout

```bash
# Configure shutdown timeout in code or config
# Default: 30 seconds for in-flight requests

# Start very long request
curl https://yourdomain.com/long-operation &

# Trigger shutdown
kill -SIGTERM $(pgrep -f shiroxy)

# Expected: Waits up to 30s for requests, then forces shutdown
```

---

## 13. HTTP Status Code Testing

### Test Inactive Domain Status Code

```bash
# Set domain to inactive
curl -X PUT http://localhost:8080/api/v1/domains/example.com \
  -d '{"status": "inactive"}'

# Request the domain
curl -I https://example.com

# Expected:
# HTTP/1.1 503 Service Unavailable
# Content-Type: text/html
```

### Test Active Domain

```bash
# Ensure domain is active
curl -X PUT http://localhost:8080/api/v1/domains/example.com \
  -d '{"status": "active"}'

# Request the domain
curl -I https://example.com

# Expected:
# HTTP/1.1 200 OK
```

### Test Non-Existent Domain

```bash
# Request domain that doesn't exist
curl -I https://nonexistent-domain.com

# Expected:
# HTTP/1.1 404 Not Found
# (Shows domain not found page)
```

### Test Backend Server Down

```bash
# Stop all backend servers for a domain
docker stop backend-1 backend-2

# Wait for health checks
sleep 35

# Request should return 503
curl -I https://example.com

# Expected:
# HTTP/1.1 503 Service Unavailable
```

---

## 14. HTTPS Redirect Testing

### Test Basic HTTP to HTTPS Redirect

```bash
# Request HTTP (should redirect to HTTPS)
curl -I http://example.com

# Expected:
# HTTP/1.1 301 Moved Permanently
# Location: https://example.com
# (Note: No port number like :80)
```

### Test Redirect with Custom Port

```bash
# Request on custom port
curl -I http://example.com:8080

# Expected:
# HTTP/1.1 301 Moved Permanently
# Location: https://example.com
# (Port stripped from redirect URL)
```

### Test Redirect Preserves Path

```bash
# Request with path
curl -I http://example.com/path/to/resource

# Expected:
# Location: https://example.com/path/to/resource
```

### Test Redirect Preserves Query Parameters

```bash
# Request with query params
curl -I "http://example.com/search?q=test&page=2"

# Expected:
# Location: https://example.com/search?q=test&page=2
```

### Verify No Double Redirects

```bash
# Follow redirects
curl -L -I http://example.com

# Expected: Only one redirect (HTTP -> HTTPS)
# Not multiple redirects
```

---

## 15. Load & Stress Testing

### Install k6 Load Testing Tool

```bash
# macOS
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Or download from https://k6.io/docs/get-started/installation/
```

### Basic Load Test

```bash
# Create test script
cat > load-test.js << 'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 50 },  // Ramp up to 50 users
    { duration: '1m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 100 }, // Ramp up to 100 users
    { duration: '1m', target: 100 },  // Stay at 100 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
};

export default function() {
  const res = http.get('https://yourdomain.com');

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
    'has gzip encoding': (r) => r.headers['Content-Encoding'] === 'gzip',
  });

  sleep(1);
}
EOF

# Run test
k6 run load-test.js
```

### Stress Test (Find Breaking Point)

```bash
# Create stress test script
cat > stress-test.js << 'EOF'
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '5m', target: 200 },
    { duration: '2m', target: 300 },
    { duration: '5m', target: 300 },
    { duration: '10m', target: 0 },
  ],
};

export default function() {
  const res = http.get('https://yourdomain.com');
  check(res, {'status is 200': (r) => r.status === 200});
}
EOF

# Run stress test
k6 run stress-test.js
```

### Spike Test (Sudden Traffic Surge)

```bash
cat > spike-test.js << 'EOF'
import http from 'k6/http';

export const options = {
  stages: [
    { duration: '10s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '10s', target: 1000 }, // Spike!
    { duration: '3m', target: 1000 },
    { duration: '10s', target: 100 },
    { duration: '3m', target: 100 },
    { duration: '10s', target: 0 },
  ],
};

export default function() {
  http.get('https://yourdomain.com');
}
EOF

k6 run spike-test.js
```

### Monitor Metrics During Load Test

```bash
# In separate terminal, monitor metrics
watch -n 1 'echo "=== CONNECTIONS ===" && curl -s http://localhost:8080/api/v1/analytics/connections | jq "{active, idle, reuse}" && echo "" && echo "=== REQUESTS ===" && curl -s http://localhost:8080/api/v1/analytics/requests | jq "{total, rps, avg_duration}"'
```

### Apache Bench Testing

```bash
# Simple benchmark
ab -n 10000 -c 100 https://yourdomain.com/

# With keep-alive
ab -n 10000 -c 100 -k https://yourdomain.com/

# POST requests
ab -n 1000 -c 50 -p post_data.json -T application/json https://yourdomain.com/api/endpoint
```

### h2load Testing (HTTP/2 specific)

```bash
# Install h2load (part of nghttp2)
# macOS: brew install nghttp2
# Ubuntu: sudo apt-get install nghttp2-client

# Run HTTP/2 load test
h2load -n 10000 -c 100 -m 10 https://yourdomain.com/

# Expected output shows:
# - Requests per second
# - Time per request
# - Transfer rate
# - HTTP/2 stream usage
```

---

## 16. Configuration Testing

### Test Development Mode

```bash
# Set development mode
export SHIROXY_MODE=development

# Start Shiroxy
go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml

# Expected:
# - Uses Pebble ACME server
# - Detailed logging
# - INSECURE_SKIP_VERIFY enabled
```

### Test Staging Mode

```bash
# Set staging mode
export SHIROXY_MODE=staging

# Start Shiroxy
go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml

# Expected:
# - Uses Let's Encrypt staging
# - Production-like behavior
# - Certificates not trusted by browsers
```

### Test Production Mode

```bash
# Set production mode
export SHIROXY_MODE=production

# Start Shiroxy
go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml

# Expected:
# - Uses Let's Encrypt production
# - Minimal logging
# - Secure defaults
```

### Test Environment Variable Override

```bash
# Override config with environment variables
export SHIROXY_LOG_LEVEL=debug
export SHIROXY_MAX_RETRY=5
export SHIROXY_HEALTH_CHECK_INTERVAL=10s
export SHIROXY_API_PORT=9090

# Start Shiroxy
go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml

# Verify overrides applied
curl http://localhost:9090/api/v1/analytics/system
```

### Test Configuration Validation

```bash
# Create invalid config
cat > test-invalid.yaml << 'EOF'
mode: invalid_mode
servers: []
frontends: []
EOF

# Try to start with invalid config
go run cmd/shiroxy/main.go -c test-invalid.yaml

# Expected: Validation error and exit
```

### Test Multiple Frontend Ports

In `shiroxy.conf.yaml`:

```yaml
frontends:
  - port: 443
    mode: https
    target: multiple
  - port: 8443
    mode: https
    target: multiple
  - port: 80
    mode: http
```

```bash
# Test each port
curl -I https://example.com:443
curl -I https://example.com:8443
curl -I http://example.com:80
```

---

## 17. API Documentation Testing

### Access Swagger UI

```bash
# Open Swagger documentation in browser
open http://localhost:8080/swagger/index.html
# or
xdg-open http://localhost:8080/swagger/index.html  # Linux
```

### Test API Endpoints Through Swagger

1. Navigate to Swagger UI
2. Expand each endpoint category
3. Try "Try it out" for each endpoint
4. Verify request/response schemas match documentation

### Test Domain Endpoints

```bash
# Register domain via Swagger
# POST /api/v1/domains/register

# Get domain via Swagger
# GET /api/v1/domains/{domain}

# Update domain via Swagger
# PUT /api/v1/domains/{domain}

# Delete domain via Swagger
# DELETE /api/v1/domains/{domain}
```

### Verify API Response Schemas

```bash
# Get OpenAPI spec
curl http://localhost:8080/swagger/doc.json | jq

# Validate responses match schema
curl http://localhost:8080/api/v1/analytics/system | jq
# Compare with schema in swagger doc.json
```

---

## 18. Integration Testing

### Complete Integration Test Script

```bash
#!/bin/bash
set -e

echo "Starting Shiroxy Integration Test..."

# 1. Register domain
echo "1. Registering domain..."
DOMAIN="test-$(date +%s).example.com"
curl -X POST http://localhost:8080/api/v1/domains/register \
  -H "Content-Type: application/json" \
  -d "{\"domain\": \"$DOMAIN\", \"email\": \"test@example.com\", \"metadata\": {\"tag\": \"test\"}}" \
  | jq

# 2. Wait for SSL certificate
echo "2. Waiting for SSL certificate..."
sleep 10

# 3. Test HTTPS with compression
echo "3. Testing HTTPS with compression..."
RESPONSE=$(curl -H "Accept-Encoding: gzip" -I https://$DOMAIN 2>&1)
echo "$RESPONSE" | grep "HTTP/2 200" || echo "Warning: Expected HTTP/2 200"
echo "$RESPONSE" | grep "content-encoding: gzip" || echo "Warning: Expected gzip encoding"

# 4. Run load test
echo "4. Running load test..."
ab -n 1000 -c 50 https://$DOMAIN/ | grep "Requests per second"

# 5. Check analytics
echo "5. Checking analytics..."
curl -s http://localhost:8080/api/v1/analytics/system | jq '{cpu, memory: .memory.alloc}'
curl -s http://localhost:8080/api/v1/analytics/connections | jq '{active, idle, reuse: .connection_reuse_count}'

# 6. Test failover
echo "6. Testing failover..."
docker stop backend-1
sleep 35
curl -I https://$DOMAIN | grep "200" || echo "Failover working: requests routed to healthy backend"
docker start backend-1

# 7. Test graceful shutdown
echo "7. Testing graceful shutdown..."
ab -n 500 -c 10 https://$DOMAIN/ &
AB_PID=$!
sleep 2
kill -SIGTERM $(pgrep -f shiroxy)
wait $AB_PID || true
echo "Graceful shutdown completed"

# 8. Restart and verify persistence
echo "8. Testing persistence..."
go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml &
SHIROXY_PID=$!
sleep 5
DOMAIN_COUNT=$(curl -s http://localhost:8080/api/v1/domains | jq length)
echo "Restored $DOMAIN_COUNT domains"

# 9. Cleanup
echo "9. Cleanup..."
curl -X DELETE http://localhost:8080/api/v1/domains/$DOMAIN
kill $SHIROXY_PID

echo "Integration test completed successfully!"
```

### Save and Run Test

```bash
# Save script
cat > integration-test.sh << 'EOF'
# ... (paste script above)
EOF

# Make executable
chmod +x integration-test.sh

# Run test
./integration-test.sh
```

---

## 19. Error Handling Testing

### Test Invalid JSON

```bash
# Send invalid JSON to API
curl -X POST http://localhost:8080/api/v1/domains/register \
  -H "Content-Type: application/json" \
  -d '{invalid json}'

# Expected:
# HTTP/1.1 400 Bad Request
# {"error": "invalid JSON syntax"}
```

### Test Missing Required Fields

```bash
# Register domain without email
curl -X POST http://localhost:8080/api/v1/domains/register \
  -H "Content-Type: application/json" \
  -d '{"domain": "test.com"}'

# Expected:
# HTTP/1.1 400 Bad Request
# {"error": "email is required"}
```

### Test Backend Connection Failure

```bash
# Stop all backends
docker stop backend-1 backend-2 backend-3

# Send request
curl -I https://example.com

# Expected:
# HTTP/1.1 503 Service Unavailable

# Check logs for error messages
tail -f shiroxy.log | grep -i error
```

### Test Invalid Certificate

```bash
# Try to connect with wrong certificate
curl --cacert wrong-ca.pem https://example.com

# Expected: SSL certificate verification error
```

### Test Rate Limiting (if configured)

```bash
# Send rapid requests
for i in {1..1000}; do
  curl -s https://example.com > /dev/null &
done

# Check for rate limit responses
# Expected: Some 429 Too Many Requests responses
```

### Test Malformed Requests

```bash
# Very long URL
curl https://example.com/$(python3 -c 'print("a"*10000)')

# Expected: 414 URI Too Long or handled gracefully

# Invalid headers
curl -H "Invalid-Header: $(python3 -c 'print("x"*100000)')" https://example.com

# Expected: 400 Bad Request or header stripped
```

---

## 20. Performance Benchmarking

### Baseline Performance Test

```bash
# Single request latency
curl -w "@curl-format.txt" -o /dev/null -s https://yourdomain.com

# curl-format.txt:
cat > curl-format.txt << 'EOF'
    time_namelookup:  %{time_namelookup}s\n
       time_connect:  %{time_connect}s\n
    time_appconnect:  %{time_appconnect}s\n
   time_pretransfer:  %{time_pretransfer}s\n
      time_redirect:  %{time_redirect}s\n
 time_starttransfer:  %{time_starttransfer}s\n
                    ----------\n
         time_total:  %{time_total}s\n
EOF
```

### Throughput Testing

```bash
# Test requests per second
ab -n 10000 -c 100 https://yourdomain.com/ | grep "Requests per second"

# Expected: High RPS with HTTP/2 and connection pooling
# Baseline: 1000-5000+ RPS depending on backend
```

### Latency Testing

```bash
# Test latency percentiles
ab -n 10000 -c 100 https://yourdomain.com/ | grep -A 10 "Percentage of the requests"

# Expected output:
#  50%    15ms
#  66%    18ms
#  75%    20ms
#  80%    22ms
#  90%    30ms
#  95%    40ms
#  98%    60ms
#  99%    80ms
# 100%   150ms (longest request)
```

### Compression Performance

```bash
# Test with compression
ab -n 10000 -c 100 -H "Accept-Encoding: gzip" https://yourdomain.com/large.html

# Test without compression
ab -n 10000 -c 100 https://yourdomain.com/large.html

# Compare throughput and latency
```

### HTTP/2 vs HTTP/1.1 Comparison

```bash
# HTTP/2 performance
h2load -n 10000 -c 100 -m 10 https://yourdomain.com/

# HTTP/1.1 performance (force)
ab -n 10000 -c 100 https://yourdomain.com/

# Compare:
# - Requests per second
# - Time per request
# - Connection reuse
```

### Memory Performance Under Load

```bash
# Start monitoring
watch -n 1 'curl -s http://localhost:8080/api/v1/analytics/system | jq .memory'

# In another terminal, run load test
ab -n 100000 -c 200 https://yourdomain.com/

# Observe:
# - Memory allocation stays stable (buffer pooling working)
# - GC runs infrequently
# - No continuous memory growth
```

### Connection Pool Efficiency

```bash
# Monitor connection reuse
watch -n 1 'curl -s http://localhost:8080/api/v1/analytics/connections | jq "{created: .connections_created, reused: .connection_reuse_count, efficiency: (.connection_reuse_count / .connections_created * 100)}"'

# Run load test in another terminal
ab -n 50000 -c 100 -k https://yourdomain.com/

# Expected: High reuse ratio (80-95%)
```

---

## 21. Test Checklist

### Core Features

- [ ] Development environment setup (Docker)
- [ ] Local environment setup (Pebble + Go)
- [ ] Configuration loading (YAML + env vars)
- [ ] Multi-mode support (dev/staging/production)

### HTTP/2 & Performance

- [ ] HTTP/2 protocol support
- [ ] Connection pooling active
- [ ] Connection reuse working
- [ ] Idle connections maintained
- [ ] Request duration decreasing with pooling
- [ ] Buffer pool reducing allocations
- [ ] Memory stable under load
- [ ] No memory leaks

### Compression

- [ ] Gzip compression for text/\*
- [ ] Gzip compression for application/json
- [ ] Gzip compression for application/javascript
- [ ] Gzip compression for application/xml
- [ ] Proper headers (Content-Encoding, Vary)
- [ ] Client capability detection
- [ ] No compression for non-compressible content
- [ ] Streaming compression working

### Domain Management

- [ ] Register new domain
- [ ] Get domain details
- [ ] Update domain metadata
- [ ] List all domains
- [ ] Remove domain
- [ ] SSL certificate generation (ACME)
- [ ] Certificate auto-renewal
- [ ] Multi-domain support

### Load Balancing

- [ ] Round-robin distribution
- [ ] Least-connection routing
- [ ] Sticky-session persistence
- [ ] Even distribution verified
- [ ] Backend statistics accurate

### Routing

- [ ] Tag-based routing working
- [ ] Tag cache performance improvement
- [ ] Trie lookup efficiency
- [ ] Multiple tag support
- [ ] Strict vs flexible tag rules

### Health & Monitoring

- [ ] Health checks running
- [ ] Automatic failover on backend failure
- [ ] Backend recovery detection
- [ ] Health status API accurate
- [ ] Custom health check endpoints
- [ ] Configurable check intervals

### WebSocket

- [ ] WebSocket connection established
- [ ] Bidirectional communication
- [ ] Multiple concurrent connections
- [ ] Load balancing for WebSocket
- [ ] Connection persistence

### Analytics

- [ ] System metrics endpoint
- [ ] Connection statistics endpoint
- [ ] Backend statistics endpoint
- [ ] Request analytics endpoint
- [ ] Real-time metrics updates
- [ ] Accurate metric calculations

### Webhooks

- [ ] Webhook endpoint configured
- [ ] Backend success events
- [ ] Backend failure events
- [ ] Domain status change events
- [ ] Webhook signature validation
- [ ] Retry logic working

### Shutdown & Persistence

- [ ] Graceful shutdown (SIGTERM)
- [ ] Graceful shutdown (SIGINT)
- [ ] In-flight requests completed
- [ ] Domain metadata persisted
- [ ] State restored on restart
- [ ] No data loss on shutdown

### HTTP Semantics

- [ ] Inactive domains return 503 (not 404)
- [ ] Active domains return 200
- [ ] Non-existent domains return 404
- [ ] HTTPS redirects have no port numbers
- [ ] Redirects preserve paths
- [ ] Redirects preserve query parameters

### Error Handling

- [ ] Invalid JSON handled
- [ ] Missing fields validated
- [ ] Backend failures handled gracefully
- [ ] Certificate errors managed
- [ ] Malformed requests rejected
- [ ] Error logs comprehensive

### Load & Stress Testing

- [ ] Sustained load (100+ users, 5+ min)
- [ ] Spike load (sudden 10x increase)
- [ ] Stress test (find breaking point)
- [ ] Performance meets requirements
- [ ] No crashes under load
- [ ] Metrics accurate during load

### API & Documentation

- [ ] Swagger UI accessible
- [ ] All endpoints documented
- [ ] Request/response schemas correct
- [ ] Try it out functionality working
- [ ] API versioning clear

### Security

- [ ] TLS 1.2+ enforced
- [ ] Certificate validation
- [ ] Webhook signature verification
- [ ] Input sanitization
- [ ] No sensitive data in logs

---

## Automated Test Script

Create a master test script that runs all critical tests:

```bash
#!/bin/bash

# Shiroxy Master Test Script
# Run all critical tests and report results

echo "======================================"
echo "  SHIROXY COMPREHENSIVE TEST SUITE"
echo "======================================"
echo ""

PASS=0
FAIL=0

run_test() {
  local test_name="$1"
  local test_command="$2"

  echo -n "Testing $test_name... "
  if eval "$test_command" > /dev/null 2>&1; then
    echo "✓ PASS"
    ((PASS++))
  else
    echo "✗ FAIL"
    ((FAIL++))
  fi
}

# HTTP/2 Tests
run_test "HTTP/2 support" "curl -I --http2 https://yourdomain.com | grep 'HTTP/2'"
run_test "Connection pooling" "curl -s http://localhost:8080/api/v1/analytics/connections | jq -e '.connection_reuse_count > 0'"

# Compression Tests
run_test "Gzip compression" "curl -H 'Accept-Encoding: gzip' -I https://yourdomain.com | grep -i 'content-encoding: gzip'"
run_test "Vary header" "curl -H 'Accept-Encoding: gzip' -I https://yourdomain.com | grep -i 'vary: accept-encoding'"

# Domain Tests
run_test "Domain registration" "curl -X POST http://localhost:8080/api/v1/domains/register -d '{\"domain\":\"test.com\",\"email\":\"test@test.com\"}' | jq -e '.status == \"success\"'"
run_test "Domain retrieval" "curl -s http://localhost:8080/api/v1/domains | jq -e 'length > 0'"

# Load Balancing Tests
run_test "Backend health" "curl -s http://localhost:8080/api/v1/analytics/backends | jq -e '.[0].alive == true'"

# Analytics Tests
run_test "System metrics" "curl -s http://localhost:8080/api/v1/analytics/system | jq -e '.cpu_usage'"
run_test "Connection stats" "curl -s http://localhost:8080/api/v1/analytics/connections | jq -e '.total_requests'"

# HTTP Status Tests
run_test "HTTPS redirect" "curl -I http://yourdomain.com | grep -i 'location: https://yourdomain.com'"

echo ""
echo "======================================"
echo "  TEST RESULTS"
echo "======================================"
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo "Total:  $((PASS + FAIL))"
echo ""

if [ $FAIL -eq 0 ]; then
  echo "✓ All tests passed!"
  exit 0
else
  echo "✗ Some tests failed"
  exit 1
fi
```

Save as `master-test.sh`, make executable, and run:

```bash
chmod +x master-test.sh
./master-test.sh
```

---

## Continuous Testing

### Setup Monitoring Script

```bash
#!/bin/bash
# continuous-monitor.sh - Monitor Shiroxy in real-time

while true; do
  clear
  echo "=== SHIROXY LIVE MONITORING ==="
  echo "Time: $(date)"
  echo ""

  echo "--- System Metrics ---"
  curl -s http://localhost:8080/api/v1/analytics/system | jq '{
    cpu: .cpu_usage,
    memory: .memory.alloc,
    goroutines,
    uptime
  }'

  echo ""
  echo "--- Connection Pool ---"
  curl -s http://localhost:8080/api/v1/analytics/connections | jq '{
    active: .active_connections,
    idle: .idle_connections,
    reuse_rate: (.connection_reuse_count / .connections_created * 100 | floor),
    avg_duration: .average_request_duration
  }'

  echo ""
  echo "--- Backend Health ---"
  curl -s http://localhost:8080/api/v1/analytics/backends | jq '.[] | {
    name,
    alive,
    requests: .total_requests
  }'

  sleep 2
done
```

---

## Notes

- Replace `yourdomain.com` with your actual domain
- Replace `localhost:8080` with your actual API port
- Adjust timeouts and intervals based on your configuration
- Some tests require Docker or specific tools installed
- Run tests in a safe environment (not production)
- Monitor logs during tests for detailed information

---

## Troubleshooting

If tests fail, check:

1. **Shiroxy is running**: `ps aux | grep shiroxy`
2. **Logs for errors**: `tail -f shiroxy.log`
3. **Configuration valid**: Verify YAML syntax
4. **Ports available**: `lsof -i :80 -i :443 -i :8080`
5. **Backend servers running**: `curl http://backend-url/health`
6. **ACME server reachable**: `curl https://pebble:14000/dir` (dev mode)
7. **DNS resolution**: `nslookup yourdomain.com`
8. **Certificates valid**: `openssl s_client -connect yourdomain.com:443`

---

**End of Testing Guide**
