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

	"strings"
	"time"

	"github.com/openai/openai-go/v3/responses"
)

type OrquestradorType struct {
}

func NewOrquestradorType() *OrquestradorType {
	return &OrquestradorType{}
}

func (service *OrquestradorType) StartPipeline(
	ctx context.Context,
	idCtxt string,
	msgs ialib.MsgGpt,
	prevID string,
	userName string) (string, []responses.ResponseOutputItemUnion, error) {

	logger.Log.Infof("\n\n[Pipeline] Início do processamento - idCtxt=%s prevID=%s\n", idCtxt, prevID)
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		logger.Log.Infof("\n\n[Pipeline] Fim do processamento - idCtxt=%s prevID=%s duração=%s\n", idCtxt, prevID, duration)
	}()

	id_ctxt := idCtxt

	objTipo, output, err := service.getNaturezaEventoSubmit(ctx, idCtxt, msgs, prevID)
	if err != nil {
		logger.Log.Errorf("Erro ao obter a natureza do submit: %v", err)
		return "", nil, erros.CreateError("Erro ao obter a natureza do submit: %s", err.Error())
	}
	logger.Log.Infof("\nEvento acionado: %d - %s\n", objTipo.Tipo.Evento, objTipo.Tipo.Descricao)
	// Verifica se é confirmação pendente (cod=300)
	if objTipo.Tipo.Evento == 300 {
		logger.Log.Infof("\n[Pipeline] Confirmação solicitada: %s\n", objTipo.Confirmacao)

		return "", output, nil
	}

	//  Executa o evento normalmente (já confirmado)
	ID, output, err := service.handleEvento(ctx, objTipo.Tipo, id_ctxt, msgs, prevID, userName)
	return ID, output, err
}

/*
Função para identificar a natureza das mensagems do usuário. Aresposta possível:
*/
func (service *OrquestradorType) getNaturezaEventoSubmit(ctx context.Context, idCtxt string, msgs ialib.MsgGpt, prevID string) (ConfirmaEvento, []responses.ResponseOutputItemUnion, error) {
	// id_ctxt, err := strconv.Atoi(idCtxt)
	// if err != nil {
	// 	logger.Log.Errorf("Erro ao converter idCtxt para int: %v", err)
	// 	return ConfirmaEvento{}, nil, erros.CreateError("Erro ao converter idCtxt para int: %s", err.Error())
	// }
	id_ctxt := (idCtxt)
	//***  IDENTIFICAÇÃO DO EVENTO
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_IDENTIFICA)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar o prompt: %v", err)
		return ConfirmaEvento{}, nil, erros.CreateError("Erro ao buscar PROMPT_FORMATA_RAG", err.Error())
	}

	var messages ialib.MsgGpt
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: prompt,
	})

	for _, msg := range msgs.Messages {
		messages.AddMessage(msg)
		//logger.Log.Infof("Mensagens: %s", msg.Text)
	}

	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages, prevID,
		config.GlobalConfig.OpenOptionModel,
		//config.GlobalConfig.OpenOptionModelSecundary, //Estou usando o GPT-5-nano
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW,
	)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar a ação desejada pelo usuário: %v", err)
		return ConfirmaEvento{}, nil, erros.CreateError("Erro ao consultar a ação desejada pelo usuário: %s", err.Error())
	}
	if resp != nil {
		usage := resp.Usage
		services.ContextoServiceGlobal.UpdateTokenUso(id_ctxt, int(usage.InputTokens), int(usage.OutputTokens))
	} else {
		logger.Log.Error("Resposta nula recebida do serviço OpenAI")
		return ConfirmaEvento{}, nil, erros.CreateError("Erro ao submeter prompt: %s", err.Error())
	}
	//Debug
	//logger.Log.Infof("Resposta do SubmitPrompt: %s", resp.OutputText())

	// mapeia JSON de retorno
	var objTipo ConfirmaEvento
	err = json.Unmarshal([]byte(resp.OutputText()), &objTipo)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar unmarshal na resposta tipoEvento: %v", err)
		return ConfirmaEvento{}, nil, erros.CreateError("Erro ao realizar unmarshal na resposta tipoEvento: %s", err.Error())
	}

	return objTipo, resp.Output, nil
}

