package models

import (
	"context"
	"fmt"
	"log"
	//"ocrserver/models"

	//"ocrserver/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PromptRow struct {
	IdPrompt  int       `json:"id_prompt"`
	IdNat     int       `json:"id_nat"`
	IdDoc     int       `json:"id_doc"`
	IdClasse  int       `json:"id_classe"`
	IdAssunto int       `json:"id_assunto"`
	NmDesc    string    `json:"nm_desc"`
	TxtPrompt string    `json:"txt_prompt"`
	DtInc     time.Time `json:"dt_inc"`
	Status    string    `json:"status"`
}

type PromptModelType struct {
	Db *pgxpool.Pool
}

// Iniciando serviços
//var PromptModel PromptModelType

func NewPromptModel() *PromptModelType {
	db, err := DBServer.GetConn()
	if err != nil {
		log.Println("NewPromptModel: Erro ao obter a conexão com o banco de dados!")
	}
	return &PromptModelType{
		Db: db,
	}
}

// func (p *PromptService) InsertReg(idNat, idDoc, idClasse, idAssunto int, nmDesc, txtPrompt string) (*PromptRow, error) {
func (model *PromptModelType) InsertReg(rowData PromptRow) (*PromptRow, error) {
	rowData.DtInc = time.Now()
	rowData.Status = "S"

	query := `INSERT INTO tab_prompts (id_nat, id_doc, id_classe, id_assunto, nm_desc, txt_prompt, dt_inc, status) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *`
	row := model.Db.QueryRow(context.Background(), query, rowData.IdNat, rowData.IdDoc, rowData.IdClasse,
		rowData.IdAssunto, rowData.NmDesc, rowData.TxtPrompt, rowData.DtInc, rowData.Status)

	var insertedRow PromptRow
	if err := row.Scan(&insertedRow.IdPrompt, &insertedRow.IdNat, &insertedRow.IdDoc, &insertedRow.IdClasse, &insertedRow.IdAssunto, &insertedRow.NmDesc, &insertedRow.TxtPrompt, &insertedRow.DtInc, &insertedRow.Status); err != nil {
		log.Printf("Erro ao inserir o registro na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao inserir registro: %w", err)
	}

	return &insertedRow, nil
}

// func (p *PromptModelType) UpdateReg(idPrompt int, nmDesc, txtPrompt string) (*PromptRow, error) {
func (model *PromptModelType) UpdateReg(dataRow PromptRow) (*PromptRow, error) {
	currentDate := time.Now()
	status := "S"

	query := `UPDATE tab_prompts SET nm_desc=$1, txt_prompt=$2, dt_inc=$3, status=$4 WHERE id_prompt=$5 RETURNING *`
	row := model.Db.QueryRow(context.Background(), query, dataRow.NmDesc, dataRow.TxtPrompt, currentDate, status, dataRow.IdPrompt)

	var updatedRow PromptRow
	if err := row.Scan(&updatedRow.IdPrompt, &updatedRow.IdNat, &updatedRow.IdDoc, &updatedRow.IdClasse, &updatedRow.IdAssunto, &updatedRow.NmDesc, &updatedRow.TxtPrompt, &updatedRow.DtInc, &updatedRow.Status); err != nil {
		log.Printf("Erro ao atualizar o registro na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao atualizar registro: %w", err)
	}

	return &updatedRow, nil
}

func (model *PromptModelType) DeleteReg(idPrompt int) (*PromptRow, error) {
	query := `DELETE FROM tab_prompts WHERE id_prompt=$1 RETURNING *`
	row := model.Db.QueryRow(context.Background(), query, idPrompt)

	var deletedRow PromptRow
	if err := row.Scan(&deletedRow.IdPrompt, &deletedRow.IdNat, &deletedRow.IdDoc, &deletedRow.IdClasse, &deletedRow.IdAssunto, &deletedRow.NmDesc, &deletedRow.TxtPrompt, &deletedRow.DtInc, &deletedRow.Status); err != nil {
		log.Printf("Erro ao deletar o registro na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao deletar registro: %w", err)
	}

	return &deletedRow, nil
}

func (model *PromptModelType) SelectById(idPrompt int) (*PromptRow, error) {
	query := `SELECT * FROM tab_prompts WHERE id_prompt=$1`
	row := model.Db.QueryRow(context.Background(), query, idPrompt)

	var selectedRow PromptRow
	if err := row.Scan(&selectedRow.IdPrompt, &selectedRow.IdNat, &selectedRow.IdDoc, &selectedRow.IdClasse, &selectedRow.IdAssunto, &selectedRow.NmDesc, &selectedRow.TxtPrompt, &selectedRow.DtInc, &selectedRow.Status); err != nil {
		log.Printf("Erro ao selecionar o registro pelo id_prompt na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &selectedRow, nil
}

func (model *PromptModelType) SelectByNatureza(idNat int) (*PromptRow, error) {
	query := `SELECT * FROM tab_prompts WHERE id_nat=$1`
	row := model.Db.QueryRow(context.Background(), query, idNat)

	var selectedRow PromptRow
	if err := row.Scan(&selectedRow.IdPrompt, &selectedRow.IdNat, &selectedRow.IdDoc, &selectedRow.IdClasse, &selectedRow.IdAssunto, &selectedRow.NmDesc, &selectedRow.TxtPrompt, &selectedRow.DtInc, &selectedRow.Status); err != nil {
		log.Printf("Erro ao selecionar o registro pelo id_nat na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &selectedRow, nil
}

func (model *PromptModelType) SelectRegs() ([]PromptRow, error) {
	query := `SELECT * FROM tab_prompts`
	rows, err := model.Db.Query(context.Background(), query)
	if err != nil {
		log.Printf("Erro ao selecionar registros na tabela prompts: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registros: %w", err)
	}
	defer rows.Close()

	var results []PromptRow
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
