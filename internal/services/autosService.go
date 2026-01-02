/*
---------------------------------------------------------------------------------------
File: userService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
---------------------------------------------------------------------------------------
*/
package services

import (
	"fmt"

	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"

	"ocrserver/internal/utils/logger"
	"sync"
)

type AutosServiceType struct {
	idx *opensearch.AutosIndexType
}

var AutosServiceGlobal *AutosServiceType
var onceInitAutosService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitAutosService(idx *opensearch.AutosIndexType) {
	onceInitAutosService.Do(func() {

		AutosServiceGlobal = &AutosServiceType{
			idx: idx,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewAutosService(idx *opensearch.AutosIndexType,
) *AutosServiceType {
	return &AutosServiceType{

		idx: idx,
	}
}

func (obj *AutosServiceType) InserirAutos(
	IdCtxt string,
	IdNatu int,
	IdPje string,
	doc string,
	docJsonRaw string, // agora recebe string
) (*consts.ResponseAutosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	// Indexa diretamente a string JSON
	row, err := obj.idx.Indexa(IdCtxt, IdNatu, IdPje, doc, docJsonRaw, nil, "")
	if err != nil {
		logger.Log.Errorf("Erro na inclusão do registro: %s - %v", IdPje, err)
		return nil, err
	}
	return row, nil
}

func (obj *AutosServiceType) UpdateAutos(data consts.ResponseAutosRow) (*consts.ResponseAutosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}
	row, err := obj.idx.Update(data.Id, data.IdCtxt, data.IdNatu, data.IdPje, data.Doc, data.DocJsonRaw, data.DocEmbedding)
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return nil, err
	}
	return row, nil
}

func (obj *AutosServiceType) DeletaAutos(id string) error {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	err := obj.idx.Delete(id)
	if err != nil {
		logger.Log.Error("Erro ao deletar documento no índice 'autos'.")
		return fmt.Errorf("Erro ao deletar documento no índice 'autos'.")
	}

	//*******************************************************************
	emb, err := AutosJsonServiceGlobal.SelectByIdDoc(id)
	if err != nil {
		logger.Log.Error("Erro ao deletar documento no índice 'autos'.")
		return fmt.Errorf("Erro ao deletar documento no índice 'autos'.")
	}
	//logger.Log.Infof("Registro: %d.", len(emb))
	for _, reg := range emb {

		//logger.Log.Infof("Registro: %s.", reg.Id)

		err := AutosJsonServiceGlobal.DeletaEmbedding(reg.Id)
		if err != nil {
			logger.Log.Error("Erro ao deletar documento no índice 'autos'.")
			return fmt.Errorf("Erro ao deletar documento no índice 'autos'.")
		}
	}

	return nil
}
func (obj *AutosServiceType) SelectById(id string) (*consts.ResponseAutosRow, error) {
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
func (obj *AutosServiceType) SelectByContexto(idCtxt string) ([]consts.ResponseAutosRow, error) {
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

func (obj *AutosServiceType) GetAutosByContexto(id string) ([]consts.ResponseAutosRow, error) {
	if obj == nil {
		logger.Log.Error("Serviço AutosServiceGlobal não inicializado.")
		return nil, fmt.Errorf("serviço AutosServiceGlobal não inicializado")
	}

	rows, err := obj.SelectByContexto(id)
	if err != nil {
		logger.Log.Errorf("[id_ctxt=%d] Erro ao buscar autos do contexto: %v", id, err)
		return nil, fmt.Errorf("erro ao buscar autos do contexto %d: %w", id, err)
	}

	if len(rows) == 0 {
		logger.Log.Warningf("[id_ctxt=%d] Nenhum registro de autos encontrado no contexto.", id)
		// retornar erro semântico ou não, dependendo do uso
		// return nil, fmt.Errorf("nenhum registro de autos encontrado para o contexto %d", id)
	}

	logger.Log.Infof("[id_ctxt=%d] Recuperados %d registros de autos.", id, len(rows))
	return rows, nil
}

func (obj *AutosServiceType) IsDocAutuado(idCtxt string, idPje string) (bool, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return false, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	exist, err := obj.idx.IsExiste(idCtxt, idPje)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return false, fmt.Errorf("CnjApi global não configurada")
	}
	return exist, nil
}
