// Author: @ShikharY10
// Docs Author: AI

// Package proxy provides functionality for handling reverse proxy operations, including
// load balancing, TLS configurations, and request routing.
package proxy

import (
	"crypto/tls" // Provides TLS (Transport Layer Security) support.
	"errors"     // Defines error handling for functions.
	"fmt"        // Implements formatted I/O operations.
	"log"        // Implements simple logging capabilities.
	"net"
	"net/http"                    // Provides HTTP client and server implementations.
	"net/url"                     // Implements URL parsing and manipulation.
	"runtime/debug"               // Provides stack traces to aid debugging.
	"shiroxy/cmd/shiroxy/domains" // Custom package for domain management.
	"shiroxy/cmd/shiroxy/webhook" // Custom package for webhook handling.
	"shiroxy/pkg/logger"          // Custom package for logging support.
	"shiroxy/pkg/models"          // Custom package for configuration models.
	"shiroxy/public"              // Custom package for public constants and assets.
	"strings"                     // Implements string manipulation functions.
	"sync"                        // Provides synchronization primitives.
	"time"
)

var DnsChallengeSolverMapped bool = false

// StartShiroxyHandler sets up and starts the Shiroxy reverse proxy with a load balancer.
// Parameters:
//   - configuration: *models.Config, contains the configuration settings.
//   - storage: *domains.Storage, holds domain metadata.
//   - webhookHandler: *webhook.WebhookHandler, handles webhooks.
//   - logHandler: *logger.Logger, custom logging utility.
//   - wg: *sync.WaitGroup, synchronization primitive to wait for goroutines.
//
// Returns:
//   - *LoadBalancer: a load balancer instance configured with the specified settings.
//   - error: error if any issues occur during setup.
func StartShiroxyHandler(configuration *models.Config, storage *domains.Storage, webhookHandler *webhook.WebhookHandler, logHandler *logger.Logger, wg *sync.WaitGroup) (*LoadBalancer, error) {

	// Initialize a BackendServers instance to hold the backend server configurations.
	var backendServers *BackendServers = &BackendServers{}
	var servers []*Server // Slice to hold individual server configurations.

	// Loop through each server specified in the configuration to set up routing.
	for _, server := range configuration.Backend.Servers {
		// Construct the URL for each server using its host and port from the configuration.
		host := url.URL{
			// Scheme could be HTTP/HTTPS based on frontend mode.
			Scheme: configuration.Frontend.Mode,
			Host:   fmt.Sprintf("%s:%s", server.Host, server.Port),
		}

		// Append the server to the servers slice.
		servers = append(servers, &Server{

			// Unique identifier for the server.
			Id: server.Id,

			// URL of the backend server.
			URL: &host,

			// Indicates if the server is alive (default to false).
			Alive: false,

			// Shiroxy structure to hold logger and request director for request URL rewriting.
			Shiroxy: &Shiroxy{
				// Logger for handling log messages.
				Logger: logHandler,
				Director: func(req *http.Request) {
					// Modifies the request URL for backend routing.
					RewriteRequestURL(req, &host)
				},
				Transport: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 60 * time.Second, // Increased keep-alive for better connection reuse
						DualStack: true,             // Enable IPv4/IPv6 dual stack
					}).DialContext,
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          300,               // Increased total idle connections
					IdleConnTimeout:       120 * time.Second, // Increased timeout for better reuse
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
					MaxIdleConnsPerHost:   100,                    // Balanced value for connection pooling
					MaxConnsPerHost:       200,                    // Setting a reasonable limit to prevent overwhelming backends
					WriteBufferSize:       int(DefaultBufferSize), // Use our buffer size constant
					ReadBufferSize:        int(DefaultBufferSize), // Use our buffer size constant
					TLSClientConfig: &tls.Config{
						MinVersion:         tls.VersionTLS12, // Ensure modern TLS
						InsecureSkipVerify: false,            // Always validate certificates in production
					},
					// HTTP/2 specific settings
					// These are new settings that enhance HTTP/2 performance
					MaxResponseHeaderBytes: 64 * 1024,

					// Disable compression because we'll handle it separately
					DisableCompression: true,
				},
				BufferPool: NewSyncBufferPool(32 * 1024),
			},
			// Splits server tags by comma for tag-based routing.
			Tags:           strings.Split(server.Tags, ","),
			Lock:           &sync.RWMutex{},
			HealthCheckUrl: server.HealthUrl,
		})
	}

	// Set the servers to the BackendServers instance.
	backendServers.Servers = servers

	// Create a new LoadBalancer instance with the provided configuration.
	loadbalancer := NewLoadBalancer(configuration, backendServers, webhookHandler, storage, wg)

	// Load error page content to be used for "domain not found" errors.
	domainNotFoundErrorResponse := LoadErrorPageHtmlContent(public.DOMAIN_NOT_FOUND_ERROR, &configuration.Default.ErrorResponses)

	// Loop through each bind (frontend port binding) specified in the configuration.
	for _, bind := range configuration.Frontend.Bind {
		// Validate the secure setting for "single" target mode.
		if bind.Target == "single" && bind.SecureSetting.SingleTargetMode == "" {
			logHandler.Log("securesetting field is required when bind target is set to 'single'", "Proxy", "Error")
			panic("") // Panic to indicate a critical configuration issue.
		}

		// Define the handler function for each frontend binding.
		frontend := Frontends{
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Recover from any panics during request handling to avoid crashing the server.
				defer func() {
					if rec := recover(); rec != nil {
						if configuration.Default.DebugMode == "dev" {
							fmt.Printf("Panic occurred: %v\n", rec)
							debug.PrintStack() // Print stack trace for debugging in development mode.
						}
						logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
						w.Header().Add("Content-Type", "text/html")
						w.WriteHeader(400) // Write a 400 Bad Request status.
						_, err := w.Write([]byte(domainNotFoundErrorResponse))
						if err != nil {
							log.Printf("failed to write response: %v", err)
						}
					}
				}()

				// Implementation for exposing dns HTTP-01 challenge solver endpoint on port 80
				if (r.URL.Port() == "80" || r.URL.Port() == "") && strings.HasPrefix(r.RequestURI, "/.well-known/acme-challenge/") {
					// Todo: Implement new custom router
					filename := strings.TrimPrefix(r.RequestURI, "/.well-known/acme-challenge/")
					if filename == "" {
						http.Error(w, "Filename not found", http.StatusBadRequest)
						return
					}
					domainName, ok := storage.DnsChallengeToken[filename]
					if !ok {
						http.Error(w, "no domain found for filename", http.StatusBadRequest)
						return
					}

					domainMetadata, ok := storage.DomainMetadata[domainName]
					if !ok {
						http.Error(w, "record not found for domain", http.StatusBadRequest)
						return
					}

					if domainMetadata.DnsChallengeKey == "" {
						if !ok {
							http.Error(w, "dns challenge authorization and solving key not found", http.StatusBadRequest)
							return
						}
					}

					fmt.Fprint(w, domainMetadata.DnsChallengeKey)
					return
				}

				// HTTP to HTTPS redirection if enabled in the configuration.
				if (configuration.Frontend.HttpToHttps) && r.URL.Port() == "80" && r.TLS == nil {
					secureFrontend := loadbalancer.Frontends["443"]
					if secureFrontend != nil {
						// Strip port from host for HTTPS redirect
						// r.Host may be "example.com:80", we need just "example.com" for HTTPS
						host := r.Host
						if h, _, err := net.SplitHostPort(r.Host); err == nil {
							host = h // Use hostname without port
						}

						redirectUrl := url.URL{
							Scheme:   "https",
							Host:     host,
							Path:     r.URL.Path,
							RawQuery: r.URL.RawQuery,
						}
						http.Redirect(w, r, redirectUrl.String(), http.StatusMovedPermanently)
					} else {
						domainName := strings.TrimSpace(r.Host)
						domainMetadata, ok := storage.DomainMetadata[domainName]

						if !ok || domainMetadata == nil {
							w.Header().Add("Content-Type", "text/html")
							w.WriteHeader(404)
							_, err := w.Write([]byte(domainNotFoundErrorResponse))
							if err != nil {
								log.Printf("failed to write response: %v", err)
							}
							return
						}

						if domainMetadata.Status == "inactive" {
							if strings.Contains(r.RequestURI, ".well-known/acme-challenge") {
								loadbalancer.ServeHTTP(w, r)
							} else {
								w.Header().Add("Content-Type", "text/html")
								w.WriteHeader(404)
								_, err := w.Write([]byte(domainNotFoundErrorResponse))
								if err != nil {
									log.Printf("failed to write response: %v", err)
								}
							}
						} else {
							loadbalancer.ServeHTTP(w, r)
						}
					}
				} else {
					loadbalancer.ServeHTTP(w, r)
				}
			}),
		}
		// Add the frontend handler to the load balancer.
		loadbalancer.Frontends[bind.Port] = &frontend

		// Check for valid target modes ("multiple" or "single").
		if bind.Target != "multiple" && bind.Target != "single" {
			logHandler.Log("Invalid target value in frontend configuration", "Proxy", "Error")
			panic("")
		}

		var server *http.Server
		var secure bool
		var err error

		// Create HTTP server based on the target mode.
		switch bind.Target {
		case "multiple":
			server, secure, err = CreateMultipleTargetServer(&bind, storage, frontend.handlerFunc)
		case "single":
			server, secure, err = CreateSingleTargetServer(&bind, storage, frontend.handlerFunc)
		}

		if err != nil {
			return nil, err // Return error if server creation fails.
		}

		// Start the server using listenAndServe function.
		ListenAndServe(server, secure, logHandler, wg)
	}
	return loadbalancer, nil
}

