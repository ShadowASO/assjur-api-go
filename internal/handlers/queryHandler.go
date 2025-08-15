package handlers

import (
	"net/http"
	"ocrserver/internal/config"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"

	"ocrserver/internal/services"
	"ocrserver/internal/services/openapi"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

type QueryHandlerType struct {
	sessionModel *models.SessionsModelType
	service      *services.QueryServiceType
}

func NewQueryHandlers(service *services.QueryServiceType) *QueryHandlerType {

	modelo, err := service.GetModel()
	if err != nil {
		logger.Log.Error("Erro ao ao obter usersModel", err.Error())
		return nil
	}
	return &QueryHandlerType{
		sessionModel: modelo,
		service:      service,
	}
}

/*
 * Faz uma consulta diretametne na api da openai
 *
 * - **Rota**: "/query"
 * - **Params**:
 * - **Método**: POST
 * - **Status**: 201/204/400
 * - **Body:
 *		{
 *			"messages": [
 * 				{
 *     				"role": string,
 *     				"content": string
 *   			}
 * 			]
 * 		}
 * - **Resposta**:
 *		{
 * 		"choices": [
 *   	{
 *     		"finish_reason": string,
 *     		"index": int,
 *     		"logprobs": {
 *       	"content": null,
 *       	"refusal": null
 *     	},
 *     		"message": {
 *       		"role": string,
 *       		"content": string
 *     		}
 *   	} ],
 * 		"created": int,
 * 		"id": string,
 * 		"message": string,
 * 		"model": string,
 * 		"object": string,
 * 		"usage": {
 *   		"completion_tokens": int64,
 *   		"prompt_tokens": int64,
 *   		"total_tokens": int64,
 *  		 "completion_tokens_details": {
 *     			"accepted_prediction_tokens": int,
 *     			"audio_tokens": int,
 *     			"reasoning_tokens": int,
 *     			"rejected_prediction_tokens": int
 *   		},
 *   	"prompt_tokens_details": {
 *     		"audio_tokens": int64,
 *     		"cached_tokens": int64
 *   }
 * }
*}
*/

// QueryHandler processa uma requisição POST para consulta na API OpenAI
// Rota: "/query"
// Body esperado:
//
//	{
//	  "messages": [
//	    { "role": "string", "content": "string" }
//	  ]
//	}
//
// Retorna JSON com dados do chat completion e status HTTP 200 (OK)
func (h *QueryHandlerType) QueryHandler(c *gin.Context) {
	//Generate request ID for tracing
	reqID, exists := c.Get(middleware.ContextKeyRequestID)
	if !exists {
		reqID = "unknown"
	}
	requestID := reqID.(string)
	var messages openapi.MsgGpt
	//--------------------------------------

	// Extrai os dados do corpo da requisição
	if err := c.ShouldBindJSON(&messages); err != nil {

		logger.Log.Errorf("Dados em body incorretos: %s", err)
		response.HandleError(c, http.StatusBadRequest, "Dados em body incorretos!", "", requestID)
		return
	}

	if len(messages.Messages) == 0 {

		logger.Log.Error("Mensagens não podem ser vazias!")
		response.HandleError(c, http.StatusBadRequest, "Mensagens não podem ser vazias!", "", requestID)
		return
	}
	msg := messages.GetMessages()
	//***********
	nrTokens, _ := services.OpenaiServiceGlobal.TokensCounter(messages)
	logger.Log.Infof("Total de tokens no prompt: %d", nrTokens)
	//**********

	retSubmit, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		c.Request.Context(),
		messages, msg[0].Id,
		config.GlobalConfig.OpenOptionModel,
		openapi.REASONING_LOW,
		openapi.VERBOSITY_LOW)
	if err != nil {

		logger.Log.Errorf("Erro no SubmitPrompt: %s", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro no SubmitPrompt!", "", requestID)
		return
	}
	usage := retSubmit.Usage

	// Crie uma estrutura de resposta que inclua os dados do ChatCompletion
	rsp := gin.H{
		"message": "Sucesso!",
		"id":      retSubmit.ID,
		"object":  retSubmit.Object,
		"created": retSubmit.CreatedAt,
		"model":   retSubmit.Model,
		"output":  retSubmit.Output,
		"usage":   usage,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}
