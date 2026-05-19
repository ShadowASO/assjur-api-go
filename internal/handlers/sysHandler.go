package handlers

import (
	"net/http"

	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

// Versao da aplicação
const AppVersion = "4.1.2"

func VersionHandler(c *gin.Context) {
	rsp := gin.H{
		"version": AppVersion,
	}

	msresponse.OK(c, http.StatusOK, "Versão da aplicação retornada com sucesso", rsp)
}
