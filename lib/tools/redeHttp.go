package tools

import "github.com/gin-gonic/gin"

// Função auxiliar para padronizar respostas

func CreateResponseMessage(message string) gin.H {
	return gin.H{
		"message": message,
	}
}

func CreateResponse(ok bool, statusCode int, message string, rows interface{}) gin.H {
	return gin.H{
		"ok":         ok,
		"statusCode": statusCode,
		"message":    message,
		"rows":       rows,
	}
}

func CreateResponseSelectRows(ok bool, statusCode int, message string, rows interface{}) gin.H {
	return gin.H{
		"ok":         ok,
		"statusCode": statusCode,
		"message":    message,
		"rows":       rows,
	}
}
func CreateResponseSelectSingle(ok bool, statusCode int, message string, row interface{}) gin.H {
	return gin.H{
		"ok":         ok,
		"statusCode": statusCode,
		"message":    message,
		"data":       row,
	}
}

func CreateResponseUserInsert(ok bool, statusCode int, message string, userID int) gin.H {
	return gin.H{
		"ok":         ok,
		"statusCode": statusCode,
		"message":    message,
		"userID":     userID,
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
