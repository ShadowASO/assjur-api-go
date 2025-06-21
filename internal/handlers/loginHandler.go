package handlers

import (
	"strconv"

	"net/http"

	"ocrserver/internal/auth"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

type LoginHandlerType struct {
	service *services.LoginServiceType
}

func NewLoginHandlers(service *services.LoginServiceType) *LoginHandlerType {

	return &LoginHandlerType{
		service: service,
	}
}

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
func (obj *LoginHandlerType) VerifyTokenHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	var body struct {
		Token string `json:"token"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		logger.Log.Errorf("JSON com Formato inválido: %v", err)
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
		logger.Log.Errorf("token inválido: %v", err)
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
func (obj *LoginHandlerType) RefreshTokenHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	var body struct {
		Token string `json:"token"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		logger.Log.Errorf("JSON com Formato inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
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
		logger.Log.Errorf("refreshToken vencido ou inválido: %v", err)
		response.HandleError(c, http.StatusUnauthorized, "Token inválido", "", requestID)
		return
	}

	// Criação do novo accessToken
	cfg, _ := obj.service.GetConfig()
	accessToken, err := auth.CreateToken(*user, cfg.AccessTokenExpire)
	logger.Log.Info(`cfg.AccessTokenExpire=` + cfg.AccessTokenExpire.String())
	if err != nil {
		logger.Log.Error("erro ao gerar o accessToken!")
		response.HandleError(c, http.StatusUnauthorized, "Erro ao gerar o Token", "", requestID)
		return
	}

	// Configura o cabeçalho de Authorization

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
func (obj *LoginHandlerType) LoginHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Errorf("JSON com Formato inválido: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Formato inválido", "", requestID)
		return
	}
	login := body

	/* Verifica se o usuário está cadastrado no banco de dados */

	userQuery, err := services.UserServiceGlobal.SelectUserByName(login.Username)
	if err != nil || userQuery == nil {
		logger.Log.Errorf("Usuário não encontrado: %v", err)
		response.HandleError(c, http.StatusNotFound, "Usuário incorreto", "", requestID)
		return
	}

	user := auth.UserAtribs{
		//UID:   fmt.Sprintf("%.0d", userQuery.UserId),
		UID:   strconv.Itoa(userQuery.UserId),
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
	cfg, _ := obj.service.GetConfig()

	accessToken, err := auth.CreateToken(user, cfg.AccessTokenExpire)
	logger.Log.Info(`cfg.AccessTokenExpire=` + cfg.AccessTokenExpire.String())
	if err != nil {
		logger.Log.Errorf("Erro ao gerar o token: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar o token", "", requestID)
		return
	}

	refreshToken, err := auth.CreateToken(user, cfg.RefreshTokenExpire)
	logger.Log.Info(`cfg.RefreshTokenExpire=` + cfg.RefreshTokenExpire.String())
	if err != nil {
		logger.Log.Errorf("Erro ao gerar um novo refreshToken: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar o Token", "", requestID)
		return
	}

	rsp := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)

}

func (obj *LoginHandlerType) OutLogin(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	c.Header("Set-Cookie", "access_token=; Path=/; HttpOnly;  Max-Age=0; Expires=Thu, 01 Jan 1970 00:00:00 GMT")

	rsp := gin.H{
		"message": "Logout bem-sucedido",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
