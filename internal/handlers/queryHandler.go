package handlers

import (
	"net/http"

	"ocrserver/internal/config"
	"ocrserver/internal/models"
	"ocrserver/internal/services"
	openaiservice "ocrserver/internal/services/openai"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type QueryHandlerType struct {
	sessionModel *models.SessionsModelType
	service      *services.QueryServiceType
}

func NewQueryHandlers(service *services.QueryServiceType) *QueryHandlerType {
	modelo, err := service.GetModel()
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao obter usersModel", err)
		return nil
	}

	return &QueryHandlerType{
		sessionModel: modelo,
		service:      service,
	}
}

/*
 * Faz uma consulta diretamente na API da OpenAI
 *
 * - **Rota**: "/query"
 * - **Método**: POST
 * - **Status**: 200/400/500
 * - **Body:
 *		{
 *			"messages": [
 * 				{
 *     				"role": string,
 *     				"content": string
 *   			}
 * 			]
 * 		}
 */

// QueryHandler processa uma requisição POST para consulta na API OpenAI.
// Rota: "/query"
// Body esperado:
//
//	{
//	  "messages": [
//	    { "role": "string", "content": "string" }
//	  ]
//	}
//
// Retorna JSON com dados da resposta da OpenAI e status HTTP 200.
func (h *QueryHandlerType) QueryHandler(c *gin.Context) {
	var messages openaiservice.MsgGpt

	if err := c.ShouldBindJSON(&messages); err != nil {
		mslogger.LoggerGlobal.Errorf("Dados em body incorretos: %s", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Dados em body incorretos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if len(messages.Messages) == 0 {
		mslogger.LoggerGlobal.Error("Mensagens não podem ser vazias")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Mensagens não podem ser vazias",
			msresponse.ErrorValidacao,
			"O campo messages deve conter ao menos uma mensagem.",
		)
		return
	}

	msg := messages.GetMessages()
	if len(msg) == 0 {
		mslogger.LoggerGlobal.Error("Nenhuma mensagem válida encontrada")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Nenhuma mensagem válida encontrada",
			msresponse.ErrorValidacao,
			"A lista de mensagens não contém itens válidos para envio ao modelo.",
		)
		return
	}

	nrTokens, err := services.OpenaiServiceGlobal.TokensCounter(messages)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao contar tokens do prompt: %v", err)
	} else {
		mslogger.LoggerGlobal.Infof("Total de tokens no prompt: %d", nrTokens)
	}

	retSubmit, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		c.Request.Context(),
		messages,
		msg[0].Id,
		config.GlobalConfig.OpenOptionModel,
		openaiservice.REASONING_LOW,
		openaiservice.VERBOSITY_LOW,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro no SubmitPrompt: %s", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro no SubmitPrompt",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"id":      retSubmit.ID,
		"object":  retSubmit.Object,
		"created": retSubmit.CreatedAt,
		"model":   retSubmit.Model,
		"output":  retSubmit.Output,
		"usage":   retSubmit.Usage,
	}

	msresponse.OK(c, http.StatusOK, "Sucesso", rsp)
}
