package models

import (
	"database/sql"
	"fmt"

	"ocrserver/internal/utils/logger"

	"time"
)

type DocsocrModelType struct {
	Db *sql.DB
}

type DocsocrRow struct {
	IdDoc     int       `json:"id_doc"`
	IdCtxt    string    `json:"id_ctxt"`
	NmFileNew string    `json:"nm_file_new"`
	NmFileOri string    `json:"nm_file_ori"`
	TxtDoc    string    `json:"txt_doc"`
	DtInc     time.Time `json:"dt_inc"`
	Status    string    `json:"status"`
}

// Iniciando serviços
//var TempautosModel TempautosModelType

func NewDocsocrModel(db *sql.DB) *DocsocrModelType {

	return &DocsocrModelType{Db: db}
}

/* Seleciona o documento indicado pelo ID*/
func (model *DocsocrModelType) SelectByIdDoc(idDoc int) (*DocsocrRow, error) {
	query := `SELECT * FROM docsocr WHERE id_doc = $1`
	row := model.Db.QueryRow(query, idDoc)

	var selectedRow DocsocrRow

	if err := row.Scan(&selectedRow.IdDoc, &selectedRow.IdCtxt, &selectedRow.NmFileNew, &selectedRow.NmFileOri,
		&selectedRow.TxtDoc, &selectedRow.DtInc, &selectedRow.Status); err != nil {
		logger.Log.Errorf("Erro ao selecionar o registro: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &selectedRow, nil
}

/* Seleciona todos os registros da tabela docsocr relativos ao contexto*/
func (model *DocsocrModelType) SelectByContexto(idCtxt int) ([]DocsocrRow, error) {
	query := `SELECT * FROM docsocr WHERE id_ctxt = $1`
	rows, err := model.Db.Query(query, idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar tabela docsocr: %v", err)
		return nil, fmt.Errorf("erro ao realizar o select na tabela docsocr: %w", err)
	}
	defer rows.Close()

	var results []DocsocrRow
	for rows.Next() {
		var row DocsocrRow
		if err := rows.Scan(&row.IdDoc, &row.IdCtxt, &row.NmFileNew, &row.NmFileOri, &row.TxtDoc, &row.DtInc, &row.Status); err != nil {
			logger.Log.Errorf("Erro ao escanear linha: %v", err)
			continue
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Errorf("Erro durante a iteração das linhas na tabela docsocr: %v", err)
		return nil, fmt.Errorf("erro durante a iteração das linhas na tabela docsocr: %w", err)
	}

	return results, nil
}

func (model *DocsocrModelType) InsertRow(row DocsocrRow) (int64, error) {
	query := `
		INSERT INTO docsocr (id_ctxt, nm_file_new, nm_file_ori, txt_doc, dt_inc, status)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id_doc;
	`
	var id int64
	ret := model.Db.QueryRow(query, row.IdCtxt, row.NmFileNew, row.NmFileOri, row.TxtDoc, row.DtInc, row.Status)
	if err := ret.Scan(&id); err != nil {
		logger.Log.Errorf("Erro ao inserir o registro na tabela docsocr: %v", err)
		return 0, fmt.Errorf("erro ao inserir o registro na tabela docsocr: %w", err)
	}

	//log.Println("Registro inserido com sucesso na tabela docsocr.")
	return id, nil
}

func (model *DocsocrModelType) DeleteRow(idDoc int) error {
	query := `DELETE FROM docsocr WHERE id_doc=$1`
	_, err := model.Db.Exec(query, idDoc)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar o registro na tabela docsocr: %v", err)
		return fmt.Errorf("erro ao deletar registro: %w", err)
	}

	return nil
}
