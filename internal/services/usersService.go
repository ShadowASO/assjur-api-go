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
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/mslogger"

	"strconv"
	"sync"
)

type UserServiceType struct {
	model *models.UsersModelType
}

var UserServiceGlobal *UserServiceType
var onceInitUserService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitUsersService(model *models.UsersModelType) {
	onceInitUserService.Do(func() {

		UserServiceGlobal = &UserServiceType{
			model: model,
		}

		mslogger.LoggerGlobal.Info("Global AutosService configurado com sucesso.")
	})
}

type bodyUsers struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewUsersService(modelo *models.UsersModelType) *UserServiceType {
	return &UserServiceType{
		model: modelo,
	}
}
func (obj *UserServiceType) GetModel() (*models.UsersModelType, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.model, nil
}

func (obj *UserServiceType) GetUser(uid string) (*models.UsersRow, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	//userID, err := strconv.ParseInt(uid, 10, 32)
	userID, err := strconv.Atoi(uid)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao fazer o parser do ID do usuário", err)
		return nil, erros.CreateError("ID do usuário inválido")
	}

	user, err := obj.model.SelectRow(userID)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Usuário não encontrado", err)
		return nil, erros.CreateError("Usuário não encontrado")
	}
	return user, nil

}
func (obj *UserServiceType) InsertUser(user models.UsersRow) (int64, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return 0, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	//key, err := h.model.Insert(uname, urole, uemail, upass, usuario)
	key, err := obj.model.InsertRow(user)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Usuário não incluído", err)
		return 0, erros.CreateError("Usuário não incluído")
	}
	return key, err

}
func (obj *UserServiceType) UpdateUser(uid, urole, upass, usuario string) error {
	//userID, err := strconv.ParseInt(uid, 10, 64)
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	userID, err := strconv.Atoi(uid)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao realizar o ParseInt do ID do usuário", err)
		return erros.CreateError("ID do usuário inválido")

	}
	//err = h.model.Update(userID, urole, upass, usuario)
	row, err := obj.model.SelectRow(userID)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Usuário não atualizado", err)
		return erros.CreateError("Usuário não atualizado")
	}
	mslogger.LoggerGlobal.Info("Usuário não atualizado " + row.Username)
	return err

}

func (obj *UserServiceType) ListUsers() ([]models.UsersRow, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	users, err := obj.model.SelectRows()
	if err != nil {

		mslogger.LoggerGlobal.ErrorErr("Erro ao listar todos os usuários", err)
		return nil, erros.CreateError("Erro ao listar todos os usuários")
	}
	return users, nil

}
func (obj *UserServiceType) SelectUserByName(username string) (*models.UsersRow, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	user, err := obj.model.SelectUserByName(username)

	if err != nil || user == nil {
		mslogger.LoggerGlobal.ErrorErr("Usuário não encontrado", err)
		return user, erros.CreateError("Erro ao listar todos os usuários")
	}
	return user, nil
}
