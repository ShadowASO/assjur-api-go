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
	IdCtxt int,
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
		logger.Log.Error("Erro na inclusão do registro", err.Error())
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
func (obj *AutosServiceType) SelectByContexto(idCtxt int) ([]consts.ResponseAutosRow, error) {
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

func (obj *AutosServiceType) GetAutosByContexto(id int) ([]consts.ResponseAutosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	//rows, err := obj.autosModel.SelectByContexto(id)
	rows, err := obj.SelectByContexto(id)
	if err != nil {
		logger.Log.Error("erro ao buscar sessão pelo ID")
		return nil, err
	}
	return rows, nil
}

func (obj *AutosServiceType) IsDocAutuado(idCtxt int, idPje string) (bool, error) {
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
