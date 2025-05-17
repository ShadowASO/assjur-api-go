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
	"ocrserver/internal/config"
	"ocrserver/internal/utils/logger"
)

type LoginServiceType struct {
	cfg *config.Config
}

func NewLoginService(cfg *config.Config) *LoginServiceType {
	return &LoginServiceType{
		cfg: cfg,
	}
}

func (obj *LoginServiceType) GetConfig() (*config.Config, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.cfg, nil
}
