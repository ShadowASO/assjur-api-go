package services

import (
	"context"

	"fmt"

	"ocrserver/internal/consts"

	"ocrserver/internal/services/tools"
	"ocrserver/internal/utils/logger"

	"github.com/openai/openai-go/v3/responses"
)

// Rotina genérica para extrair as peças do processo
func GetDocumentoAutos(idCtxt string, natDoc int) (string, error) {

	// id, err := strconv.Atoi(idCtxt)
	// if err != nil {
	// 	logger.Log.Error("ID inválidos", err.Error())
	// 	return "", fmt.Errorf("ID inválido na requisição")
	// }

	rows, err := AutosServiceGlobal.GetAutosByContexto(idCtxt)

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
func toolFuncDespacho(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}
func toolFuncDecisao(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}

//	func toolFuncDespachoOrdinatorio(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
//		return "", nil
//	}
func toolFuncPeticao(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}
func toolFuncEmbargosDeclaracao(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
	return "", nil
}
func toolFuncModeloSentenca(ctx context.Context, args FuncAutosDocumentoArgsVazio) (string, error) {
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

	tools.RegisterGenericTool(toolManager, "toolFuncDespacho", "Retorna as principais informações do despacho inicial proferido no processo", toolFuncDespacho)
	tools.RegisterGenericTool(toolManager, "toolFuncDecisao", "Retorna as principais informações das decisões interlocutórias proferidas no processo", toolFuncDecisao)
	//tools.RegisterGenericTool(toolManager, "toolFuncDespachoOrdinatorio", "Retorna as principais informações dos despachos ordinatórios proferidos no processo", toolFuncDespachoOrdinatorio)

	tools.RegisterGenericTool(toolManager, "toolFuncPeticao", "Retorna as principais informações das petições diversas apresentadas no processo", toolFuncPeticao)
	tools.RegisterGenericTool(toolManager, "toolFuncEmbargosDeclaracao", "Retorna as principais informações dos embargos de declaração interpostos no curso do processo", toolFuncEmbargosDeclaracao)

	tools.RegisterGenericTool(toolManager, "toolFuncModeloSentenca", "Retorna um modelo de sentença para ser observado na geração de minutas", toolFuncModeloSentenca)

	return toolManager
}

//****   FUNÇÕES HANDLERS

func handlerPeticaoInicial(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, consts.NATU_DOC_INICIAL)
}
func handlerContestacao(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, consts.NATU_DOC_CONTESTACAO)
}
func handlerReplica(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, consts.NATU_DOC_REPLICA)
}
func handlerDespacho(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, consts.NATU_DOC_DESPACHO)
}
func handlerDecisao(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, consts.NATU_DOC_DECISAO)
}
func handlerPeticao(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, consts.NATU_DOC_PETICAO)
}
func handlerEmbargosDeclaracao(idCtxt string) (string, error) {
	return GetDocumentoAutos(idCtxt, consts.NATU_DOC_EMBARGOS)
}

func handlerMinutaSentenca(idCtxt string) (string, error) {
	//inserir uma busca neste ponto
	return "modelo de sentença", nil
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
		case "toolFuncDespacho":
			return handlerDespacho(idCtxt)
		case "toolFuncDecisao":
			return handlerDecisao(idCtxt)
		case "toolFuncPeticao":
			return handlerPeticao(idCtxt)
		case "toolFuncEmbargosDeclaracao":
			return handlerEmbargosDeclaracao(idCtxt)
		case "toolFuncMinutaSentenca":
			return handlerMinutaSentenca(idCtxt)
		default:
			logger.Log.Warningf("Função não reconhecida: %s", output.Name)
			return "", fmt.Errorf("função desconhecida: %s", output.Name)
		}
	}
	return "", nil
}
