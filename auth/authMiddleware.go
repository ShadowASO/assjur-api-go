package auth

import (
	"log"
	"net/http"
	"ocrserver/lib/tools"

	"github.com/gin-gonic/gin"
)

// Estrutura para mensagens de erro ou sucesso na resposta
type ResponseStatus struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

// Middleware para validar o token
func AuthenticateTokenGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("ERROR: AuthenticateTokenGin - token não informado")

			response := tools.CreateResponseMessage("Token não informado!")
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		accessToken, err := ExtractToken(authHeader)
		if err != nil {
			log.Println("ERROR: AuthenticateTokenGin: ", err)

			response := tools.CreateResponseMessage("ExtractToken - Não foi possível extrair o Token!")
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		user, err := ValidateToken(accessToken)

		if err != nil {
			log.Println("ERROR: AuthenticateTokenGin - ", err)

			response := tools.CreateResponseMessage("Acesso não autorizado!")
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return

		}

		c.Set("user", user)
		c.Next()
	}
}
