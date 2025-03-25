/*
Compilação: go build -v -o server ./cmd/main.go
Execução: ./server
*/
package main

import (
	"log"

	"ocrserver/internal/auth"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"ocrserver/api/handler"
	"ocrserver/api/handler/login"
	"ocrserver/internal/config"
	"ocrserver/internal/database"

	//"ocrserver/internal/elastic"
	"ocrserver/internal/opensearch"

	"ocrserver/internal/services/cnj"
	"ocrserver/lib"

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
	fileLog := config.ConfigLog()
	defer fileLog.Close()
	// Carrego as configurações do file .env
	config.Init()

	//Cria a conexão com o banco de dados
	err := pgdb.InitializeDBServer()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer pgdb.DBServer.CloseConn()

	//Cria a conexão com o Elasticsearch
	// err = elastic.InitializeElasticServer()
	// //err = opensearch.InitializeOpenSearchServer()
	// if err != nil {
	// 	log.Fatalf("Erro ao conectar o Elasticsearch: %v", err)
	// }

	err = opensearch.InitializeOpenSearchServer()
	if err != nil {
		log.Fatalf("Erro ao conectar o Elasticsearch: %v", err)
	}

	//res, err := esCli.IndicesExists("books")
	//res, err := esCli.IndexDocumento("sentenca", elastic.BodyParamsPrompt{Natureza: "Cível", Desc: "Teste de inclusão", Conteudo: "Apenas um teste para saber"})
	// res, err := esCli.ConsultaPorConteudo("sentenca", "saber")
	// if err != nil {
	// 	log.Fatalf("Erro ao verificar existência do documento: %s", err)
	// }
	// log.Println(res)

	//Criando os Handlerss
	usersHandlers := handlers.NewUsersHandlers()
	queryHandlers := handlers.NewQueryHandlers()
	sessionHandlers := handlers.NewSessionsHandlers()
	promptHandlers := handlers.NewPromptHandlers()
	contextoHandlers := handlers.NewContextoHandlers()
	autosHandlers := handlers.NewAutosHandlers()
	uploadHandlers := handlers.NewUploadHandlers()
	docsocrHandlers := handlers.NewDocsocrHandlers()
	//elasticHandlers := handlers.NewElasticHandlers()
	openSearchHandlers := handlers.NewOpenSearchHandlers()
	contextoQueryHandlers := handlers.NewContextoQueryHandlers()

	//Cria o roteador GIN
	router := gin.Default()

	//Ativar o ReleaseMode em produção
	//gin.SetMode(gin.ReleaseMode)

	// Configura o middleware de CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     config.AllowedOrigins,                               // Origens permitidas
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
		userGroup.POST("", usersHandlers.InsertHandler)
		userGroup.GET("", usersHandlers.SelectAllHandler)
		userGroup.GET("/:id", usersHandlers.SelectHandler)
	}

	//QUERY
	router.POST("/query", queryHandlers.QueryHandler)

	//SESSIONS
	sessionGroup := router.Group("/sessions", auth.AuthenticateTokenGin())
	{
		sessionGroup.POST("", sessionHandlers.InsertHandler)
		sessionGroup.GET("", sessionHandlers.SelectAllHandler)
		sessionGroup.GET("/uso", sessionHandlers.GetTokenUsoHandler)
		sessionGroup.GET("/:id", sessionHandlers.SelectHandler)
	}

	//TABELAS
	tabelasGroup := router.Group("/tabelas", auth.AuthenticateTokenGin())
	{
		tabelasGroup.POST("/prompts", promptHandlers.InsertHandler)
		tabelasGroup.PUT("/prompts", promptHandlers.UpdateHandler)
		tabelasGroup.DELETE("/prompts/:id", promptHandlers.DeleteHandler)
		tabelasGroup.GET("/prompts", promptHandlers.SelectAllHandler)
		tabelasGroup.GET("/prompts/:id", promptHandlers.SelectByIDHandler)
	}

	//elasticGroup := router.Group("/tabelas", auth.AuthenticateTokenGin())
	//elasticGroup := router.Group("/tabelas", auth.AuthenticateTokenGin())
	// {
	// 	elasticGroup.POST("/modelos", elasticHandlers.InsertHandler)
	// 	elasticGroup.PUT("/modelos/:id", elasticHandlers.UpdateHandler)
	// 	elasticGroup.DELETE("/modelos/:id", elasticHandlers.DeleteHandler)
	// 	//Estou usando o método POST para facilitar o envio do body. Avaliar mudança para GET
	//elasticGroup.POST("/modelos/search", elasticHandlers.SearchByContentHandler)
	// 	elasticGroup.GET("/modelos/:id", elasticHandlers.SelectByIDHandler)
	// }

	openSearchGroup := router.Group("/tabelas", auth.AuthenticateTokenGin())
	{
		openSearchGroup.POST("/modelos", openSearchHandlers.InsertHandler)
		openSearchGroup.PUT("/modelos/:id", openSearchHandlers.UpdateHandler)
		openSearchGroup.DELETE("/modelos/:id", openSearchHandlers.DeleteHandler)
		// Estou usando o método POST para facilitar o envio do body. Avaliar mudança para GET
		openSearchGroup.POST("/modelos/search", openSearchHandlers.SearchByContentHandler)
		openSearchGroup.GET("/modelos/:id", openSearchHandlers.SelectByIDHandler)
	}

	//CONTEXTO
	contextoGroup := router.Group("/contexto", auth.AuthenticateTokenGin())
	{
		contextoGroup.POST("", contextoHandlers.InsertHandler)
		contextoGroup.GET("", contextoHandlers.SelectAllHandler)
		contextoGroup.GET("/:id", contextoHandlers.SelectByIDHandler)
		contextoGroup.GET("/processo/:id", contextoHandlers.SelectByProcessoHandler)

	}

	//CONTEXTO/DOCUMENTOS/UOLOAD
	uploadGroup := router.Group("/contexto/documentos/upload", auth.AuthenticateTokenGin())
	{
		uploadGroup.POST("", uploadHandlers.UploadFileHandler)
		uploadGroup.GET("/:id", uploadHandlers.SelectHandler)
		uploadGroup.DELETE("", uploadHandlers.DeleteHandler)

	}

	//CONTEXTO/DOCUMENTOS
	documentosGroup := router.Group("/contexto/documentos", auth.AuthenticateTokenGin())
	{
		documentosGroup.POST("", libocr.OcrFileHandler)
		documentosGroup.POST("/analise", autosHandlers.AutuarDocumentos)
		documentosGroup.GET("/:id", docsocrHandlers.SelectAllHandler)
		documentosGroup.DELETE("", docsocrHandlers.DeleteHandler)

	}

	//CONTEXTO/AUTOS
	autosGroup := router.Group("/contexto/autos", auth.AuthenticateTokenGin())
	{
		autosGroup.POST("", autosHandlers.InsertHandler)
		autosGroup.GET("/all/:id", autosHandlers.SelectAllHandler)
		autosGroup.GET("/:id", autosHandlers.SelectByIdHandler)
		autosGroup.DELETE("/:id", autosHandlers.DeleteHandler)

	}

	//CONTEXTO/Query
	//contextoQueryGroup := router.Group("/contexto/query", auth.AuthenticateTokenGin())
	contextoQueryGroup := router.Group("/contexto/query")
	{
		contextoQueryGroup.POST("", contextoQueryHandlers.QueryHandler)

	}

	router.POST("/upload", uploadHandlers.UploadFileHandler)
	router.GET("/ocr", libocr.OcrFileHandler)

	//router.Run(":8082")
	//router.Run(":3002")
	//Produção - A porta de execução é extraída do arquivo .env
	router.Run(config.ServerPort)

}
