package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ocrserver/internal/config"
	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/services/ialib"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/mslogger"

	"github.com/openai/openai-go/v3/responses"
)

type PipelineStatus int

const (
	StatusOK      PipelineStatus = iota
	StatusBlocked                // pré-condição não atendida / aguardando confirmação/complemento
	StatusInvalid                // não prossegue por regra/estado inválido do fluxo
)

func (s PipelineStatus) String() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusBlocked:
		return "blocked"
	case StatusInvalid:
		return "invalid"
	default:
		return "unknown"
	}
}

type PipelineResult struct {
	Status  PipelineStatus
	Message string

	// Mantém compatibilidade com seu padrão atual
	ID     string
	Output []responses.ResponseOutputItemUnion

	// Metadados opcionais úteis para frontend/telemetria
	EventCode int
	EventDesc string
}

func (r PipelineResult) IsTerminal() bool { return r.Status != StatusOK }

// Helpers de construção de resultado
func okResult(id string, out []responses.ResponseOutputItemUnion, msg string) PipelineResult {
	mslogger.LoggerGlobal.Infof("Response ID: %s", id)
	return PipelineResult{Status: StatusOK, ID: id, Output: out, Message: msg}
}

func blockedResult(id string, out []responses.ResponseOutputItemUnion, code int, msg string) PipelineResult {
	return PipelineResult{Status: StatusBlocked, ID: id, Output: out, EventCode: code, Message: msg}
}

func invalidResult(id string, out []responses.ResponseOutputItemUnion, msg string) PipelineResult {
	return PipelineResult{Status: StatusInvalid, ID: id, Output: out, Message: msg}
}

// Backward-compat: se você quiser manter chamadas antigas sem refatorar tudo agora
func (r PipelineResult) AsLegacy() (string, []responses.ResponseOutputItemUnion, error) {
	// Importante: aqui NÃO convertemos StatusBlocked/Invalid em error;
	// o chamador antigo já sabe lidar com "nil error" + output.
	return r.ID, r.Output, nil
}

type OrquestradorType struct{}

func NewOrquestradorType() *OrquestradorType { return &OrquestradorType{} }

// ==========================================
// NOVA ENTRADA (padrão PipelineResult)
// ==========================================
func (service *OrquestradorType) StartPipelineResult(
	ctx context.Context,
	idCtxt string,
	msgs ialib.MsgGpt,
	prevID string,
	userName string,
) (PipelineResult, error) {

	mslogger.LoggerGlobal.Infof("\n\n[Pipeline] Início do processamento - idCtxt=%s prevID=%s\n", idCtxt, prevID)
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		mslogger.LoggerGlobal.Infof("\n\n[Pipeline] Fim do processamento - idCtxt=%s prevID=%s duração=%s\n", idCtxt, prevID, duration)
	}()

	// 1) Identifica evento / confirmação
	//objTipo, respID, output, err := service.getNaturezaEventoSubmit(ctx, idCtxt, msgs, prevID)
	objTipo, respID, output, err := service.getNaturezaEventoSubmit(ctx, idCtxt, msgs, "")
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao obter a natureza do submit: %v", err)
		return PipelineResult{}, fmt.Errorf("getNaturezaEventoSubmit: %w", err)
	}

	mslogger.LoggerGlobal.Infof("\nEvento solicitado: %d - %s\n", objTipo.Tipo.Evento, objTipo.Tipo.Descricao)

	// Se for confirmação pendente (cod=300), isso é fluxo normal (BLOCKED)
	if objTipo.Tipo.Evento == EVENTO_CONFIRMACAO {
		mslogger.LoggerGlobal.Infof("\n[Pipeline] Confirmação solicitada: %s\n", objTipo.Confirmacao)
		res := blockedResult(respID, output, EVENTO_CONFIRMACAO, objTipo.Confirmacao)
		res.EventDesc = objTipo.Tipo.Descricao
		return res, nil
	}

	// 2) Executa evento (confirmed)
	res, err := service.handleEventoResult(ctx, objTipo.Tipo, idCtxt, msgs, prevID, userName)
	if err != nil {
		return PipelineResult{}, err
	}
	res.EventCode = objTipo.Tipo.Evento
	res.EventDesc = objTipo.Tipo.Descricao
	mslogger.LoggerGlobal.Infof("Response ID: %s", res.ID)
	return res, nil
}

