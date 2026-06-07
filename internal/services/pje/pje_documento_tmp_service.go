/*
---------------------------------------------------------------------------------------
File: pje_documento_tmp_service.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 05-06-2026
---------------------------------------------------------------------------------------
Serviço responsável por manipular documentos temporários retornados pelo MNI/PJe.

Cada registro do índice pje_documentos_tmp corresponde a um único documento processual.
O índice é temporário e serve como área intermediária para armazenar o conteúdo recebido
do MNI enquanto a aplicação realiza os processamentos necessários.
---------------------------------------------------------------------------------------
*/

package pje

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"ocrserver/internal/dominio/pje"
	"ocrserver/internal/models/opensearch"

	"ocrserver/internal/utils/mslogger"
)

type PjeDocumentoTmpService struct {
	idx *opensearch.PjeDocumentoTmpIndex
}

var PjeDocumentoTmpServiceGlobal *PjeDocumentoTmpService
var onceInitPjeDocumentoTmpService sync.Once

func InitPjeDocumentoTmpService(idx *opensearch.PjeDocumentoTmpIndex) {
	onceInitPjeDocumentoTmpService.Do(func() {
		PjeDocumentoTmpServiceGlobal = &PjeDocumentoTmpService{
			idx: idx,
		}

		mslogger.LoggerGlobal.Info("Global PJeDocumentoTmpService configurado com sucesso.")
	})
}

func NewPjeDocumentoTmpService(
	idx *opensearch.PjeDocumentoTmpIndex,
) *PjeDocumentoTmpService {
	return &PjeDocumentoTmpService{
		idx: idx,
	}
}

func (obj *PjeDocumentoTmpService) checkService() error {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de PJeDocumentoTmpService não iniciado.")
		return fmt.Errorf("PJeDocumentoTmpService não iniciado")
	}

	if obj.idx == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de PJeDocumentoTmpService sem índice OpenSearch.")
		return fmt.Errorf("índice pje_documentos_tmp não configurado")
	}

	return nil
}

func (obj *PjeDocumentoTmpService) Insert(
	data pje.PjeDocumentoTmp,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	data.IdCtxt = strings.TrimSpace(data.IdCtxt)
	data.NumeroProcesso = strings.TrimSpace(data.NumeroProcesso)
	data.IdPje = strings.TrimSpace(data.IdPje)

	if data.IdCtxt == "" {
		return nil, fmt.Errorf("id_ctxt vazio")
	}

	if data.IdPje == "" {
		return nil, fmt.Errorf("id_documento vazio")
	}

	if strings.TrimSpace(data.Status) == "" {
		data.Status = "pendente"
	}

	now := time.Now()

	if data.CriadoEm == nil {
		data.CriadoEm = &now
	}

	if data.AtualizadoEm == nil {
		data.AtualizadoEm = &now
	}

	row, err := obj.idx.Indexa(data, "")
	if err != nil {
		mslogger.LoggerGlobal.Errorf(
			"Erro na inclusão do documento temporário PJe: id_ctxt=%s id_documento=%s - %v",
			data.IdCtxt,
			data.IdPje,
			err,
		)
		return nil, err
	}

	return row, nil
}

func (obj *PjeDocumentoTmpService) InserirCampos(
	idCtxt string,
	numeroProcesso string,
	idPje string,
	descricao string,
	storageID string,
	mimetype string,
	formatoEntrega string,
	conteudoTexto string,
	conteudoHTML string,
	conteudoBase64 string,
	status string,
	erroMsg string,
	expiraEm *time.Time,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	now := time.Now()

	if strings.TrimSpace(status) == "" {
		status = "pendente"
	}

	row := pje.PjeDocumentoTmp{
		IdCtxt:         strings.TrimSpace(idCtxt),
		NumeroProcesso: strings.TrimSpace(numeroProcesso),
		IdPje:          strings.TrimSpace(idPje),
		Descricao:      strings.TrimSpace(descricao),
		StorageID:      strings.TrimSpace(storageID),
		Mimetype:       strings.TrimSpace(mimetype),
		FormatoEntrega: strings.TrimSpace(formatoEntrega),

		ConteudoTexto:  conteudoTexto,
		ConteudoHTML:   conteudoHTML,
		ConteudoBase64: conteudoBase64,

		Status: strings.TrimSpace(status),
		Erro:   erroMsg,

		CriadoEm:     &now,
		AtualizadoEm: &now,
		ExpiraEm:     expiraEm,
	}

	return obj.Insert(row)
}

