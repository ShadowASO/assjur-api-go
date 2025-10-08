package pipeline

import (
	"context"

	"fmt"

	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/services/ialib"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

type RetrieverType struct {
}

func NewRetrieverType() *RetrieverType {
	return &RetrieverType{}
}

func (service *RetrieverType) RecuperaAutosProcesso(ctx context.Context, idCtxt int) ([]consts.ResponseAutosRow, error) {

	autos, err := services.AutosServiceGlobal.GetAutosByContexto(idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao recuperar os autos: %v", err)
		return nil, err
	}
	if len(autos) == 0 {
		logger.Log.Errorf("Nenhum documento processual foi localizado nos autos: %v", err)
		return nil, err
	}
	logger.Log.Infof("Documentos do processo recuperados: %d", len(autos))

	return autos, nil
}

func (service *RetrieverType) RecuperaAutosProcessoAsMessages(ctx context.Context, idCtxt int) ([]ialib.MessageResponseItem, error) {

	autos, err := services.AutosServiceGlobal.GetAutosByContexto(idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao recuperar os autos: %v", err)
		return nil, err
	}
	messages := ialib.MsgGpt{}
	if len(autos) == 0 {
		logger.Log.Errorf("Nenhum documento processual foi localizado nos autos: %v", err)
		return messages.Messages, nil
	}

	for _, msg := range autos {
		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: msg.DocJsonRaw,
		})
	}

	return messages.Messages, nil
}

func (service *RetrieverType) RecuperaAutosSentenca(ctx context.Context, idCtxt int) ([]consts.ResponseAutosRow, error) {

	autos, err := services.AutosServiceGlobal.GetAutosByContexto(idCtxt)

	if err != nil {
		logger.Log.Errorf("Erro ao recuperar os autos: %v", err)
		return nil, err
	}
	if len(autos) == 0 {
		logger.Log.Errorf("Nenhuma análise processual foi localizada: %v", err)
		return nil, err
	}
	//Procuro todos os registros com a natureza RAG_RESPONSE_ANALISE
	documentos := []consts.ResponseAutosRow{}
	for _, row := range autos {
		if row.IdNatu == consts.NATU_DOC_SENTENCA {
			documentos = append(documentos, row)
		}
	}

	return documentos, nil
}

func (service *RetrieverType) RecuperaAnaliseJudicial(ctx context.Context, idCtxt int) ([]consts.ResponseAutosRow, error) {

	autos, err := services.AutosServiceGlobal.GetAutosByContexto(idCtxt)

	if err != nil {
		logger.Log.Errorf("Erro ao recuperar os autos: %v", err)
		return nil, err
	}
	if len(autos) == 0 {
		logger.Log.Errorf("Nenhuma análise processual foi localizada: %v", err)
		return nil, err
	}
	//Procuro todos os registros com a natureza RAG_RESPONSE_ANALISE
	documentos := []consts.ResponseAutosRow{}
	for _, row := range autos {
		if row.IdNatu == RAG_EVENTO_ANALISE {
			documentos = append(documentos, row)
		}
	}

	return documentos, nil
}

func (service *RetrieverType) RecuperaPreAnaliseJudicial(
	ctx context.Context,
	idCtxt int,
) ([]consts.ResponseAutosRow, error) {

	autos, err := services.AutosServiceGlobal.GetAutosByContexto(idCtxt)
	if err != nil {
		logger.Log.Errorf("[id_ctxt=%d] Erro ao recuperar autos do contexto: %v", idCtxt, err)
		return nil, fmt.Errorf("erro ao recuperar autos do contexto: %w", err)
	}

	if len(autos) == 0 {
		logger.Log.Warningf("[id_ctxt=%d] Nenhum registro de autos encontrado para o contexto.", idCtxt)
		return nil, fmt.Errorf("nenhuma análise processual encontrada para o contexto %d", idCtxt)
	}

	documentos := make([]consts.ResponseAutosRow, 0)
	for _, row := range autos {
		if row.IdNatu == RAG_EVENTO_PREANALISE {
			if strings.TrimSpace(row.DocJsonRaw) == "" {
				logger.Log.Warningf("[id_ctxt=%d] Pré-análise encontrada (id=%s) mas JSON está vazio.", idCtxt, row.Id)
				continue
			}
			documentos = append(documentos, row)
		}
	}

	if len(documentos) == 0 {
		logger.Log.Warningf("[id_ctxt=%d] Nenhuma pré-análise válida (com JSON) encontrada entre %d autos.", idCtxt, len(autos))
		//return nil, fmt.Errorf("nenhuma pré-análise válida encontrada")
		return nil, nil
	}

	logger.Log.Infof("[id_ctxt=%d] Recuperadas %d pré-análises válidas.", idCtxt, len(documentos))
	return documentos, nil
}

