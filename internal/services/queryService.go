/*
---------------------------------------------------------------------------------------
File: queryService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 17-05-2025
---------------------------------------------------------------------------------------
*/
package services

import (
	"fmt"

	"ocrserver/internal/models/postgres"
	"ocrserver/internal/utils/mslogger"
)

type QueryServiceType struct {
	Model *postgres.SessionsModelType
}

func NewQueryService(
	Model *postgres.SessionsModelType,

) *QueryServiceType {
	return &QueryServiceType{

		Model: Model,
	}
}

func (obj *QueryServiceType) GetModel() (*postgres.SessionsModelType, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.Model, nil
}
