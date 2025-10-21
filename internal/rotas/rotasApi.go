// internal/rotas/rotas.go
// ---------------------------------------------------------------------------------------
package rotas

import (
	"github.com/gin-gonic/gin"

	"ocrserver/internal/auth"
	"ocrserver/internal/config"
	"ocrserver/internal/database/pgdb"
	"ocrserver/internal/handlers"
	"ocrserver/internal/models"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
)

// SetRotasSistema registra todas as rotas e injeta dependências
func SetRotasSistema(router *gin.Engine, cfg *config.Config, db *pgdb.DBPool) {
	// --- JWT service ---
	jwt := auth.NewJWTService(*cfg)

	// --- MODELS ---
	userModel := models.NewUsersModel(db.Pool)
	promptModel := models.NewPromptModel(db.Pool)
	sessionsModel := models.NewSessionsModel(db.Pool)
	contextoModel := models.NewContextoModel(db.Pool)
	uploadModel := models.NewUploadModel(db.Pool)

	// --- OpenSearch Indexes ---
	indexModelos := opensearch.NewIndexModelos()
	autosIndex := opensearch.NewAutosIndex()
	autosTempIndex := opensearch.NewAutos_tempIndex()
	autosJSONEmbedding := opensearch.NewAutosJsonEmbedding()
	eventosIdx := opensearch.NewEventosIndex()
	baseIndex := opensearch.NewBaseIndex()

	// --- SERVICES ---
	userService := services.NewUsersService(userModel)
	autosService := services.NewAutosService(autosIndex)
	autosTempService := services.NewAutos_tempService(autosTempIndex)
	uploadService := services.NewUploadService(uploadModel)
	promptService := services.NewPromptService(promptModel)
	contextoService := services.NewContextoService(contextoModel)
	queryService := services.NewQueryService(sessionsModel)
	sessionService := services.NewSessionService(sessionsModel)
	cnjService := services.NewCnjService(cfg)
	loginService := services.NewLoginService(cfg)
	services.InitEventosService(eventosIdx)

	// --- HANDLERS ---
	usersHandlers := handlers.NewUsersHandlers(userService)
	queryHandlers := handlers.NewQueryHandlers(queryService)
	sessionHandlers := handlers.NewSessionsHandlers(sessionService)
	promptHandlers := handlers.NewPromptHandlers(promptService)
	contextoHandlers := handlers.NewContextoHandlers(contextoService)
	autosHandlers := handlers.NewAutosHandlers(autosService)
	autosTempHandlers := handlers.NewAutosTempHandlers(autosTempService)
	uploadHandlers := handlers.NewUploadHandlers(uploadService)
	contextoQueryHandlers := handlers.NewContextoQueryHandlers(sessionsModel)
	loginHandlers := handlers.NewLoginHandlers(loginService, jwt) // <- garante consistência do construtor
	openSearchHandlers := handlers.NewModelosHandlers(indexModelos)
	baseHandlers := handlers.NewRagHandlers(baseIndex)
	eventosHandlers := handlers.NewEventosHandlers(services.EventosServiceGlobal)

	// --- Objetos/Serviços globais (quando realmente necessários) ---
	//opensearch.InitIndexService(indexModelos)
	services.InitSessionService(sessionsModel)
	services.InitAutosService(autosIndex)
	services.InitAutos_tempService(autosTempIndex)
	services.InitUsersService(userModel)
	services.InitPromptService(promptModel)
	services.InitContextoService(contextoModel)
	services.InitUploadService(uploadModel)
	services.InitAutosJsonService(autosJSONEmbedding)
	opensearch.InitModelosService()
	opensearch.InitBaseIndex()
	services.InitBaseService(baseIndex)

	// --- ROTAS PÚBLICAS ---
	router.GET("/sys/version", handlers.VersionHandler)

	// Auth
	router.POST("/auth/login", loginHandlers.LoginHandler)
	router.POST("/auth/register", usersHandlers.InsertHandler)
	router.POST("/auth/token/refresh", loginHandlers.RefreshTokenHandler)
	router.POST("/auth/token/verify", loginHandlers.VerifyTokenHandler)

	// CNJ
	router.POST("/cnj/processo", cnjService.GetProcessoFromCnj)

	// --- ROTAS PROTEGIDAS ---

	// USERS
	userGroup := router.Group("/users", jwt.AuthMiddleware())
	{
		userGroup.POST("", usersHandlers.InsertHandler)
		userGroup.GET("", usersHandlers.SelectAllHandler)
		userGroup.GET("/:id", usersHandlers.SelectHandler)
	}

	// SESSIONS
	sessionGroup := router.Group("/sessions", jwt.AuthMiddleware())
	{
		sessionGroup.POST("", sessionHandlers.InsertHandler)
		sessionGroup.GET("", sessionHandlers.SelectAllHandler)
		sessionGroup.GET("/uso", sessionHandlers.GetTokenUsoHandler)
		sessionGroup.GET("/:id", sessionHandlers.SelectHandler)
	}

	// TABELAS (somente admin)
	tabelasGroup := router.Group("/tabelas", jwt.AuthMiddleware(), jwt.AuthorizeMiddleware("admin"))
	{
		tabelasGroup.POST("/prompts", promptHandlers.InsertHandler)
		tabelasGroup.PUT("/prompts", promptHandlers.UpdateHandler)
		tabelasGroup.GET("/prompts", promptHandlers.SelectAllHandler)
		tabelasGroup.GET("/prompts/:id", promptHandlers.SelectByIDHandler)
		tabelasGroup.DELETE("/prompts/:id", promptHandlers.DeleteHandler)
	}

	// OpenSearch (modelos)
	openSearchGroup := router.Group("/tabelas", jwt.AuthMiddleware())
	{
		openSearchGroup.POST("/modelos", openSearchHandlers.InsertHandler)
		openSearchGroup.PUT("/modelos/:id", openSearchHandlers.UpdateHandler)
		openSearchGroup.DELETE("/modelos/:id", openSearchHandlers.DeleteHandler)
		openSearchGroup.POST("/modelos/search", openSearchHandlers.SearchModelosHandler)
		openSearchGroup.GET("/modelos/:id", openSearchHandlers.SelectByIdHandler)

		// RAG
		openSearchGroup.POST("/rag", baseHandlers.InsertHandler)
		openSearchGroup.PUT("/rag/:id", baseHandlers.UpdateHandler)
		openSearchGroup.DELETE("/rag/:id", baseHandlers.DeleteHandler)
		openSearchGroup.POST("/rag/search", baseHandlers.SearchHandler)
		openSearchGroup.GET("/rag/:id", baseHandlers.SelectByIdHandler)
	}

	// CONTEXTO (somente admin)
	contextoGroup := router.Group("/contexto", jwt.AuthMiddleware())
	{
		contextoGroup.POST("", contextoHandlers.InsertHandler)
		contextoGroup.PUT("/:id", contextoHandlers.UpdateHandler)
		contextoGroup.GET("", contextoHandlers.SelectAllHandler)
		contextoGroup.GET("/:id", contextoHandlers.SelectByIDHandler)
		contextoGroup.GET("/processo/:id", contextoHandlers.SelectByProcessoHandler)
		contextoGroup.POST("/processo/search", contextoHandlers.SearchByProcessoHandler)
		contextoGroup.DELETE("/:id", contextoHandlers.DeleteHandler)
		contextoGroup.GET("/tokens/:id", contextoHandlers.SelectByProcessoHandler) // confere se este handler é o correto
	}

	// UPLOAD (somente admin)
	uploadGroup := router.Group("/contexto/documentos/upload", jwt.AuthMiddleware())
	{
		uploadGroup.POST("", uploadHandlers.UploadFileHandler)
		uploadGroup.GET("/:id", uploadHandlers.SelectHandler)
		uploadGroup.DELETE("/:id", uploadHandlers.DeleteHandlerById)
	}

	// DOCUMENTOS (somente admin)
	documentosGroup := router.Group("/contexto/documentos", jwt.AuthMiddleware())
	{
		documentosGroup.POST("", autosTempHandlers.PDFHandler)
		documentosGroup.GET("/all/:id", autosTempHandlers.SelectAllHandler)
		documentosGroup.DELETE("/:id", autosTempHandlers.DeleteHandler)
		documentosGroup.POST("/autua", autosTempHandlers.AutuarDocumentosHandler)
	}

	// AUTOS (somente admin)
	autosGroup := router.Group("/contexto/autos", jwt.AuthMiddleware())
	{
		autosGroup.POST("", autosHandlers.InsertHandler)
		autosGroup.GET("/all/:id", autosHandlers.SelectAllHandler)
		autosGroup.GET("/:id", autosHandlers.SelectByIdHandler)
		autosGroup.DELETE("/:id", autosHandlers.DeleteHandler)
	}

	// EVENTOS
	eventosGroup := router.Group("/contexto/eventos", jwt.AuthMiddleware())
	{
		eventosGroup.POST("", eventosHandlers.InsertHandler)
		eventosGroup.GET("/all/:id", eventosHandlers.SelectAllHandler)
		eventosGroup.GET("/:id", eventosHandlers.SelectByIdHandler)
		eventosGroup.DELETE("/:id", eventosHandlers.DeleteHandler)
	}

	// RAG (observação: faltava a barra, gerando rota incorreta)
	contextoQueryGroup := router.Group("/contexto/query", jwt.AuthMiddleware())
	{
		contextoQueryGroup.POST("/rag", contextoQueryHandlers.QueryHandlerTools)
	}

	// Chat
	router.POST("/query/chat", jwt.AuthMiddleware(), queryHandlers.QueryHandler)
}
