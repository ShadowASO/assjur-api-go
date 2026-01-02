package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Gin
	GinMode string

	// Servidor
	ServerPort           string // aceita "9001" ou ":9001"
	ServerHost           string
	UploadTimeoutSeconds int

	// Postgres
	PgHost     string
	PgPort     string
	PgDB       string
	PgUser     string
	PgPass     string
	DBPoolSize int

	// JWT
	JWTSecretKey       string
	AccessTokenExpire  time.Duration
	RefreshTokenExpire time.Duration

	// CNJ
	CnjPublicApiKey string
	CnjPublicApiUrl string

	// OpenAI
	OpenApiKey                    string
	OpenOptionMaxCompletionTokens int
	OpenOptionModel               string //Modelo principal 'gpt-5-mini'
	OpenOptionModelSecundary      string //Modelo secundário 'gpt-5-nano'
	OpenOptionTimeoutSeconds      int

	// Elastic (se usado)
	ElasticHost     string
	ElasticPort     string
	ElasticUser     string
	ElasticPassword string

	// OpenSearch
	OpenSearchHost     string // ex: http://192.168.0.30
	OpenSearchPort     string // ex: 9200
	OpenSearchUser     string
	OpenSearchPassword string
	//OpenSearchIndexName string
	OpenSearchRagName string

	// CORS
	AllowedOrigins []string

	// App mode
	ApplicationMode string
}

var (
	GlobalConfig   *Config
	onceLoadConfig sync.Once
	loadErr        error
)

func LoadConfig() (*Config, error) {
	onceLoadConfig.Do(func() {
		// 1) Carrega .env se existir (sem falhar em prod)
		loadDotEnvIfPresent()

		cfg := &Config{}
		loadErr = initEnv(cfg)
		if loadErr != nil {
			return
		}
		GlobalConfig = cfg

		showEnv(cfg) // não vaza segredos (usa mask)
	})

	return GlobalConfig, loadErr
}

func loadDotEnvIfPresent() {
	// Procura .env no cwd
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("⚠️  Erro ao carregar .env: %v (seguindo com variáveis de ambiente)", err)
		} else {
			log.Println("✔️  .env carregado")
		}
	} else {
		// Procura um .env no diretório do executável (deploys)
		if exe, err := os.Executable(); err == nil {
			dir := filepath.Dir(exe)
			p := filepath.Join(dir, ".env")
			if _, sErr := os.Stat(p); sErr == nil {
				if err := godotenv.Load(p); err != nil {
					log.Printf("⚠️  Erro ao carregar %s: %v", p, err)
				} else {
					log.Printf("✔️  .env carregado de %s", p)
				}
			}
		}
	}
}

// Helpers -----------------------------

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvRequired(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return "", fmt.Errorf("variável de ambiente obrigatória %s não definida", key)
	}
	return v, nil
}

func parseInt(key, val string, def, min, max int) int {
	if strings.TrimSpace(val) == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("⚠️  %s inválido (%q), usando default=%d: %v", key, val, def, err)
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// Aceita "15" (minutos) ou durações do Go: "15m", "2h"
func parseDurationFlexible(key, val string, def time.Duration) time.Duration {
	val = strings.TrimSpace(val)
	if val == "" {
		return def
	}
	// Tenta como duração Go
	if d, err := time.ParseDuration(val); err == nil {
		return d
	}
	// Tenta como número de minutos
	if n, err := strconv.Atoi(val); err == nil && n >= 0 {
		return time.Duration(n) * time.Minute
	}
	log.Printf("⚠️  %s inválido (%q), usando default=%s", key, val, def)
	return def
}

func normalizePort(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ":4001"
	}
	if p[0] != ':' {
		return ":" + p
	}
	return p
}

