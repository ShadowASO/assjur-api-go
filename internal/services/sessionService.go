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

	"ocrserver/internal/models/postgres"
	"ocrserver/internal/utils/mslogger"

	"sync"
)

type SessionServiceType struct {
	Model *postgres.SessionsModelType
}

var SessionServiceGlobal *SessionServiceType
var onceInitSessionService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitSessionService(model *postgres.SessionsModelType) {
	onceInitSessionService.Do(func() {
		SessionServiceGlobal = &SessionServiceType{
			Model: model,
		}

		mslogger.LoggerGlobal.Info("Global AutosService configurado com sucesso.")
	})
}

func NewSessionService(
	Model *postgres.SessionsModelType,

) *SessionServiceType {
	return &SessionServiceType{

		Model: Model,
	}
}

func (obj *SessionServiceType) GetSessionModel() (*postgres.SessionsModelType, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.Model, nil
}
func (obj *SessionServiceType) GetSessionByID(id int) (*postgres.SessionsRow, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	rsp, err := obj.Model.SelectSession(id)
	if err != nil {
		mslogger.LoggerGlobal.Error("erro ao buscar sessão pelo ID")
		return nil, err
	}
	return rsp, nil
}

func (obj *SessionServiceType) UpdateSession(data postgres.SessionsRow) (*postgres.SessionsRow, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	rsp, err := obj.Model.UpdateSession(data)
	if err != nil {
		mslogger.LoggerGlobal.Error("erro ao buscar sessão pelo ID")
		return nil, err
	}
	return rsp, nil

}

/*
Atualiza os campos relativos ao uso de tokens na tabela "sessions"
*/

func (obj *SessionServiceType) UpdateTokensUso(pt int64, ct int64, tt int64) error {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	const SESSIONS_ID = 1

	_, err := obj.Model.IncrementTokensAtomic(SESSIONS_ID, pt, ct, tt)
	if err != nil {
		mslogger.LoggerGlobal.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return fmt.Errorf("CnjApi global não configurada")
	}

	return err
}
