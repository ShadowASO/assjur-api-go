/*
---------------------------------------------------------------------------------------
File: eventosHandler.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 09-10-2025
---------------------------------------------------------------------------------------
*/
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ocrserver/internal/models/opensearch"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// Estrutura principal
// ============================================================================

type EventosHandlerType struct {
	service *services.EventosService
}

// Construtor
func NewEventosHandlers(service *services.EventosService) *EventosHandlerType {
	return &EventosHandlerType{
		service: service,
	}
}

// ============================================================================
// BODY REQUESTS
// ============================================================================

type BodyEventosInserir struct {
	IdCtxt     string          `json:"id_ctxt"`
	IdNatu     int             `json:"id_natu"`
	IdEvento   string          `json:"id_evento"`
	Doc        string          `json:"doc"`
	DocJsonRaw json.RawMessage `json:"doc_json_raw"`
}

// ============================================================================
// HANDLERS
// ============================================================================

// Inserir novo evento
func (obj *EventosHandlerType) InsertHandler(c *gin.Context) {
	userName := c.GetString("userName")

	var data BodyEventosInserir
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

	data.IdCtxt = strings.TrimSpace(data.IdCtxt)
	data.IdEvento = strings.TrimSpace(data.IdEvento)
	data.Doc = strings.TrimSpace(data.Doc)

	if data.IdCtxt == "" || data.IdNatu == 0 {
		mslogger.LoggerGlobal.Error("Campos obrigatórios ausentes")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Os campos id_ctxt e id_natu são obrigatórios.",
		)
		return
	}

	docJsonRaw := string(data.DocJsonRaw)

	row, err := obj.service.InserirEvento(
		data.IdCtxt,
		data.IdNatu,
		data.IdEvento,
		data.Doc,
		docJsonRaw,
		userName,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na inclusão do evento: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno no servidor durante inclusão do registro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"row": row,
	}

	msresponse.OK(c, http.StatusCreated, "Evento inserido com sucesso", rsp)
}

// Atualizar evento existente
func (obj *EventosHandlerType) UpdateHandler(c *gin.Context) {
	var requestData opensearch.ResponseEventosRow

	if err := c.ShouldBindJSON(&requestData); err != nil {
		mslogger.LoggerGlobal.Errorf("Dados do request.body inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	requestData.Id = strings.TrimSpace(requestData.Id)

	if requestData.Id == "" {
		mslogger.LoggerGlobal.Error("Campo Id inválido")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campo Id inválido",
			msresponse.ErrorValidacao,
			"O campo id é obrigatório.",
		)
		return
	}

	row, err := obj.service.UpdateEvento(requestData)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na atualização do evento: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno do servidor durante atualização",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"row": row,
	}

	msresponse.OK(c, http.StatusOK, "Evento atualizado com sucesso", rsp)
}

// Deletar evento do índice eventos e embeddings vinculados
func (obj *EventosHandlerType) DeleteHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID ausente")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID ausente",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	if err := obj.service.DeletaEvento(paramID); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao deletar evento: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao deletar evento",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Evento deletado com sucesso")
}

// Selecionar evento pelo ID
func (obj *EventosHandlerType) SelectByIdHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID ausente")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID ausente",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	row, statusCode, err := obj.service.SelectById(paramID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao consultar evento pelo ID: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao consultar evento",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	if statusCode == http.StatusNotFound {
		mslogger.LoggerGlobal.Errorf("Evento não encontrado ID: %s", paramID)

		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Evento não encontrado",
			msresponse.ErrorNaoEncontrado,
			"Não foi localizado evento para o id informado.",
		)
		return
	}

	rsp := gin.H{
		"row": row,
	}

	msresponse.OK(c, http.StatusOK, "Evento localizado com sucesso", rsp)
}

// Listar eventos de um contexto
func (obj *EventosHandlerType) SelectAllHandler(c *gin.Context) {
	ctxtID := strings.TrimSpace(c.Param("id"))

	if ctxtID == "" {
		mslogger.LoggerGlobal.Error("ID do contexto ausente")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do contexto ausente",
			msresponse.ErrorValidacao,
			"O parâmetro id do contexto é obrigatório.",
		)
		return
	}

	rows, err := obj.service.SelectByContexto(ctxtID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar eventos pelo contexto: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao buscar eventos pelo contexto",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"rows": rows,
	}

	msresponse.OK(c, http.StatusOK, "Eventos recuperados com sucesso", rsp)
}