// Se quiser manter a assinatura antiga, delegue para a nova:
func (service *OrquestradorType) StartPipeline(
	ctx context.Context,
	idCtxt string,
	msgs ialib.MsgGpt,
	prevID string,
	userName string,
) (string, []responses.ResponseOutputItemUnion, error) {

	res, err := service.StartPipelineResult(ctx, idCtxt, msgs, prevID, userName)
	if err != nil {
		// Mantém sua forma de erro padronizada
		return "", nil, erros.CreateError("Erro no pipeline: %s", err.Error())
	}
	return res.AsLegacy()
}

/*
Função para identificar a natureza das mensagens do usuário.
*/
func (service *OrquestradorType) getNaturezaEventoSubmit(
	ctx context.Context,
	idCtxt string,
	msgs ialib.MsgGpt,
	prevID string,
) (ConfirmaEvento, string, []responses.ResponseOutputItemUnion, error) {

	id_ctxt := idCtxt

	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_IDENTIFICA)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar o prompt: %v", err)
		return ConfirmaEvento{}, "", nil, erros.CreateError("Erro ao buscar PROMPT_FORMATA_RAG", err.Error())
	}

	var messages ialib.MsgGpt
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: prompt,
	})

	for _, msg := range msgs.Messages {
		messages.AddMessage(msg)
	}

	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages,
		prevID,
		config.GlobalConfig.OpenOptionModel,
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao consultar a ação desejada pelo usuário: %v", err)
		return ConfirmaEvento{}, "", nil, erros.CreateError("Erro ao consultar a ação desejada pelo usuário: %s", err.Error())
	}
	if resp == nil {
		mslogger.LoggerGlobal.Error("Resposta nula recebida do serviço OpenAI")
		return ConfirmaEvento{}, "", nil, erros.CreateError("Erro ao submeter prompt: resposta nula")
	}

	usage := resp.Usage
	services.ContextoServiceGlobal.UpdateTokenUso(id_ctxt, int(usage.InputTokens), int(usage.OutputTokens))

	// mslogger.LoggerGlobal.Infof("Response ID: %s", resp.ID)

	var objTipo ConfirmaEvento
	if err := json.Unmarshal([]byte(resp.OutputText()), &objTipo); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao realizar unmarshal na resposta tipoEvento: %v", err)
		return ConfirmaEvento{}, "", nil, erros.CreateError("Erro ao realizar unmarshal na resposta tipoEvento: %s", err.Error())
	}

	return objTipo, resp.ID, resp.Output, nil
}

// ==========================================
// handleEvento no padrão PipelineResult
// ==========================================
func (service *OrquestradorType) handleEventoResult(
	ctx context.Context,
	objTipo TipoEvento,
	id_ctxt string,
	msgs ialib.MsgGpt,
	prevID string,
	userName string,
) (PipelineResult, error) {

	switch objTipo.Evento {
	case EVENTO_ANALISE:
		return service.pipelineAnaliseProcessoResult(ctx, id_ctxt, msgs, prevID, userName)

	case EVENTO_SENTENCA:
		mslogger.LoggerGlobal.Info("\nEvento identificado: RAG_EVENTO_SENTENCA\n")
		return service.pipelineAnaliseSentencaResult(ctx, id_ctxt, msgs, prevID, userName)

	case EVENTO_COMPLEMENTO:
		mslogger.LoggerGlobal.Info("\nEvento identificado: RAG_EVENTO_COMPLEMENTO\n")
		// “não implementado” -> inválido (não é falha técnica)
		return invalidResult("", nil, "Submit de Complemento não implementado"), nil

	case EVENTO_OUTROS, EVENTO_CONCEITOS:
		mslogger.LoggerGlobal.Info("\nEvento identificado: RAG_EVENTO_OUTROS\n")
		return service.pipelineDialogoOutrosResult(ctx, id_ctxt, msgs, prevID)

	case EVENTO_ADD_BASE:
		mslogger.LoggerGlobal.Info("\nEvento identificado: RAG_EVENTO_ADD_BASE\n")
		return service.pipelineAddBaseResult(ctx, id_ctxt, userName)

	default:
		mslogger.LoggerGlobal.Warnf("Evento não reconhecido: %v", objTipo.Evento)
		return invalidResult("", nil, fmt.Sprintf("Evento não reconhecido: %d", objTipo.Evento)), nil
	}
}

