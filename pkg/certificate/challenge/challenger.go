package challenger

import (
	"fmt"
	"net/http"
	"shiroxy/pkg/loader"
	"time"

	"github.com/gin-gonic/gin"
)

func StartChallenger(port string, loaderController *loader.ProgressLoaderPayload) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.GET("/", hello)
	router.GET("/.well-known/acme-challenge/:filename", CreateChallenge)

	go func(port string, loaderController *loader.ProgressLoaderPayload) {
		time.Sleep(1 * time.Second)
		var count int = 0
		for {
			count++
			time.Sleep(1 * time.Second)
			response, _ := http.Get("http://0.0.0.0:" + port + "/")
			if response.StatusCode == 200 {
				loaderController.CloseLoader(false, "", 1000)
			} else {
				if count >= 3 {
					loaderController.CloseLoader(true, "Failed", 1000)
				}
			}
		}
	}(port, loaderController)

	host := fmt.Sprintf(":%s", port)
	err := router.Run(host)
	return err
}

func hello(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World!",
	})
}

func CreateChallenge(c *gin.Context) {
	fmt.Println("=========================================================")
	filename := c.Param("filename")
	fmt.Println("filename: ", filename)
	c.JSON(200, gin.H{
		"message": "Hello World",
	})
}
