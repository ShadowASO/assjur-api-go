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

type Autos_tempServiceType struct {
	idx *opensearch.Autos_tempIndexType
}

var Autos_tempServiceGlobal *Autos_tempServiceType
var onceInitAutos_tempService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitAutos_tempService(idx *opensearch.Autos_tempIndexType) {
	onceInitAutos_tempService.Do(func() {
		Autos_tempServiceGlobal = &Autos_tempServiceType{
			idx: idx,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewAutos_tempService(
	idx *opensearch.Autos_tempIndexType,
) *Autos_tempServiceType {
	return &Autos_tempServiceType{
		idx: idx,
	}
}

func (obj *Autos_tempServiceType) InserirAutos(IdCtxt int, IdNatu int, IdPje string, doc string) (*consts.Autos_tempRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}
	row, err := obj.idx.Indexa(IdCtxt, IdNatu, IdPje, doc, "")
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return nil, err
	}
	return row, nil
}
func (obj *Autos_tempServiceType) UpdateAutos(Id string, IdCtxt int, IdNatu int, IdPje string, doc string) (*consts.Autos_tempRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	row, err := obj.idx.Update(Id, IdCtxt, IdNatu, IdPje, doc)
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return nil, err
	}
	return row, nil
}
func (obj *Autos_tempServiceType) DeletaAutos(id string) error {
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
func (obj *Autos_tempServiceType) SelectById(id string) (*consts.Autos_tempRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	row, err := obj.idx.ConsultaById(id)
	if err != nil {
		logger.Log.Error("Erro ao consultar documento %v.", err.Error())
		return nil, fmt.Errorf("Erro ao consultar documento %v.", err.Error())
	}
	return row, nil
}
func (obj *Autos_tempServiceType) SelectByContexto(idCtxt int) ([]consts.Autos_tempRow, error) {
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

func (obj *Autos_tempServiceType) GetAutosByContexto(id int) ([]consts.Autos_tempRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	rows, err := obj.SelectByContexto(id)
	if err != nil {
		logger.Log.Error("erro ao buscar sessão pelo ID")
		return nil, err
	}
	return rows, nil
}