// createMultipleTargetServer sets up an HTTP server for the "multiple" target mode with optional TLS support.
// Parameters:
//   - bindData: *models.FrontendBind, contains the server's binding information.
//   - storage: *domains.Storage, holds domain metadata for TLS configuration.
//   - handlerFunc: http.HandlerFunc, the request handler function.
//
// Returns:
//   - *http.Server: an HTTP server instance.
//   - bool: indicates if the server is secure (using TLS).
//   - error: error if any issues occur during server creation.
func CreateMultipleTargetServer(bindData *models.FrontendBind, storage *domains.Storage, handlerFunc http.HandlerFunc) (server *http.Server, secure bool, err error) {
	if bindData.Secure {
		// Secure server with TLS configuration.
		server := &http.Server{
			Addr:    fmt.Sprintf("%s:%s", bindData.Host, bindData.Port),
			Handler: http.HandlerFunc(handlerFunc),

			// Timeouts to prevent slow clients from consuming resources
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,

			// Increase maximum header size if needed
			MaxHeaderBytes: 1 << 20,

			// Custom TLS
			TLSConfig: &tls.Config{
				// TLS configuration for HTTP/2
				MinVersion: tls.VersionTLS12,
				ClientAuth: ResolveSecurityPolicy(bindData.SecureSetting.SecureVerify),
				GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
					// Load TLS certificate based on the domain metadata.
					var cert tls.Certificate
					var err error
					domainName := strings.TrimSpace(info.ServerName)
					domainMetadata := storage.DomainMetadata[domainName]

					if domainMetadata == nil {
						return nil, fmt.Errorf("domain not found")
					}

					if domainMetadata.Status == "active" {
						cert, err = tls.X509KeyPair(domainMetadata.CertPemBlock, domainMetadata.KeyPemBlock)
						if err != nil {
							fmt.Println("tls.X509KeyPair ERROR: ", err.Error())
							return nil, fmt.Errorf("something went wrong")
						}
					} else {
						return nil, fmt.Errorf("routing deactivated")
					}

					return &cert, nil
				},
			},
		}

		return server, true, nil
	} else {
		// Unsecured server.
		server := &http.Server{
			Addr:    fmt.Sprintf("%s:%s", bindData.Host, bindData.Port),
			Handler: handlerFunc,
		}
		return server, false, nil
	}
}

