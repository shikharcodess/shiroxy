package middlewares

import (
	"encoding/json"
	"fmt"
	"shiroxy/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Middlewares struct {
	logHandler *logger.Logger
	moduleName string
}

type ApiResponse struct {
	Success  bool   `json:"success"`
	Error    any    `json:"error"`
	Data     any    `json:"data"`
	Warning  string `json:"warning"`
	Object   string `json:"object"`
	Metadata any    `json:"metadata"`
}

func InitializeMiddleware(logHandler *logger.Logger, moduleName string) (*Middlewares, error) {
	return &Middlewares{
		logHandler: logHandler,
		moduleName: moduleName,
	}, nil
}

func (m *Middlewares) WriteResponse(ginContext *gin.Context, response ApiResponse, status int) {
	data, err := json.Marshal(response)
	if err != nil {
		m.logHandler.LogError(err.Error(), "Middleware", "")
	}

	ginContext.Writer.Header().Set("Content-Type", "application/json")
	ginContext.Writer.WriteHeader(status)
	_, err = ginContext.Writer.Write(data)
	if err != nil {
		m.logHandler.LogError(err.Error(), "Middleware", "")
	}
	m.logHandler.Log(fmt.Sprintf("[%[1]d] [%[2]s]", status, ginContext.Request.URL.String()), "Middleware", m.moduleName)
}

func (m *Middlewares) CheckAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("[Middleware] [CheckAccess] [Success] ...")
		c.Next()
	}
}
