package config

import (
	"fmt"

	"sync"

	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
)

type Config struct {
	//GIN modo
	GinMode string

	// Configuração do servidor de API
	ServerPort string
	ServerHost string

	// Configurações do bando de dados
	PgHost     string
	PgPort     string
	PgDB       string
	PgUser     string
	PgPass     string
	DBPoolSize int

	// JWT_SECRET_KEY
	JWTSecretKey       string
	AccessTokenExpire  time.Duration
	RefreshTokenExpire time.Duration

	//Api do CNJ
	CnjPublicApiKey string
	CnjPublicApiUrl string

	//Api da OpenAI
	OpenApiKey string
	//OpenOptionMaxTokens           int
	OpenOptionMaxCompletionTokens int
	OpenOptionModel               string
	OpenOptionModelSecundary      string

	// Elastic
	ElasticHost     string
	ElasticPort     string
	ElasticUser     string
	ElasticPassword string

	// OpenSearch
	OpenSearchHost      string
	OpenSearchPort      string
	OpenSearchUser      string
	OpenSearchPassword  string
	OpenSearchIndexName string

	//Configuração de CORS
	AllowedOrigins []string

	// Application mode
	ApplicationMode string
}

// Variável Global com todas as configurações
var GlobalConfig *Config
var onceLoadConfig sync.Once

func LoadConfig() (*Config, error) {
	log.Println("Carregando configurações do arquivo .env")
	var loadErr error

	onceLoadConfig.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Printf("Erro ao carregar .env: %v", err)
			loadErr = err
			return
		}

		config := &Config{}
		InitEnv(config)
		GlobalConfig = config
		//showEnv(config)
	})

	return GlobalConfig, loadErr
}

func corsAllowedOrigins(cfg *Config) {
	origins := os.Getenv("CORS_ORIGINS_ALLOWED")
	if origins == "" {
		log.Println("⚠️ Nenhuma origem permitida. Usando padrão localhost.")
		cfg.AllowedOrigins = []string{"http://localhost:3002"}
		return
	}
	cfg.AllowedOrigins = strings.Split(origins, ",")
}

// getEnv retrieves an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvRequired retrieves a required environment variable
func getEnvRequired(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	return value
}

func InitEnv(cfg *Config) {

	//Configuração do release mode do GIN
	cfg.GinMode = getEnv("GIN_MODE", "release")

	corsAllowedOrigins(cfg)

	//Configurações do servidor
	cfg.ServerPort = getEnv("SERVER_PORT", ":4001")
	cfg.ServerHost = getEnv("SERVER_HOST", "localhost")

	//Configuração da conexão com o banco de dados postgresql
	cfg.PgHost = getEnv("PG_HOST", "192.168.0.30")
	cfg.PgPort = getEnv("PG_PORT", "5432")
	cfg.PgDB = getEnv("PG_DB", "assjurdb")
	cfg.PgUser = getEnv("PG_USER", "assjurpg")
	cfg.PgPass = getEnvRequired("PG_PASS")

	// Configurações do OpenSearch
	cfg.OpenSearchHost = getEnv("OPENSEARCH_HOST", "http://192.168.0.30")
	cfg.OpenSearchPort = getEnv("OPENSEARCH_PORT", "9200")
	cfg.OpenSearchUser = getEnv("OPENSEARCH_USER", "admin")
	cfg.OpenSearchPassword = getEnv("OPENSEARCH_PASSWORD", "Open@1320")
	//O nome do índice de modelos no OpenSearch
	cfg.OpenSearchIndexName = getEnv("OPENSEARCH_INDEX_NAME", "modelos")

	//CNJ
	cfg.CnjPublicApiKey = getEnvRequired("CNJ_PUBLIC_API_KEY")
	cfg.CnjPublicApiUrl = getEnvRequired("CNJ_PUBLIC_API_URL")

	cfg.ApplicationMode = os.Getenv("APPLICATION_MODE")
	//JWT_SECRET_KEY
	cfg.JWTSecretKey = getEnvRequired("JWT_SECRET")

	//OpenAI
	cfg.OpenApiKey = getEnvRequired("OPENAI_API_KEY")
	cfg.OpenOptionMaxCompletionTokens, _ = strconv.Atoi(getEnv("OPENAI_OPTION_MAX_COMPLETION_TOKENS", "16384"))
	cfg.OpenOptionModel = getEnv("OPENAI_OPTION_MODEL", openai.ChatModelGPT4_1)
	cfg.OpenOptionModelSecundary = getEnv("OPENAI_OPTION_MODEL_SECUNDARY", openai.ChatModelGPT4_1Mini)

	/*
		O número default de DBPoolSize == 25 e se houver indicação na variável de ambiente,
		modificamos
	*/
	cfg.DBPoolSize = 25
	tmp := getEnv("DB_POOLSIZE", "25")

	num, err := strconv.ParseInt(tmp, 10, 64)
	if err == nil {
		cfg.DBPoolSize = int(num)
	}

	//Tempo de expiração do acesstoken - default é 15 minutos
	tmp = getEnv("ACCESSTOKEN_EXPIRE", "10")
	num, err = strconv.ParseInt(tmp, 10, 8)
	if err == nil {
		cfg.AccessTokenExpire = time.Duration(num * int64(time.Minute))
		log.Printf("AccessTokenExpire=%s", tmp)

	}

	//Tempo de expiração do refreshtoken - defauolt é 15 minutos

	tmp = getEnv("REFRESHTOKEN_EXPIRE", "60")
	num, err = strconv.ParseInt(tmp, 10, 8)
	if err == nil {
		cfg.RefreshTokenExpire = time.Duration(num * int64(time.Minute))
		log.Printf("RefreshTokenExpire=%s", tmp)

	}

}

func showEnv(cfg *Config) {

	// Exibir as variáveis lidas
	fmt.Println("JWT_SECRET:", cfg.JWTSecretKey)
	fmt.Println("OPENAI_API_KEY:", cfg.OpenApiKey)
	fmt.Println("OPENAI_OPTION_MAX_COMPLETION_TOKENS:", cfg.OpenOptionMaxCompletionTokens)

	fmt.Println("CNJ_PUBLIC_API_KEY:", cfg.CnjPublicApiKey)
	fmt.Println("CNJ_PUBLIC_API_URL:", cfg.CnjPublicApiUrl)
	fmt.Println("SERVER_PORT:", cfg.ServerPort)
	fmt.Println("SERVER_HOST:", cfg.ServerHost)

	fmt.Println("POSTGRES_HOST:", cfg.PgHost)
	fmt.Println("POSTGRES_PORT:", cfg.PgPort)
	fmt.Println("POSTGRES_DB:", cfg.PgDB)
	fmt.Println("POSTGRES_USER:", cfg.PgUser)
	fmt.Println("POSTGRES_PASSWORD:", cfg.PgPass)

	// OpenSearch
	fmt.Println("OPENSEARCH_HOST:", cfg.OpenSearchHost)
	fmt.Println("OPENSEARCH_PORT:", cfg.OpenSearchPort)
	fmt.Println("OPENSEARCH_USER:", cfg.OpenSearchUser)
	fmt.Println("OPENSEARCH_PASSWORD:", cfg.OpenSearchPassword)

	//fmt.Println("OPENSEARCH_INDEX_NAME:", OpenSearchIndexName)
	//fmt.Println("OPENSEARCH_MODEL_ID:", OpenSearchModelId)

	fmt.Println("APPLICATION_MODE:", cfg.ApplicationMode)

}
