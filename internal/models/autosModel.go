package models

import (
	"database/sql"
	"ocrserver/internal/consts"

	"fmt"
	"log"
)

type AutosModelType struct {
	Db *sql.DB
}

func NewAutosModel(db *sql.DB) *AutosModelType {
	return &AutosModelType{Db: db}
}

func (model *AutosModelType) InsertRow(Data consts.AutosRow) (*consts.AutosRow, error) {

	query := `INSERT INTO autos (id_ctxt, id_nat, id_pje, dt_pje, autos_json, dt_inc, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`
	//row := model.Db.QueryRow(query, Data.IdCtxt, Data.IdNatu, Data.IdPje, Data.DtInc, Data.DocJson, Data.DtInc, "S")
	row := model.Db.QueryRow(query, Data.IdCtxt, Data.IdNatu, Data.IdPje, Data.DocJson, "S")

	var dataRow consts.AutosRow
	if err := row.Scan(&dataRow.Id, &dataRow.IdCtxt, &dataRow.IdNatu, &dataRow.IdPje, &dataRow.DocJson, "S"); err != nil {
		log.Printf("Erro ao inserir o registro na tabela autos: %v", err)
		return nil, fmt.Errorf("erro ao inserir registro: %w", err)
	}

	return &dataRow, nil
}

func (model *AutosModelType) UpdateRow(rowData consts.AutosRow) (*consts.AutosRow, error) {
	status := "S"
	query := `UPDATE autos SET autos_json=$1, status=$2 WHERE id_autos=$3 RETURNING *`
	row := model.Db.QueryRow(query, rowData.DocJson, status, rowData.Id)

	var updatedRow consts.AutosRow
	if err := row.Scan(&updatedRow.Id, &updatedRow.IdCtxt, &updatedRow.IdNatu, &updatedRow.IdPje, &updatedRow.DocJson, "S"); err != nil {
		log.Printf("Erro ao atualizar o registro na tabela autos: %v", err)
		return nil, fmt.Errorf("erro ao atualizar registro: %w", err)
	}

	return &updatedRow, nil
}

func (model *AutosModelType) DeleteRow(idAutos int) error {
	query := `DELETE FROM autos WHERE id_autos=$1`
	_, err := model.Db.Exec(query, idAutos)
	if err != nil {
		log.Printf("Erro ao deletar o registro na tabela autos: %v", err)
		return fmt.Errorf("erro ao deletar registro: %w", err)
	}

	return nil
}

func (model *AutosModelType) IsDocAutuado(idCtxt int, idPje string) (bool, error) {

	// Verifica os argumentos de entrada
	if idCtxt <= 0 || idPje == "" {
		return false, fmt.Errorf("parâmetros inválidos: idCtxt=%d, idPje=%q", idCtxt, idPje)
	}

	// Consulta simplificada para verificar a existência do registro
	query := `SELECT EXISTS(SELECT 1 FROM autos WHERE id_ctxt = $1 AND id_pje = $2)`
	var exists bool

	// Executa a consulta e verifica erros
	err := model.Db.QueryRow(query, idCtxt, idPje).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("erro ao verificar documento autuado: %w", err)
	}

	return exists, nil
}

func (model *AutosModelType) SelectByContexto(idCtxt int) ([]consts.AutosRow, error) {
	query := `SELECT * FROM autos WHERE id_ctxt = $1`
	rows, err := model.Db.Query(query, idCtxt)
	if err != nil {

		return nil, fmt.Errorf("falha ao executar consulta para contexto %d: %w", idCtxt, err)
	}
	defer rows.Close()

	// Armazena os resultados
	var results []consts.AutosRow

	for rows.Next() {
		var row consts.AutosRow
		if err := rows.Scan(&row.Id, &row.IdCtxt, &row.IdNatu, &row.IdPje, &row.DocJson, "S"); err != nil {
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

func (model *AutosModelType) SelectById(idAutos int) (*consts.AutosRow, error) {
	query := `SELECT * FROM autos WHERE id_autos = $1`
	//log.Printf("Executando query: %s com parâmetros: %v", query, idAutos)

	row := model.Db.QueryRow(query, idAutos)

	var selectedRow consts.AutosRow
	err := row.Scan(&selectedRow.Id, &selectedRow.IdCtxt, &selectedRow.IdNatu, &selectedRow.IdPje, &selectedRow.DocJson, "S")
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Nenhum registro encontrado para id_autos:", idAutos)
			return nil, nil
		}
		log.Printf("Erro ao selecionar o registro: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &selectedRow, nil
}
