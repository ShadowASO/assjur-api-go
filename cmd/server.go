// cmd/server.go
// ---------------------------------------------------------------------------------------
// Autor: Aldenor – refatorado com ajustes de robustez e segurança
// Inspiração: Enterprise Applications with Gin
// Data: 26-12-2024 | Refatoração: 11-08-2025
// ---------------------------------------------------------------------------------------
// Compilação: go build -v -o server ./cmd/server.go
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

	"github.com/gin-gonic/gin"

	"ocrserver/internal/config"

	"ocrserver/internal/middleware"
	"ocrserver/internal/models/opensearch"
	pgdb "ocrserver/internal/models/postgres/conn"
	"ocrserver/internal/pkg/msclientegrpc"
	"ocrserver/internal/services/grpc_services/authgrpc"
	"ocrserver/internal/services/openai"

	"ocrserver/internal/services/workers"
	"ocrserver/internal/utils/mslogger"

	"ocrserver/internal/rotas"
	"ocrserver/internal/services"
)

func main() {

	/* CONFIG: Carrega as configurações a partir do .env */
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("erro ao carregar configuração: %v", err))
	}
	/* GIN MODE: Fixa o modo de funcionamento do GIN */
	gin.SetMode(cfg.GinMode)

	/* LOGGER: Faz a inicialização do Logger global */
	err = mslogger.InitGlobal(mslogger.Options{
		FilePath:   "./logs/app.log",
		Stdout:     true,
		Rotate:     true,
		MaxSizeMB:  20,
		MaxBackups: 10,
		MaxAgeDays: 30,
		Compress:   true,
		Level:      mslogger.DebugLevel,
		JSON:       true,
		Service:    "auth-srv",
		AddSource:  true,
	})
	if err != nil {
		panic(err)
	}
	// Encerramento do Logger global deferido
	defer func() {
		if mslogger.LoggerGlobal != nil {
			mslogger.LoggerGlobal.InfoData("app encerrado", mslogger.AppLogData{
				Context: "shutdown",
			})

			_ = mslogger.LoggerGlobal.Close()
		}
	}()

	/* Insere mensagem no logger de inicialização*/
	mslogger.LoggerGlobal.InfoData("app iniciou", mslogger.AppLogData{
		Context: "startup",
		Mode:    gin.Mode(),
		Env:     config.GlobalConfig.ApplicationMode,
	})

	//************************************************************

	/* POSTGRESQL - Configura e cria uma conexta ao PostgreSQL*/
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
		mslogger.LoggerGlobal.ErrorErr("erro ao criar pool de conexões com PostgreSQL", err)
		return
	}

	defer func() {
		db.Close()

		mslogger.LoggerGlobal.InfoData("PostgreSQL desconectado com sucesso", mslogger.AppLogData{
			Context: "shutdown",
		})
	}()
	//******************************************************************

	// OPENSEARCH: Inicializa o OpenSearch
	if err := opensearch.InitOpenSearchService(); err != nil {
		log.Fatalf("erro ao conectar ao OpenSearch: %v", err)
	}

	// Serviços globais (ex.: CNJ, OpenAI)
	services.InitCnjGlobal(cfg)
	services.InitOpenaiService(cfg.OpenApiKey, cfg) // idempotente caso sem chave
	openai.InitOpenai(cfg.OpenApiKey, cfg)          // idempotente caso sem chave

	// gRPC - Cliente do microsserviço de autenticação
	authClient, err := authgrpc.New(msclientegrpc.ConfigClienteGRPC{
		Name:    "auth-srv",
		Host:    cfg.AuthGRPCHost,
		Port:    cfg.AuthGRPCPort,
		Timeout: 5 * time.Second,
		Debug:   cfg.AuthClientDebug,
	})
	if err != nil {
		panic(err)
	}
	defer authClient.Close()

	//PING no Auth-srv
	authClient.Ping(context.Background())

	defer func() {
		if err := authClient.Close(); err != nil {

			mslogger.LoggerGlobal.Errorf("erro ao fechar cliente gRPC de autenticação: %v", err)
			return
		}

		mslogger.LoggerGlobal.Info("shutdown concluído com sucesso.")
	}()

	//******************************************************

	/* Cria um router do GIN*/
	router := gin.New()
	_ = router.SetTrustedProxies(nil)

	/* Faz a atribuição do Middleware ao router. */
	router.Use(
		middleware.Logging(),
		middleware.RequestIDMiddleware(),
		middleware.ConfigureCors(cfg.AllowedOrigins),
		gin.Recovery(),
		gin.Recovery(),
	)

	// 5) Rotas de negócio (injeta cfg e DB)
	rotas.SetRotasSistema(router, cfg, db, authClient)

	/***** Serviço de limpeza de autos_temp */

	cleaner := workers.NewAutosTempCleaner(services.AutosTempServiceGlobal)

	// ctx geral do app (cancele no shutdown)
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Iniciar o serviço
	cleaner.Start(appCtx)

	/**************************************/

	// 6) Servidor HTTP com shutdown gracioso
	addr := cfg.ServerPort
	// Aceita tanto "4001" quanto ":4001" no .env
	if len(addr) > 0 && addr[0] != ':' {
		addr = ":" + addr
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}
	srv.ErrorLog = log.New(os.Stderr, "[http] ", log.LstdFlags|log.Lshortfile)

	// Canal para sinais do SO
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		mslogger.LoggerGlobal.Infof("Servidor ouvindo em %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mslogger.LoggerGlobal.Errorf("erro ao iniciar servidor: %v", err)
		}
	}()

	// Bloqueia até receber sinal de encerramento
	<-done

	mslogger.LoggerGlobal.Info("Recebido sinal de encerramento. Finalizando...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		mslogger.LoggerGlobal.Errorf("shutdown com erro: %v", err)
	} else {
		mslogger.LoggerGlobal.Info("shutdown concluído com sucesso")
	}

	fmt.Println("bye 👋")
}
