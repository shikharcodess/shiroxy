package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/cmd/shiroxy/types"

	"github.com/gin-gonic/gin"
)

type BackendController struct {
	Middlewares *middlewares.Middlewares
	Context     *types.APIContext
}

func (b *BackendController) FetchAllBackendServers(c *gin.Context) {
	response := map[string]interface{}{}

	var servers []map[string]any = []map[string]any{}
	fmt.Println("1")

	for _, server := range b.Context.LoadBalancer.Servers {
		fmt.Println("server: ", server.Id)
		serverReflect := reflect.ValueOf(server)
		if serverReflect.Kind() == reflect.Ptr {
			serverReflect = serverReflect.Elem()
		}
		serverJson := map[string]any{}
		for i := 0; i < serverReflect.NumField(); i++ {
			serverJson[serverReflect.Type().Field(i).Name] = serverReflect.Field(i).Interface()
		}

		servers = append(servers, serverJson)
	}

	response["servers"] = servers

	b.Middlewares.WriteResponse(c, middlewares.ApiResponse{
		Success: true,
		Data:    response,
	}, 200)
}

func (b *BackendController) RegisterNewBackendServer(c *gin.Context) {
	type RegisternewBackendServerRequestBody struct {
		Id        string `json:"id"`
		Host      string `json:"host"`
		Port      string `json:"port"`
		HealthUrl string `json:"health_url"`
	}
	var requestBody RegisternewBackendServerRequestBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		b.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	}

	serverUrl := url.URL{
		Scheme: b.Context.Configuration.Default.Mode,
		Host:   fmt.Sprintf("%s:%s", requestBody.Host, requestBody.Port), // The actual address where domain1's server is running
	}

	healthCheckUrl := url.URL{
		Scheme: "http",
		Host:   requestBody.HealthUrl, // The actual address where domain1's server is running
	}

	server := proxy.Server{
		Id:                            requestBody.Id,
		URL:                           &serverUrl,
		HealthCheckUrl:                &healthCheckUrl,
		Alive:                         false,
		FireWebhookOnFirstHealthCheck: true,
		Shiroxy: &proxy.Shiroxy{
			Logger: b.Context.LogHandler,
			Director: func(req *http.Request) {
				targetQuery := serverUrl.RawQuery
				req.URL.Scheme = serverUrl.Scheme
				req.URL.Host = serverUrl.Host
				req.URL.Path, req.URL.RawPath = proxy.JoinURLPath(&serverUrl, req.URL)
				if targetQuery == "" || req.URL.RawQuery == "" {
					req.URL.RawQuery = targetQuery + req.URL.RawQuery
				} else {
					req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
				}
			},
		},
	}

	b.Context.LoadBalancer.Servers = append(b.Context.LoadBalancer.Servers, &server)

	b.Middlewares.WriteResponse(c, middlewares.ApiResponse{
		Success: true,
	}, 200)
}

func (b *BackendController) RemoveBackendServer(c *gin.Context) {
	type RemoveBackendServerRequestBody struct {
		Id string `json:"id"`
	}
	var requestBody RemoveBackendServerRequestBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		b.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	}

	if requestBody.Id == "" {
		b.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   "field id is required",
		}, 400)
		return
	}
}
