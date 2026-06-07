/*
---------------------------------------------------------------------------------------
File: promptService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 17-05-2025
---------------------------------------------------------------------------------------
*/
package services

import (
	"fmt"

	"ocrserver/internal/models/postgres"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/mslogger"

	"sync"
)

type PromptServiceType struct {
	Model *postgres.PromptModelType
}

var PromptServiceGlobal *PromptServiceType
var onceInitPromptService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitPromptService(model *postgres.PromptModelType) {
	onceInitPromptService.Do(func() {
		PromptServiceGlobal = &PromptServiceType{
			Model: model,
		}

		mslogger.LoggerGlobal.Info("Global AutosService configurado com sucesso.")
	})
}

func NewPromptService(
	model *postgres.PromptModelType,

) *PromptServiceType {
	return &PromptServiceType{
		Model: model,
	}
}

func (obj *PromptServiceType) GetPromptModel() (*postgres.PromptModelType, error) {
	if obj.Model == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.Model, nil
}

func (obj *PromptServiceType) InsertPrompt(bodyParams postgres.BodyParamsPromptInsert) (*postgres.PromptRow, error) {
	if obj.Model == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.InsertReg(bodyParams)
	if err != nil {
		mslogger.LoggerGlobal.Error("Erro na inclusão de um prompt.")
		return nil, err
	}
	return row, nil
}
func (obj *PromptServiceType) UpdatePrompt(bodyParams postgres.BodyParamsPromptUpdate) (*postgres.PromptRow, error) {
	if obj.Model == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.UpdateReg(bodyParams)
	if err != nil {
		mslogger.LoggerGlobal.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
func (obj *PromptServiceType) DeletaPrompt(id int) (*postgres.PromptRow, error) {
	if obj.Model == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	row, err := obj.Model.DeleteReg(id)
	if err != nil {
		mslogger.LoggerGlobal.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
func (obj *PromptServiceType) SelectById(id int) (*postgres.PromptRow, error) {
	if obj.Model == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Model.SelectById(id)
	if err != nil {
		mslogger.LoggerGlobal.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}

func (obj *PromptServiceType) SelectAll() ([]postgres.PromptRow, error) {
	if obj.Model == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	rows, err := obj.Model.SelectRegs()
	if err != nil {
		mslogger.LoggerGlobal.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return rows, nil
}

func (obj *PromptServiceType) SelectByNatureza(prompt_natureza int) ([]postgres.PromptRow, error) {
	if obj.Model == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	/* Recupero o prompts da tabela promptsModel*/
	//prompts, err := obj.Model.SelectByNatureza(models.PROMPT_NATUREZA_IDENTIFICA)
	prompts, err := obj.Model.SelectByNatureza(prompt_natureza)
	if err != nil {
		return nil, erros.CreateErrorf("Erro ao buscar prompt %d - %v", prompt_natureza, err)
	}
	return prompts, nil
}

func (obj *PromptServiceType) GetPromptByNatureza(prompt_natureza int) (string, error) {
	if obj.Model == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return "", fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	prompt, err := obj.SelectByNatureza(prompt_natureza)
	if err != nil {
		return "", erros.CreateErrorf("Erro ao buscar prompt %d - %v", prompt_natureza, err)
	}
	if len(prompt) == 0 {
		// crie um erro semântico p/ mapear em 404
		mslogger.LoggerGlobal.Errorf("Não foi encontrado um prompt para a seguinte natureza: %d", prompt_natureza)
		return "", fmt.Errorf("não foi encontrado um prompt para a seguinte natureza: %d", prompt_natureza)
	}
	return prompt[0].TxtPrompt, nil
}
