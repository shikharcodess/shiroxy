package controllers

import (
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/storage"

	"github.com/gin-gonic/gin"
)

type DomainController struct {
	storage     *storage.Storage
	middlewares *middlewares.Middlewares
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
		d.middlewares.WriteResponse(c, gin.H{
			"error": err.Error(),
		}, 400)
		return
	}

	if requestBody.Domain == "" || requestBody.Email == "" {
		d.middlewares.WriteResponse(c, gin.H{
			"error": "fields domain and email is required",
		}, 400)
		return
	}

	err = d.storage.RegisterDomain(requestBody.Domain, requestBody.Email, requestBody.Metadata)
	if err != nil {
		d.middlewares.WriteResponse(c, gin.H{
			"error": err.Error(),
		}, 400)
		return
	} else {
		d.middlewares.WriteResponse(c, gin.H{
			"": "",
		}, 200)
	}
}