func (obj *PjeDocumentoTmpService) Update(
	id string,
	patch pje.PjeDocumentoTmpPatch,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	row, err := obj.idx.Update(id, patch)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro na alteração do documento temporário PJe.", err)
		return nil, err
	}

	return row, nil
}

func (obj *PjeDocumentoTmpService) UpdateStatus(
	id string,
	status string,
	erroMsg string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	id = strings.TrimSpace(id)
	status = strings.TrimSpace(status)

	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	if status == "" {
		return nil, fmt.Errorf("status vazio")
	}

	row, err := obj.idx.UpdateStatus(id, status, erroMsg)
	if err != nil {
		mslogger.LoggerGlobal.Errorf(
			"Erro ao atualizar status do documento temporário PJe: id=%s status=%s - %v",
			id,
			status,
			err,
		)
		return nil, err
	}

	return row, nil
}

func (obj *PjeDocumentoTmpService) MarcarComoProcessando(
	id string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	return obj.UpdateStatus(id, "processando", "")
}

func (obj *PjeDocumentoTmpService) MarcarComoProcessado(
	id string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	return obj.UpdateStatus(id, "processado", "")
}

func (obj *PjeDocumentoTmpService) MarcarComoOCRPendente(
	id string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	return obj.UpdateStatus(id, "ocr_pendente", "")
}

func (obj *PjeDocumentoTmpService) MarcarComoErro(
	id string,
	erroMsg string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if strings.TrimSpace(erroMsg) == "" {
		erroMsg = "erro não especificado"
	}

	return obj.UpdateStatus(id, "erro", erroMsg)
}

func (obj *PjeDocumentoTmpService) Deleta(
	id string,
) error {
	if err := obj.checkService(); err != nil {
		return err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("id vazio")
	}

	err := obj.idx.Delete(id)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro na deleção do documento temporário PJe.", err)
		return fmt.Errorf("erro ao deletar documento temporário PJe: %w", err)
	}

	return nil
}

func (obj *PjeDocumentoTmpService) SelectById(
	id string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	row, err := obj.idx.ConsultaById(id)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar documento temporário PJe ID=%s - %v", id, err)
		return nil, err
	}

	return row, nil
}

func (obj *PjeDocumentoTmpService) SelectByDocumento(
	idCtxt string,
	idDocumento string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	idCtxt = strings.TrimSpace(idCtxt)
	idDocumento = strings.TrimSpace(idDocumento)

	if idCtxt == "" {
		return nil, fmt.Errorf("id_ctxt vazio")
	}

	if idDocumento == "" {
		return nil, fmt.Errorf("id_documento vazio")
	}

	row, err := obj.idx.ConsultaByIdPje(idCtxt, idDocumento)
	if err != nil {
		mslogger.LoggerGlobal.Errorf(
			"Erro ao selecionar documento temporário PJe: id_ctxt=%s id_documento=%s - %v",
			idCtxt,
			idDocumento,
			err,
		)
		return nil, err
	}

	return row, nil
}

func (obj *PjeDocumentoTmpService) SelectByContexto(
	idCtxt string,
) ([]pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	idCtxt = strings.TrimSpace(idCtxt)
	if idCtxt == "" {
		return nil, fmt.Errorf("id_ctxt vazio")
	}

	rows, err := obj.idx.ConsultaByIdCtxt(idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar documentos temporários PJe do contexto ID=%s - %v", idCtxt, err)
		return nil, err
	}

	return rows, nil
}

