/*
---------------------------------------------------------------------------------------
File: eventosService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 09-10-2025
---------------------------------------------------------------------------------------
*/
package services

import (
	"fmt"
	"sync"

	"ocrserver/internal/opensearch"
	"ocrserver/internal/utils/logger"
)

// ============================================================================
// Estrutura principal
// ============================================================================

type EventosService struct {
	idx *opensearch.EventosIndex
}

var EventosServiceGlobal *EventosService
var onceInitEventosService sync.Once

// ============================================================================
// Inicialização global
// ============================================================================

func InitEventosService(idx *opensearch.EventosIndex) {
	onceInitEventosService.Do(func() {
		EventosServiceGlobal = &EventosService{
			idx: idx,
		}
		logger.Log.Info("Global EventosService configurado com sucesso.")
	})
}

func NewEventosService(idx *opensearch.EventosIndex) *EventosService {
	return &EventosService{
		idx: idx,
	}
}

// ============================================================================
// Operações CRUD
// ============================================================================

// Inserir novo evento
func (obj *EventosService) InserirEvento(
	IdCtxt string,
	IdNatu int,
	IdPje string,
	doc string,
	docJsonRaw string,
	userName string,
) (*opensearch.ResponseEventosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de EventosService não iniciado.")
		return nil, fmt.Errorf("serviço EventosService não iniciado")
	}

	row, err := obj.idx.Indexa(IdCtxt, IdNatu, IdPje, doc, docJsonRaw, nil, "", userName)
	if err != nil {
		logger.Log.Errorf("Erro na inclusão do evento: %v", err)
		return nil, err
	}
	return row, nil
}

// Atualizar evento existente
func (obj *EventosService) UpdateEvento(data opensearch.ResponseEventosRow) (*opensearch.ResponseEventosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de EventosService não iniciado.")
		return nil, fmt.Errorf("serviço EventosService não iniciado")
	}

	row, err := obj.idx.Update(
		data.Id,
		data.IdCtxt,
		data.IdNatu,
		data.IdPje,
		data.Doc,
		data.DocJsonRaw,
		data.DocEmbedding,
	)
	if err != nil {
		logger.Log.Errorf("Erro na atualização do evento: %v", err)
		return nil, err
	}
	return row, nil
}

// Deletar evento
func (obj *EventosService) DeletaEvento(id string) error {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de EventosService não iniciado.")
		return fmt.Errorf("serviço EventosService não iniciado")
	}

	err := obj.idx.Delete(id)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar documento no índice 'eventos': %v", err)
		return fmt.Errorf("erro ao deletar documento no índice 'eventos'")
	}

	// ================================================
	// Exclusão de embeddings vinculados (se existirem)
	// ================================================
	// emb, err := EventosJsonServiceGlobal.SelectByIdDoc(id)
	// if err != nil {
	// 	logger.Log.Errorf("Erro ao buscar embeddings vinculados ao evento ID=%s: %v", id, err)
	// 	return fmt.Errorf("erro ao buscar embeddings vinculados ao evento")
	// }

	// for _, reg := range emb {
	// 	err := EventosJsonServiceGlobal.DeletaEmbedding(reg.Id)
	// 	if err != nil {
	// 		logger.Log.Errorf("Erro ao deletar embedding vinculado ID=%s: %v", reg.Id, err)
	// 		return fmt.Errorf("erro ao deletar embedding vinculado")
	// 	}
	// }

	return nil
}

// Consultar evento por ID
func (obj *EventosService) SelectById(id string) (*opensearch.ResponseEventosRow, int, error) {
	if obj.idx == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, 0, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	row, statusCode, err := obj.idx.ConsultaById(id)
	if err != nil {
		logger.Log.Errorf("x: %v", err)
		return nil, statusCode, err
	}
	return row, statusCode, nil
}

// Consultar eventos por contexto (id_ctxt)
func (obj *EventosService) SelectByContexto(idCtxt string) ([]opensearch.ResponseEventosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de EventosService não iniciado.")
		return nil, fmt.Errorf("serviço EventosService não iniciado")
	}

	rows, err := obj.idx.ConsultaByIdCtxt(idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar eventos por contexto %d: %v", idCtxt, err)
		return nil, err
	}
	return rows, nil
}

// ============================================================================
// Operações derivadas e de apoio
// ============================================================================

// Retornar eventos de um contexto com log detalhado
func (obj *EventosService) GetEventosByContexto(id string) ([]opensearch.ResponseEventosRow, error) {
	if obj == nil {
		logger.Log.Error("Serviço EventosServiceGlobal não inicializado.")
		return nil, fmt.Errorf("serviço EventosServiceGlobal não inicializado")
	}

	rows, err := obj.SelectByContexto(id)
	if err != nil {
		logger.Log.Errorf("[id_ctxt=%d] Erro ao buscar eventos do contexto: %v", id, err)
		return nil, fmt.Errorf("erro ao buscar eventos do contexto %d: %w", id, err)
	}

	if len(rows) == 0 {
		logger.Log.Warningf("[id_ctxt=%d] Nenhum registro de eventos encontrado no contexto.", id)
	}

	logger.Log.Infof("[id_ctxt=%d] Recuperados %d registros de eventos.", id, len(rows))
	return rows, nil
}

// Verifica se evento já foi registrado (id_ctxt + id_evento)
func (obj *EventosService) IsEventoRegistrado(idCtxt string, idEvento string) (bool, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de EventosService não iniciado.")
		return false, fmt.Errorf("serviço EventosService não iniciado")
	}

	exist, err := obj.idx.IsExiste(idCtxt, idEvento)
	if err != nil {
		logger.Log.Errorf("Erro ao verificar existência de evento: %v", err)
		return false, err
	}
	return exist, nil
}
