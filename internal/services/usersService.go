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
	"ocrserver/internal/utils/logger"
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

		logger.Log.Info("Global AutosService configurado com sucesso.")
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
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	return obj.model, nil
}

func (obj *UserServiceType) GetUser(uid string) (*models.UsersRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	//userID, err := strconv.ParseInt(uid, 10, 32)
	userID, err := strconv.Atoi(uid)
	if err != nil {
		logger.Log.Error("Erro ao fazer o parser do ID do usuário", err.Error())
		return nil, erros.CreateError("ID do usuário inválido")
	}

	user, err := obj.model.SelectRow(userID)
	if err != nil {
		logger.Log.Error("Usuário não encontrado", err.Error())
		return nil, erros.CreateError("Usuário não encontrado")
	}
	return user, nil

}
func (obj *UserServiceType) InsertUser(user models.UsersRow) (int64, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return 0, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	//key, err := h.model.Insert(uname, urole, uemail, upass, usuario)
	key, err := obj.model.InsertRow(user)
	if err != nil {
		logger.Log.Error("Usuário não incluído", err.Error())
		return 0, erros.CreateError("Usuário não incluído")
	}
	return key, err

}
func (obj *UserServiceType) UpdateUser(uid, urole, upass, usuario string) error {
	//userID, err := strconv.ParseInt(uid, 10, 64)
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	userID, err := strconv.Atoi(uid)
	if err != nil {
		logger.Log.Error("Erro ao realizar o ParseInt do ID do usuário", err.Error())
		return erros.CreateError("ID do usuário inválido")

	}
	//err = h.model.Update(userID, urole, upass, usuario)
	row, err := obj.model.SelectRow(userID)
	if err != nil {
		logger.Log.Error("Usuário não atualizado", err.Error())
		return erros.CreateError("Usuário não atualizado")
	}
	logger.Log.Info("Usuário não atualizado " + row.Username)
	return err

}

// func (h *UserServiceType) DeleteUser(uid string) error {
// 	userID, err := strconv.ParseInt(uid, 10, 64)
// 	if err != nil {
// 		logger.Log.Error("Erro ao fazer o parser do ID do usuário", err.Error())
// 		return erros.CreateError("ID do usuário inválido")
// 	}
// 	err = h.model.Delete(userID)
// 	if err != nil {
// 		logger.Log.Error("Usuário não deletado", err.Error())
// 		return erros.CreateError("Usuário não deletado")

// 	}
// 	return err

// }
func (obj *UserServiceType) ListUsers() ([]models.UsersRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	users, err := obj.model.SelectRows()
	if err != nil {

		logger.Log.Error("Erro ao listar todos os usuários", err.Error())
		return nil, erros.CreateError("Erro ao listar todos os usuários")
	}
	return users, nil

}
func (obj *UserServiceType) SelectUserByName(username string) (*models.UsersRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	user, err := obj.model.SelectUserByName(username)

	if err != nil || user == nil {
		logger.Log.Error("Usuário não encontrado", err.Error())
		return user, erros.CreateError("Erro ao listar todos os usuários")
	}
	return user, nil
}
