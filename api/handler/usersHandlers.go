// userController.go
// Módulo que concentra as operações relacionadas à tabela 'users'
// Datas Revisão: 06/12/2024.

package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"net/http"
	"ocrserver/internal/auth"
	"ocrserver/internal/utils/msgs"
	"ocrserver/models"
	"strconv"
	"time"
)

type UsersHandlerType struct {
	usersModel *models.UsersModelType
}
type User struct {
	UserRole string `json:"userrole"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUsersHandlers() *UsersHandlerType {
	model := models.NewUsersModel()
	return &UsersHandlerType{usersModel: model}
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
	user := User{}

	if err := c.ShouldBindJSON(&user); err != nil {
		// log.Printf("user=%v", user)
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		// return
		response := msgs.CreateResponseMessage("Dados de usuário inválidos!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := service.validateUser(user); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": err.Error()})
		// return
		response := msgs.CreateResponseMessage("Dados de usuário inválidos!" + err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	hashPassword, err := auth.EncriptarSenhaBcrypt(user.Password)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao criptografar senha"})
		// return
		response := msgs.CreateResponseMessage("Erro ao criptografar senha do usuário!" + err.Error())
		c.JSON(http.StatusInternalServerError, response)
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
		// c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao inserir usuário"})
		// return
		response := msgs.CreateResponseMessage("Erro ao inserir o usuário!" + err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := msgs.CreateResponseUserInsert(true, http.StatusCreated, "Usuário incluído com sucesso", int(newUser))
	c.JSON(http.StatusCreated, response)
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

	users, err := service.usersModel.SelectRows()
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao listar usuários"})
		// return
		response := msgs.CreateResponseMessage("Usuários não encontrados!" + err.Error())
		c.JSON(http.StatusNoContent, response)
		return
	}
	c.JSON(http.StatusOK, users)
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
	// Extrai o parâmetro id da rota
	userID := c.Param("id")

	// Converte id para inteiro
	id, convErr := strconv.Atoi(userID)
	if convErr != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": "id do usuário inválido"})
		// return
		response := msgs.CreateResponseMessage("ID de usuário inválidos!" + convErr.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	users, err := service.usersModel.SelectRow(id)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao selecionar usuário"})
		// return
		response := msgs.CreateResponseMessage("Usuário não encontrado!" + err.Error())
		c.JSON(http.StatusNoContent, response)
		return
	}

	// Retorna os dados do usuário
	c.JSON(http.StatusOK, users)
}
