package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
	"shiroxy/public"
	"strings"
	"sync"
)

func StartShiroxyHandler(configuration *models.Config, storage *domains.Storage, logHandler *logger.Logger, wg *sync.WaitGroup) {

	var servers []*Server
	for _, server := range configuration.Backend.Servers {
		host := url.URL{
			Scheme: configuration.Default.Mode,
			Host:   fmt.Sprintf("%s:%s", server.Host, server.Port), // The actual address where domain1's server is running
		}

		servers = append(servers, &Server{
			URL:   &host,
			Alive: true,
			Shiroxy: &Shiroxy{
				Logger: logHandler,
				Director: func(req *http.Request) {
					rewriteRequestURL(req, &host)
				},
			},
		})
	}

	loadbalancer := NewLoadBalancer(configuration, servers, wg)

	domainNotFoundErrorResponse := loadErrorPageHtmlContent(public.DOMAIN_NOT_FOUND_ERROR, configuration.Default.ErrorResponses.ErrorPageButtonName, configuration.Default.ErrorResponses.ErrorPageButtonUrl)
	statusInactiveErrorResponse := loadErrorPageHtmlContent(public.STATUS_INACTIVE, configuration.Default.ErrorResponses.ErrorPageButtonName, configuration.Default.ErrorResponses.ErrorPageButtonUrl)

	wg.Add(1)
	go func() {
		if configuration.Frontend.Secure {
			defer wg.Done()

			var responseWriter http.ResponseWriter
			server := &http.Server{
				Addr: fmt.Sprintf("%s:%s", configuration.Frontend.Bind.Host, configuration.Frontend.Bind.Port),
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer func() {
						if rec := recover(); rec != nil {
							logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
							http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
						}
					}()
					responseWriter = w
					loadbalancer.ServeHTTP(w, r)
				}),
				TLSConfig: &tls.Config{
					ClientAuth: resolveSecurityPolicy(configuration.Frontend.SecureVerify),
					GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
						var cert tls.Certificate
						var err error

						var domainName string = strings.TrimSpace(info.ServerName)
						domainMetadata := storage.DomainMetadata[domainName]

						if domainMetadata == nil {
							http.Error(responseWriter, domainNotFoundErrorResponse, http.StatusNotFound)
							return nil, fmt.Errorf("domain not found")
						}

						if domainMetadata.Status == "active" {
							cert, err = tls.X509KeyPair(domainMetadata.CertPemBlock, domainMetadata.KeyPemBlock)
							if err != nil {
								http.Error(responseWriter, domainNotFoundErrorResponse, http.StatusNotFound)
								return nil, fmt.Errorf("something went wrong")
							}
						} else {
							http.Error(responseWriter, statusInactiveErrorResponse, http.StatusNotFound)
							return nil, fmt.Errorf("routing deactivated")
						}

						return &cert, nil
					},
				},
			}

			err := server.ListenAndServeTLS("", "")
			if err != nil {
				logHandler.LogError(err.Error(), "Proxy", "Error")
			}
		} else {
			defer wg.Done()
			server := &http.Server{
				Addr: fmt.Sprintf("%s:%s", configuration.Frontend.Bind.Host, configuration.Frontend.Bind.Port),
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer func() {
						if rec := recover(); rec != nil {
							logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
							http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
						}
					}()
					loadbalancer.ServeHTTP(w, r)
				}),
			}

			logHandler.LogSuccess("Starting Shirxoy In Unsecured Mode", "Proxy", "")
			err := server.ListenAndServe()
			if err != nil {
				logHandler.LogError(err.Error(), "Proxy", "Error")
			}
		}
	}()
}

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

func loadErrorPageHtmlContent(htmlContent, button_name, button_url string) string {

	if button_name == "" {
		button_name = "Shiroxy"
	}
	if button_url == "" {
		button_url = "https://github.com/ShikharY10/shiroxy/issues"
	}

	replacer := strings.NewReplacer(
		"{{button_name}}", button_name,
		"{{button_url}}", button_url,
	)

	// Apply the replacements
	result := replacer.Replace(htmlContent)

	return result
}
