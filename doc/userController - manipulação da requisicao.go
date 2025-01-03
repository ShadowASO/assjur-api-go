// userController.go
// Módulo que concentra as operações relacionadas à tabela 'users'
// Datas Revisão: 06/12/2024.

package usersController

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"ocrserver/auth"

	"ocrserver/models/usersModel"

	"github.com/gin-gonic/gin"
	//"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserRole string `json:"userrole"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validateUser(user User) error {
	if user.UserRole == "" || user.Username == "" || user.Email == "" || user.Password == "" {
		return fmt.Errorf("Dados inválidos")
	}
	return nil
}

func ListUsers(c *gin.Context) {
	users, err := usersModel.Services.SelectRows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao listar usuários"})
		return
	}
	c.JSON(http.StatusOK, users)
}
func SelectUser(c *gin.Context) {
	// Extrai o userID dos parâmetros da rota
	userID1 := c.Param("userID")

	// Converte userID para inteiro
	id1, convErr1 := strconv.Atoi(userID1)
	if convErr1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "userID inválido"})
		return
	}
	log.Printf("userID1 - Params=%v  - id2=%d", userID1, id1)

	// Extrai o userID da query string
	userID2 := c.Query("userID")
	if userID2 == "" {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "userID não fornecido"})
		return
	}

	// Converte userID para inteiro
	id2, convErr2 := strconv.Atoi(userID2)
	if convErr2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "userID inválido"})
		return
	}
	log.Printf("userID2 - Query=%v - id2=%d", userID2, id2)

	// Define uma estrutura para capturar o corpo da requisição
	var requestData struct {
		UserID int `json:"userID"`
	}

	// Extrai os dados do corpo da requisição
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}
	log.Printf("userID2 - Body=%v", requestData.UserID)
	// Usa o userID extraído no modelo para buscar os dados
	//users, err := usersModel.Services.SelectRow(requestData.UserID)
	users, err := usersModel.Services.SelectRow(id1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao listar usuário"})
		return
	}

	// Retorna os dados do usuário
	c.JSON(http.StatusOK, users)
}

func InsertUser(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if err := validateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": err.Error()})
		return
	}

	//hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	hashPassword, err := auth.EncriptarSenhaBcrypt(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao criptografar senha"})
		return
	}
	//user.Password = string(hashPassword)

	var userRow usersModel.UsersRow
	userRow.Userrole = user.UserRole
	userRow.Username = user.Username
	userRow.Password = string(hashPassword)
	userRow.Email = user.Email
	userRow.CreatedAt = time.Now()

	newUser, err := usersModel.Services.InsertRow(userRow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao inserir usuário"})
		return
	}
	c.JSON(http.StatusCreated, newUser)
}

