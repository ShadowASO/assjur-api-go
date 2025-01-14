package msgs

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Função imprime no log o time atual, que pode ser o início ou o fim de um processo
func CreateLogTimeMessage(message string) {
	log.Printf(message+": %s", time.Now().Format("2006-01-02 15:04:05"))
}

func CreateResponseMessage(message string) gin.H {
	log.Printf("Erro: %s", message)
	return gin.H{
		"message": message,
	}
}
func CreateResponseErrorMessage(c *gin.Context, status int, message string) {
	log.Printf("Erro: %s", message)
	c.JSON(status, gin.H{"mensagem": message})
}

func CreateResponse(ok bool, statusCode int, message string, rows interface{}) gin.H {
	return gin.H{
		// "ok":         ok,
		// "statusCode": statusCode,
		"message": message,
		"rows":    rows,
	}
}

func CreateResponseSelectRows(ok bool, statusCode int, message string, rows interface{}) gin.H {
	return gin.H{
		// "ok":         ok,
		// "statusCode": statusCode,
		"message": message,
		"rows":    rows,
	}
}
func CreateResponseSelectSingle(ok bool, statusCode int, message string, row interface{}) gin.H {
	return gin.H{
		// "ok":         ok,
		// "statusCode": statusCode,
		"message": message,
		"data":    row,
	}
}

func CreateResponseUserInsert(ok bool, statusCode int, message string, userID int) gin.H {
	return gin.H{
		// "ok":         ok,
		// "statusCode": statusCode,
		"message": message,
		"userID":  userID,
	}
}

func CreateResponseSessionsInsert(ok bool, statusCode int, message string, userID int) gin.H {
	return gin.H{
		"ok":         ok,
		"statusCode": statusCode,
		"message":    message,
		"sessionID":  userID,
	}
}