// ==========================================
// pipelineAnaliseProcesso no padrão PipelineResult
// ==========================================
func (service *OrquestradorType) pipelineAnaliseProcessoResult(
	ctx context.Context,
	id_ctxt string,
	msgs ialib.MsgGpt,
	prevID string,
	userName string,
) (PipelineResult, error) {

	mslogger.LoggerGlobal.Infof("\nIniciando pipelineAnaliseProcesso...\n")
	startTime := time.Now()
	defer func() {
		mslogger.LoggerGlobal.Infof("\nFinalizando pipelineAnaliseProcesso - duração=%s.\n", time.Since(startTime))
	}()

	retriObj := NewRetrieverType()
	genObj := NewGeneratorType()

	autos, err := retriObj.RecuperaAutosProcesso(ctx, id_ctxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao recuperar os autos do processo: %v", err)
		return PipelineResult{}, fmt.Errorf("RecuperaAutosProcesso: %w", err)
	}
	if len(autos) == 0 {
		mslogger.LoggerGlobal.Warnf("Os autos do processo estão vazios (id_ctxt=%s)", id_ctxt)
		return invalidResult("", nil, "Os autos do processo estão vazios"), nil
	}
	//***   Recupera pré-análise
	//Obs. A pré-an-análise é ncessária para identificar os pontos controvertidos e usá-los para
	//buscar na base de conhecimentos subsídios para realizar uma análise jurídica completa do
	//processo. Assim, o usuário precisa solicitar duas análises jurídicas para poder gerar uma
	//minuta de sentença, esta, sim, usará a análise jurídica.
	preAnalise, err := retriObj.RecuperaPreAnaliseJuridica(ctx, id_ctxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao realizar busca de pré-análise: %v", err)
		return PipelineResult{}, fmt.Errorf("RecuperaPreAnaliseJuridica: %w", err)
	}

	var (
		ragBase     []opensearch.ResponseBaseRow
		natuAnalise = consts.NATU_DOC_IA_ANALISE
	)

	if len(preAnalise) > 0 {
		ragBase, err = retriObj.RecuperaBaseConhecimentos(ctx, id_ctxt, preAnalise[0])
		if err != nil {
			mslogger.LoggerGlobal.Errorf("Erro ao realizar RAG de doutrina: %v", err)
			return PipelineResult{}, fmt.Errorf("RecuperaBaseConhecimentos: %w", err)
		}
		if len(ragBase) == 0 {
			mslogger.LoggerGlobal.Infof("Nenhuma doutrina recuperada (id_ctxt=%s)", id_ctxt)
		}
	} else {
		mslogger.LoggerGlobal.Infof("Será realizada uma pré-análise do processo (id_ctxt=%s)", id_ctxt)
		natuAnalise = consts.NATU_DOC_IA_PREANALISE
		ragBase = []opensearch.ResponseBaseRow{}
	}

	//***   Executa análise IA
	ID, output, err := genObj.ExecutaAnaliseProcesso(ctx, id_ctxt, msgs, prevID, autos, ragBase)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao executar análise jurídica do processo: %v", err)
		return PipelineResult{}, fmt.Errorf("ExecutaAnaliseProcesso: %w", err)
	}

	docJson := extractOutputText(output)
	if strings.TrimSpace(docJson) == "" {
		mslogger.LoggerGlobal.Warnf("Nenhum texto retornado no output da IA (id_ctxt=%s)", id_ctxt)
		return invalidResult(ID, output, "Resposta da IA não contém texto"), nil
	}

	var objAnalise AnaliseJuridicaIA
	if err := json.Unmarshal([]byte(docJson), &objAnalise); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao realizar unmarshal resposta da análise: %v", err)
		mslogger.LoggerGlobal.Infof("JSON: %s", docJson)
		return PipelineResult{}, fmt.Errorf("unmarshal AnaliseJuridicaIA: %w", err)
	}

	// ==============================================================
	// 🔹 Adiciona data de geração da análise sempre
	// ==============================================================
	objAnalise.DataGeracao = time.Now().Format("02/01/2006 15:04:05")
	mslogger.LoggerGlobal.Infof("Data de geração atribuída automaticamente: %s", objAnalise.DataGeracao)

	updatedJson, err := json.MarshalIndent(objAnalise, "", "  ")
	if err != nil {
		return PipelineResult{}, fmt.Errorf("marshal AnaliseJuridicaIA: %w", err)
	}

	ok, err := service.salvarAnalise(id_ctxt, natuAnalise, "", string(updatedJson), userName)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao salvar análise (id_ctxt=%s): %v", id_ctxt, err)
		return PipelineResult{}, fmt.Errorf("salvarAnalise: %w", err)
	}
	if !ok {
		mslogger.LoggerGlobal.Errorf("Falha ao salvar análise (id_ctxt=%s)", id_ctxt)
		return PipelineResult{}, nil // falha lógica/inesperada -> pode ser error se preferir
	}

	return okResult(ID, output, "Análise salva com sucesso"), nil
}

