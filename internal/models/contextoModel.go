package models

import (
	"database/sql"
	"fmt"
	"log"

	"time"
)

type ContextoModelType struct {
	Db *sql.DB
}

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

type BodyParamsContextoInsert struct {
	NrProc  string
	Juizo   string
	Classe  string
	Assunto string
}
type BodyParamsContextoUpdate struct {
	IdCtxt           int
	NrProc           string
	Juizo            string
	Classe           string
	Assunto          string
	PromptTokens     int
	CompletionTokens int
}

func NewContextoModel(db *sql.DB) *ContextoModelType {

	return &ContextoModelType{Db: db}
}

func (model *ContextoModelType) InsertRow(paramsData BodyParamsContextoInsert) (*ContextoRow, error) {
	currentDate := time.Now()
	promptTokens := 0
	completionTokens := 0

	query := `INSERT INTO contexto (nr_proc, juizo, classe, assunto, prompt_tokens, completion_tokens, dt_inc) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`
	insertedRow := model.Db.QueryRow(query, paramsData.NrProc, paramsData.Juizo, paramsData.Classe, paramsData.Assunto, promptTokens, completionTokens, currentDate)

	var row ContextoRow
	if err := insertedRow.Scan(&row.IdCtxt, &row.NrProc, &row.Juizo, &row.Classe, &row.Assunto, &row.PromptTokens, &row.CompletionTokens, &row.DtInc, &row.Status); err != nil {
		log.Printf("Erro ao inserir o registro na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao inserir registro: %w", err)
	}

	return &row, nil
}

func (model *ContextoModelType) UpdateRow(paramsData BodyParamsContextoUpdate) (*ContextoRow, error) {
	query := `UPDATE contexto SET nr_proc=$1, juizo=$2, classe=$3, assunto=$4, prompt_tokens=$5, completion_tokens=$6 WHERE id_ctxt=$7 RETURNING *`
	updatedRow := model.Db.QueryRow(query, paramsData.NrProc, paramsData.Juizo, paramsData.Classe, paramsData.Assunto, paramsData.PromptTokens, paramsData.CompletionTokens, paramsData.IdCtxt)

	var row ContextoRow
	if err := updatedRow.Scan(&row.IdCtxt, &row.NrProc, &row.Juizo, &row.Classe, &row.Assunto, &row.PromptTokens, &row.CompletionTokens, &row.DtInc, &row.Status); err != nil {
		log.Printf("Erro ao atualizar o registro na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao atualizar registro: %w", err)
	}

	return &row, nil
}

func (model *ContextoModelType) IncrementTokensAtomic(
	idCtxt int,
	promptTokensInc int,
	completionTokensInc int,
) (*ContextoRow, error) {

	query := `
        UPDATE contexto
        SET
            prompt_tokens = prompt_tokens + $1,
            completion_tokens = completion_tokens + $2
        WHERE id_ctxt = $3
        RETURNING id_ctxt, nr_proc, juizo, classe, assunto, prompt_tokens, completion_tokens, dt_inc, status
    `

	updatedRow := model.Db.QueryRow(
		query,
		promptTokensInc,
		completionTokensInc,
		idCtxt,
	)

	var row ContextoRow
	if err := updatedRow.Scan(
		&row.IdCtxt,
		&row.NrProc,
		&row.Juizo,
		&row.Classe,
		&row.Assunto,
		&row.PromptTokens,
		&row.CompletionTokens,
		&row.DtInc,
		&row.Status,
	); err != nil {
		log.Printf("Erro ao incrementar tokens na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao incrementar tokens: %w", err)
	}

	return &row, nil
}

func (model *ContextoModelType) DeleteReg(idCtxt int) (*ContextoRow, error) {
	query := `DELETE FROM contexto WHERE id_ctxt=$1 RETURNING *`
	deletedRow := model.Db.QueryRow(query, idCtxt)

	var row ContextoRow
	if err := deletedRow.Scan(&row.IdCtxt, &row.NrProc, &row.Juizo, &row.Classe, &row.Assunto, &row.PromptTokens, &row.CompletionTokens, &row.DtInc, &row.Status); err != nil {
		log.Printf("Erro ao deletar o registro na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao deletar registro: %w", err)
	}
	return &row, nil
}

