// userController.go
// Módulo que concentra as operações relacionadas à tabela 'users'
// Datas Revisão: 06/12/2024.

package handlers

import (
	"fmt"
	"net/http"
	"ocrserver/internal/auth"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UsersHandlerType struct {
	Model   *models.UsersModelType
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
		logger.Log.Error("Erro ao ao obter usersModel", err.Error())
		return nil
	}
	return &UsersHandlerType{
		Model:   modelo,
		service: service,
	}
}

func (service *UsersHandlerType) validateUser(user User) error {
	if user.UserRole == "" || user.Username == "" || user.Email == "" || user.Password == "" {
		return fmt.Errorf("dados inválidos")
	}
	return nil
}

/*
 * Inclui um novo usuário
 *
 * - **Rota**: "/users"
 * - **Params**:
 * - **Método**: POST
 * - **Status**: 201/400/500,
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
	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------
	user := User{}

	if err := c.ShouldBindJSON(&user); err != nil {

		logger.Log.Errorf("Dados de usuário inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Dados de usuário inválidos: ", "", requestID)
		return
	}

	if err := service.validateUser(user); err != nil {

		logger.Log.Errorf("Dados de usuário inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Dados de usuário inválidos: ", "", requestID)
		return
	}

	hashPassword, err := auth.HashPassword(user.Password)
	if err != nil {

		logger.Log.Errorf("Erro ao criptografar senha do usuário: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao criptografar senha do usuário! ", "", requestID)
		return
	}

	var userRow models.UsersRow
	userRow.Userrole = user.UserRole
	userRow.Username = user.Username
	userRow.Password = string(hashPassword)
	userRow.Email = user.Email
	userRow.CreatedAt = time.Now()

	newUser, err := service.Model.InsertRow(userRow)
	if err != nil {

		logger.Log.Errorf("Erro ao inserir o usuário: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao inserir o usuário! ", "", requestID)
		return
	}

	rsp := gin.H{
		"message": "Usuário incluído com sucesso",
		"userID":  int(newUser),
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/*
 * Lista todos os usuários cadastrados
 *
 * - **Rota**: "/users"
 * - **Params**:
 * - **Método**: GET
 * - **Body**:
 * - **Resposta**:
 *  	[{
 * 			"UserId": 1,
 *   		"Userrole": string,
 *   		"Username": string,
 *   		"Password": string,
 *   		"Email": string,
 *   		"CreatedAt": Date
 *		}]
 */
func (service *UsersHandlerType) SelectAllHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	users, err := service.Model.SelectRows()
	if err != nil {

		logger.Log.Errorf("Usuários não encontrados: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Usuários não encontrados!", "", requestID)
		return
	}

	rsp := gin.H{
		"rows": users,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/*
 * Devolve os dados do usuário indicado no parâmetro da rota
 *
 * - **Rota**: "/users/:id"
 * - **Params**:
 * - **Método**: GET
 * - **Status**: 200/204/400
 * - **Body**:
 * - **Resposta**:
 *  	[{
 * 			"UserId": 1,
 *   		"Userrole": string,
 *   		"Username": string,
 *   		"Password": string,
 *   		"Email": string,
 *   		"CreatedAt": Date
 *		}]
 */

func (service *UsersHandlerType) SelectHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------
	// Extrai o parâmetro id da rota
	userID := c.Param("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		logger.Log.Errorf("ID de usuário inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "ID de usuário inválido", "", requestID)
		return
	}

	user, err := service.Model.SelectRow(id)
	if err != nil {

		logger.Log.Errorf("Usuário não encontrado: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Usuário não encontrado!", "", requestID)
		return
	}

	// Retorna os dados do usuário

	rsp := gin.H{
		"row": user,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