// ==========================================
// pipelineAnaliseSentenca no padrão PipelineResult
// ==========================================
func (service *OrquestradorType) pipelineAnaliseSentencaResult(
	ctx context.Context,
	id_ctxt string,
	msgs ialib.MsgGpt,
	prevID string,
	userName string,
) (PipelineResult, error) {

	mslogger.LoggerGlobal.Infof("\nIniciando pipelineProcessaSentenca...\n")
	startTime := time.Now()
	defer func() {
		mslogger.LoggerGlobal.Infof("\nFinalizando pipelineProcessaSentenca - duração=%s.\n", time.Since(startTime))
	}()

	retriObj := NewRetrieverType()
	genObj := NewGeneratorType()

	analise, err := retriObj.RecuperaAnaliseJuridica(ctx, id_ctxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao realizar busca de análise jurídica: %v", err)
		return PipelineResult{}, fmt.Errorf("RecuperaAnaliseJuridica: %w", err)
	}
	if len(analise) == 0 {
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Nenhuma análise jurídica encontrada", id_ctxt)
		// Isso é pré-requisito de negócio -> INVALID
		return invalidResult("", nil, "Não foi realizada a análise jurídica."), nil
	}

	// =============================================================
	// 1️⃣ Verificação prévia das questões controvertidas. Será chamadas enquanto houve
	// questões controvertidas.
	// =============================================================
	codEvento, idVerif, outputVerif, err := genObj.VerificaQuestoesControvertidas(ctx, id_ctxt, msgs, prevID, analise)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("[id_ctxt=%s] Erro ao verificar questões controvertidas: %v", id_ctxt, err)
		return PipelineResult{}, fmt.Errorf("VerificaQuestoesControvertidas: %w", err)
	}

	switch codEvento {
	case EVENTO_COMPLEMENTO:
		mslogger.LoggerGlobal.Warnf("Há questões controvertidas — aguardando complementação: %v", codEvento)
		return blockedResult(idVerif, outputVerif, EVENTO_COMPLEMENTO, "Há questões controvertidas — aguardando complementação"), nil

	case EVENTO_SENTENCA:
		mslogger.LoggerGlobal.Infof("Verificação concluída — prosseguindo para geração da sentença: %v.", codEvento)

	default:
		msg := fmt.Sprintf("Código inesperado (%d) na verificação de controvérsias.", codEvento)
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] %s", id_ctxt, msg)
		return invalidResult(idVerif, outputVerif, msg), nil
	}

	autos, err := retriObj.RecuperaAutosProcesso(ctx, id_ctxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao recuperar os autos do processo: %v", err)
		return PipelineResult{}, fmt.Errorf("RecuperaAutosProcesso: %w", err)
	}
	if len(autos) == 0 {
		mslogger.LoggerGlobal.Warnf("Os autos do processo estão vazios (id_ctxt=%s)", id_ctxt)
		return invalidResult("", nil, "Os autos do processo estão vazios"), nil
	}

	ragBase, err := retriObj.RecuperaBaseConhecimentos(ctx, id_ctxt, analise[0])
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao realizar RAG de doutrina: %v", err)
		return PipelineResult{}, fmt.Errorf("RecuperaBaseConhecimentos: %w", err)
	}
	if len(ragBase) == 0 {
		mslogger.LoggerGlobal.Infof("Nenhuma doutrina recuperada (id_ctxt=%s)", id_ctxt)
	}

	ID, output, err := genObj.ExecutaAnaliseJulgamento(ctx, id_ctxt, msgs, prevID, autos, ragBase)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao executar análise jurídica do processo: %v", err)
		return PipelineResult{}, fmt.Errorf("ExecutaAnaliseJulgamento: %w", err)
	}

	docJson := extractOutputText(output)
	if strings.TrimSpace(docJson) == "" {
		return invalidResult(ID, output, "Resposta da IA não contém texto"), nil
	}

	var objMinuta MinutaSentenca
	if err := json.Unmarshal([]byte(docJson), &objMinuta); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao realizar unmarshal resposta da análise: %v", err)
		return PipelineResult{}, fmt.Errorf("unmarshal MinutaSentenca: %w", err)
	}

	objMinuta.DataGeracao = time.Now().Format("02/01/2006 15:04:05")
	mslogger.LoggerGlobal.Infof("[id_ctxt=%s] Data de geração da minuta definida: %s", id_ctxt, objMinuta.DataGeracao)

	updatedJson, err := json.MarshalIndent(objMinuta, "", "  ")
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao serializar minuta de sentença: %v", err)
		return PipelineResult{}, fmt.Errorf("marshal MinutaSentenca: %w", err)
	}

	ok, err := service.salvarAnalise(id_ctxt, consts.NATU_DOC_IA_SENTENCA, "", string(updatedJson), userName)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao salvar minuta (id_ctxt=%s): %v", id_ctxt, err)
		return PipelineResult{}, fmt.Errorf("salvarAnalise minuta: %w", err)
	}
	if !ok {
		mslogger.LoggerGlobal.Errorf("Falha ao salvar minuta (id_ctxt=%s)", id_ctxt)
		return invalidResult(ID, output, "Falha ao salvar minuta"), nil
	}

	return okResult(ID, output, "Minuta salva com sucesso"), nil
}

