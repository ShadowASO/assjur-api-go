package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContextKeyRequestID é a chave usada para armazenar o requestID no contexto Gin
const ContextKeyRequestID = "RequestID"

// RequestIDMiddleware gera um ID único para cada requisição HTTP e o adiciona ao contexto
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()

		// Armazena no contexto Gin
		c.Set(ContextKeyRequestID, requestID)

		// Também adiciona no header de resposta para rastreamento no cliente
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Prossegue com o próximo handler
		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	reqID, exists := c.Get(ContextKeyRequestID)
	if !exists {
		return "unknown"
	}
	if s, ok := reqID.(string); ok {
		return s
	}
	return "unknown"
}
