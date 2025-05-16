package auth

import (
	"net/http"
	"ocrserver/api/handler/response"
	"ocrserver/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Estrutura para mensagens de erro ou sucesso na resposta
type ResponseStatus struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

// Middleware para validar o token
func AuthenticateTokenGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Log.Error("AuthenticateTokenGin - token não informado!")
			response.HandleError(c, http.StatusUnauthorized, "Acesso não autorizado", "", requestID)
			c.Abort()
			return
		}

		accessToken, err := ExtractToken(authHeader)
		if err != nil {
			logger.Log.Error("ExtractToken - Não foi possível extrair o Token!", err.Error())
			response.HandleError(c, http.StatusUnauthorized, "Acesso não autorizado", "", requestID)
			c.Abort()
			return
		}

		user, err := ValidateToken(accessToken)

		if err != nil {
			logger.Log.Error("Token inválido. Acesso negado!", err.Error())
			response.HandleError(c, http.StatusUnauthorized, "Acesso não autorizado", "", requestID)
			c.Abort()
			return

		}

		c.Set("user", user)
		c.Next()
	}
}
