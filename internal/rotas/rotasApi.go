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
	//contextoModel := models.NewContextoModel(db.Pool)
	uploadModel := models.NewUploadModel(db.Pool)

	// --- OpenSearch Indexes ---
	indexModelos := opensearch.NewIndexModelos()
	autosIndex := opensearch.NewAutosIndex()
	autosTempIndex := opensearch.NewAutos_tempIndex()
	autosJSONEmbedding := opensearch.NewAutosJsonEmbedding()
	eventosIdx := opensearch.NewEventosIndex()
	baseIndex := opensearch.NewBaseIndex()
	contextoIndex := opensearch.NewContextoIndex()

	// --- SERVICES ---
	userService := services.NewUsersService(userModel)
	autosService := services.NewAutosService(autosIndex)
	autosTempService := services.NewAutos_tempService(autosTempIndex)
	uploadService := services.NewUploadService(uploadModel)
	promptService := services.NewPromptService(promptModel)

	// contextoService := services.NewContextoService(contextoModel)
	contextoService := services.NewContextoService(contextoIndex)

	queryService := services.NewQueryService(sessionsModel)
	sessionService := services.NewSessionService(sessionsModel)
	cnjService := services.NewCnjService(cfg)
	loginService := services.NewLoginService(cfg)
	services.InitEventosService(eventosIdx)
	baseService := services.NewBaseService(baseIndex)

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
	baseHandlers := handlers.NewBaseHandlers(baseService)
	eventosHandlers := handlers.NewEventosHandlers(services.EventosServiceGlobal)

	// --- Objetos/Serviços globais (quando realmente necessários) ---
	//opensearch.InitIndexService(indexModelos)
	services.InitSessionService(sessionsModel)
	services.InitAutosService(autosIndex)
	services.InitAutos_tempService(autosTempIndex)
	services.InitUsersService(userModel)
	services.InitPromptService(promptModel)
	//services.InitContextoService(contextoModel)
	services.InitContextoService(contextoIndex)
	services.InitUploadService(uploadModel)
	services.InitAutosJsonService(autosJSONEmbedding)
	opensearch.InitModelosService()
	opensearch.InitBaseIndex()
	services.InitBaseService(baseIndex)

	// --- ROTAS PÚBLICAS ---
	verApi := router.Group("/api/v1")
	{

		verApi.GET("/sys/version", handlers.VersionHandler)

		// Auth
		verApi.POST("/auth/login", loginHandlers.LoginHandler)
		verApi.POST("/auth/register", usersHandlers.InsertHandler)
		verApi.POST("/auth/token/refresh", loginHandlers.RefreshTokenHandler)
		verApi.POST("/auth/token/verify", loginHandlers.VerifyTokenHandler)

		// CNJ
		verApi.POST("/cnj/processo", cnjService.GetProcessoFromCnj)

		// --- ROTAS PROTEGIDAS ---

		// USERS
		userGroup := verApi.Group("/users", jwt.AuthMiddleware())
		{
			userGroup.POST("", usersHandlers.InsertHandler)
			userGroup.GET("", usersHandlers.SelectAllHandler)
			userGroup.GET("/:id", usersHandlers.SelectHandler)
		}

		// SESSIONS
		sessionGroup := verApi.Group("/sessions", jwt.AuthMiddleware())
		{
			sessionGroup.POST("", sessionHandlers.InsertHandler)
			sessionGroup.GET("", sessionHandlers.SelectAllHandler)
			sessionGroup.GET("/uso", sessionHandlers.GetTokenUsoHandler)
			sessionGroup.GET("/:id", sessionHandlers.SelectHandler)
		}

		// TABELAS (somente admin)
		tabelasGroup := verApi.Group("/tabelas", jwt.AuthMiddleware(), jwt.AuthorizeMiddleware("admin"))
		{
			tabelasGroup.POST("/prompts", promptHandlers.InsertHandler)
			tabelasGroup.PUT("/prompts", promptHandlers.UpdateHandler)
			tabelasGroup.GET("/prompts", promptHandlers.SelectAllHandler)
			tabelasGroup.GET("/prompts/:id", promptHandlers.SelectByIDHandler)
			tabelasGroup.DELETE("/prompts/:id", promptHandlers.DeleteHandler)
		}

		// OpenSearch (modelos)
		openSearchGroup := verApi.Group("/tabelas", jwt.AuthMiddleware())
		{
			openSearchGroup.POST("/modelos", openSearchHandlers.InsertHandler)
			openSearchGroup.PUT("/modelos/:id", openSearchHandlers.UpdateHandler)
			openSearchGroup.DELETE("/modelos/:id", openSearchHandlers.DeleteHandler)
			openSearchGroup.POST("/modelos/search", openSearchHandlers.SearchModelosHandler)
			openSearchGroup.GET("/modelos/:id", openSearchHandlers.SelectByIdHandler)

			// CRUD da Base de Conhecimentos para rag
			openSearchGroup.POST("/base", baseHandlers.InsertHandler)
			openSearchGroup.PUT("/base/:id", baseHandlers.UpdateHandler)
			openSearchGroup.DELETE("/base/:id", baseHandlers.DeleteHandler)
			openSearchGroup.POST("/base/search", baseHandlers.SearchHandler)
			openSearchGroup.GET("/base/:id", baseHandlers.SelectByIdHandler)
		}

		// CONTEXTO (somente admin)
		contextoGroup := verApi.Group("/contexto", jwt.AuthMiddleware())
		{
			contextoGroup.POST("", contextoHandlers.InsertHandler)
			contextoGroup.PUT("/:id", contextoHandlers.UpdateHandler)
			contextoGroup.GET("", contextoHandlers.SelectAllHandler)
			contextoGroup.GET("/:id", contextoHandlers.SelectByIDHandler)
			contextoGroup.GET("/search/:id", contextoHandlers.SelectByIdCtxtHandler)

			contextoGroup.GET("/processo/:id", contextoHandlers.SelectByProcessoHandler)
			contextoGroup.POST("/processo/search", contextoHandlers.SearchByProcessoHandler)
			contextoGroup.DELETE("/:id", contextoHandlers.DeleteHandler)
			contextoGroup.GET("/tokens/uso/:id", contextoHandlers.SelectTokenUsoHandler) // confere se este handler é o correto
		}

		// API para fazer o upload, listagem e exclusão do arquivo PDF extraído do PJe
		// Criada uma rota específica no NGINX

		uploadGroup := verApi.Group("/contexto/documentos/upload", jwt.AuthMiddleware())
		{
			uploadGroup.POST("", uploadHandlers.UploadFileHandler)
			uploadGroup.GET("/:id", uploadHandlers.SelectHandler)
			uploadGroup.DELETE("/:id", uploadHandlers.DeleteHandlerById)
		}

		// API para a extração das peças processuais, consulta, exclusão e autuação nos autos.
		// Atua sobre os índices "autos_temp" e "autos".
		documentosGroup := verApi.Group("/contexto/documentos", jwt.AuthMiddleware())
		{
			documentosGroup.POST("", autosTempHandlers.PDFHandler)
			documentosGroup.GET("/all/:id", autosTempHandlers.SelectAllHandler)
			documentosGroup.DELETE("/:id", autosTempHandlers.DeleteHandler)
			documentosGroup.POST("/autua", autosTempHandlers.AutuarDocumentosHandler)
		}

		// API - CRUD do index "autos"
		autosGroup := verApi.Group("/contexto/autos", jwt.AuthMiddleware())
		{
			autosGroup.POST("", autosHandlers.InsertHandler)
			autosGroup.GET("/all/:id", autosHandlers.SelectAllHandler)
			autosGroup.GET("/:id", autosHandlers.SelectByIdHandler)
			autosGroup.DELETE("/:id", autosHandlers.DeleteHandler)
		}

		// CRUD dos eventos gerados na análise jurídica: análise jurídica, minuta de sentença etc
		eventosGroup := verApi.Group("/contexto/eventos", jwt.AuthMiddleware())
		{
			eventosGroup.POST("", eventosHandlers.InsertHandler)
			eventosGroup.GET("/all/:id", eventosHandlers.SelectAllHandler)
			eventosGroup.GET("/:id", eventosHandlers.SelectByIdHandler)
			eventosGroup.DELETE("/:id", eventosHandlers.DeleteHandler)
		}

		// Análise Jurídica - O prompt da janela aciona esta API
		contextoQueryGroup := verApi.Group("/contexto/query", jwt.AuthMiddleware())
		{
			contextoQueryGroup.POST("/analise", contextoQueryHandlers.QueryHandlerPipeline)
		}

		// Chat - bate-papo
		verApi.POST("/query/chat", jwt.AuthMiddleware(), queryHandlers.QueryHandler)
	}
}
