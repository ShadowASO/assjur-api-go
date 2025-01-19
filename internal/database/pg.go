package pgdb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"ocrserver/internal/config"
)

type DBServerType struct {
	pool *pgxpool.Pool
}

// Iniciando serviços
var DBServer DBServerType

// Instância global para compartilhamento
var (
	once sync.Once
)

/*
Retorna string composta a partir das variáveis de ambiente do arquivo .env
*/
func getConfigPostgreSQL() string {
	host := config.PostgresHost
	port := config.PostgresPort
	user := config.PostgresUser
	password := config.PostgresPassword
	dbname := config.PostgresDB

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	return connStr
}

func InitializeDBServer() error {

	var err error
	once.Do(func() {
		var pool *pgxpool.Pool

		pool, err = pgxpool.New(context.Background(), getConfigPostgreSQL())
		if err != nil {
			log.Println("Failed to initialize PgServer:", err)
			return
		}
		DBServer.pool = pool
	})
	return err
}

// InitDB inicializa a conexão com o banco de dados.
func (pg *DBServerType) ConnectDB() (*pgxpool.Pool, error) {
	//var err error
	connStr := getConfigPostgreSQL()
	if connStr == "" {
		return nil, errors.New("variável de ambiente DB_CONN_STRING não definida")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Printf("Erro ao analisar a configuração do banco: %v", err)
		return nil, fmt.Errorf("erro ao analisar configuração: %w", err)
	}

	DBServer.pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Printf("Erro ao conectar ao pool do banco de dados: %v", err)
		return nil, fmt.Errorf("erro ao conectar ao pool: %w", err)
	}

	log.Println("Conexão com o pool do banco de dados estabelecida com sucesso.")
	return DBServer.pool, nil
}

func (pg *DBServerType) GetConn() (*pgxpool.Pool, error) {
	var err error
	once.Do(func() {
		DBServer.pool, err = DBServer.ConnectDB()
	})
	if err != nil {
		log.Printf("Erro ao obter conexão com o banco: %v", err)
		return nil, fmt.Errorf("erro ao obter conexão com o banco: %w", err)
	}
	return DBServer.pool, nil
}

func (pg *DBServerType) CloseConn() {
	if DBServer.pool != nil {
		DBServer.pool.Close()
		log.Println("Conexão com o banco de dados fechada com sucesso.")
	}
}
