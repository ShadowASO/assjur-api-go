package handlers

import (
	"net/http"

	"ocrserver/internal/consts"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"

	"strconv"

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
	IdCtxt   string                         `json:"id_ctxt"`
	Messages []services.MessageResponseItem `json:"messages"`
	PrevID   string                         `json:"prev_id"`
}

func (service *ContextoQueryHandlerType) QueryHandlerTools(c *gin.Context) {
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

	// Converte IdCtxt para int
	idCtxt, err := strconv.Atoi(body.IdCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao converter idCtxt para int: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao converter string para int", "", requestID)
		return
	}

	//Obtém o prompt que irá orientar a análise e elaboração da sentença
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_ANALISE_JULGAMENTO)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar o prompt: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao buscar o prompt", "", requestID)
		return
	}

	var messages services.MsgGpt
	messages.AddMessage(services.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: prompt,
	})

	//Transfiro todas as mensagens recebidas do cliente para a variável "messages" para que o modelo tenha o
	//contexto da conversa
	for _, msg := range body.Messages {
		messages.AddMessage(msg)
	}

	//Obtenho o toolManager com todas as funções que podem ser usadas pelo modelo, concentrando nele a lógica do negócio
	toolManage := services.GetRegisterToolAutos()

	//Faz a chamda ao modelo para executar o prompt
	resp, usage, err := services.OpenaiServiceGlobal.SubmitResponseFunctionRAG(
		c,
		body.IdCtxt,
		messages,
		toolManage,
		body.PrevID,
		services.REASONING_MEDIUM,
		services.VERBOSITY_MEDIUM)
	if err != nil {
		logger.Log.Errorf("Erro ao submeter o prompt e funções à análise do modelo: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao submeter o prompt e funções à análise do modelo", "", requestID)
		return
	}

	//*** Atualizo o uso de tokens para o contexto, na tabela "contexto"
	if usage != nil {
		services.ContextoServiceGlobal.UpdateTokenUso(idCtxt, int(usage.InputTokens), int(usage.OutputTokens))
	}

	rsp := gin.H{
		"message": "Sucesso!",
		"id":      resp.ID,
		"object":  resp.Object,
		"created": resp.CreatedAt,
		"model":   resp.Model,
		"output":  resp.Output,
		"usage":   resp.Usage,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}
