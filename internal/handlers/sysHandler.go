package handlers

import (
	"net/http"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

// Versao da aplicação
const AppVersion = "assjur3.0.0"

func VersionHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	rsp := gin.H{
		"version": AppVersion,
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}
