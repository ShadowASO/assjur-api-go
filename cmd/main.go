package main

import (
	"log"
	"ocrserver/auth"
	"ocrserver/config"
	"ocrserver/controllers"
	"ocrserver/controllers/login"
	"ocrserver/services/cnj"

	"ocrserver/lib"
	"ocrserver/models"

	"ocrserver/services"

	"os"
	"time"

	"github.com/gin-gonic/gin"
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
	sessionController := controllers.NewSessionController()
	promptController := controllers.NewPromptController()

	//Cria o roteador GIN
	router := gin.Default()

	//Aplico o middleware
	//router.Use(LoggerMiddleware())

	//Rotas criadas para os recursos disponíveis
	router.POST("/auth/login", login.LoginHandler)
	router.POST("/auth/token/refresh", login.RefreshTokenHandler)
	//CNJ
	router.POST("/cnj/processo", cnj.GetProcessoFromCnj)

	//USERS
	userGroup := router.Group("/users", auth.AuthenticateTokenGin())
	{
		userGroup.POST("/", usersController.InsertHandler)
		userGroup.GET("/", usersController.SelectAllHandler)
		userGroup.GET("/:id", usersController.SelectHandler)
	}

	//QUERY
	router.POST("/query", queryController.QueryHandler)

	//SESSION
	sessionGroup := router.Group("/session", auth.AuthenticateTokenGin())
	{
		sessionGroup.POST("/", sessionController.InsertHandler)
		sessionGroup.GET("/", sessionController.SelectAllHandler)
		sessionGroup.GET("/uso/:id", sessionController.SelectHandler)
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

	router.POST("/upload", uploadServices.UploadFileHandler)
	router.GET("/ocr", libocr.OcrFileHandler)
	router.GET("/lista", auth.AuthenticateTokenGin(), uploadServices.ListaUploadFileHandler)

	router.Run(":8082")
	//router.Run(":3002")

}
