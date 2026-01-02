/*
---------------------------------------------------------------------------------------
File: ContextoService.go
Autor: Aldenor
Data: 22-07-2025
---------------------------------------------------------------------------------------
*/
package services

import (
	"fmt"

	"ocrserver/internal/opensearch"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"sync"
)

type ContextoServiceType struct {
	Idx *opensearch.ContextoIndexType
}

var ContextoServiceGlobal *ContextoServiceType
var onceInitContextoService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitContextoService(model *opensearch.ContextoIndexType) {
	onceInitContextoService.Do(func() {
		ContextoServiceGlobal = &ContextoServiceType{
			Idx: model,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewContextoService(
	model *opensearch.ContextoIndexType,

) *ContextoServiceType {
	return &ContextoServiceType{
		Idx: model,
	}
}

func (obj *ContextoServiceType) InsertContexto(
	NrProc string,
	Juizo string,
	Classe string,
	Assunto string,
	userName string) (*opensearch.ResponseContextoRow, error) {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	row, err := obj.Idx.Indexa(NrProc, Juizo, Classe, Assunto, userName)
	if err != nil {
		logger.Log.Errorf("Erro ao inserir contexto: %v", err)
		return nil, erros.CreateError("Erro interno no servidor ao inserir contexto!")
	}
	return row, nil
}
func (obj *ContextoServiceType) UpdateContexto(
	id string,
	Juizo string,
	Classe string,
	Assunto string,
) (*opensearch.ResponseContextoRow, error) {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Idx.Update(id, Juizo, Classe, Assunto)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
func (obj *ContextoServiceType) DeletaContexto(idCtxt string) error {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	err := obj.Idx.Delete(idCtxt)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return err
	}
	return nil
}
func (obj *ContextoServiceType) SelectContextoById(id string) (*opensearch.ResponseContextoRow, error) {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Idx.ConsultaById(id)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
func (obj *ContextoServiceType) SelectContextoByIdCtxt(id string) ([]opensearch.ResponseContextoRow, error) {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Idx.ConsultaByIdCtxt(id)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
func (obj *ContextoServiceType) SelectContextoByProcesso(nrProc string) (*opensearch.ResponseContextoRow, error) {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Idx.ConsultaByProcesso(nrProc)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}

func (obj *ContextoServiceType) SelectContextoByProcessoLike(nrProc string) ([]opensearch.ResponseContextoRow, error) {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, err := obj.Idx.SelectContextoByProcessoStartsWith(nrProc)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}

func (obj *ContextoServiceType) SelectContextos(limit, offset int) ([]opensearch.ResponseContextoRow, error) {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	rows, err := obj.Idx.SelectContextos(limit, offset)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return rows, nil
}
func (obj *ContextoServiceType) ContextoExiste(nrProc string) (bool, error) {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return false, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	isExiste, err := obj.Idx.IsExistes(nrProc)
	if err != nil {
		logger.Log.Errorf("Erro na verificação existência!: %v", err)
		return false, err
	}
	return isExiste, nil
}

func (obj *ContextoServiceType) UpdateTokenUso(idCtxt string, pt int, ct int) (*opensearch.ResponseContextoRow, error) {
	if obj.Idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	row, err := obj.Idx.IncrementTokensAtomic(idCtxt, pt, ct)
	if err != nil {
		logger.Log.Error("Erro na alteração do registro!!")
		return nil, err
	}
	return row, nil
}
