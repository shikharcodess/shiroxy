// Author: @ShikharY10
// Docs Author: ChatGPT

// Package proxy provides functionality for handling reverse proxy operations, including
// load balancing, TLS configurations, and request routing.
package proxy

import (
	"crypto/tls"                  // Provides TLS (Transport Layer Security) support.
	"errors"                      // Defines error handling for functions.
	"fmt"                         // Implements formatted I/O operations.
	"log"                         // Implements simple logging capabilities.
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
)

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
			Scheme: configuration.Frontend.Mode, // Scheme could be HTTP/HTTPS based on frontend mode.
			Host:   fmt.Sprintf("%s:%s", server.Host, server.Port),
		}

		// Append the server to the servers slice.
		servers = append(servers, &Server{
			Id:    server.Id, // Unique identifier for the server.
			URL:   &host,     // URL of the backend server.
			Alive: false,     // Indicates if the server is alive (default to false).

			// Shiroxy structure to hold logger and request director for request URL rewriting.
			Shiroxy: &Shiroxy{
				Logger: logHandler, // Logger for handling log messages.
				Director: func(req *http.Request) {
					RewriteRequestURL(req, &host) // Modifies the request URL for backend routing.
				},
			},
			Tags:           strings.Split(server.Tags, ","), // Splits server tags by comma for tag-based routing.
			Lock:           &sync.RWMutex{},
			HealthCheckUrl: server.HealthUrl,
		})

	}

	// Set the servers to the BackendServers instance.
	backendServers.Servers = servers

	// Create a new LoadBalancer instance with the provided configuration.
	loadbalancer := NewLoadBalancer(configuration, backendServers, webhookHandler, storage, wg)

	// Load error page content to be used for "domain not found" errors.
	domainNotFoundErrorResponse := loadErrorPageHtmlContent(public.DOMAIN_NOT_FOUND_ERROR, &configuration.Default.ErrorResponses)

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
						if configuration.Default.Mode == "dev" {
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
				// HTTP to HTTPS redirection if enabled in the configuration.
				if (configuration.Frontend.HttpToHttps) && r.URL.Port() == "80" && r.TLS == nil {
					secureFrontend := loadbalancer.Frontends["443"]
					if secureFrontend != nil {
						url := fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
						http.Redirect(w, r, url, http.StatusMovedPermanently)
					} else {
						loadbalancer.ServeHTTP(w, r)
					}
				} else {
					loadbalancer.ServeHTTP(w, r) // Serve HTTP request through the load balancer.
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
		if bind.Target == "multiple" {
			server, secure, err = createMultipleTargetServer(&bind, storage, frontend.handlerFunc)
		} else if bind.Target == "single" {
			server, secure, err = createSingleTargetServer(&bind, storage, frontend.handlerFunc)
		}

		if err != nil {
			return nil, err // Return error if server creation fails.
		}

		// Start the server using listenAndServe function.
		listenAndServe(server, secure, logHandler, wg)
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
func createMultipleTargetServer(bindData *models.FrontendBind, storage *domains.Storage, handlerFunc http.HandlerFunc) (server *http.Server, secure bool, err error) {
	if bindData.Secure {
		// Secure server with TLS configuration.
		server := &http.Server{
			Addr:    fmt.Sprintf("%s:%s", bindData.Host, bindData.Port),
			Handler: http.HandlerFunc(handlerFunc),
			TLSConfig: &tls.Config{
				ClientAuth: resolveSecurityPolicy(bindData.SecureSetting.SecureVerify),
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

// createSingleTargetServer sets up an HTTP server for the "single" target mode with optional TLS support.
// Parameters are similar to createMultipleTargetServer.
// Returns similar values to createMultipleTargetServer.
func createSingleTargetServer(bindData *models.FrontendBind, storage *domains.Storage, handlerFunc http.HandlerFunc) (server *http.Server, secure bool, err error) {
	if bindData.Secure {
		var tlsConfig *tls.Config
		if bindData.SecureSetting.SingleTargetMode == "certandkey" {
			tlsConfig = &tls.Config{
				ClientAuth: resolveSecurityPolicy(bindData.SecureSetting.SecureVerify),
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
				ClientAuth: resolveSecurityPolicy(bindData.SecureSetting.SecureVerify),
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

// resolveSecurityPolicy maps the given policy string to a tls.ClientAuthType value.
// Returns the appropriate tls.ClientAuthType based on the policy.
func resolveSecurityPolicy(policy string) tls.ClientAuthType {
	if policy == "none" {
		return tls.NoClientCert
	} else if policy == "optional" {
		return tls.RequestClientCert
	} else if policy == "required" {
		return tls.RequireAndVerifyClientCert
	} else {
		return tls.NoClientCert
	}
}

// loadErrorPageHtmlContent prepares the error page HTML content by replacing placeholders
// in the provided HTML content with values from the configuration.
func loadErrorPageHtmlContent(htmlContent string, config *models.ErrorRespons) string {
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

// listenAndServe starts the HTTP server in a goroutine and handles both secured and unsecured modes.
// It also logs any server start errors.
func listenAndServe(server *http.Server, secure bool, logHandler *logger.Logger, wg *sync.WaitGroup) {
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
