package controllers

import (
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/webhook"

	"github.com/gin-gonic/gin"
)

type DomainController struct {
	Storage        *domains.Storage
	Middlewares    *middlewares.Middlewares
	WebhookHandler *webhook.WebhookHandler
}

type registerDomainRequestBody struct {
	Domain   string            `json:"domain"`
	Email    string            `json:"email"`
	Metadata map[string]string `json:"user_name"`
}

func (d *DomainController) RegisterDomain(c *gin.Context) {
	var requestBody registerDomainRequestBody
	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	}

	if requestBody.Domain == "" || requestBody.Email == "" {
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   "fields domain and email is required",
		}, 400)
		return
	}

	dnsKey, err := d.Storage.RegisterDomain(requestBody.Domain, requestBody.Email, requestBody.Metadata)
	if err != nil {
		d.WebhookHandler.Fire("domain-register-failed", map[string]string{
			"domain": requestBody.Domain,
		})
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	} else {
		d.WebhookHandler.Fire("domain-register-success", map[string]string{
			"domain": requestBody.Domain,
		})
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: true,
			Data: map[string]any{
				"dns_key": dnsKey,
			},
		}, 200)
	}
}

func (d *DomainController) ForceSSL(c *gin.Context) {
	type forceSslRequestBody struct {
		Domain string `json:"domain"`
	}

	var requestBody forceSslRequestBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	}

	if requestBody.Domain == "" {
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	}

	err = d.Storage.ForceSSL(requestBody.Domain)
	if err != nil {
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	}
	d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
		Success: true,
	}, 200)
}

func (d *DomainController) UpdateDomain(c *gin.Context) {}

func (d *DomainController) RemoveDomain(c *gin.Context) {}

func (d *DomainController) FetchDomainInfo(c *gin.Context) {}

func (d *DomainController) FetchCertificateExpiryInfo(c *gin.Context) {}
