package pipeline

import (
	"context"
	"encoding/json"
	"ocrserver/internal/config"
	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/services/ialib"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strconv"
	"strings"

	"github.com/openai/openai-go/v2/responses"
)

/* Naturezas dos submits do usuário. */
const (
	RAG_SUBMIT_ANALISE  = 101
	RAG_SUBMIT_SENTENCA = 102
	RAG_SUBMIT_DECISAO  = 103
	RAG_SUBMIT_DESPACHO = 104
	//-----  Comp
	RAG_SUBMIT_COMPLEMENTO = 201
	RAG_SUBMIT_OUTROS      = 999
)

/* Código das respostas do modelo. */
const (
	RAG_RESPONSE_ANALISE  = 201
	RAG_RESPONSE_SENTENCA = 202
	RAG_RESPONSE_DECISAO  = 203
	RAG_RESPONSE_DESPACHO = 204
	//-----  Comp
	RAG_RESPONSE_COMPLEMENTO = 301
	RAG_RESPONSE_PREANALISE  = 302
	RAG_RESPONSE_OUTROS      = 999
)

const MAX_DOC_TOKENS = 2000

type OrquestradorType struct {
}

func NewOrquestradorType() *OrquestradorType {
	return &OrquestradorType{}
}

type tipoEvento struct {
	Cod      int32  `json:"cod"`
	Natureza string `json:"natureza"`
}

type tipoResponse struct {
	Tipo_resp int    `json:"tipo_resp"`
	Texto     string `json:"texto"`
}

func (service *OrquestradorType) StartPipeline(ctx context.Context, idCtxt string, msgs ialib.MsgGpt, prevID string) (string, []responses.ResponseOutputItemUnion, error) {
	// Converte IdCtxt para int
	id_ctxt, err := strconv.Atoi(idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao converter idCtxt para int: %v", err)
		return "", nil, erros.CreateError("Erro ao converter o idCtxt para int", err.Error())
	}

	objTipo, err := service.getNaturezaSubmit(ctx, idCtxt, msgs, prevID)
	if err != nil {
		logger.Log.Errorf("Erro ao obter a natureza do submit: %v", err)
		return "", nil, erros.CreateError("Erro ao obter a natureza do submit: %s", err.Error())
	}

	// chama função auxiliar
	ID, output, err := handleSubmits(ctx, objTipo, id_ctxt, msgs, prevID)
	return ID, output, err
}

/*
Função para identificar a natureza das mensagems do usuário. Aresposta possível:

	RAG_SUBMIT_ANALISE  = 101
	RAG_SUBMIT_SENTENCA = 102
	RAG_SUBMIT_DECISAO  = 103
	RAG_SUBMIT_DESPACHO = 104
	//-----  Comp
	RAG_SUBMIT_COMPLEMENTO = 201
	RAG_SUBMIT_OUTROS      = 999
*/
func (service *OrquestradorType) getNaturezaSubmit(ctx context.Context, idCtxt string, msgs ialib.MsgGpt, prevID string) (tipoEvento, error) {
	id_ctxt, err := strconv.Atoi(idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao converter idCtxt para int: %v", err)
		return tipoEvento{}, erros.CreateError("Erro ao converter idCtxt para int: %s", err.Error())
	}
	//***  IDENTIFICAÇÃO DO EVENTO
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_IDENTIFICA)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar o prompt: %v", err)
		return tipoEvento{}, erros.CreateError("Erro ao buscar PROMPT_FORMATA_RAG", err.Error())
	}

	var messages ialib.MsgGpt
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: prompt,
	})

	for _, msg := range msgs.Messages {
		messages.AddMessage(msg)
		logger.Log.Infof("Mensagens: %s", msg.Text)
	}

	resp, err := services.OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages, prevID,
		config.GlobalConfig.OpenOptionModel,
		ialib.REASONING_LOW,
		ialib.VERBOSITY_LOW,
	)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar a ação desejada pelo usuário: %v", err)
		return tipoEvento{}, erros.CreateError("Erro ao consultar a ação desejada pelo usuário: %s", err.Error())
	}
	if resp != nil {
		usage := resp.Usage
		services.ContextoServiceGlobal.UpdateTokenUso(id_ctxt, int(usage.InputTokens), int(usage.OutputTokens))
	} else {
		logger.Log.Error("Resposta nula recebida do serviço OpenAI")
		return tipoEvento{}, erros.CreateError("Erro ao submeter prompt: %s", err.Error())
	}
	//Debug
	//logger.Log.Infof("Resposta do SubmitPrompt: %s", resp.OutputText())

	// mapeia JSON de retorno
	var objTipo tipoEvento
	err = json.Unmarshal([]byte(resp.OutputText()), &objTipo)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar unmarshal na resposta tipoEvento: %v", err)
		return tipoEvento{}, erros.CreateError("Erro ao realizar unmarshal na resposta tipoEvento: %s", err.Error())
	}

	return objTipo, nil
}

