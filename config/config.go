package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
)

var SecretKey []byte

var OpenApiKey string
var OpenOptionMaxTokens int
var OpenOptionModel string

var CnjPublicApiKey string
var CnjPublicApiUrl string
var ServerPort string
var ServerHost string

// Configuração da conexão com o banco de dados postgresql
var PostgresHost string
var PostgresPort string

func Init() {
	// Carregar as variáveis do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}
	load()
	showEnv()

}
func load() {
	/* O formato []byte é necessários para uso pelo pacote jwt do GO.*/
	SecretKey = []byte(os.Getenv("JWT_SECRET"))

	OpenApiKey = os.Getenv("OPENAI_API_KEY")
	OpenOptionMaxTokens, _ = strconv.Atoi(os.Getenv("OPENAI_OPTION_MAX_TOKENS"))

	OpenOptionModel = os.Getenv("OPENAI_OPTION_MODELO")
	if OpenOptionModel == "" {
		log.Printf("OPENAI_OPTION_MODELO: modelo incorreto")
		OpenOptionModel = openai.ChatModelGPT4oMini
	}

	CnjPublicApiKey = os.Getenv("CNJ_PUBLIC_API_KEY")
	CnjPublicApiUrl = os.Getenv("CNJ_PUBLIC_API_URL")

	ServerPort = os.Getenv("SERVER_PORT")
	ServerHost = os.Getenv("SERVER_HOST")

	// Configuração da conexão com o banco de dados postgresql
	PostgresHost = os.Getenv("POSTGRES_HOST")
	PostgresPort = os.Getenv("POSTGRES_PORT")
}

func showEnv() {

	// Exibir as variáveis lidas
	fmt.Println("JWT_SECRET:", SecretKey)
	fmt.Println("OPENAI_API_KEY:", OpenApiKey)

	fmt.Println("CNJ_PUBLIC_API_KEY:", CnjPublicApiKey)
	fmt.Println("CNJ_PUBLIC_API_URL:", CnjPublicApiUrl)
	fmt.Println("SERVER_PORT:", ServerPort)
	fmt.Println("SERVER_HOST:", ServerHost)

	fmt.Println("POSTGRES_HOST:", PostgresHost)
	fmt.Println("POSTGRES_PORT:", PostgresPort)
}
