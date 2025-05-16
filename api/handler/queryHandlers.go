package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ocrserver/api/handler/response"
	"ocrserver/internal/services/openAI"
	"ocrserver/internal/utils/msgs"
	"ocrserver/models"
)

type QueryHandlerType struct {
	sessionModel *models.SessionsModelType
}

func NewQueryHandlers() *QueryHandlerType {
	model := models.NewSessionsModel()
	return &QueryHandlerType{sessionModel: model}
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

func (service *QueryHandlerType) QueryHandler(c *gin.Context) {
	//Generate request ID for tracing
	requestID := uuid.New().String()

	var messages openAI.MsgGpt

	// Extrai os dados do corpo da requisição
	if err := c.ShouldBindJSON(&messages); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Invalid request body"})
		// return
		response := msgs.CreateResponseMessage("Dados em body incorretos!" + err.Error())
		c.JSON(http.StatusNoContent, response)
		return
	}

	if len(messages.Messages) == 0 {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Messages array cannot be empty"})
		// return
		response := msgs.CreateResponseMessage("Mensagens não podem ser vazias!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	retSubmit, err := openAI.Service.SubmitPrompt(messages)
	if err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Erro no SubmitPrompt"})
		// return
		response := msgs.CreateResponseMessage("Erro no SubmitPrompt!" + err.Error())
		c.JSON(http.StatusNoContent, response)
		return
	}

	// Crie uma estrutura de resposta que inclua os dados do ChatCompletion
	rsp := gin.H{
		"message": "Sucesso!",
		"id":      retSubmit.ID,
		"object":  retSubmit.Object,
		"created": retSubmit.Created,
		"model":   retSubmit.Model,
		"choices": retSubmit.Choices,
		"usage":   retSubmit.Usage,
	}

	//c.JSON(http.StatusOK, response)
	c.JSON(http.StatusCreated, response.NewSuccess(rsp, requestID))

}
