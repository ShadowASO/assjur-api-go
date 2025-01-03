package login

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"ocrserver/auth"
	"ocrserver/models"
)

// Estrutura de resposta para erro
type ResponseStatus struct {
	Ok         bool   `json:"ok"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

/*
Verifica se o refreshToken ainda é válido e caso afirmativo, gera um novo
acsessToken
*/
func RefreshTokenHandler(c *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Printf("token não enviado: %v", err)
		response := ResponseStatus{
			Ok:         false,
			StatusCode: http.StatusBadRequest,
			Message:    "Token não enviado!",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	refreshToken := body.Token
	if refreshToken == "" {
		log.Printf("token não enviado")
		response := ResponseStatus{
			Ok:         false,
			StatusCode: http.StatusBadRequest,
			Message:    "Token não enviado!",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Estrutura para os atributos do usuário
	//user := &auth.UserAtribs{}

	user, err := auth.ValidateToken(refreshToken)
	if err != nil {
		response := ResponseStatus{
			Ok:         false,
			StatusCode: http.StatusUnauthorized,
			Message:    "refreshToken: refreshToken inválido!",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Criação do novo accessToken
	accessToken, err := auth.CreateToken(*user, auth.AccessTokenExpire)
	if err != nil {
		response := ResponseStatus{
			Ok:         false,
			StatusCode: http.StatusInternalServerError,
			Message:    "Erro na criação do token!",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Configura o cabeçalho de Authorization
	c.Header("Authorization", accessToken)

	// Resposta de sucesso
	response := struct {
		Ok          bool   `json:"ok"`
		StatusCode  int    `json:"statusCode"`
		Message     string `json:"message"`
		AccessToken string `json:"access_token"`
	}{
		Ok:          true,
		StatusCode:  http.StatusOK,
		Message:     "Sucesso",
		AccessToken: accessToken,
	}
	c.JSON(http.StatusOK, response)
}
func LoginHandler(c *gin.Context) {
	var body struct {
		User struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"user"`
	}

	// Parse o corpo da requisição para extrair os dados do login
	if err := c.ShouldBindJSON(&body); err != nil {
		response := ResponseStatus{
			Ok:         false,
			StatusCode: http.StatusBadRequest,
			Message:    "Dados inválidos na requisição!",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	login := body.User

	// Busca o usuário no banco de dados (simulado aqui como uma função fictícia)
	usersModel := models.NewUsersModel()
	//userQuery, err := models.UsersModel.SelectUserByName(login.Username)
	userQuery, err := usersModel.SelectUserByName(login.Username)
	if err != nil || userQuery == nil {
		response := ResponseStatus{
			Ok:         false,
			StatusCode: http.StatusNotFound,
			Message:    "Usuário não encontrado!",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	user := auth.UserAtribs{
		//UID:   userQuery.UserId,
		UID:   fmt.Sprintf("%.0d", userQuery.UserId),
		Uname: userQuery.Username,
		Urole: userQuery.Userrole,
	}

	// Verifica a senha
	isMatch := auth.CompararSenhaBcrypt(login.Password, userQuery.Password)
	if !isMatch {
		response := ResponseStatus{
			Ok:         false,
			StatusCode: http.StatusUnauthorized,
			Message:    "Senha inválida!",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Cria os tokens de acesso e renovação
	accessToken, err := auth.CreateToken(user, auth.AccessTokenExpire)
	if err != nil {
		response := ResponseStatus{
			Ok:         false,
			StatusCode: http.StatusInternalServerError,
			Message:    "Erro ao criar o token de acesso!",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	refreshToken, err := auth.CreateToken(user, auth.RefreshTokenExpire)
	if err != nil {
		response := ResponseStatus{
			Ok:         false,
			StatusCode: http.StatusInternalServerError,
			Message:    "Erro ao criar o token de renovação!",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Resposta de sucesso
	response := struct {
		Ok           bool   `json:"ok"`
		StatusCode   int    `json:"statusCode"`
		Message      string `json:"message"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		Ok:           true,
		StatusCode:   http.StatusOK,
		Message:      "Sucesso",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	c.JSON(http.StatusOK, response)
}

func OutLogin(c *gin.Context) {
	c.Header("Set-Cookie", "access_token=; Path=/; HttpOnly; Expires=Thu, 01 Jan 1970 00:00:00 GMT")
	c.JSON(http.StatusOK, gin.H{"message": "Logout bem-sucedido"})
}
