package models

import (
	"context"
	"fmt"
	"log"
	"ocrserver/internal/database"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PromptModelType struct {
	Db *pgxpool.Pool
}

type PromptRow struct {
	IdPrompt  int
	IdNat     int
	IdDoc     int
	IdClasse  int
	IdAssunto int
	NmDesc    string
	TxtPrompt string
	DtInc     time.Time
	Status    string
}

type BodyParamsPromptInsert struct {
	IdNat     int
	IdDoc     int
	IdClasse  int
	IdAssunto int
	NmDesc    string
	TxtPrompt string
}

type BodyParamsPromptUpdate struct {
	IdPrompt  int
	NmDesc    string
	TxtPrompt string
}

/* Constantes relacionadas ao campos do Prompt*/
const PROMPT_NATUREZA_IDENTIFICA = 1

func NewPromptModel() *PromptModelType {
	db, err := pgdb.DBServer.GetConn()
	if err != nil {
		log.Println("NewPromptModel: Erro ao obter a conexão com o banco de dados!")
	}
	return &PromptModelType{
		Db: db,
	}
}

func (model *PromptModelType) InsertReg(paramsData BodyParamsPromptInsert) (*PromptRow, error) {
	//parâmetros default
	dtInc := time.Now()
	status := "S"

	query := `INSERT INTO prompts (id_nat, id_doc, id_classe, id_assunto, nm_desc, txt_prompt, dt_inc, status) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *`
	insertedRow := model.Db.QueryRow(context.Background(), query, paramsData.IdNat, paramsData.IdDoc, paramsData.IdClasse,
		paramsData.IdAssunto, paramsData.NmDesc, paramsData.TxtPrompt, dtInc, status)

	var row PromptRow
	if err := insertedRow.Scan(&row.IdPrompt, &row.IdNat, &row.IdDoc, &row.IdClasse, &row.IdAssunto, &row.NmDesc, &row.TxtPrompt, &row.DtInc, &row.Status); err != nil {
		log.Printf("Erro ao inserir o registro na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao inserir registro: %w", err)
	}

	return &row, nil
}

func (model *PromptModelType) UpdateReg(paramsData BodyParamsPromptUpdate) (*PromptRow, error) {

	currentDate := time.Now()
	status := "S"

	query := `UPDATE prompts SET nm_desc=$1, txt_prompt=$2, dt_inc=$3, status=$4 WHERE id_prompt=$5 RETURNING *`
	updatedRow := model.Db.QueryRow(context.Background(), query, paramsData.NmDesc, paramsData.TxtPrompt, currentDate, status, paramsData.IdPrompt)

	var row PromptRow
	if err := updatedRow.Scan(&row.IdPrompt, &row.IdNat, &row.IdDoc, &row.IdClasse, &row.IdAssunto, &row.NmDesc, &row.TxtPrompt, &row.DtInc, &row.Status); err != nil {

		log.Printf("Erro ao atualizar o registro na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao atualizar registro: %w", err)
	}

	return &row, nil
}

func (model *PromptModelType) DeleteReg(idPrompt int) (*PromptRow, error) {
	query := `DELETE FROM prompts WHERE id_prompt=$1 RETURNING *`
	deletedRow := model.Db.QueryRow(context.Background(), query, idPrompt)

	var row PromptRow
	if err := deletedRow.Scan(&row.IdPrompt, &row.IdNat, &row.IdDoc, &row.IdClasse, &row.IdAssunto, &row.NmDesc, &row.TxtPrompt, &row.DtInc, &row.Status); err != nil {
		log.Printf("Erro ao deletar o registro na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao deletar registro: %w", err)
	}

	return &row, nil
}

func (model *PromptModelType) SelectById(idPrompt int) (*PromptRow, error) {
	query := `SELECT * FROM prompts WHERE id_prompt=$1`
	selectedRow := model.Db.QueryRow(context.Background(), query, idPrompt)

	var row PromptRow
	if err := selectedRow.Scan(&row.IdPrompt, &row.IdNat, &row.IdDoc, &row.IdClasse, &row.IdAssunto, &row.NmDesc, &row.TxtPrompt, &row.DtInc, &row.Status); err != nil {
		log.Printf("Erro ao selecionar o registro pelo id_prompt na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &row, nil
}

func (model *PromptModelType) SelectByNatureza(idNat int) (*PromptRow, error) {

	query := `SELECT * FROM prompts WHERE id_nat=$1`
	selectedRow := model.Db.QueryRow(context.Background(), query, idNat)

	var row PromptRow
	if err := selectedRow.Scan(&row.IdPrompt, &row.IdNat, &row.IdDoc, &row.IdClasse, &row.IdAssunto, &row.NmDesc, &row.TxtPrompt, &row.DtInc, &row.Status); err != nil {
		log.Printf("Erro ao selecionar o registro pelo id_nat na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &row, nil
}

func (model *PromptModelType) SelectRegs() ([]PromptRow, error) {
	query := `SELECT * FROM prompts`
	rows, err := model.Db.Query(context.Background(), query)
	if err != nil {
		log.Printf("Erro ao selecionar registros na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registros: %w", err)
	}
	defer rows.Close()

	results := []PromptRow{}
	for rows.Next() {
		var row PromptRow
		if err := rows.Scan(&row.IdPrompt, &row.IdNat, &row.IdDoc, &row.IdClasse, &row.IdAssunto, &row.NmDesc, &row.TxtPrompt, &row.DtInc, &row.Status); err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			continue
		}

		results = append(results, row)
	}

	return results, nil
}
