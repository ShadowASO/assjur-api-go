package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ContextoRow struct {
	IdCtxt           int       `json:"id_ctxt"`
	NrProc           string    `json:"nr_proc"`
	Juizo            string    `json:"juizo"`
	Classe           string    `json:"classe"`
	Assunto          string    `json:"assunto"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	DtInc            time.Time `json:"dt_inc"`
	Status           string    `json:"status"`
}

type ContextoModelType struct {
	Db *pgxpool.Pool
}

// Iniciando serviços
var ContextoModel ContextoModelType

func NewContextoModel() *ContextoModelType {
	db, err := DBServer.GetConn()
	if err != nil {
		log.Println("NewPromptModel: Erro ao obter a conexão com o banco de dados!")
	}

	return &ContextoModelType{Db: db}
}

// func (p *ContextoModelType) InitService() error {
// 	//db, err := models.GetConn()
// 	db, err := DBServer.GetConn()
// 	if err != nil {
// 		return err
// 	}
// 	//Services = PromptService{Db: db}
// 	ContextoModel.Db = db
// 	return nil
// }

func (c *ContextoModelType) InsertRow(nrProc, juizo, classe, assunto string) (*ContextoRow, error) {
	currentDate := time.Now()
	promptTokens := 0
	completionTokens := 0

	query := `INSERT INTO contexto (nr_proc, juizo, classe, assunto, prompt_tokens, completion_tokens, dt_inc) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`
	row := c.Db.QueryRow(context.Background(), query, nrProc, juizo, classe, assunto, promptTokens, completionTokens, currentDate)

	var insertedRow ContextoRow
	if err := row.Scan(&insertedRow.IdCtxt, &insertedRow.NrProc, &insertedRow.Juizo, &insertedRow.Classe, &insertedRow.Assunto, &insertedRow.PromptTokens, &insertedRow.CompletionTokens, &insertedRow.DtInc, &insertedRow.Status); err != nil {
		log.Printf("Erro ao inserir o registro na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao inserir registro: %w", err)
	}

	return &insertedRow, nil
}

func (c *ContextoModelType) UpdateRow(idCtxt int, nrProc, juizo, classe, assunto string, promptTokens, completionTokens int) (*ContextoRow, error) {
	query := `UPDATE contexto SET nr_proc=$1, juizo=$2, classe=$3, assunto=$4, prompt_tokens=$5, completion_tokens=$6 WHERE id_ctxt=$7 RETURNING *`
	row := c.Db.QueryRow(context.Background(), query, nrProc, juizo, classe, assunto, promptTokens, completionTokens, idCtxt)

	var updatedRow ContextoRow
	if err := row.Scan(&updatedRow.IdCtxt, &updatedRow.NrProc, &updatedRow.Juizo, &updatedRow.Classe, &updatedRow.Assunto, &updatedRow.PromptTokens, &updatedRow.CompletionTokens, &updatedRow.DtInc, &updatedRow.Status); err != nil {
		log.Printf("Erro ao atualizar o registro na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao atualizar registro: %w", err)
	}

	return &updatedRow, nil
}

func (c *ContextoModelType) UpdateTokens(idCtxt, promptTokens, completionTokens int) (*ContextoRow, error) {
	query := `UPDATE contexto SET prompt_tokens=$1, completion_tokens=$2 WHERE id_ctxt=$3 RETURNING *`
	row := c.Db.QueryRow(context.Background(), query, promptTokens, completionTokens, idCtxt)

	var updatedRow ContextoRow
	if err := row.Scan(&updatedRow.IdCtxt, &updatedRow.NrProc, &updatedRow.Juizo, &updatedRow.Classe, &updatedRow.Assunto, &updatedRow.PromptTokens, &updatedRow.CompletionTokens, &updatedRow.DtInc, &updatedRow.Status); err != nil {
		log.Printf("Erro ao atualizar tokens na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao atualizar tokens: %w", err)
	}

	return &updatedRow, nil
}

func (c *ContextoModelType) RowExists(nrProc string) (bool, error) {
	query := `SELECT 1 FROM contexto WHERE nr_proc=$1`
	row := c.Db.QueryRow(context.Background(), query, nrProc)

	var exists int
	if err := row.Scan(&exists); err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}
		log.Printf("Erro ao verificar existência do registro na tabela contexto: %v", err)
		return false, fmt.Errorf("erro ao verificar existência do registro: %w", err)
	}

	return exists == 1, nil
}

func (c *ContextoModelType) SelectContextoById(idCtxt int) (*ContextoRow, error) {
	query := `SELECT * FROM contexto WHERE id_ctxt=$1`
	row := c.Db.QueryRow(context.Background(), query, idCtxt)

	var selectedRow ContextoRow
	if err := row.Scan(&selectedRow.IdCtxt, &selectedRow.NrProc, &selectedRow.Juizo, &selectedRow.Classe, &selectedRow.Assunto, &selectedRow.PromptTokens, &selectedRow.CompletionTokens, &selectedRow.DtInc, &selectedRow.Status); err != nil {
		log.Printf("Erro ao selecionar o registro na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &selectedRow, nil
}

func (c *ContextoModelType) SelectContextoByProcesso(nrProc string) ([]ContextoRow, error) {
	query := `SELECT * FROM contexto WHERE nr_proc=$1`
	rows, err := c.Db.Query(context.Background(), query, nrProc)
	if err != nil {
		log.Printf("Erro ao selecionar registros na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registros: %w", err)
	}
	defer rows.Close()

	var results []ContextoRow
	for rows.Next() {
		var row ContextoRow
		if err := rows.Scan(&row.IdCtxt, &row.NrProc, &row.Juizo, &row.Classe, &row.Assunto, &row.PromptTokens, &row.CompletionTokens, &row.DtInc, &row.Status); err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			continue
		}
		results = append(results, row)
	}

	return results, nil
}

func (c *ContextoModelType) SelectContextos() ([]ContextoRow, error) {
	query := `SELECT * FROM contexto`
	rows, err := c.Db.Query(context.Background(), query)
	if err != nil {
		log.Printf("Erro ao selecionar registros na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registros: %w", err)
	}
	defer rows.Close()

	var results []ContextoRow
	for rows.Next() {
		var row ContextoRow
		if err := rows.Scan(&row.IdCtxt, &row.NrProc, &row.Juizo, &row.Classe, &row.Assunto, &row.PromptTokens, &row.CompletionTokens, &row.DtInc, &row.Status); err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			continue
		}
		results = append(results, row)
	}

	return results, nil
}
