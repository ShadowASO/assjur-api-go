package models

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx"
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
//var UploadModel UploadModelType

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
	// Consulta SQL para deletar o registro
	query := `DELETE FROM temp_uploadfiles WHERE id_file = $1`

	// Executa a consulta
	result, err := model.Db.Exec(context.Background(), query, idFile)
	if err != nil {
		log.Printf("Erro ao deletar o registro na tabela temp_uploadfiles para id_file=%d: %v", idFile, err)
		return fmt.Errorf("erro ao deletar o registro na tabela temp_uploadfiles: %w", err)
	}

	// Verifica se alguma linha foi afetada
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Nenhum registro encontrado para id_file=%d na tabela temp_uploadfiles", idFile)
		return fmt.Errorf("nenhum registro encontrado para id_file=%d", idFile)
	}

	log.Printf("Registro com id_file=%d deletado com sucesso na tabela temp_uploadfiles.", idFile)
	return nil
}

func (model *UploadModelType) SelectRowById(idFile int) (*UploadRow, error) {
	// Consulta especificando as colunas necessárias
	query := `SELECT id_file, id_ctxt, nm_file_new, nm_file_ori, sn_autos, dt_inc, status 
	          FROM temp_uploadfiles 
	          WHERE id_file = $1`

	// Executa a consulta
	row := model.Db.QueryRow(context.Background(), query, idFile)

	// Prepara a estrutura para o resultado
	var result UploadRow

	// Faz o scan do resultado
	if err := row.Scan(&result.IdFile, &result.IdCtxt, &result.NmFileNew, &result.NmFileOri, &result.SnAutos, &result.DtInc, &result.Status); err != nil {
		if err == pgx.ErrNoRows { // Trata o caso de nenhum registro encontrado
			log.Printf("Nenhum registro encontrado para id_file=%d", idFile)
			return nil, nil // Ou retorne um erro específico, se preferir
		}
		log.Printf("Erro ao buscar o registro na tabela temp_uploadfiles: %v", err)
		return nil, fmt.Errorf("erro ao buscar o registro na tabela temp_uploadfiles: %w", err)
	}

	return &result, nil
}

func (model *UploadModelType) SelectRowsByContextoId(idCtxt int) ([]UploadRow, error) {
	//query := `SELECT * FROM temp_uploadfiles WHERE id_ctxt=$1`
	// Define a consulta SQL
	query := `SELECT id_file, id_ctxt, nm_file_new, nm_file_ori, sn_autos, dt_inc, status 
	          FROM temp_uploadfiles 
	          WHERE id_ctxt = $1`
	//row := model.Db.QueryRow(context.Background(), query, idCtxt)
	// Executa a consulta retornando todas as linhas
	rows, err := model.Db.Query(context.Background(), query, idCtxt)
	if err != nil {
		log.Printf("Erro ao executar a consulta: %v", err)
		return nil, fmt.Errorf("erro ao executar a consulta: %w", err)
	}
	defer rows.Close()

	// Itera sobre as linhas retornadas
	var results []UploadRow
	for rows.Next() {
		var row UploadRow
		// Faz o scan dos campos na estrutura
		if err := rows.Scan(&row.IdFile, &row.IdCtxt, &row.NmFileNew, &row.NmFileOri, &row.SnAutos, &row.DtInc, &row.Status); err != nil {
			log.Printf("Erro ao escanear os resultados: %v", err)
			return nil, fmt.Errorf("erro ao escanear os resultados: %w", err)
		}
		results = append(results, row)
	}

	// Verifica erros adicionais na iteração
	if err := rows.Err(); err != nil {
		log.Printf("Erro na iteração das linhas: %v", err)
		return nil, fmt.Errorf("erro na iteração das linhas: %w", err)
	}

	return results, nil
}
