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
func (svc *BaseServiceType) InserirDocumento(
	idCtxt string,
	idPje string,
	//hashTexto string,
	userNameInc string,
	//dtInc time.Time,
	//status string,

	classe string,
	assunto string,
	natureza string,
	tipo string,
	tema string,

	fonte string,
	texto string,
	hashTexto string,

	//textoEmbedding []float32,
) (*opensearch.ResponseBaseRow, error) {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return nil, fmt.Errorf("serviço BaseService não inicializado")
	}

	vector, err := GetDocumentoEmbeddings(texto)
	if err != nil {
		logger.Log.Errorf("Erro ao gerar embeddings: %v", err)
		//response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar embeddings", "", requestID)
		return nil, fmt.Errorf("Erro ao gerar embeddings")
	}
	//hashTexto := ""
	status := "S"

	resp, err := svc.idx.Indexa(
		idCtxt,
		idPje,
		hashTexto,
		userNameInc,
		status,
		classe,
		assunto,
		natureza,
		tipo,
		tema,
		fonte,
		texto,
		vector,
		"",
	)
	if err != nil {
		logger.Log.Errorf("Erro ao indexar documento: %v", err)
		return nil, err
	}

	if resp == nil {
		logger.Log.Errorf("Erro ao indexar documento")
		return nil, fmt.Errorf("falha ao indexar documento")
	}

	logger.Log.Infof("Documento indexado com sucesso: %s", resp.Id)
	return resp, nil
}

// UpdateDocumento atualiza o campo `data_texto` de um documento
func (svc *BaseServiceType) UpdateDocumento(id string, tema string, texto string, vector []float32) (*opensearch.ResponseBaseRow, error) {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return nil, fmt.Errorf("serviço BaseService não inicializado")
	}

	// params := opensearch.ParamsBaseUpdate{
	// 	DataTexto:     texto,
	// 	DataEmbedding: vector,
	// }
	resp, err := svc.idx.Update(id, tema, texto, vector)
	if err != nil {
		logger.Log.Errorf("Erro ao indexar documento: %v", err)
		return nil, err
	}

	if resp == nil {
		logger.Log.Errorf("Erro ao indexar documento")
		return nil, fmt.Errorf("falha ao indexar documento")
	}

	logger.Log.Infof("Documento atualizado com sucesso: %s.", resp.Id)
	return resp, nil
}

// DeletaDocumento remove um documento pelo ID
func (svc *BaseServiceType) DeletaDocumento(id string) error {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return fmt.Errorf("serviço BaseService não inicializado")
	}

	err := svc.idx.Delete(id)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar documento: %v", err)
		return err
	}

	return nil
}

// SelectById obtém um documento por ID
func (svc *BaseServiceType) SelectById(id string) (*opensearch.ResponseBaseRow, error) {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return nil, fmt.Errorf("serviço BaseService não inicializado")
	}

	doc, err := svc.idx.ConsultaById(id)
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
// func (svc *BaseServiceType) ConsultaSemantica(vetor []float32, natureza string) ([]opensearch.ResponseBaseRow, error) {
func (svc *BaseServiceType) ConsultaSemantica(texto string, natureza string) ([]opensearch.ResponseBaseRow, error) {
	if svc == nil || svc.idx == nil {
		logger.Log.Error("Tentativa de uso de BaseService não iniciado.")
		return nil, fmt.Errorf("serviço BaseService não inicializado")
	}

	vector, err := GetDocumentoEmbeddings(texto)
	if err != nil {
		logger.Log.Errorf("Erro ao gerar embeddings: %v", err)
		return nil, fmt.Errorf("Erro ao gerar embeddings")
	}

	rows, err := svc.idx.ConsultaSemantica(vector, natureza)
	if err != nil {
		logger.Log.Errorf("Erro na consulta semântica: %v", err)
		return nil, err
	}

	return rows, nil
}
func (svc *BaseServiceType) IsExist(id_ctxt string, idPje string, hash_texto string) (bool, error) {
	if svc == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return false, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	exist, err := svc.idx.IsExiste(id_ctxt, idPje, hash_texto)
	if err != nil {
		logger.Log.Error("Erro ao verificar a existência do chunk na base de conhecimentos.")
		return false, fmt.Errorf("Erro ao verificar a existência do chunk na base de conhecimentos")
	}
	return exist, nil
}
