package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ocrserver/internal/auth"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"
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
	requestID := middleware.GetRequestID(c)

	var body struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Token == "" {
		response.HandleError(c, http.StatusBadRequest, "token não enviado ou formato inválido", "", requestID)
		return
	}

	claims, err := obj.jwt.ValidateString(body.Token)
	if err != nil {
		logger.Log.Errorf("token inválido: %v", err)
		response.HandleError(c, http.StatusUnauthorized, "token inválido", "", requestID)
		return
	}

	rsp := gin.H{
		"id":    claims.ID,
		"name":  claims.Name,
		"email": claims.Email,
		"role":  claims.Role,
		"exp":   claims.ExpiresAt.Time.Unix(),
	}
	response.HandleSucesso(c, http.StatusOK, rsp, requestID)
}

/*
 * Gera novo access token a partir de um refresh token válido
 * Rota: POST /auth/token/refresh
 * Body: { "token": string }
 */
func (obj *LoginHandlerType) RefreshTokenHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var body struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Token == "" {
		response.HandleError(c, http.StatusBadRequest, "refreshToken não enviado", "", requestID)
		return
	}

	claims, err := obj.jwt.ValidateString(body.Token)
	if err != nil {
		logger.Log.Errorf("refreshToken inválido: %v", err)
		response.HandleError(c, http.StatusUnauthorized, "Token inválido", "", requestID)
		return
	}

	cfg, _ := obj.service.GetConfig()

	accessToken, err := obj.jwt.GenerateToken(
		claims.ID, claims.Name, claims.Email, claims.Role,
		cfg.AccessTokenExpire,
	)
	if err != nil {
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar o Token", "", requestID)
		return
	}

	rsp := gin.H{
		"access_token": accessToken,
	}
	response.HandleSucesso(c, http.StatusOK, rsp, requestID)
}

/*
 * Login: valida usuário/senha e entrega tokens
 * Rota: POST /auth/login
 * Body: { "username": string, "password": string }
 */
func (obj *LoginHandlerType) LoginHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	usr, err := services.UserServiceGlobal.SelectUserByName(body.Username)
	if err != nil || usr == nil {
		response.HandleError(c, http.StatusNotFound, "Usuário incorreto", "", requestID)
		return
	}

	if !auth.CheckPassword(body.Password, usr.Password) {
		response.HandleError(c, http.StatusUnauthorized, "Senha inválida", "", requestID)
		return
	}

	cfg, _ := obj.service.GetConfig()

	// supondo usr.UserId seja int; ajuste se for outro tipo
	uid, _ := strconv.Atoi(strconv.Itoa(usr.UserId))
	email := usr.Email
	if email == "" {
		email = "" // mantenha vazio se não houver no seu schema
	}

	accessToken, err := obj.jwt.GenerateToken(
		uint(uid), usr.Username, email, usr.Userrole, cfg.AccessTokenExpire,
	)
	if err != nil {
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar token", "", requestID)
		return
	}

	refreshToken, err := obj.jwt.GenerateToken(
		uint(uid), usr.Username, email, usr.Userrole, cfg.RefreshTokenExpire,
	)
	if err != nil {
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar Token", "", requestID)
		return
	}

	rsp := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}
	response.HandleSucesso(c, http.StatusCreated, rsp, requestID)
}

/*
 * Logout "stateless": apenas orientativo (se usar cookie HttpOnly, expira aqui)
 * Rota: POST /auth/logout
 */
func (obj *LoginHandlerType) OutLogin(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	// Se você estiver usando cookie de access_token:
	// c.Header("Set-Cookie", "access_token=; Path=/; HttpOnly; Max-Age=0; Expires=Thu, 01 Jan 1970 00:00:00 GMT")

	rsp := gin.H{"message": "Logout bem-sucedido"}
	response.HandleSucesso(c, http.StatusOK, rsp, requestID)
}