func normalizeURLHost(h string) (string, error) {
	h = strings.TrimSpace(h)
	if h == "" {
		return "", errors.New("host vazio")
	}
	// exige esquema para evitar confusões
	if !strings.HasPrefix(h, "http://") && !strings.HasPrefix(h, "https://") {
		return "", fmt.Errorf("host deve incluir esquema http(s)://, recebido: %q", h)
	}
	// remove barra final
	return strings.TrimRight(h, "/"), nil
}

func splitAndTrimCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		pp := strings.TrimSpace(p)
		if pp != "" {
			out = append(out, pp)
		}
	}
	return out
}

func mask(s string) string {
	if s == "" {
		return "(vazio)"
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}

// ------------------------------------

func initEnv(cfg *Config) error {
	// Gin
	cfg.GinMode = getEnv("GIN_MODE", "release")

	// CORS
	origins := getEnv("CORS_ORIGINS_ALLOWED", "")
	if origins == "" {
		log.Println("ℹ️  CORS_ORIGINS_ALLOWED vazio. Usando fallback: http://localhost:3002")
		cfg.AllowedOrigins = []string{"http://localhost:3002"}
	} else {
		list := splitAndTrimCSV(origins)
		if len(list) == 1 && list[0] == "*" {
			log.Println("⚠️  CORS com '*' detectado. Atenção: com credenciais habilitadas, alguns middlewares rejeitam '*'.")
		}
		cfg.AllowedOrigins = list
	}

	// Servidor
	cfg.ServerPort = normalizePort(getEnv("SERVER_PORT", "4001"))
	cfg.ServerHost = getEnv("SERVER_HOST", "localhost")
	cfg.UploadTimeoutSeconds = parseInt("UPLOAD_TIMEOUT_SECONDS", getEnv("UPLOAD_TIMEOUT_SECONDS", "300"), 300, 60, 300)

	// Postgres
	cfg.PgHost = getEnv("PG_HOST", "192.168.0.30")
	cfg.PgPort = getEnv("PG_PORT", "5432")
	cfg.PgDB = getEnv("PG_DB", "assjurdb")
	cfg.PgUser = getEnv("PG_USER", "assjurpg")
	var err error
	if cfg.PgPass, err = getEnvRequired("PG_PASS"); err != nil {
		return err
	}

	// OpenSearch
	rawOSHost := getEnv("OPENSEARCH_HOST", "http://192.168.0.30")
	if cfg.OpenSearchHost, err = normalizeURLHost(rawOSHost); err != nil {
		return fmt.Errorf("OPENSEARCH_HOST inválido: %w", err)
	}
	cfg.OpenSearchPort = getEnv("OPENSEARCH_PORT", "9200")
	cfg.OpenSearchUser = getEnv("OPENSEARCH_USER", "admin")
	cfg.OpenSearchPassword = getEnv("OPENSEARCH_PASSWORD", "Open@1320")
	//cfg.OpenSearchIndexName = getEnv("OPENSEARCH_INDEX_NAME", "modelos")
	cfg.OpenSearchRagName = getEnv("OPENSEARCH_RAG_NAME", "rag_doc_embedding")
	cfg.OpenOptionTimeoutSeconds = parseInt(
		"OPENAI_OPTION_TIMEOUT_SECONDS",
		getEnv("OPENAI_OPTION_TIMEOUT_SECONDS", "120"), // default 120s
		120, 30, 600, // min 30s, max 10min
	)

	// CNJ
	if cfg.CnjPublicApiKey, err = getEnvRequired("CNJ_PUBLIC_API_KEY"); err != nil {
		return err
	}
	if cfg.CnjPublicApiUrl, err = getEnvRequired("CNJ_PUBLIC_API_URL"); err != nil {
		return err
	}

	// App mode
	cfg.ApplicationMode = getEnv("APPLICATION_MODE", "production")

	// JWT
	if cfg.JWTSecretKey, err = getEnvRequired("JWT_SECRET"); err != nil {
		return err
	}

	// OpenAI (usa strings — evita acoplamento com SDK)
	if cfg.OpenApiKey, err = getEnvRequired("OPENAI_API_KEY"); err != nil {
		return err
	}
	cfg.OpenOptionMaxCompletionTokens = parseInt("OPENAI_OPTION_MAX_COMPLETION_TOKENS",
		getEnv("OPENAI_OPTION_MAX_COMPLETION_TOKENS", "16384"),
		16384, 256, 128000,
	)
	cfg.OpenOptionModel = getEnv("OPENAI_OPTION_MODEL", "gpt-5-mini")
	cfg.OpenOptionModelSecundary = getEnv("OPENAI_OPTION_MODEL_SECUNDARY", "gpt-5-mini")

	// Pool do DB
	cfg.DBPoolSize = parseInt("DB_POOLSIZE", getEnv("DB_POOLSIZE", "25"), 25, 5, 200)

	// Expiração de tokens (minutos numéricos OU duration Go)
	cfg.AccessTokenExpire = parseDurationFlexible("ACCESSTOKEN_EXPIRE", getEnv("ACCESSTOKEN_EXPIRE", "10m"), 10*time.Minute)
	cfg.RefreshTokenExpire = parseDurationFlexible("REFRESHTOKEN_EXPIRE", getEnv("REFRESHTOKEN_EXPIRE", "60m"), 60*time.Minute)

	return nil
}

func showEnv(cfg *Config) {
	// Mostra somente o necessário, mascarando segredos
	fmt.Println("--------- CONFIG ---------")
	fmt.Println("GIN_MODE:", cfg.GinMode)
	fmt.Println("SERVER_HOST:", cfg.ServerHost)
	fmt.Println("SERVER_PORT:", cfg.ServerPort)

	fmt.Println("POSTGRES_HOST:", cfg.PgHost)
	fmt.Println("POSTGRES_PORT:", cfg.PgPort)
	fmt.Println("POSTGRES_DB:", cfg.PgDB)
	fmt.Println("POSTGRES_USER:", cfg.PgUser)
	fmt.Println("POSTGRES_PASSWORD:", mask(cfg.PgPass))

	fmt.Println("OPENSEARCH_HOST:", cfg.OpenSearchHost)
	fmt.Println("OPENSEARCH_PORT:", cfg.OpenSearchPort)
	fmt.Println("OPENSEARCH_USER:", cfg.OpenSearchUser)
	fmt.Println("OPENSEARCH_PASSWORD:", mask(cfg.OpenSearchPassword))
	//fmt.Println("OPENSEARCH_INDEX_NAME:", cfg.OpenSearchIndexName)
	fmt.Println("OPENAI_OPTION_TIMEOUT_SECONDS:", cfg.OpenOptionTimeoutSeconds)

	fmt.Println("CNJ_PUBLIC_API_KEY:", mask(cfg.CnjPublicApiKey))
	fmt.Println("CNJ_PUBLIC_API_URL:", cfg.CnjPublicApiUrl)

	fmt.Println("OPENAI_API_KEY:", mask(cfg.OpenApiKey))
	fmt.Println("OPENAI_OPTION_MODEL:", cfg.OpenOptionModel)
	fmt.Println("OPENAI_OPTION_MODEL_SECUNDARY:", cfg.OpenOptionModelSecundary)
	fmt.Println("OPENAI_OPTION_MAX_COMPLETION_TOKENS:", cfg.OpenOptionMaxCompletionTokens)

	fmt.Println("DB_POOLSIZE:", cfg.DBPoolSize)
	fmt.Println("ACCESS_TOKEN_EXPIRE:", cfg.AccessTokenExpire)
	fmt.Println("REFRESH_TOKEN_EXPIRE:", cfg.RefreshTokenExpire)

	fmt.Println("APPLICATION_MODE:", cfg.ApplicationMode)
	fmt.Println("CORS_ORIGINS_ALLOWED:", strings.Join(cfg.AllowedOrigins, ","))
	fmt.Println("--------------------------")
}
