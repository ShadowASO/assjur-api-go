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
	"ocrserver/internal/models"
	"ocrserver/internal/utils/logger"
)

type QueryServiceType struct {
	Model *models.SessionsModelType
}

func NewQueryService(
	Model *models.SessionsModelType,

) *QueryServiceType {
	return &QueryServiceType{

		Model: Model,
	}
}

func (obj *QueryServiceType) GetModel() (*models.SessionsModelType, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.Model, nil
}
