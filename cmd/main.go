// cmd/server/main.go
// ---------------------------------------------------------------------------------------
// Autor: Aldenor – refatorado com ajustes de robustez e segurança
// Inspiração: Enterprise Applications with Gin
// Data: 26-12-2024 | Refatoração: 11-08-2025
// ---------------------------------------------------------------------------------------
// Compilação: go build -v -o server ./cmd/server
// Execução:   ./server
// ---------------------------------------------------------------------------------------
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"ocrserver/internal/config"
	"ocrserver/internal/database/pgdb"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/rotas"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"
)

func main() {
	// 1) Config e logger
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Inicia logger global o quanto antes
	//logger.InitLoggerGlobal("logs/app.log", true)
	// Produção (arquivo com rotação + stdout)
	logger.InitLoggerGlobal("./logs/app.log", true)

	// Opcional: ajustar nível em runtime
	logger.SetGlobalLevelFromEnv() // lê LOG_LEVEL
	// ou: logger.SetGlobalLevel(logger.DebugLevel)

	logger.Log.Info("Iniciando servidor...")

	// 2) Definição do modo do Gin **antes** de criar o router
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 3) Conexões externas (DB, OpenSearch, serviços)
	// Banco de Dados
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
		log.Fatalf("erro ao criar pool de conexões com o database: %v", err)
	}
	defer db.Close()

	// OpenSearch
	if err := opensearch.InitOpenSearchService(); err != nil {
		log.Fatalf("erro ao conectar ao OpenSearch: %v", err)
	}

	// Serviços globais (ex.: CNJ)
	services.InitCnjGlobal(cfg)
	services.InitOpenaiService(cfg.OpenApiKey, cfg) // idempotente caso sem chave

	// 4) Router e middlewares
	router := gin.New()
	// Evita warnings de proxy e reforça segurança (ajuste se usar proxy de verdade)
	_ = router.SetTrustedProxies(nil)

	// logger padrão do gin só no modo debug
	if gin.Mode() == gin.DebugMode {
		router.Use(gin.Logger())
	}

	// Middlewares essenciais
	router.Use(gin.Recovery())
	//router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RequestIDMiddleware())

	// CORS configurável
	corsCfg := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsCfg))

	// 5) Rotas de negócio (injeta cfg e DB)
	rotas.SetRotasSistema(router, cfg, db)

	// 6) Servidor HTTP com shutdown gracioso
	addr := cfg.ServerPort
	// Aceita tanto "9001" quanto ":9001" no .env
	if len(addr) > 0 && addr[0] != ':' {
		addr = ":" + addr
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Canal para sinais do SO
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Log.Infof("Servidor ouvindo em %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Errorf("erro ao iniciar servidor: %v", err)
		}
	}()

	// Bloqueia até receber sinal de encerramento
	<-done
	logger.Log.Info("Recebido sinal de encerramento. Finalizando...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Errorf("shutdown com erro: %v", err)
	} else {
		logger.Log.Info("shutdown concluído com sucesso")
	}

	fmt.Println("bye 👋")
}
