package handlers

import (
	"encoding/json"

	"net/http"

	"github.com/gin-gonic/gin"

	"ocrserver/internal/analise"
	"ocrserver/internal/services/openAI"
	"ocrserver/internal/utils/msgs"
	"ocrserver/models"
)

type ContextoQueryHandlerType struct {
	sessionModel *models.SessionsModelType
}

func NewContextoQueryHandlers() *ContextoQueryHandlerType {
	model := models.NewSessionsModel()
	return &ContextoQueryHandlerType{sessionModel: model}
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
/*
  - Insere um novo prompt na tabela 'prompts'
    *Rota: "/contexto/query"
  - Método: POST
  - Body: {
    "IdCtxt"   int    			//ID do contexto
	"Prompt"   openAI.MsgGpt	//Objeto com mensagens
	"ModeloId" string			//ID do modelo no OpenSearch
	"Tipo"     int				//Tipo de análise a ser realizada
    }
*/

func (service *ContextoQueryHandlerType) QueryHandler(c *gin.Context) {

	bodyParams := analise.BodyRequestContextoQuery{}

	decoder := json.NewDecoder(c.Request.Body)

	if err := decoder.Decode(&bodyParams); err != nil {
		response := msgs.CreateResponseMessage("Erro no SubmitPrompt!" + err.Error())
		c.JSON(http.StatusNoContent, response)
		return
	}

	if bodyParams.IdCtxt == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O IdCtxt é obrigatório"})
		return
	}
	if bodyParams.Tipo == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O Tipo  é obrigatório"})
		return
	}

	if len(bodyParams.Prompt.Messages) == 0 {
		response := msgs.CreateResponseMessage("Mensagens não podem ser vazias!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	/**
	***************************************************************************
	Constroi o vetor de mensagens para ser enviado à API da OpenAI
	*/
	rspMsgs, ctxtErr := analise.BuildAnaliseContexto(bodyParams)
	if ctxtErr != nil {
		response := msgs.CreateResponseMessage("Erro no BuildAnaliseContexto!" + ctxtErr.Error())
		c.JSON(http.StatusNoContent, response)
		return
	}
	/**
	***************************************************************************
	 */

	//Submete o vetor de mensagens à API da OpenAI
	retSubmit, err := openAI.Service.SubmitPrompt(*rspMsgs)
	if err != nil {
		response := msgs.CreateResponseMessage("Erro no SubmitPrompt!" + err.Error())
		c.JSON(http.StatusNoContent, response)
		return
	}

	// Crie uma estrutura de resposta que inclua os dados do ChatCompletion
	response := gin.H{
		"message": "Sucesso!",
		"id":      retSubmit.ID,
		"object":  retSubmit.Object,
		"created": retSubmit.Created,
		"model":   retSubmit.Model,
		"choices": retSubmit.Choices,
		"usage":   retSubmit.Usage,
	}

	c.JSON(http.StatusOK, response)

}
