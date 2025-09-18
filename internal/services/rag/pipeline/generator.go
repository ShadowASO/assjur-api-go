package pipeline

import (
	"context"

	"ocrserver/internal/config"
	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"

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

func (service *GeneratorType) ExecutaAnaliseProcesso(
	ctx context.Context,
	idCtxt int,
	msgs ialib.MsgGpt,
	prevID string,
	autos []consts.ResponseAutosRow,
	ragDoutrina []opensearch.ResponseModelos) (string, []responses.ResponseOutputItemUnion, error) {

	// Validação inicial
	if len(autos) == 0 {
		logger.Log.Warningf("Autos do processo estão vazios (id_ctxt=%d)", idCtxt)
		return "", nil, erros.CreateError("Os autos do processo estão vazios")
	}

	//Obtém o prompt que irá orientar a análise e elaboração da sentença
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_ANALISE)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar prompt (id_ctxt=%d): %v", idCtxt, err)
		return "", nil, erros.CreateError("Erro ao buscar prompt: %s", err.Error())
	}
	//logger.Log.Infof("prompt: %s", prompt)

	// Construção das mensagens
	messages := ialib.MsgGpt{}

	// Prompt inicial como "developer"
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer",
		Text: prompt,
	})

	// Mensagens do usuário
	for _, msg := range msgs.Messages {
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   msg.Id,
			Role: msg.Role,
			Text: msg.Text,
		})
	}

	// Contexto doutrinário (RAG)
	if len(ragDoutrina) > 0 {
		logger.Log.Info("Acrescentando a doutrina")
		//promptDoutrina := "Para realizar a análise, considere o seguinte contexto: "
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "developer",
			Text: "Considere também os seguintes trechos de doutrina:",
		})

		for _, doc := range ragDoutrina {
			texto := doc.Inteiro_teor
			tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
			if tokens > MAX_DOC_TOKENS { // prevenção contra prompts gigantes
				texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
				logger.Log.Infof("doutrina com %d tokens > %d: %s", tokens, MAX_DOC_TOKENS, doc.Ementa)
			}
			messages.AddMessage(ialib.MessageResponseItem{
				Id:   doc.Id,
				Role: "user",
				Text: texto,
			})
		}

	} else {
		logger.Log.Info("Doutrina está vazia")
	}

	// Autos processuais
	for _, doc := range autos {
		texto := doc.DocJsonRaw
		tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
		if tokens > MAX_DOC_TOKENS { // prevenção contra prompts gigantes
			texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
			logger.Log.Infof("peça processual com %d tokens  > %d: %s", tokens, MAX_DOC_TOKENS, doc.IdPje)
		}
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: texto,
		})
	}

	// Chamada ao serviço OpenAI
	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages, prevID,
		config.GlobalConfig.OpenOptionModel,
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW,
	)

	if err != nil {
		logger.Log.Errorf("Erro ao submeter análise (id_ctxt=%d): %v", idCtxt, err)
		return "", nil, erros.CreateError("Erro ao submeter análise: %s", err.Error())
	}

	if resp == nil {
		logger.Log.Errorf("Resposta nula recebida do serviço OpenAI (id_ctxt=%d)", idCtxt)
		return "", nil, erros.CreateError("Resposta nula recebida do serviço OpenAI")
	}

	// Atualiza uso de tokens

	services.ContextoServiceGlobal.UpdateTokenUso(
		idCtxt,
		int(resp.Usage.InputTokens),
		int(resp.Usage.OutputTokens),
	)

	return resp.ID, resp.Output, nil
}

func (service *GeneratorType) ExecutaAnaliseJulgamento(ctx context.Context,
	idCtxt int,
	msgs ialib.MsgGpt,
	prevID string,
	autos []consts.ResponseAutosRow,
	ragDoutrina []opensearch.ResponseModelos) (string, []responses.ResponseOutputItemUnion, error) {

	//Obtém o prompt que irá orientar a pré-análise e elaboração da sentença
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_JULGAMENTO)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar o prompt: %v", err)
		return "", nil, erros.CreateError("Erro ao buscar PROMPT_RAG_COMPLEMENTO", err.Error())
	}
	//logger.Log.Infof("prompt: %s", prompt)

	// Construção das mensagens
	messages := ialib.MsgGpt{}

	// Prompt inicial como "developer"
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer",
		Text: prompt,
	})

	// Mensagens do usuário
	for _, msg := range msgs.Messages {
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   msg.Id,
			Role: msg.Role,
			Text: msg.Text,
		})
	}

	// Contexto dos autos processuais

	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: "A análise deve incidir sobre os autos do processo que seguem: ",
	})

	for _, doc := range autos {
		texto := doc.DocJsonRaw
		tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
		if tokens > MAX_DOC_TOKENS { // prevenção contra documentos gigantes
			texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
			logger.Log.Infof("doutrina com %d tokens > %d: %s", tokens, MAX_DOC_TOKENS, doc.IdPje)
		}
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: texto,
		})
	}

	// Contexto doutrinário (RAG)

	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "system",
		Text: "Considere também os seguintes trechos de doutrina: ",
	})

	for _, doc := range ragDoutrina {
		texto := doc.Inteiro_teor
		tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
		if tokens > MAX_DOC_TOKENS { // prevenção contra documentos gigantes
			texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
			logger.Log.Infof("doutrina com %d tokens > %d: %s", tokens, MAX_DOC_TOKENS, doc.Ementa)
		}
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: texto,
		})
	}

	// Chamada ao serviço OpenAI
	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages, prevID,
		config.GlobalConfig.OpenOptionModel,
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW)
	if err != nil {
		logger.Log.Errorf("Erro ao submeter análise (id_ctxt=%d): %v", idCtxt, err)
		return "", nil, erros.CreateError("Erro ao submeter análise: %s", err.Error())
	}

	if resp == nil {
		logger.Log.Errorf("Resposta nula recebida do serviço OpenAI (id_ctxt=%d)", idCtxt)
		return "", nil, erros.CreateError("Resposta nula recebida do serviço OpenAI")
	}
	// Atualiza uso de tokens

	services.ContextoServiceGlobal.UpdateTokenUso(
		idCtxt,
		int(resp.Usage.InputTokens),
		int(resp.Usage.OutputTokens),
	)

	return resp.ID, resp.Output, nil
}
