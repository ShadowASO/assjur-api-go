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

func (service *ContextoQueryHandlerType) QueryHandlerTools(c *gin.Context) {
	userName := c.GetString("userName")
	requestID := middleware.GetRequestID(c)
	//--------------------------------------
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
		logger.Log.Error("A lista de mensagens está vavia")
		response.HandleError(c, http.StatusBadRequest, "A lista de mensagens está vazia", "", requestID)
		return
	}

	//Crio um novo objeto de mensagens recebidas do cliente para a variável "messages"
	var messages ialib.MsgGpt
	for _, msg := range body.Messages {
		messages.AddMessage(msg)
		//logger.Log.Infof("Mensagens: %s", msg.Text)
	}

	orch := pipeline.NewOrquestradorType()
	ID, OutPut, err := orch.StartPipeline(c.Request.Context(), body.IdCtxt, messages, body.PrevID, userName)
	if err != nil {
		logger.Log.Errorf("Erro durante o pipeline RAG: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro durante o pipeline RAG: ", err.Error(), requestID)
		return
	}

	rsp := gin.H{
		"message": "Sucesso!",
		"id":      ID,
		"output":  OutPut,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}
