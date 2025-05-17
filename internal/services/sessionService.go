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

	"github.com/openai/openai-go"
)

type SessionServiceType struct {
	Model *models.SessionsModelType
}

var SessionServiceGlobal *SessionServiceType
var onceInitSessionService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitSessionService(model *models.SessionsModelType) {
	onceInitSessionService.Do(func() {
		SessionServiceGlobal = &SessionServiceType{
			Model: model,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewSessionService(
	Model *models.SessionsModelType,

) *SessionServiceType {
	return &SessionServiceType{

		Model: Model,
	}
}

func (obj *SessionServiceType) GetSessionModel() (*models.SessionsModelType, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.Model, nil
}
func (obj *SessionServiceType) GetSessionByID(id int) (*models.SessionsRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	rsp, err := obj.Model.SelectSession(id)
	if err != nil {
		logger.Log.Error("erro ao buscar sessão pelo ID")
		return nil, err
	}
	return rsp, nil
}

func (obj *SessionServiceType) UpdateSession(data models.SessionsRow) (*models.SessionsRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	rsp, err := obj.Model.UpdateSession(data)
	if err != nil {
		logger.Log.Error("erro ao buscar sessão pelo ID")
		return nil, err
	}
	return rsp, nil

}

/*
Atualiza os campos relativos ao uso de tokens
*/
func (obj *SessionServiceType) UpdateTokensUso(retSubmit *openai.ChatCompletion) error {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	/* Calcula os valores de tokesn */
	var sessionData models.SessionsRow
	sessionData.SessionID = 1
	sessionData.UserID = 1

	//sessionsModel := models.NewSessionsModel()
	//sessionsModel := server.sessionModel
	//currentTokens, err := sessionsModel.SelectSession(sessionData.SessionID)
	currentTokens, err := obj.GetSessionByID(sessionData.SessionID)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return fmt.Errorf("CnjApi global não configurada")
	}
	sessionData.PromptTokens = retSubmit.Usage.PromptTokens + currentTokens.PromptTokens
	sessionData.CompletionTokens = retSubmit.Usage.CompletionTokens + currentTokens.CompletionTokens
	sessionData.TotalTokens = retSubmit.Usage.TotalTokens + currentTokens.TotalTokens

	//_, err = sessionsModel.UpdateSession(sessionData)
	_, err = obj.UpdateSession(sessionData)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return fmt.Errorf("CnjApi global não configurada")
	}

	return err
}
