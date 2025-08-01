package services

import (
	"context"

	"fmt"

	"ocrserver/internal/models/types"

	"ocrserver/internal/services/tools"
	"ocrserver/internal/utils/logger"
	"strconv"

	"github.com/openai/openai-go/responses"
)

// Rotina genérica para extrair as peças do processo
func GetDocumentoAutos(idCtxt string, natDoc int) (string, error) {

	id, err := strconv.Atoi(idCtxt)
	if err != nil {
		logger.Log.Error("ID inválidos", err.Error())
		return "", fmt.Errorf("ID inválido na requisição")
	}

	rows, err := AutosServiceGlobal.GetAutosByContexto(id)

	if err != nil {
		logger.Log.Error("Erro ao buscar registros dos autos.")
		return "", fmt.Errorf("erro ao buscar registros dos autos")
	}

	for _, row := range rows {
		if row.IdNatu == natDoc {
			return string(row.DocJsonRaw), nil
		}

	}
	return "", fmt.Errorf("nenhum documento encontrado")

}

// Rotinas de manipulação dos autos

type FuncAutosDocumentoArgsVazio struct {
	//Não há necessidade de passar argumentos
}

func toolFuncPeticaoInicial(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}

func toolFuncContestacao(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}
func toolFuncReplica(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}
func toolFuncDespachoInicial(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}
func toolFuncDecisaoInterlocutoria(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}
func toolFuncDespachoOrdinatorio(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}
func toolFuncPeticaoDiversa(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}
func toolFuncEmbargosDeclaracao(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}

/*
Configura e devolve um objeto ToolManager
*/
func GetRegisterToolAutos() *tools.ToolManager {
	toolManager := tools.NewToolManager()

	tools.RegisterGenericTool(toolManager, "toolFuncPeticaoInicial", "Retorna as principais informações da petição inicial do processo", toolFuncPeticaoInicial)

	tools.RegisterGenericTool(toolManager, "toolFuncContestacao", "Retorna as principais informações das contestações apresentadas no processo", toolFuncContestacao)
	tools.RegisterGenericTool(toolManager, "toolFuncReplica", "Retorna as principais informações das réplicas apresentadas no processo", toolFuncReplica)

	tools.RegisterGenericTool(toolManager, "toolFuncDespachoInicial", "Retorna as principais informações do despacho inicial proferido no processo", toolFuncDespachoInicial)
	tools.RegisterGenericTool(toolManager, "toolFuncDecisaoInterlocutoria", "Retorna as principais informações das decisões interlocutórias proferidas no processo", toolFuncDecisaoInterlocutoria)
	tools.RegisterGenericTool(toolManager, "toolFuncDespachoOrdinatorio", "Retorna as principais informações dos despachos ordinatórios proferidos no processo", toolFuncDespachoOrdinatorio)

	tools.RegisterGenericTool(toolManager, "toolFuncPeticaoDiversa", "Retorna as principais informações das petições diversas apresentadas no processo", toolFuncPeticaoDiversa)
	tools.RegisterGenericTool(toolManager, "toolFuncEmbargosDeclaracao", "Retorna as principais informações dos embargos de declaração interpostos no curso do processo", toolFuncEmbargosDeclaracao)

	return toolManager
}

//****   FUNÇÕES HANDLERS

func handlerPeticaoInicial(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, types.ItemDocumentoPeticaoInicial)
}
func handlerContestacao(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, types.ItemDocumentoContestacao)
}
func handlerReplica(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, types.ItemDocumentoReplica)
}
func handlerDespachoInicial(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, types.ItemDocumentoDespachoInicial)
}
func handlerDecisaoInterlocutoria(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, types.ItemDocumentoDecisaoInterlocutoria)
}
func handlerDespachoOrdinatorio(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, types.ItemDocumentoDespachoOrdinatorio)
}
func handlerPeticaoDiversa(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, types.ItemDocumentoPeticaoDiversa)
}
func handlerEmbargosDeclaracao(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, types.ItemDocumentoEmbargosDeclaracao)
}

/*
	Rotina que gerencia a chamada dos handlers de cada função selecionada pelo modelo. Assim, podemos passar parâmetros

personalizados.
*/
func HandlerToolsFunc(idCtxt string, output responses.ResponseOutputItemUnion) (string, error) {
	if output.Type == "function_call" {
		logger.Log.Infof("Função: %s", output.Name)
		switch output.Name {
		case "toolFuncPeticaoInicial":
			return handlerPeticaoInicial(idCtxt)
		case "toolFuncContestacao":
			return handlerContestacao(idCtxt)
		case "toolFuncReplica":
			return handlerReplica(idCtxt)
		case "toolFuncDespachoInicial":
			return handlerDespachoInicial(idCtxt)
		case "toolFuncDecisaoInterlocutoria":
			return handlerDecisaoInterlocutoria(idCtxt)
		case "toolFuncDespachoOrdinatorio":
			return handlerDespachoOrdinatorio(idCtxt)
		case "toolFuncPeticaoDiversa":
			return handlerPeticaoDiversa(idCtxt)
		case "toolFuncEmbargosDeclaracao":
			return handlerEmbargosDeclaracao(idCtxt)
		default:
			logger.Log.Warningf("Função não reconhecida: %s", output.Name)
			return "", fmt.Errorf("função desconhecida: %s", output.Name)
		}
	}
	return "", nil
}
