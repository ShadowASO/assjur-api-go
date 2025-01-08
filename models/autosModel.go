package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// type AutosRow struct {
// 	IdAutos   int       `json:"id_autos"`
// 	IdCtxt    int       `json:"id_ctxt"`
// 	IdNat     int       `json:"id_nat"`
// 	IdPje     string    `json:"id_pje"`
// 	DtPje     time.Time `json:"dt_pje"`
// 	AutosJson string    `json:"autos_json"`
// 	DtInc     time.Time `json:"dt_inc"`
// 	Status    string    `json:"status"`
// }

type AutosRow struct {
	IdAutos   int
	IdCtxt    int
	IdNat     int
	IdPje     string
	DtPje     time.Time
	AutosJson string
	DtInc     time.Time
	Status    string
}

type AutosModelType struct {
	Db *pgxpool.Pool
}

func NewAutosModel() *AutosModelType {
	db, err := DBServer.GetConn()
	if err != nil {
		log.Println("NewPromptModel: Erro ao obter a conexão com o banco de dados!")
	}

	return &AutosModelType{Db: db}
}

func (model *AutosModelType) InsertRow(rowData AutosRow) (*AutosRow, error) {
	currentDate := time.Now()
	status := "S"

	query := `INSERT INTO autos (id_ctxt, id_nat, id_pje, dt_pje, autos_json, dt_inc, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`
	row := model.Db.QueryRow(context.Background(), query, rowData.IdCtxt, rowData.IdNat, rowData.IdPje, currentDate, rowData.AutosJson, currentDate, status)

	var insertedRow AutosRow
	if err := row.Scan(&insertedRow.IdAutos, &insertedRow.IdCtxt, &insertedRow.IdNat, &insertedRow.IdPje, &insertedRow.DtPje, &insertedRow.AutosJson, &insertedRow.DtInc, &insertedRow.Status); err != nil {
		log.Printf("Erro ao inserir o registro na tabela autos: %v", err)
		return nil, fmt.Errorf("erro ao inserir registro: %w", err)
	}

	return &insertedRow, nil
}

func (model *AutosModelType) UpdateRow(rowData AutosRow) (*AutosRow, error) {
	status := "S"
	query := `UPDATE autos SET autos_json=$1, status=$2 WHERE id_autos=$3 RETURNING *`
	row := model.Db.QueryRow(context.Background(), query, rowData.AutosJson, status, rowData.IdAutos)

	var updatedRow AutosRow
	if err := row.Scan(&updatedRow.IdAutos, &updatedRow.IdCtxt, &updatedRow.IdNat, &updatedRow.IdPje, &updatedRow.DtPje, &updatedRow.AutosJson, &updatedRow.DtInc, &updatedRow.Status); err != nil {
		log.Printf("Erro ao atualizar o registro na tabela autos: %v", err)
		return nil, fmt.Errorf("erro ao atualizar registro: %w", err)
	}

	return &updatedRow, nil
}

func (model *AutosModelType) DeleteRow(idAutos int) error {
	query := `DELETE FROM autos WHERE id_autos=$1`
	_, err := model.Db.Exec(context.Background(), query, idAutos)
	if err != nil {
		log.Printf("Erro ao deletar o registro na tabela autos: %v", err)
		return fmt.Errorf("erro ao deletar registro: %w", err)
	}

	return nil
}

func (model *AutosModelType) IsDocAutuado(idCtxt int, idPje string) (bool, error) {
	query := `SELECT * FROM autos WHERE id_ctxt = $1 AND id_pje = $2`
	rows, err := model.Db.Query(context.Background(), query, idCtxt, idPje)
	if err != nil {
		log.Printf("Erro ao verificar documento autuado: %v", err)
		return false, fmt.Errorf("erro ao verificar documento: %w", err)
	}
	defer rows.Close()

	return rows.Next(), nil
}

func (model *AutosModelType) SelectByContexto(idCtxt int) ([]AutosRow, error) {
	query := `SELECT * FROM autos WHERE id_ctxt = $1`
	rows, err := model.Db.Query(context.Background(), query, idCtxt)
	if err != nil {

		return nil, fmt.Errorf("falha ao executar consulta para contexto %d: %w", idCtxt, err)
	}
	defer rows.Close()

	// Armazena os resultados
	var results []AutosRow

	for rows.Next() {
		var row AutosRow
		if err := rows.Scan(&row.IdAutos, &row.IdCtxt, &row.IdNat, &row.IdPje, &row.DtPje, &row.AutosJson, &row.DtInc, &row.Status); err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			return nil, fmt.Errorf("falha ao escanear resultados: %w", err)
		}
		results = append(results, row)
	}
	// Verifica se houve algum erro na iteração das linhas
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro durante a iteração das linhas: %w", err)
	}

	return results, nil
}

func (model *AutosModelType) SelectById(idAutos int) (*AutosRow, error) {
	query := `SELECT * FROM autos WHERE id_autos = $1`
	row := model.Db.QueryRow(context.Background(), query, idAutos)

	var selectedRow AutosRow
	if err := row.Scan(&selectedRow.IdAutos, &selectedRow.IdCtxt, &selectedRow.IdNat, &selectedRow.IdPje, &selectedRow.DtPje, &selectedRow.AutosJson, &selectedRow.DtInc, &selectedRow.Status); err != nil {
		log.Printf("Erro ao selecionar o registro: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &selectedRow, nil
}