func (service *RetrieverType) RecuperaDoutrinaRAG(ctx context.Context, idCtxt int) ([]opensearch.ResponseModelos, error) {

	//***   Recupera pré-análise
	preAnalise, err := service.RecuperaPreAnaliseJudicial(ctx, idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar busca de pré-análise: %v", err)
		return nil, erros.CreateError("Erro ao buscar pré-analise %s", err.Error())
	}

	if len(preAnalise) == 0 {
		logger.Log.Errorf("Nenhuma doutrina recuperada")
		return nil, nil
	}

	// Converte a string de busca num embedding
	vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, preAnalise[0].Doc)
	if err != nil {
		logger.Log.Errorf("Erro ao gerar embeddings: %v", err)
		return nil, erros.CreateError("Erro ao gerar embedding: %s", err.Error())
	}

	docs, err := opensearch.ModelosServiceGlobal.ConsultaSemantica(vec32, opensearch.GetNaturezaModelo(opensearch.MODELO_NATUREZA_DOUTRINA))
	if err != nil {
		logger.Log.Errorf("Erro ao consultar modelos de doutrina: %v", err)
		return nil, erros.CreateError("Erro ao consultar modelos de doutrina: %s", err.Error())
	}
	if len(docs) == 0 {
		logger.Log.Info("Nenhum modelo de doutrina retornado")
		return nil, nil
	}

	logger.Log.Infof("Documentos do doutrina recuperados: %d", len(docs))

	return docs, nil
}
func (service *RetrieverType) RecuperaAcordaoRAG(ctx context.Context, idCtxt int) ([]opensearch.ResponseModelos, error) {

	analise, err := service.RecuperaAnaliseJudicial(ctx, idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao recuperar acórdãos: %v", err)
		return nil, erros.CreateError("Erro ao recuperar acórdãos: %s", err.Error())
	}
	if len(analise) == 0 {
		logger.Log.Errorf("Nenhum acórdão localizado")
		return nil, nil
	}

	//Converte a string de busca num embedding
	vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, analise[0].Doc)
	if err != nil {
		logger.Log.Errorf("Erro ao gerar embeddings: %v", err)
		return nil, erros.CreateError("Erro ao gerar embedding: %s", err.Error())
	}

	docs, err := opensearch.ModelosServiceGlobal.ConsultaSemantica(vec32, opensearch.GetNaturezaModelo(opensearch.MODELO_NATUREZA_ACORDAO))
	if err != nil {
		logger.Log.Errorf("Erro ao consultar modelos de acórdão: %v", err)
		return nil, erros.CreateError("Erro ao consultar modelos de acórdão: %s", err.Error())
	}
	if len(docs) == 0 {
		logger.Log.Info("Nenhum modelo de acórdão retornado")
		return nil, nil
	}

	return docs, nil
}

func (service *RetrieverType) RecuperaSumulaRAG(ctx context.Context, idCtxt int) ([]opensearch.ResponseModelos, error) {

	analise, err := service.RecuperaAnaliseJudicial(ctx, idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao recuperar súmulas: %v", err)
		return nil, erros.CreateError("Erro ao recuperar súmulas: %s", err.Error())
	}
	if len(analise) == 0 {
		logger.Log.Errorf("Nenhuma súmula recuperada")
		return nil, nil
	}

	//Converte a string de busca num embedding
	vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, analise[0].Doc)
	if err != nil {
		logger.Log.Errorf("Erro ao gerar embeddings: %v", err)
		return nil, erros.CreateError("Erro ao gerar embedding: %s", err.Error())
	}

	docs, err := opensearch.ModelosServiceGlobal.ConsultaSemantica(vec32, opensearch.GetNaturezaModelo(opensearch.MODELO_NATUREZA_SUMULA))
	if err != nil {
		logger.Log.Errorf("Erro ao consultar modelos de súmula: %v", err)
		return nil, erros.CreateError("Erro ao consultar modelos de súmula: %s", err.Error())
	}
	if len(docs) == 0 {
		logger.Log.Info("Nenhum modelo de súmula retornado")
		return nil, nil
	}

	return docs, nil
}

// Consulta "rag_doc_embedding"
// FALTA CONCLUIR
func (service *RetrieverType) RecuperaBaseConhecimentos(ctx context.Context, idCtxt int) ([]opensearch.ResponseModelos, error) {

	//***   Recupera pré-análise
	preAnalise, err := service.RecuperaPreAnaliseJudicial(ctx, idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar busca de pré-análise: %v", err)
		return nil, erros.CreateError("Erro ao buscar pré-analise %s", err.Error())
	}

	if len(preAnalise) == 0 {
		logger.Log.Errorf("Nenhuma doutrina recuperada")
		return nil, nil
	}

	// Converte a string de busca num embedding
	vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, preAnalise[0].Doc)
	if err != nil {
		logger.Log.Errorf("Erro ao gerar embeddings: %v", err)
		return nil, erros.CreateError("Erro ao gerar embedding: %s", err.Error())
	}

	docs, err := opensearch.ModelosServiceGlobal.ConsultaSemantica(vec32, opensearch.GetNaturezaModelo(opensearch.MODELO_NATUREZA_DOUTRINA))
	if err != nil {
		logger.Log.Errorf("Erro ao consultar modelos de doutrina: %v", err)
		return nil, erros.CreateError("Erro ao consultar modelos de doutrina: %s", err.Error())
	}
	if len(docs) == 0 {
		logger.Log.Info("Nenhum modelo de doutrina retornado")
		return nil, nil
	}

	logger.Log.Infof("Documentos do doutrina recuperados: %d", len(docs))

	return docs, nil
}
