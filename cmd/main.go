package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"ocrserver/auth"
	"ocrserver/config"
	"ocrserver/controllers"
	"ocrserver/controllers/login"
	"ocrserver/lib"
	"ocrserver/models"
	"ocrserver/services/cnj"
	"os"
	"time"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()
		c.Next()
		duration := time.Since(start)
		log.Printf("Request - Method: %s | Status: %d | Duration: %v", c.Request.Method, c.Writer.Status(), duration)
	}
}

func main() {
	// Configurar saída do log
	log.SetOutput(os.Stdout)
	// Carrego as configurações do file .env
	config.Init()

	//Cria a conexão com o banco de dados
	err := models.InitializeDBServer()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer models.DBServer.CloseConn()

	//Criando os Controllers
	usersController := controllers.NewUsersController()
	queryController := controllers.NewQueryController()
	sessionController := controllers.NewSessionsController()
	promptController := controllers.NewPromptController()
	contextoController := controllers.NewContextoController()
	autosController := controllers.NewAutosController()
	uploadController := controllers.NewUploadController()
	tempautosController := controllers.NewTempautosController()

	//Cria o roteador GIN
	router := gin.Default()

	//Ativar o ReleaseMode em produção
	//gin.SetMode(gin.ReleaseMode)

	// Configura o middleware de CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3002"},                   // Origens permitidas
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Métodos permitidos
		AllowHeaders:     []string{"Content-Type", "Authorization"},           // Cabeçalhos permitidos
		ExposeHeaders:    []string{"Content-Length"},                          // Cabeçalhos expostos ao cliente
		AllowCredentials: true,                                                // Permitir credenciais
		MaxAge:           12 * time.Hour,                                      // Cache da política de CORS
	}))

	//AUTH - Rotas para login e geração/validação de tokens
	router.POST("/auth/login", login.LoginHandler)
	router.POST("/auth/token/refresh", login.RefreshTokenHandler)
	router.POST("/auth/token/verify", login.VerifyTokenHandler)

	//CNJ
	router.POST("/cnj/processo", cnj.GetProcessoFromCnj)

	//USERS - ok
	userGroup := router.Group("/users", auth.AuthenticateTokenGin())
	{
		userGroup.POST("", usersController.InsertHandler)
		userGroup.GET("", usersController.SelectAllHandler)
		userGroup.GET("/:id", usersController.SelectHandler)
	}

	//QUERY
	router.POST("/query", queryController.QueryHandler)

	//SESSIONS
	sessionGroup := router.Group("/sessions", auth.AuthenticateTokenGin())
	{
		sessionGroup.POST("", sessionController.InsertHandler)
		sessionGroup.GET("", sessionController.SelectAllHandler)
		sessionGroup.GET("/uso", sessionController.GetTokenUsoHandler)
		sessionGroup.GET("/:id", sessionController.SelectHandler)
	}

	//TABELAS
	tabelasGroup := router.Group("/tabelas", auth.AuthenticateTokenGin())
	{
		tabelasGroup.POST("/prompts", promptController.InsertHandler)
		tabelasGroup.PUT("/prompts", promptController.UpdateHandler)
		tabelasGroup.DELETE("/prompts/:id", promptController.DeleteHandler)
		tabelasGroup.GET("/prompts", promptController.SelectAllHandler)
		tabelasGroup.GET("/prompts/:id", promptController.SelectByIDHandler)
	}

	//CONTEXTO
	contextoGroup := router.Group("/contexto", auth.AuthenticateTokenGin())
	{
		contextoGroup.POST("", contextoController.InsertHandler)
		contextoGroup.GET("", contextoController.SelectAllHandler)
		contextoGroup.GET("/:id", contextoController.SelectByIDHandler)
		contextoGroup.GET("/processo/:id", contextoController.SelectByProcessoHandler)

	}

	//CONTEXTO/DOCUMENTOS/UOLOAD
	uploadGroup := router.Group("/contexto/documentos/upload", auth.AuthenticateTokenGin())
	{
		uploadGroup.POST("", uploadController.UploadFileHandler)
		uploadGroup.GET("/:id", uploadController.SelectHandler)
		uploadGroup.DELETE("", uploadController.DeleteHandler)

	}

	//CONTEXTO/DOCUMENTOS
	documentosGroup := router.Group("/contexto/documentos", auth.AuthenticateTokenGin())
	{
		documentosGroup.POST("", libocr.OcrFileHandler)
		documentosGroup.POST("/analise", autosController.AutuarDocumentos)

		documentosGroup.GET("/:id", tempautosController.SelectAllHandler)
		documentosGroup.DELETE("", uploadController.DeleteHandler)

	}

	//CONTEXTO/AUTOS
	autosGroup := router.Group("/contexto/autos", auth.AuthenticateTokenGin())
	{
		autosGroup.POST("", autosController.InsertHandler)
		autosGroup.GET("/:id", autosController.SelectAllHandler)

	}

	router.POST("/upload", uploadController.UploadFileHandler)
	router.GET("/ocr", libocr.OcrFileHandler)
	//router.GET("/lista", auth.AuthenticateTokenGin(), uploadServices.ListaUploadFileHandler)

	router.Run(":8082")
	//router.Run(":3002")

}
