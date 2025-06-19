package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ocrserver/internal/analise"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"
	"ocrserver/internal/services/rag"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/msgs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ContextoQueryHandlerType struct {
	Model *models.SessionsModelType
}

func NewContextoQueryHandlers(model *models.SessionsModelType) *ContextoQueryHandlerType {
	return &ContextoQueryHandlerType{Model: model}
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
	retSubmit, err := services.OpenaiServiceGlobal.SubmitPromptResponse(*rspMsgs, nil)
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
		"created": retSubmit.CreatedAt,
		"model":   retSubmit.Model,
		"output":  retSubmit.Output,
		"usage":   retSubmit.Usage,
	}

	c.JSON(http.StatusOK, response)

}

type BodyParamsQuery struct {
	IdCtxt    string `json:"id_ctxt"`
	TxtPrompt string `json:"txt_prompt"`
	PrevID    string `json:"prev_id"`
}

func (service *ContextoQueryHandlerType) QueryHandlertTools(c *gin.Context) {
	requestID := uuid.New().String()

	var body BodyParamsQuery

	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Error("Formato inválido", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao realizar bind do body", "", requestID)
		return
	}

	if body.IdCtxt == "" {
		logger.Log.Error("O ID do contexto é obrigatório")
		response.HandleError(c, http.StatusBadRequest, "O ID do contexto é obrigatório", "", requestID)
		return
	}
	if body.TxtPrompt == "" {
		logger.Log.Error("O prompt é obrigatório")
		response.HandleError(c, http.StatusBadRequest, "O prompt é obrigatório", "", requestID)
		return
	}
	//--------------------------------------------------
	//Adiciono a informação do número do contexto ao prompt
	prompt := fmt.Sprintf("%s (O contexto é igual a %v )", body.TxtPrompt, body.IdCtxt)
	//Obtenho o toolManager diretamente do pacote "rag", concentrando nele a lógica do negócio
	toolManage := rag.GetRegisterToolAutos()
	//-------------------------------------------------

	//Faço a chamada ao serviço que executa o submit utilizando tools
	resp, err := services.OpenaiServiceGlobal.SubmitResponseFunctionRAG(prompt, toolManage, body.PrevID)
	if err != nil {
		logger.Log.Error("Erro ao realizar busca pelo contexto", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao realizar busca pelo contexto", "", requestID)
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
