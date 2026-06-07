package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ocrserver/internal/services"

	"ocrserver/internal/utils/auth"
	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"
)

type LoginHandlerType struct {
	service *services.LoginServiceType
	jwt     *auth.JWTService
}

func NewLoginHandlers(service *services.LoginServiceType, jwt *auth.JWTService) *LoginHandlerType {
	return &LoginHandlerType{
		service: service,
		jwt:     jwt,
	}
}

/*
 * Verifica se o access token ainda é válido
 * Rota: POST /auth/token/verify
 * Body: { "token": string }
 */
func (obj *LoginHandlerType) VerifyTokenHandler(c *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	body.Token = strings.TrimSpace(body.Token)

	if body.Token == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Token não enviado",
			msresponse.ErrorTokenInvalido,
			"O campo token é obrigatório.",
		)
		return
	}

	claims, err := obj.jwt.ValidateString(body.Token)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Token inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Token inválido",
			msresponse.ErrorTokenInvalido,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"id":    claims.ID,
		"name":  claims.Name,
		"email": claims.Email,
		"role":  claims.Role,
		"exp":   claims.ExpiresAt.Time.Unix(),
	}

	msresponse.OK(c, http.StatusOK, "Token válido", rsp)
}

/*
 * Gera novo access token a partir de um refresh token válido
 * Rota: POST /auth/token/refresh
 * Body: { "token": string }
 */
func (obj *LoginHandlerType) RefreshTokenHandler(c *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	body.Token = strings.TrimSpace(body.Token)

	if body.Token == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Refresh token não enviado",
			msresponse.ErrorTokenInvalido,
			"O campo token é obrigatório.",
		)
		return
	}

	claims, err := obj.jwt.ValidateString(body.Token)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Refresh token inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Refresh token inválido",
			msresponse.ErrorTokenInvalido,
			err.Error(),
		)
		return
	}

	cfg, err := obj.service.GetConfig()
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao carregar configuração de autenticação: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao carregar configuração de autenticação",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	accessToken, err := obj.jwt.GenerateToken(
		claims.ID,
		claims.Name,
		claims.Email,
		claims.Role,
		cfg.AccessTokenExpire,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar access token: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao gerar token",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"access_token": accessToken,
	}

	msresponse.OK(c, http.StatusOK, "Token renovado com sucesso", rsp)
}

/*
 * Login: valida usuário/senha e entrega tokens
 * Rota: POST /auth/login
 * Body: { "username": string, "password": string }
 */
func (obj *LoginHandlerType) LoginHandler(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	body.Username = strings.TrimSpace(body.Username)
	body.Password = strings.TrimSpace(body.Password)

	if body.Username == "" || body.Password == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Usuário e senha são obrigatórios",
			msresponse.ErrorValidacao,
			"Os campos username e password são obrigatórios.",
		)
		return
	}

	usr, err := services.UserServiceGlobal.SelectUserByName(body.Username)
	if err != nil || usr == nil {
		if err != nil {
			mslogger.LoggerGlobal.Errorf("Usuário não localizado: %v", err)
		}

		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Usuário ou senha inválidos",
			msresponse.ErrorNaoAutorizado,
			"As credenciais informadas são inválidas.",
		)
		return
	}

	if !auth.CheckPassword(body.Password, usr.Password) {
		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Usuário ou senha inválidos",
			msresponse.ErrorNaoAutorizado,
			"As credenciais informadas são inválidas.",
		)
		return
	}

	cfg, err := obj.service.GetConfig()
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao carregar configuração de autenticação: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao carregar configuração de autenticação",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	uid, err := strconv.Atoi(strconv.Itoa(usr.UserId))
	if err != nil {
		mslogger.LoggerGlobal.Errorf("ID de usuário inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"ID de usuário inválido",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	email := strings.TrimSpace(usr.Email)

	accessToken, err := obj.jwt.GenerateToken(
		uint(uid),
		usr.Username,
		email,
		usr.Userrole,
		cfg.AccessTokenExpire,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar access token: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao gerar token",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	refreshToken, err := obj.jwt.GenerateToken(
		uint(uid),
		usr.Username,
		email,
		usr.Userrole,
		cfg.RefreshTokenExpire,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar refresh token: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao gerar refresh token",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	msresponse.OK(c, http.StatusCreated, "Login realizado com sucesso", rsp)
}

/*
 * Logout "stateless": apenas orientativo
 * Rota: POST /auth/logout
 */
func (obj *LoginHandlerType) OutLogin(c *gin.Context) {
	// Se você estiver usando cookie de access_token:
	// c.Header("Set-Cookie", "access_token=; Path=/; HttpOnly; Max-Age=0; Expires=Thu, 01 Jan 1970 00:00:00 GMT")

	msresponse.OK(c, http.StatusOK, "Logout bem-sucedido")
}
