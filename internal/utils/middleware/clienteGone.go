package middleware

import (
	"time"

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

// file: middleware/deadline_inspector.go

func DeadlineInspector() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if dl, ok := ctx.Deadline(); ok {
			c.Writer.WriteString("[Inspector] has deadline = true\n")
			c.Writer.WriteString("[Inspector] deadline at  = " + dl.Format(time.RFC3339Nano) + "\n")
			c.Writer.WriteString("[Inspector] time.Until  = " + time.Until(dl).String() + "\n")
		} else {
			c.Writer.WriteString("[Inspector] has deadline = false\n")
		}

		// avisa quando o contexto for cancelado (ex.: cliente fecha, WriteTimeout estoura)
		go func() {
			<-ctx.Done()
			// não escreva no response aqui (pode já ter fechado); apenas logue
			// use seu logger:
			// logger.Log.Warnf("[Inspector] ctx.Done(): %v", ctx.Err())
		}()

		c.Next()
	}
}
