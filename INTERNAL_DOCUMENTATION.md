# Shiroxy - Internal Documentation

**Version:** 1.1.0 (Kuchii Release)  
**Author:** @ShikharY10  
**Last Updated:** December 12, 2025

---

## Table of Contents

1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [Core Components](#core-components)
4. [Directory Structure Details](#directory-structure-details)
5. [Module Documentation](#module-documentation)
6. [Configuration](#configuration)
7. [Deployment](#deployment)
8. [Development Workflow](#development-workflow)

---

## Overview

Shiroxy is a **Go-based reverse proxy** designed for dynamic routing, SSL automation, and scalable domain management. It provides:

- **Automatic SSL Certificates** via ACME protocol (Let's Encrypt, Pebble)
- **HTTP/2 Support** with optimized connection pooling and multiplexing
- **Intelligent Gzip Compression** for text-based content types with automatic client detection
- **Dynamic Domain Management** through REST API
- **Advanced Load Balancing** with multiple strategies (round-robin, least-connection, sticky-session)
- **Tag-based Routing** with caching and trie-based indexing
- **Buffer Pool Management** using sync.Pool for efficient memory reuse (32KB buffers)
- **Connection Pool Statistics** tracking active/idle connections, reuse rates, and request duration
- **Health Monitoring** for backend servers with configurable intervals
- **System Analytics** with CPU, memory, and connection tracking
- **Webhook Integration** for event notifications
- **Graceful Shutdown** with data persistence and proper resource cleanup

---

## Project Structure

```
shiroxy/
├── cmd/shiroxy/                # Main application code
│   ├── main.go                 # Application entry point
│   ├── analytics/              # System analytics module
│   ├── api/                    # REST API implementation
│   ├── domains/                # Domain management & ACME integration
│   ├── logger/                 # Application-level logging
│   ├── proxy/                  # Reverse proxy core logic
│   ├── types/                  # Shared type definitions
│   ├── users/                  # User management (JWT)
│   └── webhook/                # Webhook event handling
├── pkg/                        # Reusable packages
│   ├── certificate/            # ACME certificate generation utilities
│   ├── cli/                    # Cobra CLI implementation
│   ├── color/                  # Terminal color utilities
│   ├── configuration/          # Configuration reader (Viper)
│   ├── loader/                 # Progress loader animations
│   ├── logger/                 # Package-level logging utilities
│   ├── models/                 # Configuration & data models
│   └── shutdown/               # Graceful shutdown handlers
├── utils/                      # Utility functions
│   ├── acme.go                 # ACME server selection
│   ├── check_acme_server.go    # ACME server health check
│   └── *.go                    # Other helpers
├── public/                     # HTML error pages
├── docs/                       # Swagger API documentation
├── defaults/                   # Default configuration files
├── docker/                     # Docker environments (dev/stage/prod)
├── test/                       # Test files and fixtures
├── media/                      # Images and assets
├── temp/                       # Temporary/development files
├── Dockerfile                  # Production Docker image
├── docker-compose.yaml         # Docker Compose setup
├── Makefile                    # Build automation
├── go.mod & go.sum             # Go dependency management
└── README.md                   # User-facing documentation
```

---

## Core Components

### 1. **Main Application (`cmd/shiroxy/main.go`)**

**Purpose:** Application bootstrap and orchestration.

**Key Responsibilities:**

- Initialize logger with configuration
- Parse CLI arguments and load configuration
- Select ACME server based on runtime mode (dev/stage/prod)
- Initialize storage (memory or Redis)
- Start analytics collection
- Load persisted data from previous shutdowns
- Set up graceful shutdown handlers
- Start webhook handler
- Initialize and start load balancer
- Start REST API server
- Coordinate all goroutines with `sync.WaitGroup`

**Flow:**

```
1. Logger Initialization
2. Configuration Loading (CLI)
3. ACME Server Selection
4. Storage Initialization
5. Analytics Start
6. Data Persistence Load
7. Graceful Shutdown Setup
8. Webhook Handler Start
9. Load Balancer Start
10. API Server Start
11. Wait for goroutines
```

---

### 2. **Proxy Module (`cmd/shiroxy/proxy/`)**

The proxy module is the heart of Shiroxy, handling all reverse proxy operations.

#### **2.1 Load Balancer (`balancer.go`)**

**Purpose:** Routes incoming requests to backend servers with multiple strategies.

**Key Structures:**

- **`LoadBalancer`**
  - `Ready`: Indicates if the load balancer is ready to serve requests
  - `MaxRetry`: Maximum retry attempts for failed requests (default: 3)
  - `Servers`: List of all backend servers
  - `ServerByTag`: Map of tags to backend servers
  - `Frontends`: Map of port numbers to frontend handlers
  - `RoutingDetailsByTag`: Routing state per tag (round-robin counters, connections, sticky sessions)
  - `HealthChecker`: Health monitoring component
  - `TagCache`: LRU cache for frequently accessed tags
  - `TagTrie`: Trie data structure for efficient tag lookup
  - `ConnectionStats`: HTTP/2 connection pool statistics

**Load Balancing Strategies:**

1. **Round-Robin** (`GetNextServerRoundRobin`)

   - Distributes requests sequentially across servers
   - Maintains per-tag counter for rotation
   - Skips unhealthy servers

2. **Least Connection** (`GetLeastConnectionServer`)

   - Routes to server with fewest active connections
   - Tracks connection count per server
   - Updates count on each request

3. **Sticky Session** (`GetStickySessionServer`)
   - Maps client IPs to specific servers
   - Maintains persistent sessions
   - Falls back to round-robin for new clients

**Tag-Based Routing:**

- Domains can have metadata tags (e.g., "api", "web", "cdn")
- Servers have tags defining which requests they handle
- Uses **Trie** for O(m) tag lookup (m = tag length)
- **LRU Cache** (capacity: 100) for frequently accessed tags
- `Tagrule` can be "strict" (require tags) or "flexible"

**Request Flow:**

```
1. Request arrives → ServeHTTP
2. Extract domain from Host header
3. Lookup domain metadata in storage
4. Extract tags from domain metadata
5. Check TagCache for cached servers
6. If miss, search TagTrie
7. Select server based on balance strategy
8. Forward request to selected server
9. If error, retry up to MaxRetry times
10. If all fail, mark server as unhealthy
```

#### **2.2 Proxy Handler (`proxy_handler.go`)**

**Purpose:** Sets up frontend listeners and TLS configuration.

**Key Functions:**

- **`StartShiroxyHandler`**
  - Creates backend server instances with HTTP/2 optimized transport
  - Initializes load balancer
  - Sets up frontend bindings (ports)
  - Configures TLS for HTTPS
  - Handles HTTP → HTTPS redirection
  - Implements DNS challenge solver for ACME HTTP-01

**Transport Configuration:**

```go
&http.Transport{
    ForceAttemptHTTP2:     true,
    MaxIdleConns:          300,
    IdleConnTimeout:       120s,
    MaxIdleConnsPerHost:   100,
    MaxConnsPerHost:       200,
    WriteBufferSize:       32KB,
    ReadBufferSize:        32KB,
    DisableCompression:    true,  // Handle separately
}
```

**Frontend Modes:**

1. **Multiple Target** (`CreateMultipleTargetServer`)

   - Dynamic TLS certificate loading per domain
   - Uses `GetCertificate` callback for SNI
   - Loads certs from domain metadata

2. **Single Target** (`CreateSingleTargetServer`)
   - Two modes:
     - `certandkey`: Load from file paths
     - `shiroxysinglesecure`: Load from Shiroxy storage

**Security Policies:**

- `none`: No client certificate required
- `optional`: Request but don't verify
- `required`: Require and verify client certificate

**DNS Challenge Solver:**

- Listens on port 80 for `/.well-known/acme-challenge/` paths
- Serves ACME challenge tokens for domain validation
- Maps token filenames to domain names
- Returns authorization keys for verification

#### **2.3 Reverse Proxy (`reverse_proxy.go`)**

**Purpose:** Core HTTP reverse proxy implementation (custom, not `httputil.ReverseProxy`).

**Key Components:**

- **`Shiroxy` struct**
  - Custom reverse proxy with enhanced features
  - Logger integration with structured error logging
  - Connection pool statistics for HTTP/2 monitoring
  - Buffer pooling for efficient memory reuse
  - Intelligent gzip compression with content-type detection
  - HTTP/2 upgrade handling and multiplexing
  - WebSocket support with bidirectional streaming

**Request Processing:**

1. **Header Manipulation**

   - Remove hop-by-hop headers (Connection, Proxy-Authorization, Transfer-Encoding, etc.)
   - Add X-Forwarded-\* headers (For, Host, Proto) for origin tracking
   - Handle Upgrade requests (WebSocket, HTTP/2 with proper protocol switching)
   - Strip port from HTTP to HTTPS redirects for correct URL formatting

2. **Director Function**

   - Rewrites request URL to backend server
   - Preserves or modifies Host header
   - Merges query parameters with proper escaping
   - Cleans query parameters to prevent injection

3. **Response Handling**

   - Streams response with configurable flush intervals
   - Handles 1xx informational responses
   - Supports HTTP trailers for chunked encoding
   - **Intelligent Gzip Compression:**
     - Checks client Accept-Encoding header for gzip support
     - Verifies response is not already compressed
     - Only compresses compressible content types:
       - text/\* (HTML, CSS, plain text)
       - application/json, application/javascript
       - application/xml and variants (XHTML, RSS, Atom)
     - Sets proper headers before WriteHeader:
       - Content-Encoding: gzip
       - Vary: Accept-Encoding
       - Removes Content-Length (varies with compression)
     - Implements http.Flusher for streaming compressed responses
     - Proper cleanup with deferred gzip writer close and error logging

4. **Error Handling**
   - Custom `ErrorHandler` callback with context
   - Default: 502 Bad Gateway with detailed logging
   - Proper HTTP status codes (503 for inactive domains, not 404)
   - Graceful panic recovery with http.ErrAbortHandler
   - Logs errors via Logger with severity levels

**Buffer Pool (`buffer_pool.go`):**

- **Fixed Implementation:** Stores `[]byte` directly (not pointers) to prevent invalid memory references
- Reuses byte slices to reduce GC pressure and allocation overhead
- Default size: 32KB (optimal for most HTTP responses)
- Uses `sync.Pool` for efficient, concurrent-safe allocation
- Returns buffers with zero length but full capacity for reuse
- Validates buffer capacity before returning to pool
- Auto-expands/contracts based on load patterns

**Connection Pool (`connection_pool.go`):**

- **HTTP/2 Connection Statistics Tracking:**
  - Active connections (currently in use)
  - Idle connections (available for reuse)
  - Total connections created over lifetime
  - Total connections closed
  - Connection reuse count (efficiency metric)
  - Average request duration (exponential moving average)
  - Last updated timestamp
- **Fixed IdleConnections Counter:** Guards against negative values with proper locking
- Uses `httptrace.ClientTrace` for real-time monitoring
- Thread-safe with RWMutex for concurrent access
- Exponential Moving Average (EMA) for request duration calculation
- Provides read-only stats snapshots via `GetStats()`

**Special Handlers:**

- **`handleUpgradeResponse`**: WebSocket/HTTP2 upgrade with proper connection hijacking
- **`maxLatencyWriter`**: Automatic flushing for streaming responses
- **`gzipResponseWriter`**: On-the-fly compression with flush support
- **`shouldCompress`**: Content-type based compression decision logic

#### **2.4 Health Checker (`health_check.go`)**

**Purpose:** Monitors backend server health and fires webhooks on status changes.

**Features:**

- Periodic health checks via HTTP GET/HEAD
- Configurable check interval
- Dynamic interval updates
- Webhook notifications on state changes
- Thread-safe server status updates

**Health Check Logic:**

```
1. Send HTTP GET to server.HealthCheckUrl or server.URL
2. Check response status code (200 = healthy)
3. If healthy:
   - Mark server.Alive = true
   - Fire "backendserver.register.success" webhook (first time)
4. If unhealthy:
   - Mark server.Alive = false
   - Fire "backendserver.register.failed" webhook (first time)
5. Skip server in load balancer if Alive = false
```

**Webhook Events:**

- `backendserver.register.success`: Server is up
- `backendserver.register.failed`: Server is down

---

### 3. **Domain Management (`cmd/shiroxy/domains/`)**

**Purpose:** Manage domain metadata, ACME certificates, and storage.

#### **`domain.go`**

**Key Structure:**

- **`Storage`**
  - `ACME_SERVER_URL`: ACME directory URL
  - `INSECURE_SKIP_VERIFY`: Skip TLS verification (dev mode)
  - `Storage`: Configuration (memory/Redis)
  - `RedisClient`: Redis connection
  - `DnsChallengeToken`: Maps ACME tokens to domains
  - `DomainMetadata`: Map of domain names to metadata
  - `WebhookSecret`: Secret for webhook authentication

**`DomainMetadata` (Protobuf):**

```protobuf
message DomainMetadata {
  string status = 1;                    // "active" or "inactive"
  string domain = 2;                    // Domain name
  string email = 3;                     // Owner email
  map<string, string> metadata = 4;     // Custom tags/data
  bytes acme_account_private_key = 5;   // ACME account key
  bytes cert_pem_block = 6;             // SSL certificate
  bytes key_pem_block = 7;              // Private key
  string dns_challenge_key = 8;         // ACME challenge response
}
```

**Core Functions:**

1. **`RegisterDomain(domainName, email, metadata)`**

   - Generates ACME account private key (ECDSA P256)
   - Creates domain metadata with "inactive" status
   - Calls `generateCertificate` to obtain SSL cert
   - Stores metadata in memory or Redis
   - Returns DNS challenge key for HTTP-01 validation

2. **`generateCertificate(domainMetadata)`**

   - Creates ACME client with configured directory URL
   - Creates or retrieves ACME account
   - Creates new order for domain
   - Solves HTTP-01 challenge:
     - Generates challenge token
     - Stores token in `DnsChallengeToken` map
     - Sets `DnsChallengeKey` for response
     - Polls authorization until valid
   - Generates certificate private key
   - Creates CSR (Certificate Signing Request)
   - Finalizes order with CSR
   - Downloads certificate chain
   - Encodes keys to PEM format
   - Updates domain metadata with cert/key
   - Sets status to "active"

3. **`UpdateDomain(domainName, updateBody)`**

   - Updates existing domain metadata
   - Supports memory and Redis storage

4. **`RemoveDomain(domainName)`**
   - Deletes domain from storage
   - Cleans up metadata

**Storage Backends:**

- **Memory** (`initializeMemoryStorage`)

  - Simple `map[string]*DomainMetadata`
  - Fast, no external dependencies
  - Data lost on restart (unless persisted)

- **Redis** (`ConnectRedis`)
  - Uses Protocol Buffers for serialization
  - Persistent across restarts
  - Supports Redis connection string or host:port

**ACME Integration:**

- Uses `github.com/mholt/acmez` library
- Supports Let's Encrypt (prod/staging) and Pebble (dev)
- Automatic certificate renewal (TODO: implement)
- HTTP-01 challenge only (DNS-01 not implemented)

#### **`domain.proto`**

Protobuf definitions for domain persistence and wire format.

---

### 4. **API Module (`cmd/shiroxy/api/`)**

**Purpose:** REST API for domain, backend, and analytics management.

#### **`api.go`**

**API Structure:**

- Framework: **Gin** (HTTP router)
- Authentication: **Basic Auth** (username: user email, password: user secret)
- Base Path: `/v1`
- Swagger Docs: `/docs/swagger/*`
- Health Endpoint: `/v1/health`

**Routes:**

1. **Domain Routes** (`/v1/domain`)

   - `POST /` - Register new domain
   - `PATCH /:domain` - Update domain metadata
   - `PATCH /:domain/retryssl` - Force SSL retry
   - `GET /:domain` - Fetch domain info
   - `DELETE /:domain` - Remove domain

2. **Analytics Routes** (`/v1/analytics`)

   - `GET /domain` - Domain analytics
   - `GET /system` - System analytics

3. **Backend Routes** (`/v1/backends`)
   - `GET /` - List all backend servers
   - `POST /` - Add backend server
   - `DELETE /:id` - Remove backend server

#### **Controllers**

**`domains.go`**

- **`RegisterDomain`**

  - Validates domain and email
  - Calls `storage.RegisterDomain()`
  - Fires webhooks on success/failure
  - Returns DNS challenge key

- **`ForceSSL`**

  - Triggers SSL certificate regeneration
  - Useful for retrying failed cert generation

- **`UpdateDomain`**

  - Updates domain metadata (tags, etc.)

- **`FetchDomainInfo`**

  - Returns domain details and certificate info

- **`RemoveDomain`**
  - Deletes domain and its certificates

**`analytics.go`**

- **`FetchDomainAnalytics`**

  - Returns total domains, active, inactive counts

- **`FetchSystemAnalytics`**
  - Returns CPU, memory, GC, connection pool stats
  - Uses `AnalyticsConfiguration.ReadAnalytics()`

**`backends.go`**

- **`FetchAllBackendServers`**

  - Lists all servers with health status
  - Excludes internal Shiroxy instance

- **`AddBackendServer`**

  - Dynamically adds server to load balancer
  - Updates tag indexing

- **`RemoveBackendServer`**
  - Removes server from pool

---

### 5. **Analytics (`cmd/shiroxy/analytics/`)**

**Purpose:** Collect system metrics (CPU, memory, GC, connection stats).

#### **`analytics.go`**

**Key Structure:**

- **`AnalyticsConfiguration`**
  - `RequestAnalytics`: Channel to trigger on-demand collection
  - `ReadAnalyticsData`: Channel to receive analytics
  - `TriggerInterval`: Collection interval (seconds)
  - `latestShiroxyAnalytics`: Cached latest stats
  - `collectingAnalytics`: Mutex flag to prevent concurrent collection

**`ShiroxyAnalytics` Structure:**

```go
type ShiroxyAnalytics struct {
    TotalDomain             int                    `json:"total_domain"`
    TotalCertSize           int                    `json:"total_cert_size"`
    TotalCerts              int                    `json:"total_certs"`
    TotalUnSecuredDomains   int                    `json:"total_unsecured_domains"`
    TotalSecuredDomains     int                    `json:"total_secured_domain"`
    TotalFailedSSLAttempts  int                    `json:"total_failed_ssl_attempts"`
    TotalSuccessSSLAttempts int                    `json:"total_success_ssl_attempts"`
    TotalUser               int                    `json:"total_user"`
    Memory_ALLOC            int                    `json:"memory_alloc"`       // MB
    Memory_TOTAL_ALLOC      int                    `json:"memory_total_alloc"` // MB
    Memory_SYS              int                    `json:"memory_sys"`         // MB
    CPU_Usage               float64                `json:"cpu_usage"`          // %
    GC_Count                int                    `json:"gc_count"`
    Metadata                map[string]interface{} `json:"metadata"`
}
```

**Collection Process:**

1. **Periodic Collection** (ticker-based)

   - Runs every `TriggerInterval` seconds
   - Reads `runtime.MemStats` for memory metrics
   - Uses `gopsutil/process` for CPU usage
   - Converts bytes to MB (`bToMb`)

2. **On-Demand Collection**

   - Triggered via `RequestAnalytics` channel
   - Used by API endpoints

3. **Concurrency Safety**
   - Uses `sync.RWMutex` for safe concurrent access
   - Prevents duplicate collections with `collectingAnalytics` flag

**Functions:**

- **`StartAnalytics(interval, logger, wg)`**: Initialize and start collection goroutine
- **`collectAnalytics(logger)`**: Gather current metrics
- **`ReadAnalytics(forced)`**: Get latest analytics (optionally force new collection)
- **`UpdateTriggerInterval(duration)`**: Change collection frequency
- **`StopAnalytics()`**: Graceful shutdown

---

### 6. **Webhook Handler (`cmd/shiroxy/webhook/`)**

**Purpose:** Fire HTTP webhooks on application events.

#### **`webhook.go`**

**Event Types:**

- `domain-register-success`
- `domain-register-failed`
- `backendserver.register.success`
- `backendserver.register.failed`
- Custom events (configurable)

**Webhook Flow:**

1. **Event Triggered**
   - Component calls `webhookHandler.Fire(eventName, data)`
2. **Event Filtering**
   - Checks if event is in configured `Webhook.Events` list
3. **Payload Construction**

   ```json
   {
     "eventname": "domain-register-success",
     "data": {
       "domain": "example.com"
     }
   }
   ```

4. **HTTP Request**

   - Method: POST
   - URL: `Webhook.Url` from config
   - Headers: `Content-Type: application/json`
   - Body: JSON payload

5. **Response Handling**
   - 200 OK: Log success
   - Other: Log error

**Secret Generation:**

- Auto-generated if not provided
- Format: UUID + random numbers + random letters (shuffled)
- Stored in domain storage for persistence

**Configuration:**

```yaml
webhook:
  enable: true
  url: "https://your-webhook-receiver.com/events"
  events:
    - "domain-register-success"
    - "domain-register-failed"
```

---

### 7. **Logger (`pkg/logger/` & `cmd/shiroxy/logger/`)**

**Purpose:** Multi-output logging system with color support.

#### **`pkg/logger/logger.go`**

**Features:**

- Console output with ANSI colors
- Remote syslog support (UDP)
- Structured logging with timestamp, package, module
- Log levels: Error, Success, Warning, Info

**Log Format:**

```
[DD/MM/YYYY HH:MM:SS] [PackageName] [ModuleName] => Message
```

**Outputs:**

1. **Console** (when `Logging.Enable = true`)

   - Color-coded by log level
   - Success: Green
   - Error: Red
   - Warning: Yellow
   - Info: Default

2. **Remote Syslog** (when `Logging.EnableRemote = true`)
   - Protocol: UDP
   - Maps to syslog levels:
     - Success → Notice
     - Error → Err
     - Warning → Warning
     - Info → Info

**Color Utilities** (`pkg/logger/colorprint.go`):

- `RedPrint`, `GreenPrint`, `YellowPrint`, `BluePrint`, `CyanPrint`
- Uses ANSI escape codes

---

### 8. **Configuration (`pkg/configuration/` & `pkg/models/`)**

#### **`models.go`**

**Configuration Structure:**

```yaml
runtime: # Runtime settings
  mode: dev # dev, stage, prod
  instancename: shiroxy-1
  acmeserverurl: "" # Override ACME server
  aCMEserverinsecureskipverify: "yes"

default:
  debugmode: dev
  logpath: "/var/log/shiroxy"
  datapersistancepath: "/var/lib/shiroxy"
  enablednschallengesolver: true

  timeout:
    connect: "10s"
    client: "10s"
    server: "10s"

  storage:
    location: memory # memory or redis
    redishost: ""
    redisport: ""
    redispassword: ""
    redisconnectionstring: ""

  analytics:
    collectioninterval: 10 # seconds
    routename: "/analytics"

  errorresponses:
    errorpagebuttonname: "Shiroxy"
    errorpagebuttonurl: "https://github.com/shikharcodess/shiroxy"

  user:
    email: "admin@example.com"
    secret: "your-secret"

  adminapi:
    port: "2210"

frontend:
  mode: http # http or https
  httptohttps: true # Auto-redirect
  bind:
    - port: "80"
      host: "0.0.0.0"
      target: multiple # single or multiple
      secure: false
    - port: "443"
      host: "0.0.0.0"
      target: multiple
      secure: true
      securesetting:
        secureverify: none # none, optional, required
        singletargetmode: "" # certandkey, shiroxysinglesecure

backend:
  name: my-backends
  balance: round-robin # round-robin, least-count, sticky-session
  healthcheckmode: active
  healthchecktriggerduration: 5 # seconds
  tagrule: flexible # strict or flexible
  servers:
    - id: server-1
      host: 127.0.0.1
      port: "8080"
      healthurl: "http://127.0.0.1:8080/health"
      tags: "api,web"

logging:
  enable: true
  enableremote: false
  remotebind:
    host: "127.0.0.1"
    port: "514"
  mode: console
  schema: []
  include: []

webhook:
  enable: false
  url: ""
  events: []
```

#### **`configuration/reader.go`**

**Configuration Loading:**

1. Uses **Viper** for YAML parsing
2. Loads from file path passed via CLI (`-c` flag)
3. Unmarshals into `models.Config` struct
4. Shows progress loader during loading
5. Validates required fields

---

### 9. **Graceful Shutdown (`pkg/shutdown/`)**

**Purpose:** Handle SIGINT/SIGTERM signals and persist data before exit.

#### **`graceful_shutdown.go`**

**Shutdown Process:**

1. **Signal Handling**

   - Listens for `SIGINT` (Ctrl+C) or `SIGTERM` (Docker stop)
   - Also triggered by panic recovery

2. **Data Collection**

   - Request latest analytics
   - Collect all domain metadata
   - Gather webhook secrets

3. **Marshaling**

   - Analytics → JSON
   - Domain data → Protobuf (`DataPersistance` message)
   - Combined → `ShutdownMetadata` Protobuf
   - Base64 encode for storage

4. **Persistence**

   - Write to file: `{env}-persistence.shiroxy`
   - Location: `DataPersistancePath` from config
   - Example: `/var/lib/shiroxy/dev-persistence.shiroxy`

5. **Exit**
   - Log success/failure
   - Call `os.Exit(0)`

#### **`load_persistence.go`**

**Startup Data Loading:**

1. Read persistence file
2. Base64 decode
3. Unmarshal Protobuf
4. Restore domain metadata to storage
5. Restore webhook secrets
6. Log success/failure

**`ShutdownMetadata` Protobuf:**

```protobuf
message ShutdownMetadata {
  bytes domain_metadata = 1;    // Serialized DataPersistance
  bytes system_data = 2;        // JSON analytics
  string webhook_secret = 3;
}

message DataPersistance {
  string datetime = 1;
  repeated DomainMetadata domains = 2;
}
```

---

### 10. **CLI (`pkg/cli/`)**

**Purpose:** Command-line interface using Cobra.

#### **`cobra.go`**

**Commands:**

1. **Root Command** (`shiroxy`)

   - Runs main application
   - Flags:
     - `-c, --config`: Path to config file (required)

2. **Cert Command** (`shiroxy cert`)
   - Manual certificate generation
   - Flags:
     - `-d, --domain`: Domain name
     - `-e, --email`: Email address
   - Calls `pkg/certificate/certificate.go`

**Usage:**

```bash
# Run Shiroxy
shiroxy -c /path/to/shiroxy.conf.yaml

# Generate certificate manually
shiroxy cert -d example.com -e admin@example.com
```

---

### 11. **Utilities (`utils/`)**

#### **`acme.go`**

**ACME Server Selection:**

```go
const ACME_DEV_SERVER_URL = "https://127.0.0.1:14000/dir"        // Pebble
const ACME_STAGE_SERVER_URL = "https://acme-staging-v02.api.letsencrypt.org/directory"
const ACME_PROD_SERVER_URL = "https://acme-v02.api.letsencrypt.org/directory"

func ChooseAcmeServer(mode string) string {
    switch mode {
    case "dev": return ACME_DEV_SERVER_URL
    case "stage": return ACME_STAGE_SERVER_URL
    case "prod": return ACME_PROD_SERVER_URL
    default: return ACME_DEV_SERVER_URL
    }
}
```

#### **`check_acme_server.go`**

**Health Check for ACME Server:**

- Sends HTTP GET to directory URL
- Validates response
- Logs warnings if unreachable

---

### 12. **Public Assets (`public/`)**

**Error Pages:**

1. **`domain_not_found_html.go`**

   - HTML template for 404 errors
   - Customizable button (name, URL)
   - Responsive design
   - Shiroxy logo

2. **`shiroxy_not_ready.go`**

   - Shown when load balancer not ready
   - During startup or maintenance

3. **`status_inactive.go`**
   - Shown for inactive domains
   - SSL not yet provisioned

**Template Variables:**

- `{{button_name}}`: Custom button text
- `{{button_url}}`: Custom button link

---

## Configuration

### Environment-Based Configuration

Shiroxy supports three runtime modes:

1. **Development** (`mode: dev`)

   - ACME: Pebble (local testing)
   - TLS: InsecureSkipVerify = true
   - Logging: Verbose
   - Debug mode: enabled

2. **Staging** (`mode: stage`)

   - ACME: Let's Encrypt Staging
   - TLS: Verify certificates
   - Rate limits: Higher than prod

3. **Production** (`mode: prod`)
   - ACME: Let's Encrypt Production
   - TLS: Full verification
   - Rate limits: Strict

### Configuration Files

**Default Locations:**

- `defaults/shiroxy.conf.yaml` - Local development
- `defaults/shiroxy.conf.remote.yaml` - Remote deployment

**Override via CLI:**

```bash
shiroxy -c /custom/path/config.yaml
```

### Critical Configuration Fields

**Must Configure:**

- `default.user.email` - Admin email
- `default.user.secret` - API password
- `backend.servers[]` - At least one backend server

**Optional but Recommended:**

- `default.storage.location` - Set to "redis" for production
- `webhook.url` - For event notifications
- `frontend.httptohttps` - Enable for automatic HTTPS redirect

---

## Deployment

### Local Development

**Prerequisites:**

- Go 1.24+
- Pebble ACME server (for SSL testing)

**Steps:**

1. Clone repository

   ```bash
   git clone https://github.com/shikharcodess/shiroxy.git
   cd shiroxy
   ```

2. Start Pebble (in separate terminal)

   ```bash
   cd pebble
   go build -o pebble cmd/pebble/main.go
   ./pebble
   ```

3. Run Shiroxy
   ```bash
   sudo go run cmd/shiroxy/main.go -c defaults/shiroxy.conf.yaml
   ```

### Docker Deployment

**Using Docker Compose:**

```bash
docker-compose up -d --build
```

**Exposed Ports:**

- `80` - HTTP (SSL challenges)
- `443` - HTTPS (TLS proxy)
- `2210` - Admin API

**Environment Variables:**

```dockerfile
SHIROXY_ENVIRONMENT=dev  # dev, stage, prod
```

### Build Modes

**Using Makefile:**

```bash
# Development build
make MODE=dev build

# Staging build
make MODE=stage build

# Production build
make MODE=prod build
```

**Output:** `build/shiroxy` binary

**LDFLAGS Injection:**

- `ACME_SERVER_URL` - Set at compile time
- `INSECURE_SKIP_VERIFY` - Set based on mode
- `MODE` - Runtime mode

### Multi-Stage Deployment

**Dockerfile Variants:**

- `docker/dev/Dockerfile` - Development
- `docker/stage/Dockerfile` - Staging
- `docker/prod/Dockerfile` - Production

**Image Optimization:**

- Multi-stage builds
- Minimal base image (Alpine)
- Separate build and runtime stages

---

## Development Workflow

### Project Conventions

**Package Organization:**

- `cmd/shiroxy/*` - Application-specific code (main, API, proxy)
- `pkg/*` - Reusable packages (logger, config, shutdown)
- `utils/*` - Helper functions

**Code Style:**

- Format: `make fmt` (gofmt)
- Lint: `make lint` (golangci-lint)
- Test: `make test`

### Testing

**Test Files:**

- `cmd/shiroxy/main_test.go`
- `cmd/shiroxy/proxy/*_test.go`
- `cmd/shiroxy/domains/domain_test.go`

**K6 Load Testing:**

- `test/k6/dummy-test/index.js`
- Simulates concurrent requests
- Measures throughput and latency

**Run Tests:**

```bash
make test
```

### Adding New Features

**Example: Adding a New Load Balancing Strategy**

1. Add strategy to `proxy/balancer.go`

   ```go
   func (lb *LoadBalancer) GetWeightedRoundRobin(tag string) *Server {
       // Implementation
   }
   ```

2. Update `selectServerFromList` switch

   ```go
   case "weighted-round-robin":
       return lb.GetWeightedRoundRobin(tag, servers.Servers)
   ```

3. Update config model (`pkg/models/models.go`)

   ```go
   type Backend struct {
       Balance string `json:"balance"` // Add "weighted-round-robin"
   }
   ```

4. Document in `configuration.md`

### Protocol Buffers

**Generating Code:**

```bash
# Domain metadata
protoc --go_out=. cmd/shiroxy/domains/domain.proto

# Shutdown metadata
protoc --go_out=. pkg/shutdown/shutdown.proto

# User models
protoc --go_out=. cmd/shiroxy/users/user_models.proto
```

**Proto Files:**

- `cmd/shiroxy/domains/domain.proto`
- `pkg/shutdown/shutdown.proto`
- `cmd/shiroxy/users/user_models.proto`

### Swagger Documentation

**Generate Swagger:**

```bash
swag init -g cmd/shiroxy/main.go
```

**Access Docs:**

```
http://localhost:2210/docs/swagger/index.html
```

**Annotations:**

- Located in `cmd/shiroxy/api/controllers/*.go`
- Follows Swaggo format

---

## Monitoring & Observability

### Health Checks

**Backend Health:**

- Endpoint: Server's configured `healthurl`
- Interval: `backend.healthchecktriggerduration` seconds
- Method: HTTP GET
- Success: 200 OK

**Application Health:**

- Endpoint: `http://localhost:2210/v1/health`
- Returns: `{"status": "healthy"}`

### Analytics

**Available Metrics:**

1. **Domain Metrics** (`/v1/analytics/domain`)

   - Total domains
   - Active domains
   - Inactive domains

2. **System Metrics** (`/v1/analytics/system`)
   - Memory allocated (MB)
   - Total memory allocated (MB)
   - System memory (MB)
   - CPU usage (%)
   - GC count
   - Connection pool stats:
     - Active connections
     - Idle connections
     - Connections created/closed
     - Connection reuse count
     - Average request duration

### Logging

**Log Locations:**

- Console: `stdout` (when `logging.enable = true`)
- Remote: Syslog server (when `logging.enableremote = true`)
- File: Not currently implemented (TODO)

**Log Levels:**

- Info: General operations
- Success: Successful operations (domain registration, SSL generation)
- Warning: Non-critical issues
- Error: Failures requiring attention

---

## Security Considerations

### TLS Configuration

**Minimum TLS Version:** TLS 1.2

**Cipher Suites:** Go's default secure suites

**Client Certificate Options:**

- `none`: No client cert required (default)
- `optional`: Request but don't enforce
- `required`: Require and verify client cert

### API Authentication

**Current:** Basic Auth

- Username: `default.user.email`
- Password: `default.user.secret`

**Future:** JWT tokens (partial implementation in `cmd/shiroxy/users/user.go`)

### ACME Security

**Account Keys:** ECDSA P256 (256-bit)
**Certificate Keys:** ECDSA P256

**Challenge Types:**

- HTTP-01: ✅ Implemented
- DNS-01: ❌ Not implemented
- TLS-ALPN-01: ❌ Not implemented

### Webhook Security

**Authentication:** Shared secret
**Transport:** HTTPS recommended
**Validation:** Response status code check

---

## Performance Optimization

### Connection Pooling

**HTTP/2 Optimizations:**

- `MaxIdleConns: 300` - Total pool size
- `MaxIdleConnsPerHost: 100` - Per-backend limit
- `IdleConnTimeout: 120s` - Keep connections longer
- `MaxConnsPerHost: 200` - Prevent backend overload

### Buffer Pooling

**Purpose:** Reduce GC pressure during request/response copying

**Implementation:**

- Uses `sync.Pool`
- Default buffer size: 32KB
- Auto-scales based on request size

### Tag Caching

**LRU Cache:**

- Capacity: 100 entries
- Stores frequently accessed tags
- O(1) lookup for cached entries

**Trie Index:**

- O(m) lookup (m = tag length)
- Efficient for prefix matching
- Falls back when cache misses

### Compression

**Gzip Support:**

- Enabled for responses with `Accept-Encoding: gzip`
- Disabled on transport (handled at proxy layer)
- Configurable buffer size

---

## Troubleshooting

### Common Issues

**1. "Domain not found" Error**

- Check domain is registered via API
- Verify domain status is "active"
- Check domain metadata has correct tags

**2. SSL Certificate Generation Fails**

- Ensure ACME server is reachable
- Verify port 80 is accessible for HTTP-01 challenge
- Check DNS points to Shiroxy IP
- Review ACME server logs

**3. Backend Server Marked Unhealthy**

- Verify backend is running
- Check `healthurl` is correct
- Ensure health endpoint returns 200 OK
- Review health check interval

**4. API Returns 401 Unauthorized**

- Verify `default.user.email` and `default.user.secret` in config
- Check Basic Auth header format: `Authorization: Basic <base64(email:secret)>`

### Debug Mode

**Enable:**

```yaml
default:
  debugmode: dev
```

**Features:**

- Prints stack traces on panic
- Verbose logging
- Detailed error messages

### Log Analysis

**Key Log Patterns:**

```
# Successful startup
[*] [STARTUP] [INFO] => Running shiroxy in dev MODE

# Domain registered
[*] [Domains] [RegisterDomain] => Domain registered: example.com

# SSL generated
[*] [Domains] [generateCertificate] => Certificate Generated Successfully For Domain: example.com

# Backend unhealthy
[*] [HealthChecker] [CheckHealth] => Server 127.0.0.1:8080 is unhealthy
```

---

## Future Enhancements

### Planned Features

1. **Certificate Renewal**

   - Automatic renewal before expiry
   - Configurable renewal threshold
   - Webhook notifications

2. **DNS-01 Challenge Support**

   - For wildcard certificates
   - Integration with DNS providers

3. **Rate Limiting**

   - Per-domain rate limits
   - Per-IP rate limits
   - Configurable strategies

4. **Metrics Export**

   - Prometheus endpoint
   - Grafana dashboards
   - Custom metrics

5. **Advanced Caching**

   - HTTP response caching
   - CDN integration
   - Cache invalidation API

6. **Multi-Tenancy**

   - Organization-based isolation
   - RBAC (Role-Based Access Control)
   - Per-tenant quotas

7. **Configuration Hot Reload**
   - Reload without restart
   - SIGHUP signal support
   - Incremental config updates

### Known Limitations

1. **ACME Challenges:**

   - Only HTTP-01 supported
   - Requires port 80 accessibility

2. **Storage:**

   - Redis connection not pooled
   - No Redis cluster support

3. **Health Checks:**

   - Basic HTTP only
   - No custom health check logic
   - No circuit breaker pattern

4. **Load Balancing:**
   - No weighted round-robin
   - No priority-based routing
   - No geographic routing

---

## Contributing

### Development Setup

1. Fork repository
2. Create feature branch
3. Make changes
4. Run tests: `make test`
5. Format code: `make fmt`
6. Lint code: `make lint`
7. Submit pull request

### Code Review Guidelines

- All code must pass linting
- Tests required for new features
- Update documentation
- Follow existing code style
- Add comments for complex logic

### Documentation

**Files to Update:**

- `README.md` - User-facing docs
- `INTERNAL_DOCUMENTATION.md` - This file
- `configuration.md` - Config reference
- `CONTRIBUTION.md` - Contribution guide
- `docs/api.md` - API documentation

---

## References

### External Dependencies

**Core:**

- `github.com/gin-gonic/gin` - HTTP router
- `github.com/mholt/acmez` - ACME client
- `github.com/go-redis/redis/v8` - Redis client
- `github.com/spf13/viper` - Configuration
- `github.com/spf13/cobra` - CLI framework
- `go.uber.org/zap` - ACME logging
- `google.golang.org/protobuf` - Protobuf serialization

**Utilities:**

- `github.com/golang-jwt/jwt/v4` - JWT tokens
- `github.com/google/uuid` - UUID generation
- `github.com/shirou/gopsutil` - System metrics
- `github.com/joho/godotenv` - Environment variables

**Documentation:**

- `github.com/swaggo/swag` - Swagger generation
- `github.com/swaggo/gin-swagger` - Swagger UI

### ACME Specification

- RFC 8555: Automatic Certificate Management Environment (ACME)
- Let's Encrypt Documentation: https://letsencrypt.org/docs/
- Pebble Test Server: https://github.com/letsencrypt/pebble

### HTTP/2 Resources

- RFC 7540: Hypertext Transfer Protocol Version 2 (HTTP/2)
- Go HTTP/2 Transport: https://pkg.go.dev/golang.org/x/net/http2

---

## Version 1.1.0 (Kuchii Release) - Changelog

### New Features

#### 1. **HTTP/2 Support**

- Full HTTP/2 transport implementation with `ForceAttemptHTTP2: true`
- Connection pooling with configurable limits:
  - MaxIdleConns: 300
  - MaxIdleConnsPerHost: 100
  - MaxConnsPerHost: 200
  - IdleConnTimeout: 120s
- HTTP/2 connection statistics tracking via `httptrace.ClientTrace`
- Real-time monitoring of connection reuse and multiplexing efficiency

#### 2. **Intelligent Gzip Compression**

- Automatic content-type detection for compression eligibility
- Supports: text/\*, application/json, application/javascript, application/xml, and variants
- Client capability detection via Accept-Encoding header
- Proper header management (Content-Encoding, Vary, Content-Length removal)
- Streaming compression with http.Flusher interface support
- Resource cleanup with deferred gzip writer close

#### 3. **Buffer Pool Management**

- Efficient memory reuse using sync.Pool
- 32KB default buffer size optimized for HTTP responses
- Direct slice storage (not pointers) for memory safety
- Automatic buffer validation and capacity management
- Reduces GC pressure and allocation overhead

#### 4. **Connection Pool Statistics**

- Track active and idle HTTP/2 connections
- Monitor connection creation, closure, and reuse rates
- Calculate average request duration using Exponential Moving Average (EMA)
- Thread-safe statistics with RWMutex
- Real-time metrics for performance monitoring

### Bug Fixes

#### 1. **Buffer Pool Pointer Issue**

- **Problem:** Creating pointers to local variables caused invalid memory references
- **Fix:** Store `[]byte` directly in sync.Pool instead of `*[]byte`
- **Impact:** Prevents crashes and memory corruption

#### 2. **IdleConnections Counter**

- **Problem:** Counter could go negative when connections reused before being marked idle
- **Fix:** Added guard condition `if stats.IdleConnections > 0` before decrement
- **Impact:** Accurate connection pool statistics

#### 3. **HTTP to HTTPS Redirect URL**

- **Problem:** Port number included in redirect (e.g., `https://example.com:80`)
- **Fix:** Use `net.SplitHostPort()` to extract hostname without port
- **Impact:** Correct redirect URLs for HTTPS

#### 4. **Gzip Header Timing**

- **Problem:** Headers set after WriteHeader() call, preventing proper compression
- **Fix:** Set all compression headers before calling WriteHeader()
- **Impact:** Compression actually works with proper HTTP headers

#### 5. **HTTP Status Codes**

- **Problem:** Inactive domains returned 404 (Not Found)
- **Fix:** Changed to 503 (Service Unavailable) for correct HTTP semantics
- **Impact:** Proper status code semantics for monitoring and caching

#### 6. **Duplicate Header Setting**

- **Problem:** Gzip headers set multiple times in different places
- **Fix:** Consolidated header setting to single location before WriteHeader
- **Impact:** Cleaner code and no duplicate headers

### Performance Improvements

#### 1. **Compressible Types Optimization**

- **Change:** Moved from function-local to package-level variable
- **Impact:** Eliminates repeated slice allocations on every request
- **Benefit:** Reduced memory allocations and GC pressure

#### 2. **Gzip Writer Resource Management**

- **Change:** Proper deferred cleanup with error logging
- **Impact:** No resource leaks even during panics
- **Benefit:** Better stability under error conditions

#### 3. **Response Writer Simplification**

- **Change:** Removed complex lazy initialization pattern
- **Impact:** Simpler, more predictable code flow
- **Benefit:** Easier debugging and maintenance

### Code Quality Improvements

- Enhanced error logging with context and severity levels
- Improved code documentation and inline comments
- Better separation of concerns in gzip handling
- Thread-safety improvements with proper mutex usage
- Consistent error handling patterns throughout

### Testing & Validation

All changes have been validated for:

- Memory safety (no invalid pointers)
- Thread safety (proper locking mechanisms)
- HTTP specification compliance
- Resource cleanup (no leaks)
- Performance under load

---

## Appendix

### File Size Summary

**Total Lines of Code:** ~7,000+ lines (proxy module: 3,606 lines)

**Largest Files:**

- `cmd/shiroxy/proxy/reverse_proxy.go`: 913 lines (updated with compression)
- `cmd/shiroxy/proxy/balancer.go`: 577 lines
- `cmd/shiroxy/domains/domain.go`: 458 lines
- `cmd/shiroxy/proxy/proxy_handler.go`: 439 lines
- `cmd/shiroxy/proxy/buffer_pool.go`: 55 lines (new)
- `cmd/shiroxy/proxy/connection_pool.go`: 114 lines (new)

### Dependency Graph

```
main.go
├── cli (configuration)
├── logger
├── analytics
├── domains
│   ├── certificate (ACME)
│   └── storage (Redis/Memory)
├── webhook
├── shutdown
│   └── persistence (Protobuf)
├── proxy
│   ├── balancer (load balancing)
│   ├── proxy_handler (TLS, frontend)
│   ├── reverse_proxy (HTTP forwarding)
│   ├── health_check (monitoring)
│   ├── buffer_pool (memory optimization)
│   └── connection_pool (HTTP/2 stats)
└── api
    ├── controllers (domain, backend, analytics)
    ├── routes (routing)
    └── middlewares (auth, response)
```

### Glossary

- **ACME:** Automatic Certificate Management Environment (Let's Encrypt protocol)
- **CSR:** Certificate Signing Request
- **ECC:** Elliptic Curve Cryptography (ECDSA)
- **HTTP-01:** ACME challenge type using HTTP server
- **LRU:** Least Recently Used (caching strategy)
- **PEM:** Privacy-Enhanced Mail (certificate format)
- **SNI:** Server Name Indication (TLS extension)
- **Trie:** Prefix tree data structure

---

**End of Internal Documentation**

For user-facing documentation, see `README.md` and `docs/`.

For API reference, see `docs/api.md` or Swagger UI at `/docs/swagger/`.

For configuration reference, see `configuration.md`.
