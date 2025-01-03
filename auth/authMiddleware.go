package auth

import (
	"log"
	"net/http"

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
			c.JSON(http.StatusUnauthorized, ResponseStatus{StatusCode: 401, Message: "Token não informado"})
			c.Abort()
			return
		}

		accessToken, err := ExtractToken(authHeader)
		if err != nil {
			log.Println("ERROR: AuthenticateTokenGin: ", err)
			c.JSON(http.StatusUnauthorized, ResponseStatus{StatusCode: 401, Message: err.Error()})
			c.Abort()
			return
		}

		user, err := ValidateToken(accessToken)

		if err != nil {
			log.Println("ERROR: AuthenticateTokenGin - ", err)
			c.JSON(http.StatusUnauthorized, ResponseStatus{StatusCode: 401, Message: "Acesso não autorizado"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
