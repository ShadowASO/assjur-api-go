package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionsRow struct {
	SessionID        int
	UserID           int
	Model            string
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
	SessionStart     time.Time
	SessionEnd       *time.Time
}

type SessionsModelType struct {
	Db *pgxpool.Pool
}

// Iniciando serviços
var SessionsModel SessionsModelType

func NewSessionsModel() *SessionsModelType {
	db, err := DBServer.GetConn()
	if err != nil {
		log.Println("NewPromptModel: Erro ao obter a conexão com o banco de dados!")
	}

	return &SessionsModelType{Db: db}
}

func (model *SessionsModelType) SelectSessions() ([]SessionsRow, error) {
	query := `SELECT * FROM sessions`
	rows, err := model.Db.Query(context.Background(), query)
	if err != nil {
		log.Printf("Erro na seleção de sessões: %v", err)
		return nil, fmt.Errorf("erro ao selecionar sessões: %w", err)
	}
	defer rows.Close()

	var sessions []SessionsRow
	for rows.Next() {
		var session SessionsRow
		if err := rows.Scan(&session.SessionID, &session.UserID, &session.Model, &session.PromptTokens, &session.CompletionTokens, &session.TotalTokens, &session.SessionStart, &session.SessionEnd); err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (model *SessionsModelType) SelectSession(id int) (*SessionsRow, error) {
	query := `SELECT * FROM sessions WHERE session_id = $1`
	row := model.Db.QueryRow(context.Background(), query, id)

	var session SessionsRow
	if err := row.Scan(&session.SessionID, &session.UserID, &session.Model, &session.PromptTokens, &session.CompletionTokens, &session.TotalTokens, &session.SessionStart, &session.SessionEnd); err != nil {
		log.Printf("Erro na seleção da sessão: %v", err)
		return nil, fmt.Errorf("erro ao selecionar sessão: %w", err)
	}

	return &session, nil
}

func (model *SessionsModelType) InsertSession(data SessionsRow) (int, error) {
	query := `INSERT INTO sessions (user_id, model, prompt_tokens, completion_tokens, total_tokens) VALUES ($1, $2, $3, $4, $5) RETURNING session_id`
	row := model.Db.QueryRow(context.Background(), query, data.UserID, data.Model, data.PromptTokens, data.CompletionTokens, data.TotalTokens)

	var sessionID int
	if err := row.Scan(&sessionID); err != nil {
		log.Printf("Erro na inserção da sessão: %v", err)
		return 0, fmt.Errorf("erro ao inserir sessão: %w", err)
	}

	return sessionID, nil
}

func (model *SessionsModelType) UpdateSession(data SessionsRow) (*SessionsRow, error) {

	query := `UPDATE sessions SET  prompt_tokens = $1, completion_tokens = $2, total_tokens = $3 WHERE session_id = $4`
	_, err := model.Db.Exec(context.Background(), query, data.PromptTokens, data.CompletionTokens, data.TotalTokens, data.SessionID)
	if err != nil {
		log.Printf("Erro na atualização da sessão: %v", err)
		return nil, fmt.Errorf("erro ao atualizar sessão: %w", err)
	}

	// Retornar sessão atualizada
	return &SessionsRow{
		SessionID:        data.SessionID,
		UserID:           data.UserID,
		Model:            data.Model,
		PromptTokens:     data.PromptTokens,
		CompletionTokens: data.CompletionTokens,
		TotalTokens:      data.TotalTokens,
	}, nil
}

func (model *SessionsModelType) SelectSessionTokensUsage(id int) (*SessionsRow, error) {
	return model.SelectSession(id)
}
