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

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// Estrutura principal
// ============================================================================

type EventosHandlerType struct {
	service *services.EventosService
	//idx     *opensearch.EventosIndex
}

// // Estrutura base para o JSON (mantida como referência genérica)
// type DocumentoBase struct {
// 	Tipo *struct {
// 		Key         int    `json:"key"`
// 		Description string `json:"description"`
// 	} `json:"tipo"`
// 	Processo string `json:"processo"`
// 	IdEvento string `json:"id_evento"`
// }

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
	requestID := middleware.GetRequestID(c)

	var data BodyEventosInserir
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Dados inválidos", "", requestID)
		return
	}

	if data.IdCtxt == "" || data.IdNatu == 0 {
		logger.Log.Error("Campos obrigatórios ausentes!")
		response.HandleError(c, http.StatusBadRequest, "Campos obrigatórios ausentes!", "", requestID)
		return
	}

	docJsonRaw := string(data.DocJsonRaw)

	row, err := obj.service.InserirEvento(data.IdCtxt, data.IdNatu, data.IdEvento, data.Doc, docJsonRaw, userName)
	if err != nil {
		logger.Log.Errorf("Erro na inclusão do evento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no servidor durante inclusão do registro", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Evento inserido com sucesso!",
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

// Atualizar evento existente
func (obj *EventosHandlerType) UpdateHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var requestData opensearch.ResponseEventosRow
	if err := c.ShouldBindJSON(&requestData); err != nil {
		logger.Log.Errorf("Dados do request.body inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	if requestData.Id == "" {
		logger.Log.Error("Campo Id inválido")
		response.HandleError(c, http.StatusBadRequest, "Campo Id inválido", "", requestID)
		return
	}

	row, err := obj.service.UpdateEvento(requestData)
	if err != nil {
		logger.Log.Errorf("Erro na atualização do evento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno do servidor durante atualização", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Evento atualizado com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

// Deletar evento (índice 'eventos' e embeddings vinculados)
func (obj *EventosHandlerType) DeleteHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID ausente")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}

	err := obj.service.DeletaEvento(paramID)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar evento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao deletar evento", "", requestID)
		return
	}

	rsp := gin.H{
		"ok":      true,
		"message": "Evento deletado com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

// Selecionar evento pelo ID
func (obj *EventosHandlerType) SelectByIdHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID ausente na requisição")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}

	row, err := obj.service.SelectById(paramID)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar evento por ID: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao consultar evento por ID", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Evento localizado com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

// Listar eventos de um contexto (GET /contexto/eventos/:id)
func (obj *EventosHandlerType) SelectAllHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	ctxtID := c.Param("id")
	if ctxtID == "" {
		logger.Log.Error("ID do contexto ausente")
		response.HandleError(c, http.StatusBadRequest, "ID do contexto ausente", "", requestID)
		return
	}

	// idKey, err := strconv.Atoi(ctxtID)
	// if err != nil {
	// 	logger.Log.Errorf("ID de contexto inválido: %v", err)
	// 	response.HandleError(c, http.StatusBadRequest, "ID de contexto inválido", "", requestID)
	// 	return
	// }
	idKey := ctxtID

	rows, err := obj.service.SelectByContexto(idKey)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar eventos pelo contexto: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao buscar eventos pelo contexto", "", requestID)
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Eventos recuperados com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