// ==========================================
// pipelineDialogoOutros no padrão PipelineResult
// ==========================================
func (service *OrquestradorType) pipelineDialogoOutrosResult(
	ctx context.Context,
	id_ctxt string,
	msgs ialib.MsgGpt,
	prevID string,
) (PipelineResult, error) {

	mslogger.LoggerGlobal.Infof("\nIniciando pipelineDialogoOutros...\n")
	startTime := time.Now()
	defer func() {
		mslogger.LoggerGlobal.Infof("\nFinalizando pipelineDialogoOutros - duração=%s.\n", time.Since(startTime))
	}()

	var messages ialib.MsgGpt

	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_OUTROS)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar prompt (id_ctxt=%s): %v", id_ctxt, err)
		return PipelineResult{}, fmt.Errorf("GetPromptByNatureza: %w", err)
	}

	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer",
		Text: prompt,
	})

	appendUserMessages(&messages, msgs)

	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages,
		prevID,
		config.GlobalConfig.OpenOptionModel,
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao consultar a ação desejada pelo usuário: %v", err)
		return PipelineResult{}, fmt.Errorf("SubmitPromptResponse: %w", err)
	}
	if resp == nil {
		mslogger.LoggerGlobal.Error("Resposta nula recebida do serviço OpenAI")
		return PipelineResult{}, fmt.Errorf("SubmitPromptResponse: resposta nula")
	}

	usage := resp.Usage
	services.ContextoServiceGlobal.UpdateTokenUso(id_ctxt, int(usage.InputTokens), int(usage.OutputTokens))

	mslogger.LoggerGlobal.Infof("Response ID: %s", resp.ID)

	return okResult(resp.ID, resp.Output, "Resposta gerada com sucesso"), nil
}

// ==========================================
// pipelineAddBase no padrão PipelineResult
// ==========================================
func (service *OrquestradorType) pipelineAddBaseResult(
	ctx context.Context,
	id_ctxt string,
	userName string,
) (PipelineResult, error) {

	mslogger.LoggerGlobal.Infof("\nIniciando pipelineAddBase...\n")
	startTime := time.Now()
	defer func() {
		mslogger.LoggerGlobal.Infof("\nFinalizando pipelineAddBase - duração=%s.\n", time.Since(startTime))
	}()

	retriObj := NewRetrieverType()

	sentenca, err := retriObj.RecuperaAutosSentenca(ctx, id_ctxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao recuperar a sentença dos autos: %v", err)
		return PipelineResult{}, fmt.Errorf("RecuperaAutosSentenca: %w", err)
	}
	if len(sentenca) == 0 {
		mslogger.LoggerGlobal.Warnf("Não existe sentença nos autos (id_ctxt=%s)", id_ctxt)
		return invalidResult("", nil, "Não existe sentença nos autos"), nil
	}

	ingestObj := NewIngestorType()
	if err := ingestObj.StartAddSentencaBase(ctx, sentenca, id_ctxt, userName); err != nil {
		return PipelineResult{}, fmt.Errorf("StartAddSentencaBase: %w", err)
	}

	output, err := createOutPutEventoBase(EVENTO_ADD_BASE, "Sentença adicionada à base de conhecimento!")
	if err != nil {
		return PipelineResult{}, fmt.Errorf("createOutPutEventoBase: %w", err)
	}

	return okResult("", output, "Sentença adicionada à base de conhecimento"), nil
}

// ==========================================
// Util: extrair texto do output (DRY)
// ==========================================
func extractOutputText(output []responses.ResponseOutputItemUnion) string {
	var sb strings.Builder
	for _, item := range output {
		for _, c := range item.Content {
			if c.Text != "" {
				sb.WriteString(c.Text)
				sb.WriteString("\n")
			}
		}
	}
	return strings.TrimSpace(sb.String())
}
