package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ocrserver/internal/config"
	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	openaiservice "ocrserver/internal/services/openai"

	"ocrserver/internal/services"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/mslogger"

	"github.com/openai/openai-go/v3/responses"
)

type GeneratorType struct {
}

func NewGeneratorType() *GeneratorType {
	return &GeneratorType{}
}

func (service *GeneratorType) ExecutaAnaliseProcesso(
	ctx context.Context,
	idCtxt string,
	msgs openaiservice.MsgGpt,
	prevID string,
	autos []consts.ResponseAutosRow,
	ragBase []opensearch.ResponseBaseRow,
) (string, []responses.ResponseOutputItemUnion, error) {

	if len(autos) == 0 {
		mslogger.LoggerGlobal.Warnf("Autos do processo estão vazios (id_ctxt=%s)", idCtxt)
		return "", nil, erros.CreateError("Os autos do processo estão vazios")
	}

	messages := openaiservice.MsgGpt{}

	// ============================================================
	// 01 - Developer Prompt
	// ============================================================
	service.appendDeveloperAnalise(&messages)

	// ============================================================
	// 02 - RAG Base
	// ============================================================
	service.appendBaseAnalise(&messages, ragBase)

	// ============================================================
	// 03 - Prompt Jurídico
	// ============================================================
	if err := service.appendPromptAnalise(&messages, idCtxt); err != nil {
		return "", nil, err
	}

	// ============================================================
	// 04 - Autos Processuais
	// ============================================================
	service.appendAutos(&messages, autos)

	// ============================================================
	// 05 - Mensagens do Usuário (continuação)
	// ============================================================
	// for _, msg := range msgs.Messages {
	// 	messages.AddMessage(openaiservice.MessageResponseItem{
	// 		Id:   msg.Id,
	// 		Role: msg.Role,
	// 		Text: msg.Text,
	// 	})
	// }
	// ============================================================
	// 05 - Mensagens do usuário
	// ============================================================
	appendUserMessages(&messages, msgs)

	// ============================================================
	// 06 - Envio ao modelo OpenAI
	// ============================================================
	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages,
		prevID,
		config.GlobalConfig.OpenOptionModelTop, //Usando o modelo 'OPENAI_OPTION_MODEL_TOP'
		//config.GlobalConfig.OpenOptionModel, //Usando o modelo 'OPENAI_OPTION_MODEL'
		openaiservice.REASONING_LOW,
		openaiservice.VERBOSITY_LOW,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao submeter análise (id_ctxt=%s): %v", idCtxt, err)
		return "", nil, erros.CreateError("Erro ao submeter análise: %s", err.Error())
	}
	if resp == nil {
		return "", nil, erros.CreateError("Resposta nula recebida do serviço OpenAI")
	}

	// ============================================================
	// 07 - Atualização de tokens
	// ============================================================
	totalTokens := int(resp.Usage.InputTokens + resp.Usage.OutputTokens)
	mslogger.LoggerGlobal.Infof("\n\n[id_ctxt=%s] Tokens usados: input=%d, output=%d, total=%d\n\n",
		idCtxt, resp.Usage.InputTokens, resp.Usage.OutputTokens, totalTokens)

	services.ContextoServiceGlobal.UpdateTokenUso(
		idCtxt,
		int(resp.Usage.InputTokens),
		int(resp.Usage.OutputTokens),
	)

	return resp.ID, resp.Output, nil
}

// ============================================================
// Função principal (Pipeline modularizado — Julgamento / Sentença)
// ============================================================
func (service *GeneratorType) ExecutaAnaliseJulgamento(
	ctx context.Context,
	idCtxt string,
	msgs openaiservice.MsgGpt,
	prevID string,
	autos []consts.ResponseAutosRow,
	ragBase []opensearch.ResponseBaseRow,
) (string, []responses.ResponseOutputItemUnion, error) {

	messages := openaiservice.MsgGpt{}

	// ============================================================
	// 01 - Developer Prompt
	// ============================================================
	service.appendDeveloperJulgamento(&messages)

	// ============================================================
	// 02 - RAG Base
	// ============================================================
	service.appendBaseJulgamento(&messages, ragBase)

	// ============================================================
	// 03 - Prompt Jurídico (modelo da sentença)
	// ============================================================
	if err := service.appendPromptJulgamento(&messages, idCtxt); err != nil {
		return "", nil, err
	}

	// ============================================================
	// 04 - Autos processuais
	// ============================================================
	service.appendAutos(&messages, autos)

	// ============================================================
	// 05 - Mensagens do usuário
	// ============================================================
	appendUserMessages(&messages, msgs)

	// ============================================================
	// 06 - Execução do modelo OpenAI
	// ============================================================
	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages,
		prevID,
		config.GlobalConfig.OpenOptionModelTop, //Usando o modelo 'OPENAI_OPTION_MODEL_TOP'
		openaiservice.REASONING_LOW,
		openaiservice.VERBOSITY_LOW,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao submeter análise (id_ctxt=%s): %v", idCtxt, err)
		return "", nil, erros.CreateError("Erro ao submeter análise: %s", err.Error())
	}
	if resp == nil {
		return "", nil, erros.CreateError("Resposta nula recebida do serviço OpenAI")
	}

	// ============================================================
	// 07 - Atualiza uso de tokens
	// ============================================================
	totalTokens := resp.Usage.InputTokens + resp.Usage.OutputTokens
	mslogger.LoggerGlobal.Infof("\n\n[CTX=%s] Julgamento concluído — input=%d, output=%d, total=%d tokens\n\n",
		idCtxt, resp.Usage.InputTokens, resp.Usage.OutputTokens, totalTokens)

	services.ContextoServiceGlobal.UpdateTokenUso(
		idCtxt,
		int(resp.Usage.InputTokens),
		int(resp.Usage.OutputTokens),
	)

	return resp.ID, resp.Output, nil
}

