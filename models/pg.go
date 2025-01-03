package models

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBServerType struct {
	pool *pgxpool.Pool
}

// Iniciando serviços
var DBServer DBServerType

// Instância global para compartilhamento
// var pool *pgxpool.Pool
var (
	//pool *pgxpool.Pool
	once sync.Once
)

// Coneção com a VPS
const connStr = "host=191.101.71.18 port=7432 user=assjurpg dbname=assjurdb password=Assjur@vps sslmode=disable"

// Conexão com a PS local - O acesso ao Postgres no container o host=localhost
//const connStr = "host=localhost port=5432 user=assjurpg dbname=assjurdb password=Assjur@vps sslmode=disable"

func InitializeDBServer() error {
	var err error
	once.Do(func() {
		var pool *pgxpool.Pool
		pool, err = pgxpool.New(context.Background(), connStr)
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
