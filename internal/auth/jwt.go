/*
---------------------------------------------------------------------------------------
File: jwt.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
---------------------------------------------------------------------------------------
*/
package auth

import (
	"errors"
	"net/http"

	"ocrserver/internal/config"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/utils/logger"

	"time"

	"slices"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secretKey []byte
	//Config    config.Config
	Config config.Config
}

// Custom claims structure extending jwt.StandardClaims
type Claims struct {
	UserID    uint   `json:"user_id"`
	UserEmail string `json:"user_email"`
	UserRole  string `json:"user_role"`
	UserName  string `json:"user_name"`
	jwt.RegisteredClaims
}

func NewJWTService(jwtSecretKey string, config config.Config) *JWTService {
	return &JWTService{
		secretKey: []byte(jwtSecretKey),
		Config:    config,
	}
}

// GenerateToken creates a new JWT token for a user
func (j *JWTService) GenerateToken(userID uint, name, email, role string, timeExpire time.Duration) (string, error) {
	// Create claims with multiple fields
	claims := &Claims{
		UserID:    userID,
		UserEmail: email,
		UserRole:  role,
		UserName:  name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(timeExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			//Issuer:    utils.GenerateUUID().String()},
			Issuer: uuid.New().String()},
	}

	//Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//Generate encoded token
	return token.SignedString(j.secretKey)
}

// ValidateToken validates the JWT token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return j.secretKey, nil
		},
	)
	if err != nil {
		return nil, err
	}
	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token claims")
}

// AutenticaMiddleware protects routes requiring authentication
func (j *JWTService) AutenticaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}
		// Remove "Bearer " prefix if present
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}
		// Validate token
		claims, err := j.ValidateToken(tokenString)

		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		// Set claims in context for later use
		c.Set("userID", claims.UserID)
		c.Set("userName", claims.UserName)
		c.Set("userEmail", claims.UserEmail)
		c.Set("userRole", claims.UserRole)
		c.Next()
	}
}

func (j *JWTService) AuthorizaMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//Generate request ID for tracing
		requestID := uuid.New().String()

		userRoleInterface, exist := c.Get("userRole")
		if !exist {

			response.HandleError(c, http.StatusUnauthorized, "Usuário não autenticado", "", requestID)
			logger.Log.Error("Usuário não autenticado")
			c.Abort()
			return
		}

		userRole, ok := userRoleInterface.(string)
		if !ok {

			response.HandleError(c, http.StatusUnauthorized, "Formato inválido do token ou da permissão do usuário", "", requestID)
			logger.Log.Error("Formato inválido")
			c.Abort()
			return
		}

		// Admin tem permissão total
		if userRole == "admin" {
			c.Next()
			return
		}

		// Verifica se o usuário tem uma das roles permitidas
		if slices.Contains(allowedRoles, userRole) {
			c.Next()
			return
		}

		response.HandleError(c, http.StatusForbidden, "Usuário sem permissão suficiente para esta ação", "", requestID)
		logger.Log.Error("Usuário não autorizado para essa ação")
		c.Abort()
	}
}