func (obj *PjeDocumentoTmpService) SelectByNumeroProcesso(
	numeroProcesso string,
) ([]pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	numeroProcesso = strings.TrimSpace(numeroProcesso)
	if numeroProcesso == "" {
		return nil, fmt.Errorf("numero_processo vazio")
	}

	rows, err := obj.idx.ConsultaByNumeroProcesso(numeroProcesso)
	if err != nil {
		mslogger.LoggerGlobal.Errorf(
			"Erro ao selecionar documentos temporários PJe do processo %s - %v",
			numeroProcesso,
			err,
		)
		return nil, err
	}

	return rows, nil
}

func (obj *PjeDocumentoTmpService) SelectByStatus(
	status string,
) ([]pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	status = strings.TrimSpace(status)
	if status == "" {
		return nil, fmt.Errorf("status vazio")
	}

	rows, err := obj.idx.ConsultaByStatus(status)
	if err != nil {
		mslogger.LoggerGlobal.Errorf(
			"Erro ao selecionar documentos temporários PJe por status=%s - %v",
			status,
			err,
		)
		return nil, err
	}

	return rows, nil
}

func (obj *PjeDocumentoTmpService) GetDocumentosByContexto(
	idCtxt string,
) ([]pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	idCtxt = strings.TrimSpace(idCtxt)
	if idCtxt == "" {
		return nil, fmt.Errorf("id_ctxt vazio")
	}

	rows, err := obj.SelectByContexto(idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar documentos temporários do contexto ID=%s - %v", idCtxt, err)
		return nil, fmt.Errorf("erro ao buscar documentos temporários do contexto %s: %w", idCtxt, err)
	}

	if len(rows) == 0 {
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Nenhum documento temporário PJe encontrado no contexto.", idCtxt)
		return rows, nil
	}

	mslogger.LoggerGlobal.Infof("[id_ctxt=%s] Recuperados %d documentos temporários PJe.", idCtxt, len(rows))

	return rows, nil
}

func (obj *PjeDocumentoTmpService) IsExiste(
	idCtxt string,
	idDocumento string,
) (bool, error) {
	if err := obj.checkService(); err != nil {
		return false, err
	}

	idCtxt = strings.TrimSpace(idCtxt)
	idDocumento = strings.TrimSpace(idDocumento)

	if idCtxt == "" {
		return false, fmt.Errorf("id_ctxt vazio")
	}

	if idDocumento == "" {
		return false, fmt.Errorf("id_documento vazio")
	}

	existe, err := obj.idx.IsExiste(idCtxt, idDocumento)
	if err != nil {
		mslogger.LoggerGlobal.Infof(
			"Documento temporário PJe não encontrado ou erro na consulta: id_ctxt=%s id_documento=%s - %v",
			idCtxt,
			idDocumento,
			err,
		)
		return false, err
	}

	return existe, nil
}

func (obj *PjeDocumentoTmpService) Upsert(
	data pje.PjeDocumentoTmp,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	data.IdCtxt = strings.TrimSpace(data.IdCtxt)
	data.IdPje = strings.TrimSpace(data.IdPje)

	if data.IdCtxt == "" {
		return nil, fmt.Errorf("id_ctxt vazio")
	}

	if data.IdPje == "" {
		return nil, fmt.Errorf("id_documento vazio")
	}

	id := opensearch.MakePjeDocumentoTmpID(data.IdCtxt, data.IdPje)
	if id == "" {
		return nil, fmt.Errorf("não foi possível montar o id do documento temporário")
	}

	return obj.idx.Indexa(data, id)
}

func (obj *PjeDocumentoTmpService) LimparExpirados() error {
	if err := obj.checkService(); err != nil {
		return err
	}

	err := obj.idx.DeleteExpirados()
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao limpar documentos temporários PJe expirados.", err)
		return err
	}

	mslogger.LoggerGlobal.Info("Limpeza de documentos temporários PJe expirados executada com sucesso.")

	return nil
}
