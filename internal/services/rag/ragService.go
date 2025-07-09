package rag

import (
	"context"

	"fmt"

	"ocrserver/internal/database/pgdb"
	"ocrserver/internal/models"
	"ocrserver/internal/models/types"
	"ocrserver/internal/services/tools"
	"ocrserver/internal/utils/logger"
	"strconv"
)

// Função de teste
type GetNomeClienteArgs struct {
	Premiado string `json:"premiado" description:"Número do bilhete premiado"`
}

func GetNomeCliente(ctx context.Context, args GetNomeClienteArgs) (string, error) {
	logger.Log.Info("entrei")
	logger.Log.Info(args.Premiado)
	return "Aldenor Sombra de Oliveira", nil
}

// Rotina genérica para extrair as peças do processo
func GetDocumentoAutos(idCtxt string, natDoc int) (string, error) {

	logger.Log.Info(types.GetDocumentoDescriptionByKey(natDoc))

	id, err := strconv.Atoi(idCtxt)
	if err != nil {
		logger.Log.Error("ID inválidos", err.Error())
		return "", fmt.Errorf("ID inválido na requisição")

	}
	autosModel := models.NewAutosModel(pgdb.DBPoolGlobal.Pool)

	rows, err := autosModel.SelectByContexto(id)
	if err != nil {
		logger.Log.Error("Erro ao buscar registros dos autos.")
		return "", fmt.Errorf("erro ao buscar registros dos autos")
	}

	for _, row := range rows {
		if row.IdNatu == natDoc {
			return string(row.Doc), nil
		}

	}
	return "", fmt.Errorf("nenhum documento encontrado")

}

// Rotinas de manipulação dos autos
type FuncAutosDocumentoArgs struct {
	IdCtxt string `json:"IdCtxt" description:"Identifica o contexto criado para reunir as informações e peças de um processo judicial"`
}

func GetPeticaoInicial(ctx context.Context, args FuncAutosDocumentoArgs) (string, error) {
	return GetDocumentoAutos(args.IdCtxt, types.ItemDocumentoPeticaoInicial)
}
func GetContestacao(ctx context.Context, args FuncAutosDocumentoArgs) (string, error) {
	return GetDocumentoAutos(args.IdCtxt, types.ItemDocumentoContestacao)
}
func GetReplica(ctx context.Context, args FuncAutosDocumentoArgs) (string, error) {
	return GetDocumentoAutos(args.IdCtxt, types.ItemDocumentoReplica)
}
func GetDespachoInicial(ctx context.Context, args FuncAutosDocumentoArgs) (string, error) {
	return GetDocumentoAutos(args.IdCtxt, types.ItemDocumentoDespachoInicial)
}
func GetDecisaoInterlocutoria(ctx context.Context, args FuncAutosDocumentoArgs) (string, error) {
	return GetDocumentoAutos(args.IdCtxt, types.ItemDocumentoDecisaoInterlocutoria)
}
func GetDespachoOrdinatorio(ctx context.Context, args FuncAutosDocumentoArgs) (string, error) {
	return GetDocumentoAutos(args.IdCtxt, types.ItemDocumentoDespachoOrdinatorio)
}
func GetPeticaoDiversa(ctx context.Context, args FuncAutosDocumentoArgs) (string, error) {
	return GetDocumentoAutos(args.IdCtxt, types.ItemDocumentoPeticaoDiversa)
}
func GetEmbargosDeclaracao(ctx context.Context, args FuncAutosDocumentoArgs) (string, error) {
	return GetDocumentoAutos(args.IdCtxt, types.ItemDocumentoEmbargosDeclaracao)
}
func GetSentenca(ctx context.Context, args FuncAutosDocumentoArgs) (string, error) {
	return GetDocumentoAutos(args.IdCtxt, types.ItemDocumentoSentenca)
}

/*
Configura e devolve um objeto ToolManager
*/
func GetRegisterToolAutos() *tools.ToolManager {
	toolManage := tools.NewToolManager()

	tools.RegisterGenericTool(toolManage, "GetPeticaoInicial", "Retorna as principais informação da petição inicial do processo", GetPeticaoInicial)
	tools.RegisterGenericTool(toolManage, "GetContestacao", "Retorna as principais informação das contestações apresentadas no processo", GetContestacao)
	tools.RegisterGenericTool(toolManage, "GetReplica", "Retorna as principais informação das réplicas apresentadas no processo", GetReplica)

	tools.RegisterGenericTool(toolManage, "GetDespachoInicial", "Retorna as principais informação do despacho inicial proferido no processo", GetDespachoInicial)
	tools.RegisterGenericTool(toolManage, "GetDecisaoInterlocutoria", "Retorna as principais informação das decisões interlocutórias proferidas no processo", GetDecisaoInterlocutoria)
	tools.RegisterGenericTool(toolManage, "GetDespachoOrdinatorio", "Retorna as principais informação dos despachos ordinatórios proferidos no processo", GetDespachoOrdinatorio)
	tools.RegisterGenericTool(toolManage, "GetPeticaoDiversa", "Retorna as principais informação das petições diversas apresentadas no processo", GetPeticaoDiversa)
	tools.RegisterGenericTool(toolManage, "GetEmbargosDeclaracao", "Retorna as principais informação dos embargos de declaração interpostos no curso do processo", GetEmbargosDeclaracao)
	tools.RegisterGenericTool(toolManage, "GetSentenca", "Retorna as principais informação das sentenças proferidas no curso do processo", GetSentenca)

	return toolManage
}
