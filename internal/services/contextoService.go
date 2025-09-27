/*
---------------------------------------------------------------------------------------
File: ContextoService.go
Autor: Aldenor
Data: 22-07-2025
---------------------------------------------------------------------------------------
*/
package services

import (
	"fmt"

	"ocrserver/internal/models"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"sync"
)

type ContextoServiceType struct {
	Model *models.ContextoModelType
}

var ContextoServiceGlobal *ContextoServiceType
var onceInitContextoService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitContextoService(model *models.ContextoModelType) {
	onceInitContextoService.Do(func() {
		ContextoServiceGlobal = &ContextoServiceType{
			Model: model,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewContextoService(
	model *models.ContextoModelType,

) *ContextoServiceType {
	return &ContextoServiceType{
		Model: model,
	}
}

func (obj *ContextoServiceType) GetContextoModel() (*models.ContextoModelType, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.Model, nil
}

func (obj *ContextoServiceType) InsertContexto(bodyParams models.BodyParamsContextoInsert) (*models.ContextoRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	row, err := obj.Model.InsertRow(bodyParams)
	if err != nil {
		logger.Log.Errorf("Erro ao inserir contexto: %v", err)
		return nil, erros.CreateError("Erro interno no servidor ao inserir contexto!")
	}
	return row, nil
}
func (obj *ContextoServiceType) UpdateContexto(bodyParams models.BodyParamsContextoUpdate) (*models.ContextoRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.UpdateRow(bodyParams)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
func (obj *ContextoServiceType) DeletaContexto(id int) (*models.ContextoRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	row, err := obj.Model.DeleteReg(id)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
func (obj *ContextoServiceType) SelectContextoById(id int) (*models.ContextoRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.SelectContextoById(id)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
func (obj *ContextoServiceType) SelectContextoByProcesso(nrProc string) (*models.ContextoRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.SelectContextoByProcesso(nrProc)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}

func (obj *ContextoServiceType) SelectContextoByProcessoLike(nrProc string) ([]models.ContextoRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.SelectContextoByProcessoStartsWith(nrProc)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}

func (obj *ContextoServiceType) SelectContextos(limit, offset int) ([]models.ContextoRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	rows, err := obj.Model.SelectContextos(limit, offset)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return rows, nil
}
func (obj *ContextoServiceType) ContextoExiste(nrProc string) (bool, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return false, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	isExiste, err := obj.Model.RowExists(nrProc)
	if err != nil {
		logger.Log.Errorf("Erro na verificação existência!: %v", err)
		return false, err
	}
	return isExiste, nil
}

func (obj *ContextoServiceType) UpdateTokenUso(idCtxt int, pt int, ct int) (*models.ContextoRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	row, err := obj.Model.IncrementTokensAtomic(idCtxt, pt, ct)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
