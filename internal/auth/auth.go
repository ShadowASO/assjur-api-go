package auth

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"ocrserver/internal/config"
)

// Configurações de expiração dos tokens
// const (
// 	// AccessTokenExpire  = 2 * time.Minute
// 	// RefreshTokenExpire = 2 * time.Hour

// )

// Estrutura para os atributos do usuário
type UserAtribs struct {
	UID      string `json:"uid"`
	Uname    string `json:"uname"`
	Urole    string `json:"urole"`
	Password string `json:"-"`
}

// Verificar e decodificar o token JWT
func verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(config.GlobalConfig.JWTSecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}

// Função para validar o token e extrair os dados do usuário
func ValidateToken(token string) (*UserAtribs, error) {
	decoded, err := verifyToken(token)

	if err != nil {
		log.Println("Erro ao decodificar o token: ", err)
		return nil, err
	}
	user := &UserAtribs{}

	/* Validar valores */

	//UID
	if uid, ok := decoded["uid"].(string); ok {
		user.UID = uid
	} else if uidFloat, ok := decoded["uid"].(float64); ok {
		user.UID = fmt.Sprintf("%.0f", uidFloat) // Converte float64 para string
	} else {
		return nil, errors.New("UID ausente ou inválido no token")
	}
	//UNAME
	if uname, ok := decoded["uname"].(string); ok {
		user.Uname = uname
	} else {
		return nil, errors.New("uname ausente ou inválido no token")
	}
	//UROLE
	if urole, ok := decoded["urole"].(string); ok {
		user.Urole = urole
	} else {
		return nil, errors.New("urole ausente ou inválido no token")
	}

	return user, nil
}

// Encriptar a senha utilizando bcrypt
func EncriptarSenhaBcrypt(password string) (string, error) {
	if password == "" {
		return "", errors.New("senha não fornecida")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Erro ao encriptar senha:", err)
		return "", err
	}
	return string(hash), nil
}

// Comparar a senha fornecida com o hash armazenado
func CompararSenhaBcrypt(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Criar um token JWT
func CreateToken(user UserAtribs, expiration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"uid":   user.UID,
		"uname": user.Uname,
		"urole": user.Urole,
		"exp":   time.Now().Add(expiration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte([]byte(config.GlobalConfig.JWTSecretKey)))
}

// Função para extrair o token
func ExtractToken(authHeader string) (string, error) {
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 {
		return "", errors.New("formato de token inválido")
	}
	return tokenParts[1], nil
}