func (service *OrquestradorType) handleEvento(
	ctx context.Context,
	objTipo TipoEvento,
	id_ctxt string,
	msgs ialib.MsgGpt,
	prevID string,
	userName string) (string, []responses.ResponseOutputItemUnion, error) {
	switch objTipo.Evento {
	case RAG_EVENTO_ANALISE:
		return service.pipelineAnaliseProcesso(ctx, id_ctxt, msgs, prevID, userName)
	case RAG_EVENTO_SENTENCA:
		logger.Log.Info("\nEvento identificado: RAG_EVENTO_SENTENCA\n")
		return service.pipelineAnaliseSentenca(ctx, id_ctxt, msgs, prevID, userName)
	case RAG_EVENTO_COMPLEMENTO:
		logger.Log.Info("\nEvento identificado: RAG_EVENTO_COMPLEMENTO\n")
		return "", nil, erros.CreateError("Submit de Complemento não implementado", "")
	case RAG_EVENTO_OUTROS, RAG_EVENTO_CONCEITOS:
		logger.Log.Info("\nEvento identificado: RAG_EVENTO_OUTROS\n")
		return service.pipelineDialogoOutros(ctx, id_ctxt, msgs, prevID)
	case RAG_EVENTO_ADD_BASE:
		logger.Log.Info("\nEvento identificado: RAG_EVENTO_ADD_BASE\n")
		return service.pipelineAddBase(ctx, id_ctxt, userName)
	default:
		logger.Log.Warningf("Evento não reconhecido: %v", objTipo.Evento)
		return "", nil, erros.CreateErrorf("Evento não reconhecido: %d", objTipo.Evento)
	}
}