func (service *GeneratorType) VerificaQuestoesControvertidas(
	ctx context.Context,
	id_ctxt string,
	msgs openaiservice.MsgGpt,
	prevID string,
	rawsAnalise []opensearch.ResponseEventosRow,
) (int, string, []responses.ResponseOutputItemUnion, error) {

	if rawsAnalise == nil {
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Nenhuma análise jurídica encontrada", id_ctxt)
		return -1, "", nil, erros.CreateError("Não foi realizada uma análise jurídica.")
	}
	if len(rawsAnalise) == 0 {
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Nenhuma análise jurídica encontrada", id_ctxt)
		return -1, "", nil, erros.CreateError("Não foi realizada uma análise jurídica.")
	}

	// 🔹 Obtém o prompt de verificação
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_COMPLEMENTA_JULGAMENTO)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("[id_ctxt=%s] Erro ao buscar prompt: %v", id_ctxt, err)
		return -1, "", nil, erros.CreateError("Erro ao buscar prompt: %s", err.Error())
	}

	// 🧱 Cria novo objeto de mensagens preservando histórico
	var msgsAtual openaiservice.MsgGpt
	for _, m := range msgs.Messages {
		msgsAtual.AddMessage(m) // adiciona histórico anterior
	}

	// 🔹 Adiciona o prompt (como system ou developer) mantendo histórico anterior
	msgsAtual.AddMessage(openaiservice.MessageResponseItem{
		Role: "developer",
		Text: prompt,
	})

	// 🔹 Converte o registro de análise jurídica para struct Go
	jsonObj := rawsAnalise[0].DocJsonRaw
	var objAnalise AnaliseJuridicaIA
	if err := json.Unmarshal([]byte(jsonObj), &objAnalise); err != nil {
		mslogger.LoggerGlobal.Errorf("[id_ctxt=%s] Erro ao realizar unmarshal da análise jurídica: %v", id_ctxt, err)
		return -1, "", nil, erros.CreateError("Erro ao decodificar análise jurídica.")
	}

	// 🔹 Adiciona questões controvertidas como mensagens de usuário
	for _, q := range objAnalise.QuestoesControvertidas {
		texto := fmt.Sprintf("Pergunta: %s", q.PerguntaAoUsuario)
		tokens, _ := openaiservice.OpenaiGlobal.StringTokensCounter(texto)
		if tokens > MAX_DOC_TOKENS {
			texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
			mslogger.LoggerGlobal.Infof("[id_ctxt=%s] Questão truncada (%d tokens > %d)", id_ctxt, tokens, MAX_DOC_TOKENS)
		}
		msgsAtual.AddMessage(openaiservice.MessageResponseItem{
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
		openaiservice.REASONING_LOW,
		openaiservice.VERBOSITY_LOW,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("[id_ctxt=%s] Erro ao submeter prompt de verificação: %v", id_ctxt, err)
		return -1, "", nil, erros.CreateError("Erro ao submeter prompt: %s", err.Error())
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

	//---------------   EXTRAI APENAS O OBJETO JSON DA RESPOSTA

	// Extrai texto da resposta
	var textoVerif strings.Builder
	output := resp.Output
	for _, item := range output {
		for _, c := range item.Content {
			if c.Text != "" {
				textoVerif.WriteString(c.Text)
			}
		}
	}
	respVerif := strings.TrimSpace(textoVerif.String())
	//logger.Log.Infof("respVerif: %s", respVerif)

	var verif ComplementoEvento

	if err := json.Unmarshal([]byte(respVerif), &verif); err != nil {
		mslogger.LoggerGlobal.Errorf("[id_ctxt=%s] Erro ao interpretar resposta da verificação: %v", id_ctxt, err)
		return -1, resp.ID, resp.Output, erros.CreateError("Erro ao decodificar retorno da verificação das controvérsias.")
	}

	//---------------------------------------------------------

	// 🔹 Retorna resultado do modelo
	if resp == nil {
		return -1, "", nil, erros.CreateError("Resposta nula recebida do modelo")
	}

	return verif.Tipo.Evento, resp.ID, resp.Output, err
}
