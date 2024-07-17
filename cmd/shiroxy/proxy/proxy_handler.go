package proxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
	"shiroxy/public"
	"shiroxy/utils"
	"strings"
	"sync"
)

func StartShiroxyHandler(configuration *models.Config, storage *domains.Storage, logHandler *logger.Logger, wg *sync.WaitGroup) (*LoadBalancer, error) {

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

	domainNotFoundErrorResponse := loadErrorPageHtmlContent(public.DOMAIN_NOT_FOUND_ERROR, &configuration.Default.ErrorResponses)
	// statusInactiveErrorResponse := loadErrorPageHtmlContent(public.STATUS_INACTIVE, &configuration.Default.ErrorResponses)

	// ONLY FOR TESTING ++++++ Load Pebble's CA certificate +++++++++++++
	// fileContent, err := os.ReadFile("/home/shikharcode/Main/opensource/shiroxy/temp/pebble/test/certs/pebble.minica.pem")
	// if err != nil {
	// 	log.Fatalf("Failed to read Pebble CA certificate: %v", err)
	// }
	// caCertPool := x509.NewCertPool()
	// caCertPool.AppendCertsFromPEM(fileContent)

	for _, bind := range configuration.Frontend.Bind {
		utils.LogStruct(bind)
		if bind.Target == "single" && bind.SecureSetting.SingleTargetMode == "" {
			logHandler.Log("securesetting field is required when bind target is set to 'single'", "Proxy", "Error")
			panic("")
		}

		frontend := Frontends{
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if rec := recover(); rec != nil {
						logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
						http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
					}
				}()
				if (configuration.Frontend.HttpToHttps) && r.URL.Port() == "80" && r.TLS == nil {
					secureFrontend := loadbalancer.Frontends["443"]
					if secureFrontend != nil {
						url := fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
						http.Redirect(w, r, url, http.StatusMovedPermanently)
					} else {
						loadbalancer.ServeHTTP(w, r)
					}
				} else {
					loadbalancer.ServeHTTP(w, r)
				}
			}),
		}
		loadbalancer.Frontends[bind.Port] = &frontend

		fmt.Println("target: ", bind.Target)
		fmt.Println("secure: ", bind.Secure)

		wg.Add(1)
		go func(bindData *models.FrontendBind) {
			defer wg.Done()
			if bindData.Target == "multiple" {
				if bindData.Secure {
					// frontend := Frontends{
					// 	handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// 		defer func() {
					// 			if rec := recover(); rec != nil {
					// 				logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
					// 				http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
					// 			}
					// 		}()
					// 		loadbalancer.ServeHTTP(w, r)
					// 	}),
					// }
					// loadbalancer.Frontends[bindData.Port] = &frontend
					server := &http.Server{
						Addr:    fmt.Sprintf("%s:%s", bindData.Host, bindData.Port),
						Handler: http.HandlerFunc(frontend.handlerFunc),
						// Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// 	defer func() {
						// 		if rec := recover(); rec != nil {
						// 			logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
						// 			http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
						// 		}
						// 	}()
						// 	loadbalancer.ServeHTTP(w, r)
						// }),
						TLSConfig: &tls.Config{
							ClientAuth: resolveSecurityPolicy(bindData.SecureSetting.SecureVerify),
							GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
								var cert tls.Certificate
								var err error

								var domainName string = strings.TrimSpace(info.ServerName)
								domainMetadata := storage.DomainMetadata[domainName]

								if domainMetadata == nil {
									// http.Error(responseWriter, domainNotFoundErrorResponse, http.StatusNotFound)
									return nil, fmt.Errorf("domain not found")
								}

								if domainMetadata.Status == "active" {
									cert, err = tls.X509KeyPair(domainMetadata.CertPemBlock, domainMetadata.KeyPemBlock)
									if err != nil {
										fmt.Println("tls.X509KeyPair ERROR: ", err.Error())
										// http.Error(responseWriter, domainNotFoundErrorResponse, http.StatusNotFound)
										return nil, fmt.Errorf("something went wrong")
									}
								} else {
									// http.Error(responseWriter, statusInactiveErrorResponse, http.StatusNotFound)
									return nil, fmt.Errorf("routing deactivated")
								}

								return &cert, nil
							},
							// RootCAs: caCertPool,
						},
					}

					err := server.ListenAndServeTLS("", "")
					if err != nil {
						logHandler.LogError(err.Error(), "Proxy", "Error")
					}
				} else {
					// frontend := Frontends{
					// 	handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// 		defer func() {
					// 			if rec := recover(); rec != nil {
					// 				logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
					// 				http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
					// 			}
					// 		}()
					// 		loadbalancer.ServeHTTP(w, r)
					// 	}),
					// }
					// loadbalancer.Frontends[bindData.Port] = &frontend
					server := &http.Server{
						Addr:    fmt.Sprintf("%s:%s", bindData.Host, bindData.Port),
						Handler: frontend.handlerFunc,
						// Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// 	defer func() {
						// 		if rec := recover(); rec != nil {
						// 			logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
						// 			http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
						// 		}
						// 	}()
						// 	loadbalancer.ServeHTTP(w, r)
						// }),
					}

					logHandler.LogSuccess("Starting Shirxoy In Unsecured Mode", "Proxy", "Success")
					err := server.ListenAndServe()
					if err != nil {
						logHandler.LogError(err.Error(), "Proxy", "Error")
					}
				}
			} else if bindData.Target == "single" {
				fmt.Println("Single mode Reached =========")
				if bindData.Secure {
					fmt.Println("single and secured")
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
					// frontend := Frontends{
					// 	handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// 		defer func() {
					// 			if rec := recover(); rec != nil {
					// 				logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
					// 				http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
					// 			}
					// 		}()
					// 		loadbalancer.ServeHTTP(w, r)
					// 	}),
					// }
					// loadbalancer.Frontends[bindData.Port] = &frontend
					server := &http.Server{
						Addr: fmt.Sprintf("%s:%s", bindData.Host, bindData.Port),
						// Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// 	defer func() {
						// 		if rec := recover(); rec != nil {
						// 			logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
						// 			http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
						// 		}
						// 	}()
						// 	loadbalancer.ServeHTTP(w, r)
						// }),
						Handler:   frontend.handlerFunc,
						TLSConfig: tlsConfig,
					}

					err := server.ListenAndServeTLS("", "")
					if err != nil {
						logHandler.LogError(err.Error(), "Proxy", "Error")
					}
				} else {
					fmt.Println("single and unsecured")
					// frontend := Frontends{
					// 	handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// 		defer func() {
					// 			if rec := recover(); rec != nil {
					// 				logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
					// 				http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
					// 			}
					// 		}()
					// 		loadbalancer.ServeHTTP(w, r)
					// 	}),
					// }
					// loadbalancer.Frontends[bindData.Port] = &frontend
					server := &http.Server{
						Addr:    fmt.Sprintf("%s:%s", bindData.Host, bindData.Port),
						Handler: frontend.handlerFunc,
						// Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// 	defer func() {
						// 		if rec := recover(); rec != nil {
						// 			logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
						// 			http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
						// 		}
						// 	}()
						// 	loadbalancer.ServeHTTP(w, r)
						// }),
					}

					logHandler.LogSuccess("Starting Shirxoy In Unsecured Mode", "Proxy", "Success")
					err := server.ListenAndServe()
					if err != nil {
						logHandler.LogError(err.Error(), "Proxy", "Error")
					}
				}
			} else {
				logHandler.Log("Invalid target value in frontend configuraton", "Proxy", "Error")
				panic("")
			}
		}(&bind)
	}

	// wg.Add(1)
	// go func() {
	// 	if configuration.Frontend.Secure {
	// 		defer wg.Done()

	// 		// var responseWriter http.ResponseWriter
	// 		server := &http.Server{
	// 			Addr: fmt.Sprintf("%s:%s", configuration.Frontend.Bind.Host, configuration.Frontend.Bind.Port),
	// 			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 				defer func() {
	// 					if rec := recover(); rec != nil {
	// 						logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
	// 						http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
	// 					}
	// 				}()
	// 				// responseWriter = w
	// 				loadbalancer.ServeHTTP(w, r)
	// 			}),
	// 			TLSConfig: &tls.Config{
	// 				ClientAuth: resolveSecurityPolicy(configuration.Frontend.SecureVerify),
	// 				ServerName: "shikharcode.in",
	// 				GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	// 					var cert tls.Certificate
	// 					var err error

	// 					var domainName string = strings.TrimSpace(info.ServerName)
	// 					domainMetadata := storage.DomainMetadata[domainName]

	// 					if domainMetadata == nil {
	// 						// http.Error(responseWriter, domainNotFoundErrorResponse, http.StatusNotFound)
	// 						return nil, fmt.Errorf("domain not found")
	// 					}

	// 					if domainMetadata.Status == "active" {
	// 						cert, err = tls.X509KeyPair(domainMetadata.CertPemBlock, domainMetadata.KeyPemBlock)
	// 						if err != nil {
	// 							fmt.Println("tls.X509KeyPair ERROR: ", err.Error())
	// 							// http.Error(responseWriter, domainNotFoundErrorResponse, http.StatusNotFound)
	// 							return nil, fmt.Errorf("something went wrong")
	// 						}
	// 						// fmt.Println("")
	// 						// utils.LogStruct(cert)
	// 					} else {
	// 						// http.Error(responseWriter, statusInactiveErrorResponse, http.StatusNotFound)
	// 						return nil, fmt.Errorf("routing deactivated")
	// 					}

	// 					return &cert, nil
	// 				},
	// 				RootCAs: caCertPool,
	// 			},
	// 		}

	// 		err := server.ListenAndServeTLS("", "")
	// 		if err != nil {
	// 			logHandler.LogError(err.Error(), "Proxy", "Error")
	// 		}
	// 	} else {
	// 		defer wg.Done()
	// 		server := &http.Server{
	// 			Addr: fmt.Sprintf("%s:%s", configuration.Frontend.Bind.Host, configuration.Frontend.Bind.Port),
	// 			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 				defer func() {
	// 					if rec := recover(); rec != nil {
	// 						logHandler.LogError(fmt.Sprintf("Recovered from panic: %v", rec), "Proxy", "Error")
	// 						http.Error(w, domainNotFoundErrorResponse, http.StatusInternalServerError)
	// 					}
	// 				}()
	// 				loadbalancer.ServeHTTP(w, r)
	// 			}),
	// 		}

	// 		logHandler.LogSuccess("Starting Shirxoy In Unsecured Mode", "Proxy", "Success")
	// 		err := server.ListenAndServe()
	// 		if err != nil {
	// 			logHandler.LogError(err.Error(), "Proxy", "Error")
	// 		}
	// 	}
	// }()

	return loadbalancer, nil
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

func loadErrorPageHtmlContent(htmlContent string, config *models.ErrorRespons) string {

	if config.ErrorPageButtonName == "" {
		config.ErrorPageButtonName = "Shiroxy"
	}
	if config.ErrorPageButtonUrl == "" {
		config.ErrorPageButtonUrl = "https://github.com/ShikharY10/shiroxy/issues"
	}

	replacer := strings.NewReplacer(
		"{{button_name}}", config.ErrorPageButtonName,
		"{{button_url}}", config.ErrorPageButtonUrl,
	)

	// Apply the replacements
	result := replacer.Replace(htmlContent)

	return result
}
