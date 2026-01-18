package handlers

import (
	"net/http"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

// Versao da aplicação
const AppVersion = "assjur3.3.0"

func VersionHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	rsp := gin.H{
		"version": AppVersion,
	}
	response.HandleSucesso(c, http.StatusOK, rsp, requestID)

}