/*
O pipeline da análise do processo está concentrado nesta função.
*/
func (service *OrquestradorType) pipelineAnaliseProcesso(
	ctx context.Context,
	id_ctxt string,
	msgs ialib.MsgGpt,
	prevID string,
	userName string) (string, []responses.ResponseOutputItemUnion, error) {

	//------------------ Registra no log o início do pipeline
	logger.Log.Infof("\nIniciando pipelineAnaliseProcesso...\n")
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		logger.Log.Infof("\nFinalizando pipelineAnaliseProcesso - duração=%s.\n", duration)
	}()
	//----------------------

	retriObj := NewRetrieverType()
	genObj := NewGeneratorType()

	//*** Recupera AUTOS
	autos, err := retriObj.RecuperaAutosProcesso(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao recuperar os autos do processo: %v", err)
		return "", nil, erros.CreateError("Erro ao recuperar os autos do processo: %s", err.Error())
	}
	if len(autos) == 0 {
		logger.Log.Warningf("Os autos do processo estão vazios (id_ctxt=%d)", id_ctxt)
		return "", nil, erros.CreateError("Os autos do processo estão vazios")
	}

	//***   Recupera pré-análise
	//Obs. A pré-an-análise é ncessária para identificar os pontos controvertidos e usá-los para
	//buscar na base de conhecimentos subsídios para realizar uma análise jurídica completa do
	//processo. Assim, o usuário precisa solicitar duas análises jurídicas para poder gerar uma
	//minuta de sentença, esta, sim, usará a análise jurídica.
	preAnalise, err := retriObj.RecuperaPreAnaliseJuridica(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar busca de pré-análise: %v", err)
		return "", nil, erros.CreateError("Erro ao buscar pré-analise %s", err.Error())
	}

	//***   Define natureza da análise
	var (
		ragBase     []opensearch.ResponseBaseRow
		natuAnalise = consts.NATU_DOC_IA_ANALISE
	)

	//Sempre buscar a base de conhecimentos
	if len(preAnalise) > 0 {

		// Recupera base de conhecimento
		ragBase, err = retriObj.RecuperaBaseConhecimentos(ctx, id_ctxt, preAnalise[0])
		if err != nil {
			logger.Log.Errorf("Erro ao realizar RAG de doutrina: %v", err)
			return "", nil, erros.CreateError("Erro ao realizar RAG de doutrina %s", err.Error())
		}
		if len(ragBase) == 0 {
			logger.Log.Infof("Nenhuma doutrina recuperada (id_ctxt=%d)", id_ctxt)
		}
	} else {
		logger.Log.Infof("Será realizada uma pré-análise do processo (id_ctxt=%d)", id_ctxt)
		natuAnalise = consts.NATU_DOC_IA_PREANALISE
		ragBase = []opensearch.ResponseBaseRow{}
	}

	//***   Executa análise IA
	ID, output, err := genObj.ExecutaAnaliseProcesso(ctx, id_ctxt, msgs, prevID, autos, ragBase)
	if err != nil {
		logger.Log.Errorf("Erro ao executar análise jurídica do processo: %v", err)
		return "", nil, erros.CreateError("Erro ao executar análise jurídica do processo: %s", err.Error())
	}

	//***   Extrai resposta em texto
	var sb strings.Builder
	for _, item := range output {
		for _, c := range item.Content {
			if c.Text != "" {
				sb.WriteString(c.Text)
				sb.WriteString("\n")
			}
		}
	}
	docJson := strings.TrimSpace(sb.String())

	if docJson == "" {
		logger.Log.Warningf("Nenhum texto retornado no output da IA (id_ctxt=%d)", id_ctxt)
		return "", output, erros.CreateError("Resposta da IA não contém texto")
	}

	//*** Converte objeto JSON para um objeto GO(tipoResponse)
	var objAnalise AnaliseJuridicaIA

	err = json.Unmarshal([]byte(docJson), &objAnalise)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar unmarshal resposta da análise: %v", err)
		return ID, output, erros.CreateError("Erro ao unmarshal resposta da análise")
	}

	// ==============================================================
	// 🔹 Adiciona data de geração da análise sempre
	// ==============================================================

	objAnalise.DataGeracao = time.Now().Format("02/01/2006 15:04:05")
	logger.Log.Infof("Data de geração atribuída automaticamente: %s", objAnalise.DataGeracao)

	//*** Regrava JSON atualizado com data_geracao
	updatedJson, err := json.MarshalIndent(objAnalise, "", "  ")
	if err != nil {
		return ID, output, erros.CreateError("Erro ao serializar análise atualizada: %s", err.Error())
	}

	//***  Salva análise/pré-análise

	ok, err := service.salvarAnalise(id_ctxt, natuAnalise, "", string(updatedJson), userName)
	if err != nil {
		logger.Log.Errorf("Erro ao salvar análise (id_ctxt=%d): %v", id_ctxt, err)
		return ID, output, err
	}
	if !ok {
		logger.Log.Errorf("Falha ao salvar análise (id_ctxt=%d)", id_ctxt)
		return ID, output, erros.CreateError("Erro ao salvar análise")
	}

	return ID, output, nil
}

