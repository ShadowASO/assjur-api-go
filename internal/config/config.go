package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
)

var SecretKey []byte

var OpenApiKey string
var OpenOptionMaxTokens int
var OpenOptionMaxCompletionTokens int
var OpenOptionModel string

var CnjPublicApiKey string
var CnjPublicApiUrl string
var ServerPort string
var ServerHost string

// Configuração da conexão com o banco de dados postgresql
var PostgresHost string
var PostgresPort string
var PostgresDB string
var PostgresUser string
var PostgresPassword string

// Elastic
var ElasticHost string
var ElasticPort string
var ElasticUser string
var ElasticPassword string

var AllowedOrigins []string

// GIN
var GinMode string

func corsAllowedOrigins() {
	origins := os.Getenv("CORS_ORIGINS_ALLOWED")
	log.Println(origins)
	if origins == "" {
		log.Println("⚠️ Nenhuma origem permitida definida no .env. Usando padrão localhost.")
		AllowedOrigins = []string{"http://localhost:3002"}
	}
	AllowedOrigins = strings.Split(origins, ",")
}

func ConfigLog() *os.File {
	// Nome do arquivo de log
	logFileName := "application.log"

	// Abre o arquivo de log (ou cria caso não exista)
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Erro ao abrir/criar o arquivo de log: %v", err)
	}
	//defer file.Close()

	// Configura o pacote log para gravar no arquivo
	//log.SetOutput(file)
	// Configura o log para escrever no terminal e no arquivo
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)
	return file
}

func Init() {
	// Configurar saída do log
	//log.SetOutput(os.Stdout)

	// Carregar as variáveis do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}
	corsAllowedOrigins()
	load()
	//showEnv()

}
func load() {
	/* O formato []byte é necessários para uso pelo pacote jwt do GO.*/
	SecretKey = []byte(os.Getenv("JWT_SECRET"))

	OpenApiKey = os.Getenv("OPENAI_API_KEY")
	OpenOptionMaxTokens, _ = strconv.Atoi(os.Getenv("OPENAI_OPTION_MAX_TOKENS"))
	OpenOptionMaxCompletionTokens, _ = strconv.Atoi(os.Getenv("OPENAI_OPTION_MAX_COMPLETION_TOKENS"))

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
	PostgresDB = os.Getenv("POSTGRES_DB")
	PostgresUser = os.Getenv("POSTGRES_USER")
	PostgresPassword = os.Getenv("POSTGRES_PASSWORD")

	// Configurações do Elastic
	ElasticHost = os.Getenv("ELASTIC_HOST")
	ElasticPort = os.Getenv("ELASTIC_PORT")
	ElasticUser = os.Getenv("ELASTIC_USER")
	ElasticPassword = os.Getenv("ELASTIC_PASSWORD")

	//Gin - verifica se a variável de ambiente GIN_MODE está
	//em release mode
	GinMode = os.Getenv("GIN_MODE")
	if GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
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
	fmt.Println("POSTGRES_DB:", PostgresDB)
	fmt.Println("POSTGRES_USER:", PostgresUser)
	fmt.Println("POSTGRES_PASSWORD:", PostgresPassword)

	// Elasticsearch
	fmt.Println("ELASTIC_HOST:", ElasticHost)
	fmt.Println("ELASTIC_PORT:", ElasticPort)
	fmt.Println("ELASTIC_USER:", ElasticUser)
	fmt.Println("ELASTIC_PASSWORD:", ElasticPassword)
}
