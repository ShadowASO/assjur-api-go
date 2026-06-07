// userController.go
// Módulo que concentra as operações relacionadas à tabela 'users'
// Datas Revisão: 06/12/2024.

package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ocrserver/internal/models/postgres"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/auth"
	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type UsersHandlerType struct {
	Model   *postgres.UsersModelType
	service *services.UserServiceType
}

type User struct {
	UserRole string `json:"userrole"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUsersHandlers(service *services.UserServiceType) *UsersHandlerType {
	modelo, err := service.GetModel()
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao obter usersModel", err)
		return nil
	}

	return &UsersHandlerType{
		Model:   modelo,
		service: service,
	}
}

func (service *UsersHandlerType) validateUser(user User) error {
	user.UserRole = strings.TrimSpace(user.UserRole)
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	user.Password = strings.TrimSpace(user.Password)

	if user.UserRole == "" || user.Username == "" || user.Email == "" || user.Password == "" {
		return fmt.Errorf("os campos userrole, username, email e password são obrigatórios")
	}

	return nil
}

/*
 * Inclui um novo usuário
 *
 * - **Rota**: "/users"
 * - **Método**: POST
 * - **Status**: 201/400/500
 * - **Body:
 *		{
 * 			"userrole": string
 *    		"username": string
 *    		"email": string
 *    		"password": string
 * 		}
 * - **Resposta**:
 *  	{
 * 			"userID": int
 *		}
 */
func (service *UsersHandlerType) InsertHandler(c *gin.Context) {
	user := User{}

	if err := c.ShouldBindJSON(&user); err != nil {
		mslogger.LoggerGlobal.Errorf("Dados de usuário inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Dados de usuário inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	user.UserRole = strings.TrimSpace(user.UserRole)
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	user.Password = strings.TrimSpace(user.Password)

	if err := service.validateUser(user); err != nil {
		mslogger.LoggerGlobal.Errorf("Dados de usuário inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Dados de usuário inválidos",
			msresponse.ErrorValidacao,
			err.Error(),
		)
		return
	}

	hashPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao criptografar senha do usuário: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao criptografar senha do usuário",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	userRow := postgres.UsersRow{
		Userrole:  user.UserRole,
		Username:  user.Username,
		Password:  string(hashPassword),
		Email:     user.Email,
		CreatedAt: time.Now(),
	}

	newUser, err := service.Model.InsertRow(userRow)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao inserir o usuário: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao inserir o usuário",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"user_id": int(newUser),
	}

	msresponse.OK(c, http.StatusCreated, "Usuário incluído com sucesso", rsp)
}

/*
 * Lista todos os usuários cadastrados
 *
 * - **Rota**: "/users"
 * - **Método**: GET
 */
func (service *UsersHandlerType) SelectAllHandler(c *gin.Context) {
	users, err := service.Model.SelectRows()
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Usuários não encontrados: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao selecionar usuários",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"rows": users,
	}

	msresponse.OK(c, http.StatusOK, "Usuários selecionados com sucesso", rsp)
}

/*
 * Devolve os dados do usuário indicado no parâmetro da rota
 *
 * - **Rota**: "/users/:id"
 * - **Método**: GET
 * - **Status**: 200/400/404/500
 */
func (service *UsersHandlerType) SelectHandler(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))

	if userID == "" {
		mslogger.LoggerGlobal.Error("ID de usuário não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID de usuário não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("ID de usuário inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID de usuário inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	user, err := service.Model.SelectRow(id)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Usuário não encontrado: %v", err)

		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Usuário não encontrado",
			msresponse.ErrorNaoEncontrado,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"row": user,
	}

	msresponse.OK(c, http.StatusOK, "Usuário selecionado com sucesso", rsp)
}
