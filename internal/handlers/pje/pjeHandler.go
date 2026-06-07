package pje

import (
	"net/http"

	pjeservices "ocrserver/internal/services/pje"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type PjeHandler struct {
	srv *pjeservices.PjeService
}

func NewPjeHandler(service *pjeservices.PjeService) *PjeHandler {
	return &PjeHandler{
		srv: service,
	}
}

type BodyPjeListaRequest struct {
	Numero_processo string   `json:"numero_processo"`
	Usuario_cpf     string   `json:"usuario_cpf"`
	Usuario_senha   string   `json:"usuario_senha"`
	Documentos      []string `json:"documentos"`
	Formato         string   `json:"formato"`
}

func (obj *PjeHandler) ListaDocumentos(c *gin.Context) {
	var data BodyPjeListaRequest

	if err := c.ShouldBindJSON(&data); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao decodificar JSON: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Dados inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if data.Numero_processo == "" || data.Usuario_cpf == "" || data.Usuario_senha == "" {
		mslogger.LoggerGlobal.Errorf("Campos obrigatórios ausentes!: %s", data.Numero_processo)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Os campos id_ctxt e id_natu são obrigatórios.",
		)
		return
	}

	rsp, err := obj.srv.ListaDocumentos(c, data.Numero_processo, data.Usuario_cpf, data.Usuario_senha, data.Documentos, data.Formato)

	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na inclusão do registro %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno no servidor durante inclusão do registro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusCreated, "Consulta realizada com sucesso", rsp)
}
