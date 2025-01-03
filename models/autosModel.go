package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AutosRow struct {
	IdAutos   int       `json:"id_autos"`
	IdCtxt    int       `json:"id_ctxt"`
	IdNat     int       `json:"id_nat"`
	IdPje     string    `json:"id_pje"`
	DtPje     time.Time `json:"dt_pje"`
	AutosJson string    `json:"autos_json"`
	DtInc     time.Time `json:"dt_inc"`
	Status    string    `json:"status"`
}

type AutosModelType struct {
	Db *pgxpool.Pool
}

// Iniciando serviços
var AutosModel AutosModelType

func NewAutosModel() *AutosModelType {
	db, err := DBServer.GetConn()
	if err != nil {
		log.Println("NewPromptModel: Erro ao obter a conexão com o banco de dados!")
	}

	return &AutosModelType{Db: db}
}

// func (p *AutosModelType) InitService() error {
// 	//db, err := models.GetConn()
// 	db, err := DBServer.GetConn()
// 	if err != nil {
// 		return err
// 	}
// 	//Services = PromptService{Db: db}
// 	AutosModel.Db = db
// 	return nil
// }

func (a *AutosModelType) InsertRow(idCtxt int, idNat int, idPje string, autosJson string) (*AutosRow, error) {
	currentDate := time.Now()
	status := "S"

	query := `INSERT INTO autos (id_ctxt, id_nat, id_pje, dt_pje, autos_json, dt_inc, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`
	row := a.Db.QueryRow(context.Background(), query, idCtxt, idNat, idPje, currentDate, autosJson, currentDate, status)

	var insertedRow AutosRow
	if err := row.Scan(&insertedRow.IdAutos, &insertedRow.IdCtxt, &insertedRow.IdNat, &insertedRow.IdPje, &insertedRow.DtPje, &insertedRow.AutosJson, &insertedRow.DtInc, &insertedRow.Status); err != nil {
		log.Printf("Erro ao inserir o registro na tabela autos: %v", err)
		return nil, fmt.Errorf("erro ao inserir registro: %w", err)
	}

	return &insertedRow, nil
}

func (a *AutosModelType) UpdateRow(idAutos int, autosJson string) (*AutosRow, error) {
	status := "S"
	query := `UPDATE autos SET autos_json=$1, status=$2 WHERE id_autos=$3 RETURNING *`
	row := a.Db.QueryRow(context.Background(), query, autosJson, status, idAutos)

	var updatedRow AutosRow
	if err := row.Scan(&updatedRow.IdAutos, &updatedRow.IdCtxt, &updatedRow.IdNat, &updatedRow.IdPje, &updatedRow.DtPje, &updatedRow.AutosJson, &updatedRow.DtInc, &updatedRow.Status); err != nil {
		log.Printf("Erro ao atualizar o registro na tabela autos: %v", err)
		return nil, fmt.Errorf("erro ao atualizar registro: %w", err)
	}

	return &updatedRow, nil
}

func (a *AutosModelType) DeleteRow(idAutos int) error {
	query := `DELETE FROM autos WHERE id_autos=$1`
	_, err := a.Db.Exec(context.Background(), query, idAutos)
	if err != nil {
		log.Printf("Erro ao deletar o registro na tabela autos: %v", err)
		return fmt.Errorf("erro ao deletar registro: %w", err)
	}

	return nil
}

func (a *AutosModelType) IsDocAutuado(idCtxt int, idPje string) (bool, error) {
	query := `SELECT * FROM autos WHERE id_ctxt = $1 AND id_pje = $2`
	rows, err := a.Db.Query(context.Background(), query, idCtxt, idPje)
	if err != nil {
		log.Printf("Erro ao verificar documento autuado: %v", err)
		return false, fmt.Errorf("erro ao verificar documento: %w", err)
	}
	defer rows.Close()

	return rows.Next(), nil
}

func (a *AutosModelType) SelectByContexto(idCtxt int) ([]AutosRow, error) {
	query := `SELECT * FROM autos WHERE id_ctxt = $1`
	rows, err := a.Db.Query(context.Background(), query, idCtxt)
	if err != nil {
		log.Printf("Erro ao selecionar documentos por contexto: %v", err)
		return nil, fmt.Errorf("erro ao selecionar documentos: %w", err)
	}
	defer rows.Close()

	var results []AutosRow
	for rows.Next() {
		var row AutosRow
		if err := rows.Scan(&row.IdAutos, &row.IdCtxt, &row.IdNat, &row.IdPje, &row.DtPje, &row.AutosJson, &row.DtInc, &row.Status); err != nil {
			log.Printf("Erro ao escanear linha: %v", err)
			continue
		}
		results = append(results, row)
	}

	return results, nil
}

func (a *AutosModelType) SelectById(idAutos int) (*AutosRow, error) {
	query := `SELECT * FROM autos WHERE id_autos = $1`
	row := a.Db.QueryRow(context.Background(), query, idAutos)

	var selectedRow AutosRow
	if err := row.Scan(&selectedRow.IdAutos, &selectedRow.IdCtxt, &selectedRow.IdNat, &selectedRow.IdPje, &selectedRow.DtPje, &selectedRow.AutosJson, &selectedRow.DtInc, &selectedRow.Status); err != nil {
		log.Printf("Erro ao selecionar o registro: %v", err)
		return nil, fmt.Errorf("erro ao selecionar registro: %w", err)
	}

	return &selectedRow, nil
}
