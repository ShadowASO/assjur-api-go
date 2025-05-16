package login

import (
	"fmt"

	"net/http"
	"ocrserver/api/handler/response"
	"ocrserver/internal/auth"

	"ocrserver/internal/utils/logger"

	"ocrserver/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
 * 			"UserId":   user.UID,
 *			"Username": user.Uname,
 *			"Urole":    user.Urole,
 * 		}
 */
func VerifyTokenHandler(c *gin.Context) {
	requestID := uuid.New().String()
	var body struct {
		Token string `json:"token"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		logger.Log.Error("JSON com Formato inválido", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	bodyParamToken := body.Token
	if bodyParamToken == "" {
		logger.Log.Error("token não enviado no body")
		response.HandleError(c, http.StatusBadRequest, "token não enviado", "", requestID)
		return
	}

	user, err := auth.ValidateToken(bodyParamToken)
	if err != nil {
		logger.Log.Error("token inválido!")
		response.HandleError(c, http.StatusUnauthorized, "token inválido", "", requestID)
		return
	}

	rsp := gin.H{
		"user_id":  user.UID,
		"username": user.Uname,
		"userrole": user.Urole,
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
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
	requestID := uuid.New().String()

	var body struct {
		Token string `json:"token"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		logger.Log.Error("JSON com Formato inválido", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Formato inválido", err.Error(), requestID)
		return
	}

	refreshToken := body.Token
	if refreshToken == "" {
		logger.Log.Error("refreshToken não enviado")
		response.HandleError(c, http.StatusBadRequest, "refreshToken não enviado", "", requestID)
		return
	}

	// Estrutura para os atributos do usuário
	user, err := auth.ValidateToken(refreshToken)
	if err != nil {
		logger.Log.Error("refreshToken vencido ou inválido!")
		response.HandleError(c, http.StatusUnauthorized, "Token inválido", "", requestID)
		return
	}

	// Criação do novo accessToken
	accessToken, err := auth.CreateToken(*user, auth.AccessTokenExpire)
	if err != nil {
		logger.Log.Error("erro ao gerar o accessToken!")
		response.HandleError(c, http.StatusUnauthorized, "Erro ao gerar o Token", "", requestID)
		return
	}

	// Configura o cabeçalho de Authorization
	//c.Header("Authorization", accessToken)

	rsp := gin.H{
		"access_token": accessToken,
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
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
	requestID := uuid.New().String()
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Error("JSON com Formato inválido", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Formato inválido", err.Error(), requestID)
		return
	}
	login := body

	/* Verifica se o usuário está cadastrado no banco de dados */
	usersModel := models.NewUsersModel()

	userQuery, err := usersModel.SelectUserByName(login.Username)
	if err != nil || userQuery == nil {
		logger.Log.Error("Usuário não encontrado", err.Error())
		response.HandleError(c, http.StatusNotFound, "Usuário incorreto", "", requestID)
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
		logger.Log.Error("Senha inválida!")
		response.HandleError(c, http.StatusUnauthorized, "Senha inválida", "", requestID)
		return
	}

	// Cria os tokens de acesso e renovação
	accessToken, err := auth.CreateToken(user, auth.AccessTokenExpire)
	if err != nil {
		logger.Log.Error("Erro ao gerar o token", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar o token", "", requestID)
		return
	}

	refreshToken, err := auth.CreateToken(user, auth.RefreshTokenExpire)
	if err != nil {
		logger.Log.Error("Erro ao gerar um novo refreshToken!", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar o Token", "", requestID)
		return
	}

	rsp := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)

}

func OutLogin(c *gin.Context) {
	requestID := uuid.New().String()
	c.Header("Set-Cookie", "access_token=; Path=/; HttpOnly; Expires=Thu, 01 Jan 1970 00:00:00 GMT")

	rsp := gin.H{
		"message": "Logout bem-sucedido",
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}
