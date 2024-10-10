package controllers

import (
	"fmt"
	"net/http"
	"shiroxy/cmd/shiroxy/domains"

	"github.com/gin-gonic/gin"
)

func CreateChallenge(c *gin.Context) {
	name := c.Param("name")
	c.String(http.StatusOK, "Hello, %s", name)
}

func GetDNSChallengeSolver(storage *domains.Storage) *DNSChallengeSolver {
	return &DNSChallengeSolver{
		domainStorage: storage,
	}
}

type DNSChallengeSolver struct {
	challenges    map[string]string
	domainStorage *domains.Storage
}

func (d *DNSChallengeSolver) SolveDNSChallenge(c *gin.Context) {
	filename := c.Param("filename")
	fmt.Println("filename: ", filename)
	if filename == "" {
		c.Status(404)
		return
	}

	domainName, ok := d.domainStorage.DnsChallengeToken[filename]
	if !ok {
		c.Status(404)
		return
	}

	domainMetadata, ok := d.domainStorage.DomainMetadata[domainName]
	if !ok {
		c.Status(404)
		return
	}

	challengeKeyAuthorization := domainMetadata.DnsChallengeKey
	if challengeKeyAuthorization == "" {
		c.Status(404)
		return
	}

	fmt.Fprint(c.Writer, challengeKeyAuthorization)
	c.AbortWithStatus(200)
}

func (d *DNSChallengeSolver) SolveDNSChallengeHttpVersion(w http.ResponseWriter, r http.Request) {

}
