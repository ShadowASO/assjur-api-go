package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

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
	ragBase []opensearch.ResponseBase) (string, []responses.ResponseOutputItemUnion, error) {

	messages := ialib.MsgGpt{}

	// Validação inicial
	if len(autos) == 0 {
		logger.Log.Warningf("Autos do processo estão vazios (id_ctxt=%d)", idCtxt)
		return "", nil, erros.CreateError("Os autos do processo estão vazios")
	}
	//01 - DEVELOPER
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer",
		Text: "Você é um assistente jurídico especializado em análise de processos judiciais. Siga estritamente o formato JSON definido.",
	})

	// 02 - RAG: Base de conhecimento (RAG)

	if len(ragBase) > 0 {
		logger.Log.Info("Acrescentando a base de conhecimento")
		// txtRag := `A seguir, apresento informações jurídicas relevantes e casos semelhantes, extraídos de nossa
		// base de conhecimento. Use essas informações apenas como referência para fundamentar a análise do processo,
		// sem criar novos fatos.`
		const RAGHeader = `As informações a seguir foram recuperadas de nossa base de conhecimento jurídica (RAG).
			Elas contêm fundamentos e temas relevantes de casos semelhantes.
			Utilize-as apenas como referência para análise jurídica, sem criar novos fatos.`
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: RAGHeader,
		})

		for _, doc := range ragBase {
			//texto := doc.DataTexto
			texto := doc.Tema + ": " + doc.DataTexto
			tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
			if tokens > MAX_DOC_TOKENS { // prevenção contra prompts gigantes
				texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
				logger.Log.Infof("doutrina com %d tokens > %d: %s", tokens, MAX_DOC_TOKENS, doc.Tema)
			}
			messages.AddMessage(ialib.MessageResponseItem{
				Id:   doc.Id,
				Role: "user",
				Text: texto,
			})
			// logger.Log.Infof("\nTema: %s", doc.Tema)
			// logger.Log.Infof("\nTexto: %s", doc.DataTexto)
		}

	} else {
		logger.Log.Info("Doutrina está vazia")
	}

	// 03 - PROMPT: Obtém o prompt que irá orientar a análise e elaboração da sentença
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_ANALISE)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar prompt (id_ctxt=%d): %v", idCtxt, err)
		return "", nil, erros.CreateError("Erro ao buscar prompt: %s", err.Error())
	}
	//logger.Log.Infof("prompt: %s", prompt)

	// Adiciona como a primeira mensagem
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: prompt,
	})

	// 04 - AUTOS: Autos processuais
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

	// PROMPT DO USUÁRIO
	for _, msg := range msgs.Messages {
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   msg.Id,
			Role: msg.Role,
			Text: msg.Text,
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
	ragBase []opensearch.ResponseBase) (string, []responses.ResponseOutputItemUnion, error) {

	// Construção das mensagens
	messages := ialib.MsgGpt{}

	//01 - Contexto base de conhecimento (RAG)
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer",
		Text: `Você é um assistente jurídico especializado em análise de processos judiciais. 
		Siga estritamente o formato JSON e as regras fornecidas.`,
	})

	//02 - RAG: Acrescento a base de conhecimento RAG
	const RAGHeader = `As informações a seguir foram recuperadas de nossa base de conhecimento jurídica (RAG).
	Elas contêm fundamentos e temas relevantes de casos semelhantes.
	Utilize-as apenas como referência para análise jurídica, sem criar novos fatos.`

	// txtRag := `A seguir, apresento informações jurídicas relevantes e casos semelhantes, extraídos de nossa
	// 	base de conhecimento. Use essas informações apenas como referência para fundamentar a análise do processo,
	// 	sem criar novos fatos.`
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: RAGHeader,
	})
	for _, doc := range ragBase {
		texto := doc.DataTexto
		tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
		if tokens > MAX_DOC_TOKENS { // prevenção contra documentos gigantes
			texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
			logger.Log.Infof("doutrina com %d tokens > %d: %s", tokens, MAX_DOC_TOKENS, doc.Tema)
		}
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: texto,
		})
		logger.Log.Infof("\nTema: %s", doc.Tema)
		logger.Log.Infof("\nTexto: %s", doc.DataTexto)
	}
	if len(ragBase) == 0 {
		logger.Log.Info("Não foram obtidos registros da base de conhecimento: ragBase==0")
	}

	//03 - PROMPTO: Obtém o prompt que irá orientar a pré-análise e elaboração da sentença
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_JULGAMENTO)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar o prompt: %v", err)
		return "", nil, erros.CreateError("Erro ao buscar PROMPT_RAG_COMPLEMENTO", err.Error())
	}
	//logger.Log.Infof("prompt: %s", prompt)
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: prompt,
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

	// Mensagens do usuário
	for _, msg := range msgs.Messages {
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   msg.Id,
			Role: msg.Role,
			Text: msg.Text,
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

func (service *GeneratorType) VerificaQuestoesControvertidas(
	ctx context.Context,
	id_ctxt int,
	msgs ialib.MsgGpt,
	prevID string,
) (string, []responses.ResponseOutputItemUnion, error) {

	retriObj := NewRetrieverType()

	// 🔹 Recupera pré-análise
	preAnalise, err := retriObj.RecuperaPreAnaliseJudicial(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("[id_ctxt=%d] Erro ao realizar busca de pré-análise: %v", id_ctxt, err)
		return "", nil, erros.CreateError("Erro ao buscar pré-análise: %s", err.Error())
	}
	if len(preAnalise) == 0 {
		logger.Log.Warningf("[id_ctxt=%d] Nenhuma pré-análise encontrada", id_ctxt)
		return "", nil, erros.CreateError("Não foi realizada a pré-análise.")
	}

	// 🔹 Obtém o prompt de verificação
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_COMPLEMENTA_JULGAMENTO)
	if err != nil {
		logger.Log.Errorf("[id_ctxt=%d] Erro ao buscar prompt: %v", id_ctxt, err)
		return "", nil, erros.CreateError("Erro ao buscar prompt: %s", err.Error())
	}

	// 🧱 Cria novo objeto de mensagens preservando histórico
	var msgsAtual ialib.MsgGpt
	for _, m := range msgs.Messages {
		msgsAtual.AddMessage(m) // adiciona histórico anterior
	}

	// 🔹 Adiciona o prompt (como system ou developer) mantendo histórico anterior
	msgsAtual.AddMessage(ialib.MessageResponseItem{
		Role: "developer",
		Text: prompt,
	})

	// 🔹 Converte pré-análise para struct Go
	jsonObj := preAnalise[0].DocJsonRaw
	var objAnalise AnaliseJuridicaIA
	if err := json.Unmarshal([]byte(jsonObj), &objAnalise); err != nil {
		logger.Log.Errorf("[id_ctxt=%d] Erro ao realizar unmarshal da pré-análise: %v", id_ctxt, err)
		return "", nil, erros.CreateError("Erro ao decodificar pré-análise.")
	}

	// 🔹 Adiciona questões controvertidas como mensagens de usuário
	for _, q := range objAnalise.QuestoesControvertidas {
		texto := fmt.Sprintf("Pergunta: %s", q.PerguntaAoUsuario)
		tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
		if tokens > MAX_DOC_TOKENS {
			texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
			logger.Log.Infof("[id_ctxt=%d] Questão truncada (%d tokens > %d)", id_ctxt, tokens, MAX_DOC_TOKENS)
		}
		msgsAtual.AddMessage(ialib.MessageResponseItem{
			Role: "user",
			Text: texto,
		})
	}

	// 🔹 Submete o histórico completo (sem sobrescrever msgs)
	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		msgsAtual, // ← mantém todas as mensagens acumuladas
		prevID,
		config.GlobalConfig.OpenOptionModel,
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW,
	)
	if err != nil {
		logger.Log.Errorf("[id_ctxt=%d] Erro ao submeter prompt de verificação: %v", id_ctxt, err)
		return "", nil, erros.CreateError("Erro ao submeter prompt: %s", err.Error())
	}

	// 🔹 Atualiza uso de tokens
	if resp != nil {
		usage := resp.Usage
		services.ContextoServiceGlobal.UpdateTokenUso(
			id_ctxt,
			int(usage.InputTokens),
			int(usage.OutputTokens),
		)
	}

	// 🔹 Retorna resultado do modelo
	if resp == nil {
		return "", nil, erros.CreateError("Resposta nula recebida do modelo")
	}

	return resp.ID, resp.Output, err
}