// /Em implementação
func (service *OrquestradorType) pipelineAnaliseSentenca(
	ctx context.Context,
	id_ctxt string,
	msgs ialib.MsgGpt,
	prevID string, userName string) (string, []responses.ResponseOutputItemUnion, error) {

	//------------------ Registra o início e fim no log
	logger.Log.Infof("\nIniciando pipelineProcessaSentenca...\n")
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		logger.Log.Infof("\nFinalizando pipelineProcessaSentenca - duração=%s.\n", duration)
	}()
	//----------------------

	retriObj := NewRetrieverType()
	genObj := NewGeneratorType()

	//***   Recupera Análise Jurídica
	analise, err := retriObj.RecuperaAnaliseJuridica(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar busca de análise jurídica: %v", err)
		return "", nil, erros.CreateErrorf("Erro ao buscar analise jurídica %s", err.Error())
	}
	if analise == nil {
		logger.Log.Warningf("[id_ctxt=%d] Nenhuma análise jurídica encontrada", id_ctxt)
		return "", nil, erros.CreateError("Não foi realizada uma análise jurídica.")
	}
	if len(analise) == 0 {
		logger.Log.Warningf("[id_ctxt=%d] Nenhuma análise jurídica encontrada", id_ctxt)
		return "", nil, erros.CreateError("Não foi realizada uma análise jurídica.")
	}

	// =============================================================
	// 1️⃣ Verificação prévia das questões controvertidas. Será chamadas enquanto houve
	// questões controvertidas.
	// =============================================================
	codEvento, idVerif, outputVerif, err := genObj.VerificaQuestoesControvertidas(ctx, id_ctxt, msgs, prevID, analise)
	if err != nil {
		logger.Log.Errorf("[id_ctxt=%d] Erro ao verificar questões controvertidas: %v", id_ctxt, err)
		return idVerif, outputVerif, erros.CreateErrorf("Erro na verificação das questões controvertidas: %s", err.Error())
	}

	// Avalida o código de evento retornado
	switch codEvento {
	case 301:
		logger.Log.Warningf("Há questões controvertidas — aguardando complementação: %v", codEvento)
		return idVerif, outputVerif, nil

	case 202:
		logger.Log.Infof("Verificação concluída — prosseguindo para geração da sentença: %v.", codEvento)

	default:
		msg := fmt.Sprintf("Código inesperado (%d) na verificação de controvérsias.", codEvento)
		logger.Log.Warningf("[id_ctxt=%d] %s", id_ctxt, msg)
		return idVerif, outputVerif, erros.CreateError(msg)
	}

	// =============================================================
	// 2️⃣ Recupera autos do processo
	// =============================================================
	autos, err := retriObj.RecuperaAutosProcesso(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao recuperar os autos do processo: %v", err)
		return "", nil, erros.CreateError("Erro ao recuperar os autos do processo: %s", err.Error())
	}
	if len(autos) == 0 {
		logger.Log.Warningf("Os autos do processo estão vazios (id_ctxt=%d)", id_ctxt)
		return "", nil, erros.CreateError("Os autos do processo estão vazios")
	}

	// =============================================================
	// 3️⃣ Recupera doutrina via RAG
	// =============================================================
	ragBase, err := retriObj.RecuperaBaseConhecimentos(ctx, id_ctxt, analise[0])
	if err != nil {
		logger.Log.Errorf("Erro ao realizar RAG de doutrina: %v", err)
		return "", nil, erros.CreateError("Erro ao realizar RAG de doutrina %s", err.Error())
	}
	if len(ragBase) == 0 {
		logger.Log.Infof("Nenhuma doutrina recuperada (id_ctxt=%d)", id_ctxt)
	}

	// =============================================================
	// 4️⃣ Executa a geração da minuta de sentença via IA
	// =============================================================
	ID, output, err := genObj.ExecutaAnaliseJulgamento(ctx, id_ctxt, msgs, prevID, autos, ragBase)
	if err != nil {
		logger.Log.Errorf("Erro ao executar análise jurídica do processo: %v", err)
		return "", nil, erros.CreateError("Erro ao executar análise jurídica do processo: %s", err.Error())
	}

	// =============================================================
	// 5️⃣ Extrai texto do retorno da IA
	// =============================================================
	var sb strings.Builder
	for _, item := range output {
		for _, c := range item.Content {
			if c.Text != "" {
				sb.WriteString(c.Text)
				sb.WriteString("\n")
			}
		}
	}
	docJson := strings.TrimSpace(sb.String())
	if docJson == "" {
		return "", output, erros.CreateError("Resposta da IA não contém texto")
	}

	// =============================================================
	// 6️⃣ Converte JSON em objeto Go (MinutaSentenca)
	// =============================================================
	var objMinuta MinutaSentenca
	if err := json.Unmarshal([]byte(docJson), &objMinuta); err != nil {
		logger.Log.Errorf("Erro ao realizar unmarshal resposta da análise: %v", err)
		return ID, output, erros.CreateError("Erro ao unmarshal resposta da análise")
	}

	// =============================================================
	// 7️⃣ Adiciona data de geração da sentença sempre
	// =============================================================

	objMinuta.DataGeracao = time.Now().Format("02/01/2006 15:04:05")
	logger.Log.Infof("[id_ctxt=%d] Data de geração da minuta definida: %s", id_ctxt, objMinuta.DataGeracao)

	// Recria JSON com o campo atualizado
	updatedJson, err := json.MarshalIndent(objMinuta, "", "  ")
	if err != nil {
		logger.Log.Errorf("Erro ao serializar minuta de sentença: %v", err)
		return ID, output, erros.CreateError("Erro ao serializar minuta de sentença: %s", err.Error())
	}

	// =============================================================
	// 8️⃣ Salva minuta
	// =============================================================
	//ok, err := service.salvarMinutaSentenca(ctx, id_ctxt, consts.NATU_DOC_IA_SENTENCA, "", string(updatedJson))
	ok, err := service.salvarAnalise(id_ctxt, consts.NATU_DOC_IA_SENTENCA, "", string(updatedJson), userName)
	if err != nil {
		logger.Log.Errorf("Erro ao salvar minuta (id_ctxt=%d): %v", id_ctxt, err)
		return ID, output, err
	}
	if !ok {
		logger.Log.Errorf("Falha ao salvar minuta (id_ctxt=%d)", id_ctxt)
		return ID, output, erros.CreateError("Erro ao salvar minuta ")
	}

	return ID, output, nil
}