// CreateSingleTargetServer sets up an HTTP server for the "single" target mode with optional TLS support.
// Parameters are similar to createMultipleTargetServer.
// Returns similar values to createMultipleTargetServer.
func CreateSingleTargetServer(bindData *models.FrontendBind, storage *domains.Storage, handlerFunc http.HandlerFunc) (server *http.Server, secure bool, err error) {
	if bindData.Secure {
		var tlsConfig *tls.Config
		if bindData.SecureSetting.SingleTargetMode == "certandkey" {
			tlsConfig = &tls.Config{
				ClientAuth: ResolveSecurityPolicy(bindData.SecureSetting.SecureVerify),
				ServerName: bindData.SecureSetting.CertAndKey.Domain,
				GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
					cert, err := tls.LoadX509KeyPair(bindData.SecureSetting.CertAndKey.Cert, bindData.SecureSetting.CertAndKey.Key)
					if err != nil {
						return nil, err
					}
					return &cert, nil
				},
			}
		} else if bindData.SecureSetting.SingleTargetMode == "shiroxyshinglesecure" {
			tlsConfig = &tls.Config{
				ClientAuth: ResolveSecurityPolicy(bindData.SecureSetting.SecureVerify),
				ServerName: bindData.SecureSetting.CertAndKey.Domain,
				GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
					domainMetadata, ok := storage.DomainMetadata[info.ServerName]
					if !ok {
						return nil, errors.New("certificate not found")
					}
					if domainMetadata.Status == "active" {
						cert, err := tls.X509KeyPair(domainMetadata.CertPemBlock, domainMetadata.KeyPemBlock)
						if err != nil {
							fmt.Println("tls.X509KeyPair ERROR: ", err.Error())
							return nil, fmt.Errorf("something went wrong")
						}
						return &cert, nil
					} else {
						return nil, fmt.Errorf("routing deactivated")
					}
				},
			}
		}
		server := &http.Server{
			Addr:      fmt.Sprintf("%s:%s", bindData.Host, bindData.Port),
			Handler:   handlerFunc,
			TLSConfig: tlsConfig,
		}

		return server, true, nil
	} else {
		// Unsecured server.
		server := &http.Server{
			Addr:    fmt.Sprintf("%s:%s", bindData.Host, bindData.Port),
			Handler: handlerFunc,
		}

		return server, false, nil
	}
}

