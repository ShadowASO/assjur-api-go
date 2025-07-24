package handlers

import (
	"fmt"
	"net/http"
	"ocrserver/internal/analise"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"
	"ocrserver/internal/services/rag"
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
	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	bodyParams := analise.BodyRequestContextoQuery{}

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Parâmetros inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	if bodyParams.IdCtxt == 0 {

		logger.Log.Error("O campo IdCtxt é obrigatório")
		response.HandleError(c, http.StatusBadRequest, "O campo IdCtxt é obrigatório", "", requestID)
		return
	}
	if bodyParams.Tipo == 0 {

		logger.Log.Error("O campo Tipo  é obrigatório")
		response.HandleError(c, http.StatusBadRequest, "O campo Tipo  é obrigatório", "", requestID)
		return
	}

	if len(bodyParams.Prompt.Messages) == 0 {

		logger.Log.Error("Mensagens não podem ser vazias!")
		response.HandleError(c, http.StatusBadRequest, "Mensagens não podem ser vazias!", "", requestID)
		return
	}

	/**
	***************************************************************************
	Constroi o vetor de mensagens para ser enviado à API da OpenAI
	*/
	rspMsgs, err := analise.BuildAnaliseContexto(bodyParams)
	if err != nil {
		logger.Log.Errorf("Erro interno no BuildAnaliseContexto: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no BuildAnaliseContexto!", "", requestID)
		return
	}
	/**
	***************************************************************************
	 */

	//Submete o vetor de mensagens à API da OpenAI
	retSubmit, usage, err := services.OpenaiServiceGlobal.SubmitPromptResponse(c, *rspMsgs, nil, "")
	if err != nil {

		logger.Log.Errorf("Erro interno ao realizar o submit: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno ao realizar o submit", "", requestID)
		return
	}

	//*** Atualizo o uso de tokens para o contexto

	idCtxt := bodyParams.IdCtxt
	services.ContextoServiceGlobal.UpdateTokenUso(idCtxt, int(usage.InputTokens), int(usage.OutputTokens))
	//************************************************

	// Crie uma estrutura de resposta que inclua os dados do ChatCompletion
	rsp := gin.H{
		"message": "Sucesso!",
		"id":      retSubmit.ID,
		"object":  retSubmit.Object,
		"created": retSubmit.CreatedAt,
		"model":   retSubmit.Model,
		"output":  retSubmit.Output,
		"usage":   retSubmit.Usage,
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

type BodyParamsQuery struct {
	IdCtxt    string `json:"id_ctxt"`
	TxtPrompt string `json:"txt_prompt"`
	PrevID    string `json:"prev_id"`
}

func (service *ContextoQueryHandlerType) QueryHandlerTools(c *gin.Context) {

	//Generate request ID for tracing
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
	resp, usage, err := services.OpenaiServiceGlobal.SubmitResponseFunctionRAG(c, prompt, toolManage, body.PrevID)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar busca pelo contexto: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao realizar busca pelo contexto", "", requestID)
		return
	}

	//*** Atualizo o uso de tokens para o contexto

	idCtxt, err := strconv.Atoi(body.IdCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao converte idCtxt de string para int: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao converter string para int", "", requestID)
		return
	}
	services.ContextoServiceGlobal.UpdateTokenUso(idCtxt, int(usage.InputTokens), int(usage.OutputTokens))

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
