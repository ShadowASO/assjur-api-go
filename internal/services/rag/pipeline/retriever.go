package pipeline

import (
	"context"
	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/services/ialib"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
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
		if row.IdNatu == RAG_RESPONSE_ANALISE {
			documentos = append(documentos, row)
		}
	}

	return documentos, nil
}

func (service *RetrieverType) RecuperaDoutrinaRAG(ctx context.Context, idCtxt int) ([]opensearch.ResponseModelos, error) {

	analise, err := service.RecuperaAnaliseJudicial(ctx, idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao recuperar análise jurídica: %v", err)
		return nil, erros.CreateError("Erro ao recuperar análise jurídica: %s", err.Error())
	}
	if len(analise) == 0 {
		logger.Log.Errorf("Nenhuma analise recuperada")
		return nil, nil
	}

	//Converte a string de busca num embedding
	vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, analise[0].Doc)
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

	return docs, nil
}
