package middleware

import (
	"fmt"
	"ocrserver/internal/utils/logger"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()
		c.Next()
		duration := time.Since(start)
		//msg := fmt.Sprintf("Request - Method: %s | Status: %d | Duration: %v | URL: %s", c.Request.Method, c.Writer.Status(), duration, c.Request.URL.Path)
		//_ = fmt.Sprintf("Request - Method: %s | Status: %d | Duration: %v", c.Request.Method, c.Writer.Status(), duration)
		msg := fmt.Sprintf("| %d |  %v | %s  | %s : %s", c.Writer.Status(), duration, c.Request.Method, c.RemoteIP(), c.Request.URL.Path)
		logger.Log.Info(msg)
	}
}
