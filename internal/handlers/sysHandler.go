package handlers

import (
	"net/http"

	"ocrserver/internal/config"
	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

// Versao da aplicação
const AppVersion = "4.1.11"

type ModeloOpenai struct {
	OpenModelTop       string `json:"open_model_top"`       //Modelo principal 'gpt-5.4-mini'
	OpenModel          string `json:"open_model"`           //Modelo principal 'gpt-5-mini'
	OpenModelSecundary string `json:"open_model_secundary"` //Modelo secundário 'gpt-5-nano'
}

func VersionHandler(c *gin.Context) {
	rsp := gin.H{
		"version": AppVersion,
	}

	msresponse.OK(c, http.StatusOK, "Versão da aplicação retornada com sucesso", rsp)
}

/*
OpenOptionModelTop            string //Modelo principal 'gpt-5.4-mini'

	OpenOptionModel               string //Modelo principal 'gpt-5-mini'
	OpenOptionModelSecundary      string //Modelo secundário 'gpt-5-nano'
*/
func ModelosHandler(c *gin.Context) {
	cfg := config.GlobalConfig
	if cfg == nil {

		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Configuração inválida",
			msresponse.ErrorInterno,
			"Configuração inválida",
		)
		return

	}

	rsp := ModeloOpenai{
		OpenModelTop:       cfg.OpenOptionModelTop,
		OpenModel:          cfg.OpenOptionModel,
		OpenModelSecundary: cfg.OpenOptionModelSecundary,
	}

	msresponse.OK(c, http.StatusOK, "Versão da aplicação retornada com sucesso", rsp)
}
