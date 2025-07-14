/*
---------------------------------------------------------------------------------------
File: autosEmbeddingService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 11-07-2025
---------------------------------------------------------------------------------------
*/
//package embedding
package services

import (
	"context"
	"fmt"

	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	//"ocrserver/internal/services"

	"ocrserver/internal/utils/logger"
	"sync"
)

type AutosJsonServiceType struct {
	idx *opensearch.AutosJsonEmbeddingType
}

var AutosJsonServiceGlobal *AutosJsonServiceType
var onceInitAutosJsonService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitAutosJsonService(idx *opensearch.AutosJsonEmbeddingType) {
	onceInitAutosJsonService.Do(func() {

		AutosJsonServiceGlobal = &AutosJsonServiceType{
			idx: idx,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewAutosJsonService(idx *opensearch.AutosJsonEmbeddingType,
) *AutosJsonServiceType {
	return &AutosJsonServiceType{

		idx: idx,
	}
}

func (obj *AutosJsonServiceType) InserirEmbedding(idDoc string, IdCtxt int, IdNatu int, doc_embedding []float32) (*consts.ResponseAutosJsonEmbeddingRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	row, err := obj.idx.Indexa(idDoc, IdCtxt, IdNatu, doc_embedding)
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return nil, err
	}
	return row, nil
}
func (obj *AutosJsonServiceType) UpdateEmbedding(id string, idDoc string, IdCtxt int, IdNatu int, doc_embedding []float32) (*consts.ResponseAutosJsonEmbeddingRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	row, err := obj.idx.Update(id, idDoc, IdCtxt, IdNatu, doc_embedding)
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return nil, err
	}
	return row, nil
}
func (obj *AutosJsonServiceType) DeletaEmbedding(id string) error {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	err := obj.idx.Delete(id)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return fmt.Errorf("CnjApi global não configurada")
	}
	return nil
}
func (obj *AutosJsonServiceType) SelectById(id string) (*consts.ResponseAutosJsonEmbeddingRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	row, err := obj.idx.ConsultaById(id)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return nil, fmt.Errorf("CnjApi global não configurada")
	}
	return row, nil
}
func (obj *AutosJsonServiceType) SelectByContexto(idCtxt int) ([]consts.ResponseAutosJsonEmbeddingRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	rows, err := obj.idx.ConsultaByIdCtxt(idCtxt)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return nil, fmt.Errorf("CnjApi global não configurada")
	}
	return rows, nil
}

// Inclui um novo documento no índice autos_embedding
func (obj *AutosJsonServiceType) IncluirDocumento(idDoc string, idCtxt int, idNatu int, idPje string, doc string) (string, error) {
	ctx := context.Background()
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar AutosEmbeddingType global sem inicializá-la.")
		return "", fmt.Errorf("AutosEmbeddingType global não configurada")
	}

	// Gera o embedding do documento
	embeddingResp, err := OpenaiServiceGlobal.GetEmbeddingFromText(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar embedding do texto: %w", err)
	}
	//Converte o vetor para 32
	vector32 := OpenaiServiceGlobal.Float64ToFloat32Slice(embeddingResp)

	resp, err := obj.InserirEmbedding(idDoc, idCtxt, idNatu, vector32)
	if err != nil {
		logger.Log.Errorf("Erro ao indexar documento: %v", err)
		return "", err
	}
	logger.Log.Infof("Documento inserido em %v: %v", "Autos", resp.Id)

	return resp.Id, nil
}
