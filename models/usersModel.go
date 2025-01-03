package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersModelType struct {
	Db *pgxpool.Pool
}

type UsersRow struct {
	UserId    int       `json:"user_id"`
	Userrole  string    `json:"userrole"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUsersModel() *UsersModelType {
	db, err := DBServer.GetConn()
	if err != nil {
		log.Println("NewUsersModel: Erro ao obter a conexão com o banco de dados!")
	}
	return &UsersModelType{Db: db}
}

func (model *UsersModelType) SelectRows() ([]UsersRow, error) {
	querySql := "SELECT * FROM users"
	rows, err := model.Db.Query(context.Background(), querySql)
	if err != nil {
		log.Printf("Erro ao consultar tabela users: %v", err)
		return nil, fmt.Errorf("erro ao realizar o select na tabela users: %w", err)
	}
	defer rows.Close()

	var results []UsersRow
	for rows.Next() {
		var row UsersRow
		if err := rows.Scan(&row.UserId, &row.Userrole, &row.Username, &row.Password, &row.Email, &row.CreatedAt); err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			continue
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Erro durante a iteração das linhas na tabela users: %v", err)
		return nil, fmt.Errorf("erro durante a iteração das linhas na tabela users: %w", err)
	}

	return results, nil
}

func (model *UsersModelType) SelectRow(userID int) (*UsersRow, error) {
	querySql := "SELECT user_id, userrole, username, password, email, created_at FROM users WHERE user_id = $1"
	row := model.Db.QueryRow(context.Background(), querySql, userID)

	var user UsersRow
	if err := row.Scan(&user.UserId, &user.Userrole, &user.Username, &user.Password, &user.Email, &user.CreatedAt); err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("nenhum usuário encontrado com o ID %d", userID)
		}
		log.Printf("Erro ao escanear linha: %v", err)
		return nil, fmt.Errorf("erro ao escanear linha: %w", err)
	}

	return &user, nil
}

func (model *UsersModelType) SelectUserByName(username string) (*UsersRow, error) {
	querySql := "SELECT user_id, userrole, username, password, email, created_at FROM users WHERE username = $1"
	row := model.Db.QueryRow(context.Background(), querySql, username)

	var user UsersRow
	if err := row.Scan(&user.UserId, &user.Userrole, &user.Username, &user.Password, &user.Email, &user.CreatedAt); err != nil {
		if err.Error() == "no rows in result set" {
			log.Printf("Nenhum usuário encontrado com o nome '%s'", username)
			return nil, nil
		}
		log.Printf("Erro ao escanear linha: %v", err)
		return nil, fmt.Errorf("erro ao escanear linha: %w", err)
	}

	return &user, nil
}

func (model *UsersModelType) InsertRow(row UsersRow) (int64, error) {
	query := `
		INSERT INTO users (userrole, username, password, email, created_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING user_id;
	`
	var userID int64

	ret := model.Db.QueryRow(context.Background(), query, row.Userrole, row.Username, row.Password, row.Email, row.CreatedAt)
	if err := ret.Scan(&userID); err != nil {
		log.Printf("Erro ao inserir o registro na tabela users: %v", err)
		return 0, fmt.Errorf("erro ao inserir o registro na tabela users: %w", err)
	}

	log.Println("Registro inserido com sucesso na tabela users.")
	return userID, nil
}
