package handlers

import (
	"net/http"

	"ocrserver/internal/consts"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services/ialib"
	"ocrserver/internal/services/rag/pipeline"

	"ocrserver/internal/services"

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
	ID, OutPut, err := orch.StartPipeline(c.Request.Context(), body.IdCtxt, messages, body.PrevID)
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

func (service *ContextoQueryHandlerType) QueryHandlerTools_ant(c *gin.Context) {
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
	// idCtxt, err := strconv.Atoi(body.IdCtxt)
	// if err != nil {
	// 	logger.Log.Errorf("Erro ao converter idCtxt para int: %v", err)
	// 	response.HandleError(c, http.StatusBadRequest, "Erro ao converter string para int", "", requestID)
	// 	return
	// }
	idCtxt := (body.IdCtxt)

	//Obtém o prompt que irá orientar a análise e elaboração da sentença
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_ANALISE_JULGAMENTO)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar o prompt: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao buscar o prompt", "", requestID)
		return
	}

	var messages ialib.MsgGpt
	messages.AddMessage(ialib.MessageResponseItem{
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
	// resp, usage, err := services.OpenaiServiceGlobal.SubmitResponseFunctionRAG(
	// 	c,
	// 	body.IdCtxt,
	// 	messages,
	// 	toolManage,
	// 	body.PrevID,
	// 	services.REASONING_MEDIUM,
	// 	services.VERBOSITY_MEDIUM)
	resp, err := services.OpenaiServiceGlobal.SubmitPromptTools(
		c,
		body.IdCtxt,
		messages,
		toolManage,
		body.PrevID,
		ialib.REASONING_MEDIUM,
		ialib.VERBOSITY_MEDIUM)
	if err != nil {
		logger.Log.Errorf("Erro ao submeter o prompt e funções à análise do modelo: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao submeter o prompt e funções à análise do modelo", "", requestID)
		return
	}
	if resp != nil {
		usage := resp.Usage
		services.ContextoServiceGlobal.UpdateTokenUso(idCtxt, int(usage.InputTokens), int(usage.OutputTokens))
	}
	params, ok, err := services.OpenaiServiceGlobal.ExtraiResponseTools(body.IdCtxt, resp, services.HandlerToolsFunc)
	if err != nil {
		logger.Log.Errorf("Erro ao extrai as funções: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao extrai as funções", "", requestID)
		return
	}

	if ok {
		resp, err = services.OpenaiServiceGlobal.SubmitResponseTools(
			c,
			resp.ID,
			params,
			ialib.REASONING_MEDIUM,
			ialib.VERBOSITY_MEDIUM,
		)
		if resp != nil {
			usage := resp.Usage
			services.ContextoServiceGlobal.UpdateTokenUso(idCtxt, int(usage.InputTokens), int(usage.OutputTokens))
		} else {
			logger.Log.Error("SubmitResponseTools: nenhuma resposta devolvida!")
		}
	}

	if resp == nil {
		logger.Log.Error("Resposta nula recebida do serviço OpenAI")
		response.HandleError(c, http.StatusInternalServerError, "Resposta nula recebida do serviço", "", requestID)
		return
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
