/*
---------------------------------------------------------------------------------------
File: jwt.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
Alteração: 11-08-2025
---------------------------------------------------------------------------------------
*/
package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"slices"

	"ocrserver/internal/config"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/utils/logger"
)

/*
=========================

	Claims padronizadas
	=========================
*/
// type Claims struct {
// 	ID    uint   `json:"id"`
// 	Email string `json:"email"`
// 	Role  string `json:"role"`
// 	Name  string `json:"name"`
// 	jwt.RegisteredClaims
// }

type Claims struct {
	ID    uint   `json:"user_id"`
	Email string `json:"user_email"`
	Role  string `json:"user_role"`
	Name  string `json:"user_name"`
	jwt.RegisteredClaims
}

/*
=========================

	Serviço JWT

=========================
*/
type JWTService struct {
	secretKey []byte
	issuer    string
	leeway    time.Duration
}

func NewJWTService(cfg config.Config) *JWTService {
	return &JWTService{
		secretKey: []byte(cfg.JWTSecretKey),
		issuer:    "assjur",         // ajuste se quiser: cfg.AppName
		leeway:    30 * time.Second, // tolerância para clock skew
	}
}

func (j *JWTService) GenerateToken(id uint, name, email, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		ID:    id,
		Email: email,
		Role:  role,
		Name:  name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-j.leeway)),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(j.secretKey)
}

// Valida o token
func (j *JWTService) ParseToken(tokenString string) (*Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithLeeway(j.leeway),
	)
	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}
	if claims.Issuer != j.issuer {
		return nil, errors.New("issuer inválido")
	}
	return claims, nil
}

// ValidateString é um alias conveniente para endpoints que recebem o token no body
func (j *JWTService) ValidateString(token string) (*Claims, error) {
	return j.ParseToken(token)
}

/*
=========================

	Helpers HTTP

=========================
*/
func ExtractBearerToken(authHeader string) (string, error) {
	parts := strings.Fields(authHeader)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1], nil
	}
	return "", errors.New("authorization header inválido (esperado: Bearer <token>)")
}

/*
=========================

	Middlewares
=========================
*/
//MiddleWare para validar a autenticação de usuário de uma requisição http
func (j *JWTService) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		h := c.GetHeader("Authorization")
		if h == "" {
			response.HandleError(c, http.StatusUnauthorized, "Cabeçalho Authorization ausente", "", requestID)
			c.Abort()
			return
		}
		token, err := ExtractBearerToken(h)
		if err != nil {
			response.HandleError(c, http.StatusUnauthorized, "Token mal formatado", "", requestID)
			c.Abort()
			return
		}
		claims, err := j.ParseToken(token)
		if err != nil {
			response.HandleError(c, http.StatusUnauthorized, "Token inválido ou expirado", "", requestID)
			c.Abort()
			return
		}

		// Injeta no contexto
		c.Set("userID", claims.ID)
		c.Set("userName", claims.Name)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		//logger.Log.Infof("JWT ok: id=%d email=%s role=%q jti=%d", claims.ID, claims.Email, claims.Role, claims.ID)
		c.Next()
	}
}

// MiddleWare para validar a autorização do usuário de uma requisição http para um determinado serviço
func (j *JWTService) AuthorizeMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()

		roleVal, ok := c.Get("userRole")
		if !ok {
			response.HandleError(c, http.StatusUnauthorized, "Usuário não autenticado", "", requestID)
			c.Abort()
			return
		}
		role, _ := roleVal.(string)

		// Admin sempre pode
		if role == "admin" || slices.Contains(allowedRoles, role) {
			c.Next()
			return
		}
		logger.Log.Infof("Acesso negado: role=%q precisa de %v", role, allowedRoles)
		response.HandleError(c, http.StatusForbidden, "Usuário sem permissão suficiente para esta ação", "", requestID)
		c.Abort()
	}
}

/*
=========================

	Senhas (bcrypt)
=========================
*/
//Gera um hash encriptado com bcrypt de uma senha de usuário para salvar no banco de dados
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("senha não fornecida")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

// Verifica a validade de uma senha comparando com o seu hash encriptado com bcrypt
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
