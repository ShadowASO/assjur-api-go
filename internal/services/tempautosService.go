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
	"encoding/json"
	"fmt"
	"ocrserver/internal/models"
	"ocrserver/internal/utils/logger"
	"sync"
	"time"
)

type RegKeys struct {
	IdContexto int
	IdDoc      int
}

type TempautosServiceType struct {
	autosModel     *models.AutosModelType
	promptModel    *models.PromptModelType
	tempautosModel *models.TempautosModelType
}

// Estrutura base para o JSON
type DocumentoBase struct {
	Tipo *struct {
		Key         int    `json:"key"`
		Description string `json:"description"`
	} `json:"tipo"`
	Processo string `json:"processo"`
	IdPje    string `json:"id_pje"`
}

var TempautosServiceGlobal *TempautosServiceType
var onceInitTempautosService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitTempautosService(autosModel *models.AutosModelType,
	promptModel *models.PromptModelType,
	tempautosModel *models.TempautosModelType) {
	onceInitAutosService.Do(func() {

		TempautosServiceGlobal = &TempautosServiceType{
			autosModel:     autosModel,
			promptModel:    promptModel,
			tempautosModel: tempautosModel,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewTempautosService(
	autosModel *models.AutosModelType,
	promptModel *models.PromptModelType,
	tempautosModel *models.TempautosModelType) *TempautosServiceType {
	return &TempautosServiceType{
		autosModel:     autosModel,
		promptModel:    promptModel,
		tempautosModel: tempautosModel,
	}
}

func (obj *TempautosServiceType) GetPromptModel() (*models.TempautosModelType, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.tempautosModel, nil
}

func (obj *TempautosServiceType) ProcessarDocumento(reg RegKeys) error {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	msg := fmt.Sprintf("Processando documento: IdDoc=%d, IdContexto=%d", reg.IdDoc, reg.IdContexto)
	logger.Log.Info(msg)

	//REcupero o registro da tabela temp_autos
	dataTempautos, err := obj.tempautosModel.SelectByIdDoc(reg.IdDoc)
	if err != nil {

		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%d - IdContexto=%d", reg.IdDoc, reg.IdContexto)
	}

	/* Recupero o prompt da tabela promptsModel*/
	dataPrompt, err := obj.promptModel.SelectByNatureza(models.PROMPT_NATUREZA_IDENTIFICA)
	if err != nil {

		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%d - IdContexto=%d", reg.IdDoc, reg.IdContexto)
	}
	//var messages openAI.MsgGpt
	var messages MsgGpt
	messages.CreateMessage("user", dataTempautos.TxtDoc)
	messages.CreateMessage("user", dataPrompt.TxtPrompt)

	//retSubmit, err := openAI.OpenAIServiceGlobal.SubmitPrompt(messages)
	retSubmit, err := OpenaiServiceGlobal.SubmitPrompt(messages)
	if err != nil {
		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%d - IdContexto=%d", reg.IdDoc, reg.IdContexto)
	}

	/* Atualiza o uso de tokens na tabela 'sessions' */
	// sessionService := NewSessionsHandlers()
	// err = sessionService.UpdateTokensUso(retSubmit)
	// if err != nil {
	// 	return fmt.Errorf("ERROR: Erro na atualização do uso de tokens")
	// }

	/* Verifico se a resposta é um json válido*/
	rspJson := retSubmit.Choices[0].Message.Content

	var objJson = DocumentoBase{}
	err = json.Unmarshal([]byte(rspJson), &objJson)
	if err != nil {

		return fmt.Errorf("ERROR: Erro ao fazer o parse do JSON")
	}

	//isAutuado, err := service.autosModel.IsDocAutuado(context.Background(), reg.IdContexto, objJson.IdPje)
	isAutuado, err := obj.autosModel.IsDocAutuado(reg.IdContexto, objJson.IdPje)
	if err != nil {
		return fmt.Errorf("ERROR: Erro ao verificar se documento já existe")

	}
	if isAutuado {
		return fmt.Errorf("ERROR: Documento já existe na tabela autosModel")
	}

	//Faz a inclusão do documentos na tabela autos
	autosParams := models.AutosRow{}
	autosParams.IdCtxt = reg.IdContexto
	autosParams.IdNat = objJson.Tipo.Key
	autosParams.IdPje = objJson.IdPje

	autosParams.AutosJson = json.RawMessage(rspJson) // Suporte para JSON nativo no Go

	autosParams.DtInc = time.Now()
	autosParams.Status = "S"

	_, err = obj.autosModel.InsertRow(autosParams)
	if err != nil {

		return fmt.Errorf("ERROR: Erro na inclusão do registro na tabela autosModel")

	}

	//Faz a deleção do registro na tabela temp_autos
	err = obj.tempautosModel.DeleteRow(dataTempautos.IdDoc)
	if err != nil {

		return fmt.Errorf("ERROR: Erro ao deletar registro na tabela temp_autos")
	}

	msg = "Concluído com sucesso!"
	logger.Log.Info(msg)
	return nil

}