// ResolveSecurityPolicy maps the given policy string to a tls.ClientAuthType value.
// Returns the appropriate tls.ClientAuthType based on the policy.
func ResolveSecurityPolicy(policy string) tls.ClientAuthType {
	switch policy {
	case "none":
		return tls.NoClientCert
	case "optional":
		return tls.RequestClientCert
	case "required":
		return tls.RequireAndVerifyClientCert
	default:
		return tls.NoClientCert
	}
}

// LoadErrorPageHtmlContent prepares the error page HTML content by replacing placeholders
// in the provided HTML content with values from the configuration.
func LoadErrorPageHtmlContent(htmlContent string, config *models.ErrorRespons) string {
	if config.ErrorPageButtonName == "" {
		config.ErrorPageButtonName = "Shiroxy"
	}
	if config.ErrorPageButtonUrl == "" {
		config.ErrorPageButtonUrl = "https://github.com/ShikharY10/shiroxy"
	}

	replacer := strings.NewReplacer(
		"{{button_name}}", config.ErrorPageButtonName,
		"{{button_url}}", config.ErrorPageButtonUrl,
	)

	result := replacer.Replace(htmlContent)
	return result
}

// ListenAndServe starts the HTTP server in a goroutine and handles both secured and unsecured modes.
// It also logs any server start errors.
func ListenAndServe(server *http.Server, secure bool, logHandler *logger.Logger, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if secure {
			err := server.ListenAndServeTLS("", "")
			if err != nil {
				logHandler.LogError(fmt.Sprintf("while starting secured server: %s", err.Error()), "Proxy", "Error")
			}
		} else {
			err := server.ListenAndServe()
			if err != nil {
				logHandler.LogError(fmt.Sprintf("while starting unsecured server: %s", err.Error()), "Proxy", "Error")
			}
		}
	}()
}
