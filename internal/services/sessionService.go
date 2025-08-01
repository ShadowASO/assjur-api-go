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
Atualiza os campos relativos ao uso de tokens na tabela "sessions"
*/

func (obj *SessionServiceType) UpdateTokensUso(pt int64, ct int64, tt int64) error {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	const SESSIONS_ID = 1

	_, err := obj.Model.IncrementTokensAtomic(SESSIONS_ID, pt, ct, tt)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return fmt.Errorf("CnjApi global não configurada")
	}

	return err
}
