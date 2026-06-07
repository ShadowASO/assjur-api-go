package handlers

import (
	"net/http"
	"time"

	"ocrserver/internal/models"

	"ocrserver/internal/services/openai"
	"ocrserver/internal/services/rag/pipeline"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type ContextoQueryHandlerType struct {
	Model *models.SessionsModelType
}

func NewContextoQueryHandlers(model *models.SessionsModelType) *ContextoQueryHandlerType {
	return &ContextoQueryHandlerType{Model: model}
}

type BodyParamsQuery struct {
	IdCtxt   string                       `json:"id_ctxt"`
	Messages []openai.MessageResponseItem `json:"messages"`
	PrevID   string                       `json:"prev_id"`
}

func (service *ContextoQueryHandlerType) QueryHandlerPipeline(c *gin.Context) {
	userName := c.GetString("userName")

	start := time.Now()
	defer func() {
		mslogger.LoggerGlobal.Infof("Pipeline de análise concluída: %v", time.Since(start))
	}()

	var body BodyParamsQuery
	if err := c.ShouldBindJSON(&body); err != nil {
		mslogger.LoggerGlobal.Errorf("Parâmetros inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Parâmetros do body inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if body.IdCtxt == "" {
		mslogger.LoggerGlobal.Error("O ID do contexto é obrigatório")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"O ID do contexto é obrigatório",
			msresponse.ErrorValidacao,
			"O campo id_ctxt deve ser informado.",
		)
		return
	}

	if len(body.Messages) == 0 {
		mslogger.LoggerGlobal.Error("A lista de mensagens está vazia")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"A lista de mensagens está vazia",
			msresponse.ErrorValidacao,
			"O campo messages deve conter ao menos uma mensagem.",
		)
		return
	}

	var messages openai.MsgGpt
	for _, msg := range body.Messages {
		messages.AddMessage(msg)
	}

	orch := pipeline.NewOrquestradorType()

	res, err := orch.StartPipelineResult(
		c.Request.Context(),
		body.IdCtxt,
		messages,
		body.PrevID,
		userName,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro durante o pipeline RAG: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro durante o pipeline RAG",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	mslogger.LoggerGlobal.Infof("Response ID: %s", res.ID)

	data := gin.H{
		"message":     res.Message,
		"status":      res.Status.String(),
		"ok":          res.Status == pipeline.StatusOK,
		"blocked":     res.Status == pipeline.StatusBlocked,
		"invalid":     res.Status == pipeline.StatusInvalid,
		"id":          res.ID,
		"output":      res.Output,
		"response_id": res.ID,
		"eventCode":   res.EventCode,
		"eventDesc":   res.EventDesc,
	}

	switch res.Status {
	case pipeline.StatusOK:
		msresponse.OK(
			c,
			http.StatusOK,
			res.Message,
			data,
		)
		return

	case pipeline.StatusBlocked:
		msresponse.Result(
			c,
			http.StatusOK,
			false,
			"Aguardando ação do usuário",
			data,
			&msresponse.ErrorBody{
				Code:        msresponse.ErrorValidacao,
				Message:     "Aguardando ação do usuário",
				Description: "O pipeline foi bloqueado porque depende de confirmação ou complementação do usuário.",
			},
		)
		return

	case pipeline.StatusInvalid:
		msresponse.Result(
			c,
			http.StatusUnprocessableEntity,
			false,
			"Pré-condição não atendida",
			data,
			&msresponse.ErrorBody{
				Code:        msresponse.ErrorValidacao,
				Message:     "Pré-condição não atendida",
				Description: "O pipeline identificou que os dados fornecidos ainda não permitem concluir a operação.",
			},
		)
		return

	default:
		mslogger.LoggerGlobal.Errorf("Status de pipeline desconhecido: %v", res.Status)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Status de pipeline desconhecido",
			msresponse.ErrorInterno,
			"Status retornado pelo pipeline não foi reconhecido pelo handler.",
		)
		return
	}
}
