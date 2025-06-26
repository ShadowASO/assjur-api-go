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
	"ocrserver/internal/models"
	"ocrserver/internal/utils/logger"
	"sync"
)

type AutosServiceType struct {
	autosModel     *models.AutosModelType
	promptModel    *models.PromptModelType
	tempautosModel *models.DocsocrModelType
}

var AutosService *AutosServiceType
var onceInitAutosService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitAutosService(autosModel *models.AutosModelType,
	promptModel *models.PromptModelType,
	tempautosModel *models.DocsocrModelType) {
	onceInitAutosService.Do(func() {

		AutosService = &AutosServiceType{
			autosModel:     autosModel,
			promptModel:    promptModel,
			tempautosModel: tempautosModel,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewAutosService(autosModel *models.AutosModelType,
	promptModel *models.PromptModelType,
	tempautosModel *models.DocsocrModelType,
) *AutosServiceType {
	return &AutosServiceType{
		autosModel:     autosModel,
		promptModel:    promptModel,
		tempautosModel: tempautosModel,
	}
}

func (obj *AutosServiceType) GetAutosByContexto(id int) ([]models.AutosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	// AUTOS - Recupera os registros dos autos
	// var autos = models.NewAutosModel()
	rows, err := obj.autosModel.SelectByContexto(id)
	if err != nil {
		logger.Log.Error("erro ao buscar sessão pelo ID")
		return nil, err
	}
	return rows, nil
}
func (obj *AutosServiceType) InserirAutos(Data models.AutosRow) (*models.AutosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}
	row, err := obj.autosModel.InsertRow(Data)
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return nil, err
	}
	return row, nil
}
func (obj *AutosServiceType) UpdateAutos(Data models.AutosRow) (*models.AutosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}
	row, err := obj.autosModel.UpdateRow(Data)
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return nil, err
	}
	return row, nil
}
func (obj *AutosServiceType) DeletaAutos(id int) error {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}
	err := obj.autosModel.DeleteRow(id)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return fmt.Errorf("CnjApi global não configurada")
	}
	return nil
}
func (obj *AutosServiceType) SelectById(id int) (*models.AutosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}
	// row, err := service.autosModel.SelectById(id)
	row, err := obj.autosModel.SelectById(id)

	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return nil, fmt.Errorf("CnjApi global não configurada")
	}
	return row, nil
}
func (obj *AutosServiceType) SelectByContexto(idCtxt int) ([]models.AutosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}
	//rows, err := service.autosModel.SelectByContexto(idKey)
	rows, err := obj.autosModel.SelectByContexto(idCtxt)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return nil, fmt.Errorf("CnjApi global não configurada")
	}
	return rows, nil
}
