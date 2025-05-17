/*
---------------------------------------------------------------------------------------
File: config-db.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
---------------------------------------------------------------------------------------
*/
package pgdb

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

//var DBConn *sql.DB

type DBPool struct {
	Pool *sql.DB
}

var DBPoolGlobal *DBPool

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	PoolSize int
}

func NewDBConn(cfg DBConfig) (*DBPool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	//Configurando o pool de conexões
	conn.SetMaxOpenConns(cfg.PoolSize)
	conn.SetMaxIdleConns(cfg.PoolSize)
	conn.SetConnMaxIdleTime(5 * time.Minute)

	log.Printf("Conexão realizada com sucesso: %s", cfg.DBName)
	DBPoolGlobal = &DBPool{
		Pool: conn,
	}
	return DBPoolGlobal, nil
}

// Close cleanly shuts down the connection pool
func (db *DBPool) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
