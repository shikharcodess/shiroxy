package controllers

import (
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/types"
	"shiroxy/utils"

	"github.com/gin-gonic/gin"
)

type DomainController struct {
	Context     *types.APIContext
	Middlewares *middlewares.Middlewares
}

func (d *DomainController) RegisterDomain(c *gin.Context) {
	type registerDomainRequestBody struct {
		Domain   string            `json:"domain"`
		Email    string            `json:"email"`
		Metadata map[string]string `json:"metadata"`
	}

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

	dnsKey, err := d.Context.DomainStorage.RegisterDomain(requestBody.Domain, requestBody.Email, requestBody.Metadata)
	if err != nil {
		d.Context.WebhookHandler.Fire("domain-register-failed", map[string]string{
			"domain": requestBody.Domain,
		})
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	} else {
		d.Context.WebhookHandler.Fire("domain-register-success", map[string]string{
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
			Error:   "field domain is required",
		}, 400)
		return
	}

	err = d.Context.DomainStorage.ForceSSL(requestBody.Domain)
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

func (d *DomainController) UpdateDomain(c *gin.Context) {
	type UpdateDomainRequestBody struct {
		Metadata map[string]string `json:"metadata"`
	}

	domainName := c.Param("domain")

	var requestBody UpdateDomainRequestBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	}

	domainData := d.Context.DomainStorage.DomainMetadata[domainName]
	if requestBody.Metadata != nil {
		for key, value := range requestBody.Metadata {
			domainData.Metadata[key] = value
		}
	}

	data := utils.DestructureStruct(domainData)

	d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
		Success: true,
		Error:   "",
		Data:    data,
	}, 200)
}

func (d *DomainController) RemoveDomain(c *gin.Context) {
	domainName := c.Param("domain")
	domainData := d.Context.DomainStorage.DomainMetadata[domainName]

	err := d.Context.DomainStorage.RemoveDomain(domainName)
	if err != nil {
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   err.Error(),
		}, 400)
		return
	}

	data := utils.DestructureStruct(domainData)

	d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
		Success: true,
		Error:   "",
		Data:    data,
	}, 200)
}

func (d *DomainController) FetchDomainInfo(c *gin.Context) {
	domainName := c.Param("domain")

	domainData := d.Context.DomainStorage.DomainMetadata[domainName]
	if domainData == nil {
		d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
			Success: false,
			Error:   "domain not found",
			Data:    nil,
		}, 400)
	}
	data := utils.DestructureStruct(&domainData)
	d.Middlewares.WriteResponse(c, middlewares.ApiResponse{
		Success: true,
		Error:   "",
		Data:    data,
	}, 200)
}
