package pipeline

import (
	"context"

	"ocrserver/internal/config"
	"ocrserver/internal/consts"

	"ocrserver/internal/services"
	"ocrserver/internal/services/ialib"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/openai/openai-go/v2/responses"
)

type GeneratorType struct {
}

func NewGeneratorType() *GeneratorType {
	return &GeneratorType{}
}

func (service *GeneratorType) ExecutaAnaliseProcesso(ctx context.Context, idCtxt int, msgs ialib.MsgGpt, prevID string, autos []consts.ResponseAutosRow) (string, []responses.ResponseOutputItemUnion, error) {

	//Obtém o prompt que irá orientar a análise e elaboração da sentença
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_ANALISE)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar o prompt: %v", err)
		return "", nil, erros.CreateError("Erro ao buscar PROMPT_ANALISE_JULGAMENTO", err.Error())
	}
	//logger.Log.Infof("prompt: %s", prompt)

	//PROMPT_RAG: Acrescento o prompt como primeira mensagem
	messages := ialib.MsgGpt{}
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: prompt,
	})

	//USUÁRIO: Acrescento as mensagens de prompt do usuário
	for _, msg := range msgs.Messages {
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   msg.Id,
			Role: msg.Role,
			Text: msg.Text,
		})

	}

	//AUTOS: Acrescento cada documento dos autos como uma mensagem
	for _, msg := range autos {
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: msg.DocJsonRaw,
		})
	}
	//Verifica qual a ação requerida pelo usuário
	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages, prevID,
		config.GlobalConfig.OpenOptionModel,
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW)
	if err != nil {
		logger.Log.Errorf("Erro ao submeter análise do processo: %v", err)
		return "", nil, erros.CreateError("Erro ao submeter análise do processo: %s", err.Error())
	}

	if resp != nil {
		usage := resp.Usage
		services.ContextoServiceGlobal.UpdateTokenUso(idCtxt, int(usage.InputTokens), int(usage.OutputTokens))
	} else {
		logger.Log.Error("Resposta nula recebida do serviço OpenAI")
		return "", nil, erros.CreateError("Erro ao submeter análise do processo: %s", err.Error())
	}

	return resp.ID, resp.Output, nil
}

func (service *GeneratorType) ExecutaPreAnaliseJulgamento(ctx context.Context, idCtxt int, msgs ialib.MsgGpt, prevID string, autos []consts.ResponseAutosRow) (string, []responses.ResponseOutputItemUnion, error) {

	//Obtém o prompt que irá orientar a pré-análise e elaboração da sentença
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_COMPLEMENTO)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar o prompt: %v", err)
		return "", nil, erros.CreateError("Erro ao buscar PROMPT_RAG_COMPLEMENTO", err.Error())
	}
	//logger.Log.Infof("prompt: %s", prompt)

	//PROMPT_RAG: Acrescento o prompt como primeira mensagem
	messages := ialib.MsgGpt{}
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: prompt,
	})

	//USUÁRIO: Acrescento as mensagens de prompt do usuário
	for _, msg := range msgs.Messages {
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   msg.Id,
			Role: msg.Role,
			Text: msg.Text,
		})
	}

	//AUTOS: Acrescento cada documento dos autos como uma mensagem
	for _, msg := range autos {
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: msg.DocJsonRaw,
		})
	}

	//Verifica qual a ação requerida pelo usuário
	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages, prevID,
		config.GlobalConfig.OpenOptionModel,
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW)
	if err != nil {
		logger.Log.Errorf("Erro ao submeter pré-análise do processo: %v", err)
		return "", nil, erros.CreateError("Erro ao submeter pré-análise do processo: %s", err.Error())
	}

	if resp != nil {
		usage := resp.Usage
		services.ContextoServiceGlobal.UpdateTokenUso(idCtxt, int(usage.InputTokens), int(usage.OutputTokens))
	} else {
		logger.Log.Error("Resposta nula recebida do serviço OpenAI")
		return "", nil, erros.CreateError("Erro ao submeter análise do processo: %s", err.Error())
	}

	return resp.ID, resp.Output, nil
}
