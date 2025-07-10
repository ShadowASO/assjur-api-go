/*
---------------------------------------------------------------------------------------
File: main.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 26-12-2024
Alterações:
---------------------------------------------------------------------------------------
Compilação: go build -v -o server ./cmd/main.go
Execução: ./server
*/
package main

import (
	"fmt"
	"log"

	"ocrserver/internal/auth"
	"ocrserver/internal/lib/pje_lib"
	"ocrserver/internal/models"
	"ocrserver/internal/services"
	"ocrserver/internal/services/embedding"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"ocrserver/internal/config"

	"ocrserver/internal/handlers"
	"ocrserver/internal/opensearch"

	"ocrserver/internal/database/pgdb"

	"time"
)

// Versao da aplicação
const AppVersion = "1.0.1"

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()
		c.Next()
		duration := time.Since(start)
		//msg := fmt.Sprintf("Request - Method: %s | Status: %d | Duration: %v", c.Request.Method, c.Writer.Status(), duration)
		_ = fmt.Sprintf("Request - Method: %s | Status: %d | Duration: %v", c.Request.Method, c.Writer.Status(), duration)
		//logger.Log.Info(msg)
	}
}

func main() {

	//Configuração inicial
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	//Inicializa Objetos Globais
	logger.InitLoggerGlobal("logs/app.log", true)

	//Inicializando a api do CNJ globalmente
	services.InitCnjGlobal(cfg)

	//Exibe o número da versão
	ver := fmt.Sprintf("Versão da aplicação: %s\n", AppVersion)
	logger.Log.Info(ver)

	//Conexão com o Banco de Dados
	dbConfig := pgdb.DBConfig{
		Host:     cfg.PgHost,
		Port:     cfg.PgPort,
		User:     cfg.PgUser,
		Password: cfg.PgPass,
		DBName:   cfg.PgDB,
		PoolSize: cfg.DBPoolSize,
	}
	db, err := pgdb.NewDBConn(dbConfig)
	if err != nil {
		log.Fatalf("Erro ao criar o pool de conexões como o database: %v", err)
	}
	defer db.Close()

	//Inicializa a conexão com o OpenSearch
	err = opensearch.InitOpenSearchService()
	if err != nil {
		log.Fatalf("Erro ao conectar o OpenSearch: %v", err)
	}

	//** MODELS -- Instanciando os MODELOS
	userModel := models.NewUsersModel(db.Pool)
	autosModel := models.NewAutosModel(db.Pool)
	promptModel := models.NewPromptModel(db.Pool)
	tempautosModel := models.NewDocsocrModel(db.Pool)
	sessionsModel := models.NewSessionsModel(db.Pool)
	contextoModel := models.NewContextoModel(db.Pool)
	uploadModel := models.NewUploadModel(db.Pool)

	//** INDEX - OpenSearch
	indexModelos := opensearch.NewIndexModelos()
	autosIndex := opensearch.NewAutosIndex()
	autos_tempIndex := opensearch.NewAutos_tempIndex()

	//** SERVICES -- Instancia os SERVICES
	userService := services.NewUsersService(userModel)
	autosService := services.NewAutosService(autosIndex)
	promptService := services.NewPromptService(promptModel)
	queryService := services.NewQueryService(sessionsModel)
	sessionService := services.NewSessionService(sessionsModel)
	cnjService := services.NewCnjService(cfg)
	loginService := services.NewLoginService(cfg)
	embeddingService := embedding.NewAutosEmbedding()
	autos_tempService := services.NewAutos_tempService(autos_tempIndex)
	//Instancia o JWT service
	jwt := auth.NewJWTService(cfg.JWTSecretKey, *cfg)

	//** HANDLERS -- Criando os Handlerss
	usersHandlers := handlers.NewUsersHandlers(userService)
	queryHandlers := handlers.NewQueryHandlers(queryService)
	sessionHandlers := handlers.NewSessionsHandlers(sessionService)
	promptHandlers := handlers.NewPromptHandlers(promptService)
	contextoHandlers := handlers.NewContextoHandlers(contextoModel)
	autosHandlers := handlers.NewAutosHandlers(autosService)
	uploadHandlers := handlers.NewUploadHandlers(uploadModel)
	docsocrHandlers := handlers.NewDocsocrHandlers(autos_tempService)
	contextoQueryHandlers := handlers.NewContextoQueryHandlers(sessionsModel)
	loginHandlers := handlers.NewLoginHandlers(loginService)
	openSearchHandlers := handlers.NewModelosHandlers(indexModelos)
	embeddingHandlers := handlers.NewEmbeddingHandlers(embeddingService)

	// GLOBAIS -- Inicializando

	//** Iniciando Variáveis de Serviços Globais **
	opensearch.InitIndexService(indexModelos)
	//Inicializa o OpenAIService global
	services.InitOpenaiService(cfg.OpenApiKey, cfg)
	services.InitTempautosService(autosModel, promptModel, tempautosModel)
	//services.InitOpenaiService(userModel)
	services.InitSessionService(sessionsModel)
	//Inicializando o global AutoService
	services.InitAutosService(autosIndex)
	services.InitAutos_tempService(autos_tempIndex)
	services.InitUsersService(userModel)

	//Cria o roteador GIN
	router := gin.Default()

	//Gin - verifica se a variável de ambiente GIN_MODE está
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	//Registra os loggins no sistema
	router.Use(LoggerMiddleware())
	router.Use(middleware.RequestIDMiddleware())

	// Configura o middleware de CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,                                  // Origens permitidas
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Métodos permitidos
		AllowHeaders:     []string{"Content-Type", "Authorization"},           // Cabeçalhos permitidos
		ExposeHeaders:    []string{"Content-Length"},                          // Cabeçalhos expostos ao cliente
		AllowCredentials: true,                                                // Permitir credenciais
		MaxAge:           12 * time.Hour,                                      // Cache da política de CORS
	}))

	//AUTH - Rotas para login e geração/validação de tokens
	router.POST("/auth/login", loginHandlers.LoginHandler)
	router.POST("/auth/register", usersHandlers.InsertHandler)
	router.POST("/auth/token/refresh", loginHandlers.RefreshTokenHandler)
	router.POST("/auth/token/verify", loginHandlers.VerifyTokenHandler)

	//CNJ
	router.POST("/cnj/processo", cnjService.GetProcessoFromCnj)

	//USERS - ok
	userGroup := router.Group("/users", jwt.AutenticaMiddleware())
	{
		userGroup.POST("", usersHandlers.InsertHandler)
		userGroup.GET("", usersHandlers.SelectAllHandler)
		userGroup.GET("/:id", usersHandlers.SelectHandler)
	}

	//QUERY
	router.POST("/query/chat", jwt.AutenticaMiddleware(), queryHandlers.QueryHandler)

	//SESSIONS
	sessionGroup := router.Group("/sessions", jwt.AutenticaMiddleware())
	{
		sessionGroup.POST("", sessionHandlers.InsertHandler)
		sessionGroup.GET("", sessionHandlers.SelectAllHandler)
		sessionGroup.GET("/uso", sessionHandlers.GetTokenUsoHandler)
		sessionGroup.GET("/:id", sessionHandlers.SelectHandler)
	}

	//TABELAS
	tabelasGroup := router.Group("/tabelas", jwt.AutenticaMiddleware())
	{
		tabelasGroup.POST("/prompts", promptHandlers.InsertHandler)
		tabelasGroup.PUT("/prompts", promptHandlers.UpdateHandler)
		tabelasGroup.DELETE("/prompts/:id", promptHandlers.DeleteHandler)
		tabelasGroup.GET("/prompts", promptHandlers.SelectAllHandler)
		tabelasGroup.GET("/prompts/:id", promptHandlers.SelectByIDHandler)
	}

	openSearchGroup := router.Group("/tabelas", jwt.AutenticaMiddleware())
	{
		openSearchGroup.POST("/modelos", openSearchHandlers.InsertHandler)
		openSearchGroup.PUT("/modelos/:id", openSearchHandlers.UpdateHandler)
		openSearchGroup.DELETE("/modelos/:id", openSearchHandlers.DeleteHandler)
		// Estou usando o método POST para facilitar o envio do body. Avaliar mudança para GET
		openSearchGroup.POST("/modelos/search", openSearchHandlers.SearchModelosHandler)
		openSearchGroup.GET("/modelos/:id", openSearchHandlers.SelectByIdHandler)

		//Inserir todo o contexto no banco vetorial
		openSearchGroup.POST("/modelos/autos/:id", embeddingHandlers.InsertHandler)

		//Inserir um único documento no banco vetorial
		openSearchGroup.POST("/modelos/autos/doc", embeddingHandlers.InsertDocumentoHandler)
	}

	//CONTEXTO
	contextoGroup := router.Group("/contexto", jwt.AutenticaMiddleware())
	{
		contextoGroup.POST("", contextoHandlers.InsertHandler)
		contextoGroup.GET("", contextoHandlers.SelectAllHandler)
		contextoGroup.GET("/:id", contextoHandlers.SelectByIDHandler)
		contextoGroup.GET("/processo/:id", contextoHandlers.SelectByProcessoHandler)

	}

	//CONTEXTO/DOCUMENTOS/UOLOAD
	uploadGroup := router.Group("/contexto/documentos/upload", jwt.AutenticaMiddleware())
	{
		uploadGroup.POST("", uploadHandlers.UploadFileHandler)
		uploadGroup.GET("/:id", uploadHandlers.SelectHandler)
		uploadGroup.DELETE("/:id", uploadHandlers.DeleteHandlerById)

	}

	//CONTEXTO/DOCUMENTOS/OCR
	documentosGroup := router.Group("/contexto/documentos/ocr", jwt.AutenticaMiddleware())
	{
		documentosGroup.POST("/juntada/:id", pje_lib.JuntadaByContextHandler)
		documentosGroup.POST("", pje_lib.PDFHandler)
		documentosGroup.POST("/:id", pje_lib.OcrByContextHandler)
		documentosGroup.GET("/all/:id", docsocrHandlers.SelectAllHandler)
		documentosGroup.DELETE("/:id", docsocrHandlers.DeleteHandlerById)

	}

	//CONTEXTO/AUTOS
	autosGroup := router.Group("/contexto/autos", jwt.AutenticaMiddleware())
	{
		autosGroup.POST("/analise", autosHandlers.AutuarDocumentos)
		autosGroup.POST("", autosHandlers.InsertHandler)
		autosGroup.GET("/all/:id", autosHandlers.SelectAllHandler)
		autosGroup.GET("/:id", autosHandlers.SelectByIdHandler)
		autosGroup.DELETE("/:id", autosHandlers.DeleteHandler)

	}

	//CONTEXTO/Query
	contextoQueryGroup := router.Group("/contexto/query", jwt.AutenticaMiddleware())
	{

		contextoQueryGroup.POST("rag", contextoQueryHandlers.QueryHandlerTools)
		contextoQueryGroup.POST("", contextoQueryHandlers.QueryHandler)
	}

	//Produção - A porta de execução é extraída do arquivo .env
	router.Run(cfg.ServerPort)

}
