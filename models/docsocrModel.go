package models

import (
	"context"
	"fmt"
	"log"
	"ocrserver/internal/database"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TempautosModelType struct {
	Db *pgxpool.Pool
}

type TempAutosRow struct {
	IdDoc     int
	IdCtxt    int
	NmFileNew string
	NmFileOri string
	TxtDoc    string
	DtInc     time.Time
	Status    string
}

// Iniciando serviços
//var TempautosModel TempautosModelType

func NewTempautosModel() *TempautosModelType {
	db, err := pgdb.DBServer.GetConn()
	if err != nil {
		log.Println("NewPromptModel: Erro ao obter a conexão com o banco de dados!")
	}

	return &TempautosModelType{Db: db}
}

/* Seleciona o documento indicado pelo ID*/
func (model *TempautosModelType) SelectByIdDoc(idDoc int) (*TempAutosRow, error) {
	query := `SELECT * FROM docsocr WHERE id_doc = $1`
	row := model.Db.QueryRow(context.Background(), query, idDoc)

	var selectedRow TempAutosRow

	if err := row.Scan(&selectedRow.IdDoc, &selectedRow.IdCtxt, &selectedRow.NmFileNew, &selectedRow.NmFileOri,
		&selectedRow.TxtDoc, &selectedRow.DtInc, &selectedRow.Status); err != nil {
		log.Printf("Erro ao selecionar o registro: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &selectedRow, nil
}

/* Seleciona todos os registros da tabela docsocr relativos ao contexto*/
func (model *TempautosModelType) SelectByContexto(idCtxt int) ([]TempAutosRow, error) {
	query := `SELECT * FROM docsocr WHERE id_ctxt = $1`
	rows, err := model.Db.Query(context.Background(), query, idCtxt)
	if err != nil {
		log.Printf("Erro ao consultar tabela docsocr: %v", err)
		return nil, fmt.Errorf("erro ao realizar o select na tabela docsocr: %w", err)
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
		log.Printf("Erro durante a iteração das linhas na tabela docsocr: %v", err)
		return nil, fmt.Errorf("erro durante a iteração das linhas na tabela docsocr: %w", err)
	}

	return results, nil
}

func (model *TempautosModelType) InsertRow(row TempAutosRow) (int64, error) {
	query := `
		INSERT INTO docsocr (id_ctxt, nm_file_new, nm_file_ori, txt_doc, dt_inc, status)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id_doc;
	`
	var id int64
	ret := model.Db.QueryRow(context.Background(), query, row.IdCtxt, row.NmFileNew, row.NmFileOri, row.TxtDoc, row.DtInc, row.Status)
	if err := ret.Scan(&id); err != nil {
		log.Printf("Erro ao inserir o registro na tabela docsocr: %v", err)
		return 0, fmt.Errorf("erro ao inserir o registro na tabela docsocr: %w", err)
	}

	log.Println("Registro inserido com sucesso na tabela docsocr.")
	return id, nil
}

func (model *TempautosModelType) DeleteRow(idDoc int) error {
	query := `DELETE FROM docsocr WHERE id_doc=$1`
	_, err := model.Db.Exec(context.Background(), query, idDoc)
	if err != nil {
		log.Printf("Erro ao deletar o registro na tabela docsocr: %v", err)
		return fmt.Errorf("erro ao deletar registro: %w", err)
	}

	return nil
}
