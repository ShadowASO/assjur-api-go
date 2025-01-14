package login

import (
	"fmt"
	"log"
	"net/http"
	"ocrserver/internal/auth"

	"ocrserver/internal/utils/msgs"
	"ocrserver/models"

	"github.com/gin-gonic/gin"
)

/*
 * Verifica se o acessToken ainda é válido
 *
 * - **Rota**: "/auth/token/verify"
 * - **Params**:
 * - **Método**: POST
 * - **Status**: 200/401
 * - **Body:
 *		{
 * 			"token": string
 * 		}
 * - **Resposta**:
 *  	{
 * 			"message": string,
 * 		}
 */
func VerifyTokenHandler(c *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Printf("token não enviado: %v", err)
		response := msgs.CreateResponseMessage("informações de token inválidas!")
		c.JSON(http.StatusBadRequest, response)
	}

	bodyParamToken := body.Token
	if bodyParamToken == "" {
		log.Printf("token não enviado")

		response := msgs.CreateResponseMessage("token não enviado!")
		c.JSON(http.StatusBadRequest, response)
	}

	_, err = auth.ValidateToken(bodyParamToken)
	if err != nil {

		response := msgs.CreateResponseMessage("token inválido!")
		c.JSON(http.StatusUnauthorized, response)
	}

	response := msgs.CreateResponseMessage("token válido!")
	c.JSON(http.StatusOK, response)
}

/*
 * Verifica se o refreshToken é valido e caso positivo, gera um novo acessToken.
 *
 * - **Rota**: "/auth/token/refresh"
 * - **Params**:
 * - **Método**: POST
 * - **Status**: 200/401/500
 * - **Body:
 *		{
 * 			"token": string
 * 		}
 * - **Resposta**:
 *  	{
 * 			"AccessToken": string
 *		}
 */
func RefreshTokenHandler(c *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Printf("token não enviado: %v", err)

		response := msgs.CreateResponseMessage("Dados inválidos no token!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	refreshToken := body.Token
	if refreshToken == "" {
		log.Printf("token não enviado")

		response := msgs.CreateResponseMessage("Token não enviado!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Estrutura para os atributos do usuário

	user, err := auth.ValidateToken(refreshToken)
	if err != nil {

		log.Printf("refreshToken inválido/vencido!")
		response := msgs.CreateResponseMessage("Token inválido/vencido!")
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Criação do novo accessToken
	accessToken, err := auth.CreateToken(*user, auth.AccessTokenExpire)
	if err != nil {

		log.Printf("acessToken inválido/vencido!")
		response := msgs.CreateResponseMessage("Token inválido/vencido!")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Configura o cabeçalho de Authorization
	c.Header("Authorization", accessToken)

	response := gin.H{
		"AccessToken": accessToken,
	}
	c.JSON(http.StatusOK, response)
}

/*
 * Faz o login no sistema, validando o nome e senha do usuário,
 * gerando e devolvendo um accessToken e um refreshToken
 *
 * - **Rota**: "/auth/login"
 * - **Params**:
 * - **Método**: POST
 * - **Status**: 200/401
 * - **Body:
 *		{
 *    		"username": string,
 *   		"password": string"
 *		}
 * - **Resposta**:
 *  	{
 * 			"AccessToken": string
 * 			"RefreshToken": string
*		}
*/
func LoginHandler(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response := msgs.CreateResponseMessage("Dados inválidos na requisição!")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	login := body

	/* Verifica se o usuário está cadastrado no banco de dados */
	usersModel := models.NewUsersModel()

	userQuery, err := usersModel.SelectUserByName(login.Username)
	if err != nil || userQuery == nil {

		response := msgs.CreateResponseMessage("Usuário não cadastrado!")
		c.JSON(http.StatusNotFound, response)
		return
	}

	user := auth.UserAtribs{

		UID:   fmt.Sprintf("%.0d", userQuery.UserId),
		Uname: userQuery.Username,
		Urole: userQuery.Userrole,
	}

	// Verifica a senha
	isMatch := auth.CompararSenhaBcrypt(login.Password, userQuery.Password)
	if !isMatch {

		response := msgs.CreateResponseMessage("Senha inválida!")
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Cria os tokens de acesso e renovação
	accessToken, err := auth.CreateToken(user, auth.AccessTokenExpire)
	if err != nil {

		response := msgs.CreateResponseMessage("Erro na criação do acessToken!")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	refreshToken, err := auth.CreateToken(user, auth.RefreshTokenExpire)
	if err != nil {

		response := msgs.CreateResponseMessage("Erro na criação do refreshToken!")
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := gin.H{
		"AccessToken":  accessToken,
		"RefreshToken": refreshToken,
	}
	c.JSON(http.StatusOK, response)
}

func OutLogin(c *gin.Context) {
	c.Header("Set-Cookie", "access_token=; Path=/; HttpOnly; Expires=Thu, 01 Jan 1970 00:00:00 GMT")
	c.JSON(http.StatusOK, gin.H{"message": "Logout bem-sucedido"})
}
