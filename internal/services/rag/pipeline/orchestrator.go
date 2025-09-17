package pipeline

import (
	"context"
	"encoding/json"
	"ocrserver/internal/config"
	"ocrserver/internal/consts"
	"ocrserver/internal/services"
	"ocrserver/internal/services/ialib"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strconv"

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
	RAG_RESPONSE_OUTROS      = 999
)

type OrquestradorType struct {
}

func NewOrquestradorType() *OrquestradorType {
	return &OrquestradorType{}
}

type tipoEvento struct {
	Cod      int32  `json:"cod"`
	Natureza string `json:"natureza"`
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
		return submitAnaliseProcesso(ctx, id_ctxt, msgs, prevID)

	case RAG_SUBMIT_SENTENCA:
		logger.Log.Info("Resposta do SubmitPrompt: RAG_SUBMIT_SENTENCA")
		//return "", nil, erros.CreateError("Submit de Sentença não implementado", "")
		return submitProcessaSentenca(ctx, id_ctxt, msgs, prevID)
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
Rotina que efetivamente inicia a análise do processo, recuperando o contexto e disparando
a análise.
*/
func submitAnaliseProcesso(ctx context.Context, id_ctxt int, msgs ialib.MsgGpt, prevID string) (string, []responses.ResponseOutputItemUnion, error) {
	retriObj := NewRetrieverType()
	genObj := NewGeneratorType()

	autos, err := retriObj.RecuperaAutosProcesso(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao RecuperaAutosParaAnalise: %v", err)
		return "", nil, erros.CreateError("Erro ao RecuperaAutosParaAnalise: %s", err.Error())
	}

	ID, output, err := genObj.ExecutaAnaliseProcesso(ctx, id_ctxt, msgs, prevID, autos)
	if err != nil {
		logger.Log.Errorf("Erro ao ExecutaAnaliseProcesso: %v", err)
		return "", nil, erros.CreateError("Erro ao ExecutaAnaliseProcesso: %s", err.Error())
	}

	return ID, output, nil
}

// /Em implementação
func submitProcessaSentenca(ctx context.Context, id_ctxt int, msgs ialib.MsgGpt, prevID string) (string, []responses.ResponseOutputItemUnion, error) {
	//logger.Log.Infof("Resposta do SubmitPrompt: %s", retSub.OutputText())

	retriObj := NewRetrieverType()
	genObj := NewGeneratorType()

	//Recupera as análises jurídicas feitas previamente
	// analises, err := retriObj.RecuperaAnaliseJudicialParaJulgamento(ctx, id_ctxt)
	// if err != nil {
	// 	logger.Log.Errorf("Erro ao RecuperaAutosParaAnalise: %v", err)
	// 	return "", nil, erros.CreateError("Erro ao RecuperaAutosParaAnalise: %s", err.Error())
	// }
	// if len(analises) == 0 {
	// 	logger.Log.Errorf("Nenhuma análises jurídicas realizadas: %v", err)
	// 	return "", nil, erros.CreateError("Nenhuma análises jurídicas realizadas: %s", err.Error())
	// }

	autos, err := retriObj.RecuperaAnaliseJudicial(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao RecuperaAutos autos do processo: %v", err)
		return "", nil, erros.CreateError("Erro ao RecuperaAutos autos do processo: %s", err.Error())
	}
	if len(autos) == 0 {
		logger.Log.Error("Nenhuma documento processual recuperado")
		return "", nil, erros.CreateError("Nenhuma documento processual recuperado")
	}
	doutrina, err := retriObj.RecuperaDoutrinaRAG(ctx, id_ctxt)
	if err != nil {
		logger.Log.Errorf("Erro ao RecuperaAutos de doutrina: %v", err)
		return "", nil, erros.CreateError("Erro ao RecuperaAutos de doutrina %s", err.Error())
	}
	if len(doutrina) == 0 {
		logger.Log.Error("Nenhuma doutrina recuperado")
		return "", nil, erros.CreateError("Nenhuma doutrina recuperado")
	}
	///************************************
	ID, output, err := genObj.ExecutaPreAnaliseJulgamento(ctx, id_ctxt, msgs, prevID, autos)
	if err != nil {
		logger.Log.Errorf("Erro ao ExecutaAnaliseProcesso: %v", err)
		return "", nil, erros.CreateError("Erro ao ExecutaAnaliseProcesso: %s", err.Error())
	}
	//Debug
	logger.Log.Infof("Resposta do SubmitPrompt: %s", output[0].Output)

	return ID, output, nil
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