func (model *ContextoModelType) RowExists(nrProc string) (bool, error) {
	query := `SELECT * FROM contexto WHERE nr_proc=$1`
	selectedRow := model.Db.QueryRow(query, nrProc)

	var row ContextoRow
	if err := selectedRow.Scan(&row.IdCtxt, &row.NrProc, &row.Juizo, &row.Classe, &row.Assunto, &row.PromptTokens, &row.CompletionTokens, &row.DtInc, &row.Status); err != nil {

		log.Printf("Processo não cadastrado!")
		return false, nil
	}
	log.Printf("Processo já cadastrado!")
	return true, nil
}

func (model *ContextoModelType) SelectContextoById(idCtxt int) (*ContextoRow, error) {
	query := `SELECT * FROM contexto WHERE id_ctxt=$1`
	selectedRow := model.Db.QueryRow(query, idCtxt)

	var row ContextoRow
	if err := selectedRow.Scan(&row.IdCtxt, &row.NrProc, &row.Juizo, &row.Classe, &row.Assunto, &row.PromptTokens, &row.CompletionTokens, &row.DtInc, &row.Status); err != nil {
		log.Printf("Erro ao selecionar o registro na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &row, nil
}

func (model *ContextoModelType) SelectContextoByProcesso(nrProc string) (*ContextoRow, error) {
	query := `SELECT * FROM contexto WHERE nr_proc=$1`
	rows := model.Db.QueryRow(query, nrProc)

	var row ContextoRow

	if err := rows.Scan(&row.IdCtxt, &row.NrProc, &row.Juizo, &row.Classe, &row.Assunto, &row.PromptTokens, &row.CompletionTokens, &row.DtInc, &row.Status); err != nil {

		log.Printf("Erro ao selecionar o contexto pelo processo: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &row, nil
}

func (model *ContextoModelType) SelectContextoByProcessoStartsWith(nrProcPart string) ([]ContextoRow, error) {
	query := `SELECT id_ctxt, nr_proc, juizo, classe, assunto, prompt_tokens, completion_tokens, dt_inc, status
	          FROM contexto
	          WHERE nr_proc LIKE $1`

	// Busca por registros cujo nr_proc começa com nrProcPart
	rows, err := model.Db.Query(query, nrProcPart+"%")
	if err != nil {
		log.Printf("Erro ao executar consulta LIKE no contexto: %v", err)
		return nil, fmt.Errorf("erro ao executar consulta: %w", err)
	}
	defer rows.Close()

	var resultados []ContextoRow

	for rows.Next() {
		var row ContextoRow
		if err := rows.Scan(
			&row.IdCtxt,
			&row.NrProc,
			&row.Juizo,
			&row.Classe,
			&row.Assunto,
			&row.PromptTokens,
			&row.CompletionTokens,
			&row.DtInc,
			&row.Status,
		); err != nil {
			log.Printf("Erro ao ler linha do contexto: %v", err)
			return nil, fmt.Errorf("erro ao ler resultado: %w", err)
		}
		resultados = append(resultados, row)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Erro durante iteração das linhas: %v", err)
		return nil, fmt.Errorf("erro na iteração dos resultados: %w", err)
	}

	return resultados, nil
}

func (model *ContextoModelType) SelectContextos_ant() ([]ContextoRow, error) {
	query := `SELECT * FROM contexto`
	rows, err := model.Db.Query(query)
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

func (model *ContextoModelType) SelectContextos(limit, offset int) ([]ContextoRow, error) {
	query := `SELECT id_ctxt, nr_proc, juizo, classe, assunto, prompt_tokens, completion_tokens, dt_inc, status
	          FROM contexto
	          ORDER BY dt_inc DESC
	          LIMIT $1 OFFSET $2`

	rows, err := model.Db.Query(query, limit, offset)
	if err != nil {
		log.Printf("Erro ao selecionar registros na tabela contexto: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registros: %w", err)
	}
	defer rows.Close()

	var results []ContextoRow
	for rows.Next() {
		var row ContextoRow
		if err := rows.Scan(
			&row.IdCtxt,
			&row.NrProc,
			&row.Juizo,
			&row.Classe,
			&row.Assunto,
			&row.PromptTokens,
			&row.CompletionTokens,
			&row.DtInc,
			&row.Status,
		); err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			continue
		}
		results = append(results, row)
	}

	return results, nil
}