func handleSubmits(ctx context.Context, objTipo tipoEvento, id_ctxt int, msgs ialib.MsgGpt, prevID string) (string, []responses.ResponseOutputItemUnion, error) {
	switch objTipo.Cod {
	case RAG_SUBMIT_ANALISE:
		return pipelineAnaliseProcesso(ctx, id_ctxt, msgs, prevID)

	case RAG_SUBMIT_SENTENCA:
		logger.Log.Info("Resposta do SubmitPrompt: RAG_SUBMIT_SENTENCA")
		//return "", nil, erros.CreateError("Submit de Sentença não implementado", "")
		return pipelineProcessaSentenca(ctx, id_ctxt, msgs, prevID)
	case RAG_SUBMIT_COMPLEMENTO:
		logger.Log.Info("Resposta do SubmitPrompt: RAG_SUBMIT_COMPLEMENTO")
		return "", nil, erros.CreateError("Submit de Complemento não implementado", "")
	case RAG_SUBMIT_OUTROS:
		logger.Log.Info("Resposta do SubmitPrompt: RAG_SUBMIT_OUTROS")
		return "", nil, erros.CreateError("Submit de Outros não implementado", "")

	default:
		logger.Log.Warningf("Evento não reconhecido: %v", objTipo.Cod)
		return "", nil, erros.CreateError("Evento não reconhecido: %v", string(objTipo.Cod))
	}
}

/*
O pipeline da análise do processo está concentrado nesta função.
*/
func pipelineAnaliseProcesso(
	ctx context.Context,
	id_ctxt int,
	msgs ialib.MsgGpt,
	prevID string) (string, []responses.ResponseOutputItemUnion, error) {

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
	preAnalise, err := retriObj.RecuperaPreAnalise(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar busca de pré-análise: %v", err)
		return "", nil, erros.CreateError("Erro ao buscar pré-analise %s", err.Error())
	}

	//***   Define natureza da análise
	var (
		ragDoutrina []opensearch.ResponseModelos
		natuAnalise = RAG_RESPONSE_ANALISE
	)

	if len(preAnalise) > 0 {

		// Recupera doutrina via RAG
		ragDoutrina, err = retriObj.RecuperaDoutrinaRAG(ctx, id_ctxt)
		if err != nil {
			logger.Log.Errorf("Erro ao realizar RAG de doutrina: %v", err)
			return "", nil, erros.CreateError("Erro ao realizar RAG de doutrina %s", err.Error())
		}
		if len(ragDoutrina) == 0 {
			logger.Log.Infof("Nenhuma doutrina recuperada (id_ctxt=%d)", id_ctxt)
		}
	} else {
		logger.Log.Infof("Nenhuma pré-análise encontrada (id_ctxt=%d)", id_ctxt)
		natuAnalise = RAG_RESPONSE_PREANALISE
		ragDoutrina = []opensearch.ResponseModelos{}
	}

	//***   Executa análise IA
	ID, output, err := genObj.ExecutaAnaliseProcesso(ctx, id_ctxt, msgs, prevID, autos, ragDoutrina)
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
	resposta := strings.TrimSpace(sb.String())

	if resposta == "" {
		logger.Log.Warningf("Nenhum texto retornado pela IA (id_ctxt=%d)", id_ctxt)
		return "", output, erros.CreateError("Resposta da IA não contém texto")
	}

	if resposta == "" {
		logger.Log.Warningf("Nenhum texto retornado no output da IA (id_ctxt=%d)", id_ctxt)
		return "", output, erros.CreateError("Resposta da IA não contém texto")
	}

	//*** Converte objeto JSON para um objeto GO(tipoResponse)
	var objAnalise tipoResponse
	err = json.Unmarshal([]byte(resposta), &objAnalise)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar unmarshal resposta da análise: %v", err)
		return ID, output, erros.CreateError("Erro ao unmarshal resposta da análise")
	}

	//***  Salva análise/pré-análise
	ok, err := salvarAnaliseProcesso(ctx, id_ctxt, natuAnalise, objAnalise.Texto, "")

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

