/*
---------------------------------------------------------------------------------------
File: baseService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 05-10-2025
---------------------------------------------------------------------------------------
*/

package services

import (
	"fmt"
	"sync"

	"ocrserver/internal/opensearch"
	"ocrserver/internal/utils/logger"
)

type BaseServiceType struct {
	idx *opensearch.BaseIndexType
}

var BaseServiceGlobal *BaseServiceType
var onceInitBaseService sync.Once

// InitBaseService inicializa o serviço global de índice base
func InitBaseService(idx *opensearch.BaseIndexType) {
	onceInitBaseService.Do(func() {
		BaseServiceGlobal = &BaseServiceType{idx: idx}
		logger.Log.Info("Global BaseService configurado com sucesso.")
	})
}

// NewBaseService cria uma nova instância independente do serviço base
func NewBaseService(idx *opensearch.BaseIndexType) *BaseServiceType {
	return &BaseServiceType{idx: idx}
}

// InserirDocumento indexa um novo documento no índice base
func (svc *BaseServiceType) InserirDocumento(doc opensearch.ParamsBaseInsert) error {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return fmt.Errorf("serviço BaseService não inicializado")
	}

	resp, err := svc.idx.IndexaDocumento(doc)
	if err != nil {
		logger.Log.Errorf("Erro ao indexar documento: %v", err)
		return err
	}

	if resp.Inspect().Response.StatusCode >= 400 {
		logger.Log.Errorf("Erro ao indexar: %s", resp.Inspect().Response.Status())
		return fmt.Errorf("falha ao indexar documento: %s", resp.Inspect().Response.Status())
	}

	logger.Log.Infof("Documento indexado com sucesso no índice %s", resp.Result)
	return nil
}

// UpdateDocumento atualiza o campo `data_texto` de um documento
func (svc *BaseServiceType) UpdateDocumento(id string, texto string, vector []float32) error {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return fmt.Errorf("serviço BaseService não inicializado")
	}

	params := opensearch.ParamsBaseUpdate{
		DataTexto:     texto,
		DataEmbedding: vector,
	}
	resp, err := svc.idx.UpdateDocumento(id, params)
	if err != nil {
		logger.Log.Errorf("Erro ao atualizar documento: %v", err)
		return err
	}

	if resp.Inspect().Response.StatusCode >= 400 {
		logger.Log.Errorf("Erro na atualização: %s", resp.Inspect().Response.Status())
		return fmt.Errorf("falha ao atualizar documento: %s", resp.Inspect().Response.Status())
	}

	logger.Log.Infof("Documento %s atualizado com sucesso.", id)
	return nil
}

// DeletaDocumento remove um documento pelo ID
func (svc *BaseServiceType) DeletaDocumento(id string) error {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return fmt.Errorf("serviço BaseService não inicializado")
	}

	resp, err := svc.idx.DeleteDocumento(id)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar documento: %v", err)
		return err
	}

	if resp.Inspect().Response.StatusCode >= 400 {
		logger.Log.Errorf("Erro ao deletar: %s", resp.Inspect().Response.Status())
		return fmt.Errorf("falha ao deletar documento: %s", resp.Inspect().Response.Status())
	}

	logger.Log.Infof("Documento %s deletado do índice %s", id, resp.Result)
	return nil
}

// SelectById obtém um documento por ID
func (svc *BaseServiceType) SelectById(id string) (*opensearch.ResponseBase, error) {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return nil, fmt.Errorf("serviço BaseService não inicializado")
	}

	doc, err := svc.idx.ConsultaDocumentoById(id)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar documento por ID: %v", err)
		return nil, err
	}
	if doc == nil {
		logger.Log.Warningf("Documento %s não encontrado no índice base.", id)
		return nil, nil
	}
	return doc, nil
}

// ConsultaSemantica executa uma busca vetorial no índice base
func (svc *BaseServiceType) ConsultaSemantica(vetor []float32, natureza string) ([]opensearch.ResponseBase, error) {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return nil, fmt.Errorf("serviço BaseService não inicializado")
	}

	resultados, err := svc.idx.ConsultaSemantica(vetor, natureza)
	if err != nil {
		logger.Log.Errorf("Erro na consulta semântica: %v", err)
		return nil, err
	}

	return resultados, nil
}
func (svc *BaseServiceType) IsExist(idPje string) (bool, error) {
	if svc == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return false, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	exist, err := svc.idx.IsExiste(idPje)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return false, fmt.Errorf("CnjApi global não configurada")
	}
	return exist, nil
}