func (service *OrquestradorType) pipelineDialogoOutros(
	ctx context.Context,
	id_ctxt string,
	msgs ialib.MsgGpt,
	prevID string) (string, []responses.ResponseOutputItemUnion, error) {

	//------------------
	logger.Log.Infof("\nIniciando pipelineDialogoOutros...\n")
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		logger.Log.Infof("\nFinalizando pipelineDialogoOutros - duração=%s.\n", duration)
	}()
	//----------------------
	var messages ialib.MsgGpt

	//Obtém o prompt que irá orientar a análise e elaboração da sentença
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_OUTROS)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar prompt (id_ctxt=%d): %v", id_ctxt, err)
		return "", nil, erros.CreateError("Erro ao buscar prompt: %s", err.Error())
	}
	//logger.Log.Infof("prompt: %s", prompt)

	//SYSTEM PROMPT: Adiciona o prompt do sistema
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer",
		Text: prompt,
	})

	//USER PROMPT: as mensagem que o Usuário inseriu no prompt
	appendUserMessages(&messages, msgs)

	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages, prevID,
		config.GlobalConfig.OpenOptionModel,
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW,
	)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar a ação desejada pelo usuário: %v", err)
		return "", nil, erros.CreateError("Erro ao consultar a ação desejada pelo usuário: %s", err.Error())
	}
	if resp != nil {
		usage := resp.Usage
		services.ContextoServiceGlobal.UpdateTokenUso(id_ctxt, int(usage.InputTokens), int(usage.OutputTokens))
	} else {
		logger.Log.Error("Resposta nula recebida do serviço OpenAI")
		return "", nil, erros.CreateError("Erro ao submeter prompt: %s", err.Error())
	}
	//Debug
	//logger.Log.Infof("Resposta do SubmitPrompt: %s", resp.OutputText())

	return resp.ID, resp.Output, err
}

//---------**************************************************************************

// --*********************************************************************************
// Faz a inclusão da sentença na base de precedentes
func (service *OrquestradorType) pipelineAddBase(
	ctx context.Context,
	id_ctxt string,
	userName string) (string, []responses.ResponseOutputItemUnion, error) {

	//------------------
	logger.Log.Infof("\nIniciando pipelineAddBase...\n")
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		logger.Log.Infof("\nFinalizando pipelineAddBase - duração=%s.\n", duration)
	}()
	//----------------------

	retriObj := NewRetrieverType()

	//01 - AUTOS: *** Recupera a SENTENÇA PROFERIDA  DOS AUTOS
	sentenca, err := retriObj.RecuperaAutosSentenca(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao recuperar a sentença dos autos: %v", err)
		return "", nil, erros.CreateError("Erro ao recuperar a sentença dos autos: %s", err.Error())
	}
	if len(sentenca) == 0 {
		logger.Log.Warningf("Não existe sentença nos autos (id_ctxt=%d)", id_ctxt)
		return "", nil, erros.CreateError("Não existe sentença nos autos")
	}
	ingestObj := NewIngestorType()

	err = ingestObj.StartAddSentencaBase(ctx, sentenca, id_ctxt,
		userName)
	if err != nil {
		return "", nil, erros.CreateError("Erro ao adicionar a sentença à base de conhecimento!")
	}

	output, err := createOutPutEventoBase(RAG_EVENTO_ADD_BASE, "Sentença adicionada à base de conhecimento!")
	if err != nil {
		return "", nil, err
	}

	return "", output, nil

}