func salvarAnaliseProcesso(ctx context.Context, idCtxt int, natu int, doc string, docJson string) (bool, error) {

	row, err := services.AutosServiceGlobal.InserirAutos(idCtxt, natu, "", doc, docJson)
	if err != nil {
		logger.Log.Errorf("Erro na inclusão da análise %v", err)
		return false, erros.CreateError("Erro na inclusão do registro: %s", err.Error())
	}
	logger.Log.Infof("ID do registro: %s", row.Id)
	return true, nil
}

// /Em implementação
func pipelineProcessaSentenca(
	ctx context.Context,
	id_ctxt int,
	msgs ialib.MsgGpt,
	prevID string) (string, []responses.ResponseOutputItemUnion, error) {

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

	var ragDoutrina []opensearch.ResponseModelos

	// Recupera doutrina via RAG
	ragDoutrina, err = retriObj.RecuperaDoutrinaRAG(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar RAG de doutrina: %v", err)
		return "", nil, erros.CreateError("Erro ao realizar RAG de doutrina %s", err.Error())
	}
	if len(ragDoutrina) == 0 {
		logger.Log.Infof("Nenhuma doutrina recuperada (id_ctxt=%d)", id_ctxt)
	}

	//***   Executa análise IA
	ID, output, err := genObj.ExecutaAnaliseJulgamento(ctx, id_ctxt, msgs, prevID, autos, ragDoutrina)
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
	resposta := strings.TrimSpace(sb.String())

	if resposta == "" {
		logger.Log.Warningf("Nenhum texto retornado pela IA (id_ctxt=%d)", id_ctxt)
		return "", output, erros.CreateError("Resposta da IA não contém texto")
	}

	if resposta == "" {
		logger.Log.Warningf("Nenhum texto retornado no output da IA (id_ctxt=%d)", id_ctxt)
		return "", output, erros.CreateError("Resposta da IA não contém texto")
	}

	//*** Converte objeto JSON para um objeto GO(tipoResponse)
	var objAnalise tipoResponse
	err = json.Unmarshal([]byte(resposta), &objAnalise)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar unmarshal resposta da análise: %v", err)
		return ID, output, erros.CreateError("Erro ao unmarshal resposta da análise")
	}

	//***  Salva análise/pré-análise
	ok, err := salvarMinutaSentenca(ctx, id_ctxt, RAG_RESPONSE_SENTENCA, objAnalise.Texto, "")

	if err != nil {
		logger.Log.Errorf("Erro ao salvar minuta (id_ctxt=%d): %v", id_ctxt, err)
		return ID, output, err
	}
	if !ok {
		logger.Log.Errorf("Falha ao salvar minuta (id_ctxt=%d)", id_ctxt)
		return ID, output, erros.CreateError("Erro ao salvar minuta")
	}

	return ID, output, nil
}

func salvarMinutaSentenca(ctx context.Context, idCtxt int, natu int, doc string, docJson string) (bool, error) {

	row, err := services.AutosServiceGlobal.InserirAutos(idCtxt, natu, "", doc, docJson)
	if err != nil {
		logger.Log.Errorf("Erro na inclusão da minuta %v", err)
		return false, erros.CreateError("Erro na inclusão da minuta: %s", err.Error())
	}
	logger.Log.Infof("ID do registro: %s", row.Id)
	return true, nil
}

func (service *OrquestradorType) StarterOld(ctx context.Context, idCtxt string, msgs ialib.MsgGpt, prevID string) (string, []responses.ResponseOutputItemUnion, error) {
	// Converte IdCtxt para int
	id_ctxt, err := strconv.Atoi(idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao converter idCtxt para int: %v", err)
		return "", nil, erros.CreateError("Erro ao converter o idCtxt para int", err.Error())
	}

	//***  IDENTIFICAÇÃO DO EVENTO
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_IDENTIFICA)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar o prompt: %v", err)
		return "", nil, erros.CreateError("Erro ao buscar PROMPT_FORMATA_RAG", err.Error())
	}

	var messages ialib.MsgGpt
	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: prompt,
	})

	for _, msg := range msgs.Messages {
		messages.AddMessage(msg)
		logger.Log.Infof("Mensagens: %s", msg.Text)
	}

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

	// mapeia JSON de retorno
	var objTipo tipoEvento
	err = json.Unmarshal([]byte(resp.OutputText()), &objTipo)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar unmarshal na resposta tipoEvento: %v", err)
		return "", nil, erros.CreateError("Erro ao realizar unmarshal na resposta tipoEvento: %s", err.Error())
	}

	// chama função auxiliar
	//ID, output, err := handleSubmits(ctx, objTipo, id_ctxt, retSub)
	ID, output, err := handleSubmits(ctx, objTipo, id_ctxt, msgs, prevID)
	return ID, output, err
}
