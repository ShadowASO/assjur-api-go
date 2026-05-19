/*
---------------------------------------------------------------------------------------
File: msresponse.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 06-05-2026
---------------------------------------------------------------------------------------
*/
package msresponse

import (
	"fmt"
	"ocrserver/internal/utils/mslogger"

	"time"

	"github.com/gin-gonic/gin"
)

type ErrorCode int

const (
	ErrorFormatoInvalido ErrorCode = 1
	ErrorTokenInvalido   ErrorCode = 2
	ErrorNaoAutorizado   ErrorCode = 3
	ErrorNaoEncontrado   ErrorCode = 4
	ErrorValidacao       ErrorCode = 5
	ErrorInterno         ErrorCode = 500
)

type APIResponse struct {
	RequestID string     `json:"request_id,omitempty" example:"7f8c2a9b-5f8a-4b0e-91e5-3b9c2d9a1234"`
	OK        bool       `json:"ok" example:"true"`
	Message   string     `json:"message,omitempty" example:"Sucesso"`
	Data      any        `json:"data,omitempty"`
	Error     *ErrorBody `json:"error,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

type ErrorBody struct {
	Code        ErrorCode `json:"code,omitempty" example:"1"`
	Message     string    `json:"message,omitempty" example:"Erro de validação"`
	Description string    `json:"description,omitempty" example:"Campo username é obrigatório"`
}

func LogTime(msg string) {
	now := time.Now().Format("2006-01-02 15:04:05")

	if mslogger.LoggerGlobal != nil {
		mslogger.LoggerGlobal.InfoData("time_marker", mslogger.AppLogData{
			Context: fmt.Sprintf("%s: %s", msg, now),
		})
		return
	}

	fmt.Printf("%s: %s\n", msg, now)
}

func getRequestID(c *gin.Context) string {
	if c == nil {
		return ""
	}

	if v, ok := c.Get("request_id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	if rid := c.Writer.Header().Get("X-Request-Id"); rid != "" {
		return rid
	}

	return ""
}

func respond(c *gin.Context, status int, resp APIResponse) {
	if resp.RequestID == "" {
		resp.RequestID = getRequestID(c)
	}

	if resp.Timestamp.IsZero() {
		resp.Timestamp = time.Now().UTC()
	}

	c.JSON(status, resp)
}

func OK(c *gin.Context, status int, message string, data ...any) {
	resp := APIResponse{
		OK:      true,
		Message: message,
	}

	if len(data) > 0 {
		resp.Data = data[0]
	}

	respond(c, status, resp)
}

func Fail(
	c *gin.Context,
	status int,
	message string,
	code ErrorCode,
	description string,
) {
	rid := getRequestID(c)

	if status >= 500 && mslogger.LoggerGlobal != nil {
		mslogger.LoggerGlobal.ErrorData("response_fail", mslogger.AppLogData{
			RequestID:   rid,
			Status:      status,
			Code:        int(code),
			Context:     message,
			Description: description,
		})
	}

	respond(c, status, APIResponse{
		OK:        false,
		RequestID: rid,
		Message:   message,
		Error: &ErrorBody{
			Code:        code,
			Message:     message,
			Description: description,
		},
	})
}

func Result(
	c *gin.Context,
	status int,
	ok bool,
	message string,
	data any,
	errBody *ErrorBody,
) {
	resp := APIResponse{
		OK:        ok,
		Message:   message,
		Data:      data,
		Error:     errBody,
		Timestamp: time.Now().UTC(),
	}

	if resp.RequestID == "" {
		resp.RequestID = getRequestID(c)
	}

	c.JSON(status, resp)
}
