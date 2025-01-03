// userController.go
// Módulo que concentra as operações relacionadas à tabela 'users'
// Datas Revisão: 06/12/2024.

package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"ocrserver/auth"
	"ocrserver/models"
	"strconv"
	"time"
)

type UsersControllerType struct {
	usersModel *models.UsersModelType
}

type User struct {
	UserRole string `json:"userrole"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUsersController() *UsersControllerType {
	model := models.NewUsersModel()
	return &UsersControllerType{usersModel: model}
}

func (service *UsersControllerType) validateUser(user User) error {
	if user.UserRole == "" || user.Username == "" || user.Email == "" || user.Password == "" {
		return fmt.Errorf("dados inválidos")
	}
	return nil
}

func (service *UsersControllerType) SelectAllHandler(c *gin.Context) {

	users, err := service.usersModel.SelectRows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao listar usuários"})
		return
	}
	c.JSON(http.StatusOK, users)
}
func (service *UsersControllerType) SelectHandler(c *gin.Context) {
	// Extrai o parâmetro id da rota
	userID := c.Param("id")

	// Converte id para inteiro
	id, convErr := strconv.Atoi(userID)
	if convErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "userID inválido"})
		return
	}
	log.Printf("id - Params=%v", userID)

	users, err := service.usersModel.SelectRow(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao listar usuário"})
		return
	}

	// Retorna os dados do usuário
	c.JSON(http.StatusOK, users)
}

func (service *UsersControllerType) InsertHandler(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if err := service.validateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": err.Error()})
		return
	}

	hashPassword, err := auth.EncriptarSenhaBcrypt(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao criptografar senha"})
		return
	}

	var userRow models.UsersRow
	userRow.Userrole = user.UserRole
	userRow.Username = user.Username
	userRow.Password = string(hashPassword)
	userRow.Email = user.Email
	userRow.CreatedAt = time.Now()

	newUser, err := service.usersModel.InsertRow(userRow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao inserir usuário"})
		return
	}
	c.JSON(http.StatusCreated, newUser)
}
