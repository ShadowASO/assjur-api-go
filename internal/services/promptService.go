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
)

type PromptServiceType struct {
	Model *models.PromptModelType
}

func NewPromptService(
	promptModel *models.PromptModelType,

) *PromptServiceType {
	return &PromptServiceType{
		Model: promptModel,
	}
}

func (obj *PromptServiceType) GetPromptModel() (*models.PromptModelType, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.Model, nil
}

func (obj *PromptServiceType) InsertPrompt(bodyParams models.BodyParamsPromptInsert) (*models.PromptRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.InsertReg(bodyParams)
	if err != nil {
		logger.Log.Error("Erro na inclusão de um prompt.")
		return nil, err
	}
	return row, nil
}
func (obj *PromptServiceType) UpdatePrompt(bodyParams models.BodyParamsPromptUpdate) (*models.PromptRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.UpdateReg(bodyParams)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
func (obj *PromptServiceType) DeletaPrompt(id int) (*models.PromptRow, error) {
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
func (obj *PromptServiceType) SelectById(id int) (*models.PromptRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.SelectById(id)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}

func (obj *PromptServiceType) SelectAll() ([]models.PromptRow, error) {
	if obj.Model == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	rows, err := obj.Model.SelectRegs()
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return rows, nil
}
