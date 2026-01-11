/*
---------------------------------------------------------------------------------------
File: response.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
---------------------------------------------------------------------------------------
*/
package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorDetail struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}

// StandardResponse represents our unified API response structure
type StandardResponse struct {
	Ok        bool         `json:"ok"`
	Data      interface{}  `json:"data,omitempty"`
	Error     *ErrorDetail `json:"error,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	RequestID string       `json:"request_id"`
}

func NewSuccess(data interface{}, requestID string) StandardResponse {
	return StandardResponse{
		Ok:        true,
		Data:      data,
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	}
}

// NewError creates an error response
func NewError(statusCode int, message, description, requestID string) StandardResponse {
	return StandardResponse{
		Ok: false,
		Error: &ErrorDetail{
			Code:        statusCode,
			Message:     message,
			Description: description,
		},
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	}
}

func HandleError(c *gin.Context, statusCode int, message, description, requestID string) {

	c.JSON(statusCode, StandardResponse{
		Ok: false,
		Error: &ErrorDetail{
			Code:        statusCode,
			Message:     message,
			Description: description,
		},
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	})
}
func HandleSucesso(c *gin.Context, statusCode int, data interface{}, requestID string) {
	c.JSON(statusCode, StandardResponse{
		Ok:        true,
		Data:      data,
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	})
}

// HandleResult devolve Ok=true/false e pode incluir Data mesmo quando Ok=false.
func HandleResult(c *gin.Context, statusCode int, ok bool, data interface{}, errDetail *ErrorDetail, requestID string) {
	c.JSON(statusCode, StandardResponse{
		Ok:        ok,
		Data:      data,
		Error:     errDetail,
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	})
}
