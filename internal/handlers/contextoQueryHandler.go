package handlers

import (
	"net/http"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services/ialib"
	"ocrserver/internal/services/rag/pipeline"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

type ContextoQueryHandlerType struct {
	Model *models.SessionsModelType
}

func NewContextoQueryHandlers(model *models.SessionsModelType) *ContextoQueryHandlerType {
	return &ContextoQueryHandlerType{Model: model}
}

type BodyParamsQuery struct {
	IdCtxt   string                      `json:"id_ctxt"`
	Messages []ialib.MessageResponseItem `json:"messages"`
	PrevID   string                      `json:"prev_id"`
}

func (service *ContextoQueryHandlerType) QueryHandlerPipeline(c *gin.Context) {
	userName := c.GetString("userName")
	requestID := middleware.GetRequestID(c)

	var body BodyParamsQuery
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Errorf("Parâmetros inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	if body.IdCtxt == "" {
		logger.Log.Error("O ID do contexto é obrigatório")
		response.HandleError(c, http.StatusBadRequest, "O ID do contexto é obrigatório", "", requestID)
		return
	}

	if len(body.Messages) == 0 {
		logger.Log.Error("A lista de mensagens está vazia")
		response.HandleError(c, http.StatusBadRequest, "A lista de mensagens está vazia", "", requestID)
		return
	}

	var messages ialib.MsgGpt
	for _, msg := range body.Messages {
		messages.AddMessage(msg)
	}

	orch := pipeline.NewOrquestradorType()

	// ✅ novo método
	res, err := orch.StartPipelineResult(c.Request.Context(), body.IdCtxt, messages, body.PrevID, userName)
	if err != nil {
		logger.Log.Errorf("Erro durante o pipeline RAG: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro durante o pipeline RAG", err.Error(), requestID)
		return
	}

	// Data rica, sempre igual (front não sofre)
	data := gin.H{
		"message":   res.Message,
		"status":    res.Status.String(),
		"ok":        res.Status == pipeline.StatusOK,
		"blocked":   res.Status == pipeline.StatusBlocked,
		"invalid":   res.Status == pipeline.StatusInvalid,
		"id":        res.ID,
		"output":    res.Output,
		"eventCode": res.EventCode,
		"eventDesc": res.EventDesc,
	}

	// Map status -> HTTP + Ok + ErrorDetail (quando não OK)
	switch res.Status {

	case pipeline.StatusOK:
		response.HandleSucesso(c, http.StatusOK, data, requestID)
		return

	case pipeline.StatusBlocked:
		// Fluxo normal: cliente precisa confirmar/complementar.
		// HTTP 200 e Ok=false (não concluiu), com "error" sem description técnica.
		response.HandleResult(
			c,
			http.StatusOK,
			false,
			data,
			&response.ErrorDetail{
				Code:    http.StatusOK,
				Message: "Aguardando ação do usuário",
			},
			requestID,
		)
		return

	case pipeline.StatusInvalid:
		// Pré-condição/regra não atendida: 422 é bem apropriado.
		response.HandleResult(
			c,
			http.StatusUnprocessableEntity,
			false,
			data,
			&response.ErrorDetail{
				Code:    http.StatusUnprocessableEntity,
				Message: "Pré-condição não atendida",
			},
			requestID,
		)
		return

	default:
		// defensivo
		logger.Log.Errorf("Status de pipeline desconhecido: %v", res.Status)
		response.HandleError(c, http.StatusInternalServerError, "Status de pipeline desconhecido", "", requestID)
		return
	}
}
