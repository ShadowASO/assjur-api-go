package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	//"ocrserver/models"
)

type TempAutosRow struct {
	IdDoc     int       `json:"id_doc"`
	IdCtxt    int       `json:"id_ctxt"`
	NmFileNew string    `json:"nm_file_new"`
	NmFileOri string    `json:"nm_file_ori"`
	TxtDoc    string    `json:"txt_doc"`
	DtInc     time.Time `json:"dt_inc"`
	Status    string    `json:"status"`
}

type TempautosModelType struct {
	Db *pgxpool.Pool
}

// Iniciando serviços
var TempautosModel TempautosModelType

func NewTempautosModel() *TempautosModelType {
	db, err := DBServer.GetConn()
	if err != nil {
		log.Println("NewPromptModel: Erro ao obter a conexão com o banco de dados!")
	}

	return &TempautosModelType{Db: db}
}

func (model *TempautosModelType) SelectRows() ([]TempAutosRow, error) {
	query := "SELECT * FROM temp_autos"
	rows, err := model.Db.Query(context.Background(), query)
	if err != nil {
		log.Printf("Erro ao consultar tabela temp_autos: %v", err)
		return nil, fmt.Errorf("erro ao realizar o select na tabela temp_autos: %w", err)
	}
	defer rows.Close()

	var results []TempAutosRow
	for rows.Next() {
		var row TempAutosRow
		if err := rows.Scan(&row.IdDoc, &row.IdCtxt, &row.NmFileNew, &row.NmFileOri, &row.TxtDoc, &row.DtInc, &row.Status); err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			continue
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Erro durante a iteração das linhas na tabela temp_autos: %v", err)
		return nil, fmt.Errorf("erro durante a iteração das linhas na tabela temp_autos: %w", err)
	}

	return results, nil
}

func (model *TempautosModelType) InsertRow(row TempAutosRow) (int64, error) {
	query := `
		INSERT INTO temp_autos (id_ctxt, nm_file_new, nm_file_ori, txt_doc, dt_inc, status)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id_doc;
	`
	var id int64
	ret := model.Db.QueryRow(context.Background(), query, row.IdCtxt, row.NmFileNew, row.NmFileOri, row.TxtDoc, row.DtInc, row.Status)
	if err := ret.Scan(&id); err != nil {
		log.Printf("Erro ao inserir o registro na tabela temp_autos: %v", err)
		return 0, fmt.Errorf("erro ao inserir o registro na tabela temp_autos: %w", err)
	}

	log.Println("Registro inserido com sucesso na tabela temp_autos.")
	return id, nil
}
