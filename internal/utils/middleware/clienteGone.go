package middleware

import (
	"github.com/gin-gonic/gin"
)

// ClientGoneMiddleware aborta a requisição com 499 ("Client Closed Request")
// se o cliente/proxy encerrar a conexão (ctx.Done()) enquanto o handler ainda roda.
//
// Observação importante: o gin.Context não é thread-safe. Este middleware
// só chama métodos de escrita (AbortWithStatus) a partir da goroutine de
// observação. Em workloads críticos, prefira também checar ctx.Done() dentro
// dos seus próprios handlers/serviços e retornar cedo.
func ClientGoneMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Contexto do request (cancela quando cliente desconecta)
		reqCtx := c.Request.Context()

		// Canal para sinalizar que a chain acabou (c.Next() retornou)
		finished := make(chan struct{})

		// Goroutine que observa cancelamento do cliente
		go func() {
			select {
			case <-reqCtx.Done():
				// Cliente encerrou: aborta pipeline. 499 é um código
				// comum para "Client Closed Request".
				// Use apenas um dos dois (com ou sem JSON).
				// c.AbortWithStatus(499)
				c.AbortWithStatusJSON(499, gin.H{
					"error": "client closed request",
				})
			case <-finished:
				// Handler terminou normalmente
			}
		}()

		// Prossegue para os próximos middlewares/handlers
		c.Next()

		// Sinaliza que terminamos
		close(finished)
	}
}
