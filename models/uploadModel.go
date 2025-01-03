package models

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	//"ocrserver/models"
	"time"
)

type UploadRow struct {
	IdFile    int
	IdCtxt    int
	NmFileNew string
	NmFileOri string
	SnAutos   string
	DtInc     time.Time
	Status    string
}

type UploadModelType struct {
	Db *pgxpool.Pool
}

// Iniciando serviços
var UploadModel UploadModelType

func NewUploadModel() *UploadModelType {
	db, err := DBServer.GetConn()
	if err != nil {
		log.Println("NewPromptModel: Erro ao obter a conexão com o banco de dados!")
	}

	return &UploadModelType{Db: db}
}

func (model *UploadModelType) SelectRows() ([]UploadRow, error) {
	querySql := "SELECT * FROM temp_uploadfiles"
	rows, err := model.Db.Query(context.Background(), querySql)
	if err != nil {
		log.Println("Erro ao realizar o SELECT na tabela temp_uploadfiles:", err)
		return nil, err
	}
	defer rows.Close() // Garante o fechamento dos recursos

	var results []UploadRow
	for rows.Next() {
		var row UploadRow

		// Mapeia os campos do resultado para a estrutura
		err = rows.Scan(&row.IdFile, &row.IdCtxt, &row.NmFileNew, &row.NmFileOri, &row.SnAutos, &row.DtInc, &row.Status)
		if err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			continue
		}

		results = append(results, row)
	}

	// Verifica erros durante a iteração
	if err = rows.Err(); err != nil {
		log.Printf("Erro durante a iteração das linhas na tabela temp_upload: %v", err)
		return nil, fmt.Errorf("erro durante a iteração das linhas na tabela temp_upload: %w", err)
	}

	return results, nil
}

func (model *UploadModelType) InsertRow(row UploadRow) (int64, error) {
	query := `
		INSERT INTO temp_uploadfiles ( id_ctxt, nm_file_new, nm_file_ori, sn_autos, dt_inc, status)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id_file;
	`
	var id int64

	ret := model.Db.QueryRow(context.Background(), query, row.IdCtxt, row.NmFileNew, row.NmFileOri, row.SnAutos, row.DtInc, row.Status)
	err := ret.Scan(&id)
	if err != nil {
		log.Printf("Erro ao inserir o registro na tabela temp_uploadfiles: %v", err)
		return 0, fmt.Errorf("erro ao inserir o registro na tabela temp_uploadfiles: %w", err)
	}

	log.Println("Registro inserido com sucesso na tabela temp_uploadfiles.")
	return id, err
}

func (model *UploadModelType) UpdateRow(idFile int, nmFileNew, nmFileOri, snAutos string, status string) error {
	query := `UPDATE temp_uploadfiles SET nm_file_new=$1, nm_file_ori=$2, sn_autos=$3, status=$4 WHERE id_file=$5`

	_, err := model.Db.Exec(context.Background(), query, nmFileNew, nmFileOri, snAutos, status, idFile)
	if err != nil {
		log.Printf("Erro ao atualizar o registro na tabela temp_uploadfiles: %v", err)
		return fmt.Errorf("erro ao atualizar o registro na tabela temp_uploadfiles: %w", err)
	}

	log.Println("Registro atualizado com sucesso na tabela temp_uploadfiles.")
	return nil
}

func (model *UploadModelType) DeleteRow(idFile int) error {
	query := `DELETE FROM temp_uploadfiles WHERE id_file=$1`

	_, err := model.Db.Exec(context.Background(), query, idFile)
	if err != nil {
		log.Printf("Erro ao deletar o registro na tabela temp_uploadfiles: %v", err)
		return fmt.Errorf("erro ao deletar o registro na tabela temp_uploadfiles: %w", err)
	}

	log.Println("Registro deletado com sucesso na tabela temp_uploadfiles.")
	return nil
}

func (model *UploadModelType) SelectRowById(idFile int) (*UploadRow, error) {
	query := `SELECT * FROM temp_uploadfiles WHERE id_file=$1`
	row := model.Db.QueryRow(context.Background(), query, idFile)

	var result UploadRow
	if err := row.Scan(&result.IdFile, &result.IdCtxt, &result.NmFileNew, &result.NmFileOri, &result.SnAutos, &result.DtInc, &result.Status); err != nil {
		log.Printf("Erro ao buscar o registro na tabela temp_uploadfiles: %v", err)
		return nil, fmt.Errorf("erro ao buscar o registro na tabela temp_uploadfiles: %w", err)
	}

	return &result, nil
}
