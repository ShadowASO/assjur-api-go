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

	"ocrserver/internal/models"
	"ocrserver/internal/services"

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
	promptModel := models.NewPromptModel(db.Pool)
	sessionsModel := models.NewSessionsModel(db.Pool)
	contextoModel := models.NewContextoModel(db.Pool)
	uploadModel := models.NewUploadModel(db.Pool)

	//** INDEX - OpenSearch
	indexModelos := opensearch.NewIndexModelos()
	autosIndex := opensearch.NewAutosIndex()
	autos_tempIndex := opensearch.NewAutos_tempIndex()
	autos_json_embedding := opensearch.NewAutosJsonEmbedding()

	//** SERVICES -- Instancia os SERVICES
	userService := services.NewUsersService(userModel)
	autosService := services.NewAutosService(autosIndex)
	autos_tempService := services.NewAutos_tempService(autos_tempIndex)

	uploadService := services.NewUploadService(uploadModel)

	promptService := services.NewPromptService(promptModel)
	contextoService := services.NewContextoService(contextoModel)
	queryService := services.NewQueryService(sessionsModel)
	sessionService := services.NewSessionService(sessionsModel)
	cnjService := services.NewCnjService(cfg)
	loginService := services.NewLoginService(cfg)

	//Instancia o JWT service
	jwt := auth.NewJWTService(cfg.JWTSecretKey, *cfg)

	//** HANDLERS -- Criando os Handlerss
	usersHandlers := handlers.NewUsersHandlers(userService)
	queryHandlers := handlers.NewQueryHandlers(queryService)
	sessionHandlers := handlers.NewSessionsHandlers(sessionService)
	promptHandlers := handlers.NewPromptHandlers(promptService)
	contextoHandlers := handlers.NewContextoHandlers(contextoService)
	autosHandlers := handlers.NewAutosHandlers(autosService)
	autos_tempHandlers := handlers.NewAutosTempHandlers(autos_tempService)
	uploadHandlers := handlers.NewUploadHandlers(uploadService)
	contextoQueryHandlers := handlers.NewContextoQueryHandlers(sessionsModel)
	loginHandlers := handlers.NewLoginHandlers(loginService)
	openSearchHandlers := handlers.NewModelosHandlers(indexModelos)

	// GLOBAIS -- Inicializando

	//** Iniciando Variáveis de Serviços Globais **
	opensearch.InitIndexService(indexModelos)
	//Inicializa o OpenAIService global
	services.InitOpenaiService(cfg.OpenApiKey, cfg)
	services.InitSessionService(sessionsModel)
	//Inicializando o global AutoService
	services.InitAutosService(autosIndex)
	services.InitAutos_tempService(autos_tempIndex)
	services.InitUsersService(userModel)
	services.InitPromptService(promptModel)
	services.InitContextoService(contextoModel)
	services.InitUploadService(uploadModel)
	services.InitAutosJsonService(autos_json_embedding)

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
		tabelasGroup.GET("/prompts", promptHandlers.SelectAllHandler)
		tabelasGroup.GET("/prompts/:id", promptHandlers.SelectByIDHandler)
		tabelasGroup.DELETE("/prompts/:id", promptHandlers.DeleteHandler)
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
		//openSearchGroup.POST("/modelos/autos/:id", embeddingHandlers.InsertHandler)

		//Inserir um único documento no banco vetorial
		//openSearchGroup.POST("/modelos/autos/doc", embeddingHandlers.InsertDocumentoHandler)
	}

	//CONTEXTO
	contextoGroup := router.Group("/contexto", jwt.AutenticaMiddleware())
	{
		contextoGroup.POST("", contextoHandlers.InsertHandler)
		contextoGroup.GET("", contextoHandlers.SelectAllHandler)
		contextoGroup.GET("/:id", contextoHandlers.SelectByIDHandler)
		contextoGroup.GET("/processo/:id", contextoHandlers.SelectByProcessoHandler)

	}

	//CONTEXTO/DOCUMENTOS/UPLOAD
	uploadGroup := router.Group("/contexto/documentos/upload", jwt.AutenticaMiddleware())
	{
		uploadGroup.POST("", uploadHandlers.UploadFileHandler)       //ok
		uploadGroup.GET("/:id", uploadHandlers.SelectHandler)        //ok
		uploadGroup.DELETE("/:id", uploadHandlers.DeleteHandlerById) //ok

	}

	//CONTEXTO/DOCUMENTOS
	documentosGroup := router.Group("/contexto/documentos", jwt.AutenticaMiddleware())
	{

		documentosGroup.POST("", autos_tempHandlers.PDFHandler)              //Divide o PDF completo dos autos em seus vários documentos
		documentosGroup.GET("/all/:id", autos_tempHandlers.SelectAllHandler) //ok
		documentosGroup.DELETE("/:id", autos_tempHandlers.DeleteHandler)     //ok
		//documentosGroup.POST("/saneador/:id", autos_tempHandlers.SanearByContextHandler) //Identifica natureza e dEleta os documentos inúteis. Apenas isso.

		documentosGroup.POST("/autua", autos_tempHandlers.AutuarDocumentosHandler) //Interpretação pela IA e geração do JSON

	}

	//CONTEXTO/AUTOS
	autosGroup := router.Group("/contexto/autos", jwt.AutenticaMiddleware())
	{

		autosGroup.POST("", autosHandlers.InsertHandler)
		autosGroup.GET("/all/:id", autosHandlers.SelectAllHandler) //ok
		autosGroup.GET("/:id", autosHandlers.SelectByIdHandler)
		autosGroup.DELETE("/:id", autosHandlers.DeleteHandler)

	}

	//CONTEXTO/Query
	contextoQueryGroup := router.Group("/contexto/query", jwt.AutenticaMiddleware())
	{
		contextoQueryGroup.POST("rag", contextoQueryHandlers.QueryHandlerTools)
	}

	//Produção - A porta de execução é extraída do arquivo .env
	router.Run(cfg.ServerPort)

}
