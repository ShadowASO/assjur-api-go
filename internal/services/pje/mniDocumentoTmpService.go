/*
---------------------------------------------------------------------------------------
File: mniDocumentoTmpService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 05-06-2026
---------------------------------------------------------------------------------------
Serviço responsável por manipular documentos temporários retornados pelo MNI/PJe.

Cada registro do índice mni_documentos_tmp corresponde a um único documento processual.
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

	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/utils/mslogger"
)

type MniDocumentoTmpService struct {
	idx *opensearch.MniDocumentosTmpIndex
}

var MniDocumentoTmpServiceGlobal *MniDocumentoTmpService
var onceInitMniDocumentoTmpService sync.Once

func InitMniDocumentoTmpService(idx *opensearch.MniDocumentosTmpIndex) {
	onceInitMniDocumentoTmpService.Do(func() {
		MniDocumentoTmpServiceGlobal = &MniDocumentoTmpService{
			idx: idx,
		}

		mslogger.LoggerGlobal.Info("Global MniDocumentoTmpService configurado com sucesso.")
	})
}

func NewMniDocumentoTmpService(
	idx *opensearch.MniDocumentosTmpIndex,
) *MniDocumentoTmpService {
	return &MniDocumentoTmpService{
		idx: idx,
	}
}

func (obj *MniDocumentoTmpService) checkService() error {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de MniDocumentoTmpService não iniciado.")
		return fmt.Errorf("MniDocumentoTmpService não iniciado")
	}

	if obj.idx == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de MniDocumentoTmpService sem índice OpenSearch.")
		return fmt.Errorf("índice mni_documentos_tmp não configurado")
	}

	return nil
}

func (obj *MniDocumentoTmpService) Insert(
	data consts.MniDocumentoTmp,
) (*consts.ResponseMniDocumentoTmp, error) {
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
			"Erro na inclusão do documento temporário MNI: id_ctxt=%s id_documento=%s - %v",
			data.IdCtxt,
			data.IdPje,
			err,
		)
		return nil, err
	}

	return row, nil
}

// func (obj *MniDocumentoTmpService) InserirDocumentoTmpCampos(
func (obj *MniDocumentoTmpService) InserirCampos(
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
) (*consts.ResponseMniDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	now := time.Now()

	if strings.TrimSpace(status) == "" {
		status = "pendente"
	}

	row := consts.MniDocumentoTmp{
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

func (obj *MniDocumentoTmpService) Update(
	id string,
	patch consts.MniDocumentoTmpPatch,
) (*consts.ResponseMniDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	row, err := obj.idx.Update(id, patch)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro na alteração do documento temporário MNI.", err)
		return nil, err
	}

	return row, nil
}

func (obj *MniDocumentoTmpService) UpdateStatus(
	id string,
	status string,
	erroMsg string,
) (*consts.ResponseMniDocumentoTmp, error) {
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
			"Erro ao atualizar status do documento temporário MNI: id=%s status=%s - %v",
			id,
			status,
			err,
		)
		return nil, err
	}

	return row, nil
}

func (obj *MniDocumentoTmpService) MarcarComoProcessando(
	id string,
) (*consts.ResponseMniDocumentoTmp, error) {
	return obj.UpdateStatus(id, "processando", "")
}

func (obj *MniDocumentoTmpService) MarcarComoProcessado(
	id string,
) (*consts.ResponseMniDocumentoTmp, error) {
	return obj.UpdateStatus(id, "processado", "")
}

func (obj *MniDocumentoTmpService) MarcarComoOCRPendente(
	id string,
) (*consts.ResponseMniDocumentoTmp, error) {
	return obj.UpdateStatus(id, "ocr_pendente", "")
}

func (obj *MniDocumentoTmpService) MarcarComoErro(
	id string,
	erroMsg string,
) (*consts.ResponseMniDocumentoTmp, error) {
	if strings.TrimSpace(erroMsg) == "" {
		erroMsg = "erro não especificado"
	}

	return obj.UpdateStatus(id, "erro", erroMsg)
}

func (obj *MniDocumentoTmpService) Deleta(
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
		mslogger.LoggerGlobal.ErrorErr("Erro na deleção do documento temporário MNI.", err)
		return fmt.Errorf("erro ao deletar documento temporário MNI: %w", err)
	}

	return nil
}

func (obj *MniDocumentoTmpService) SelectById(
	id string,
) (*consts.ResponseMniDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	row, err := obj.idx.ConsultaById(id)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar documento temporário MNI ID=%s - %v", id, err)
		return nil, err
	}

	return row, nil
}

func (obj *MniDocumentoTmpService) SelectByDocumento(
	idCtxt string,
	idDocumento string,
) (*consts.ResponseMniDocumentoTmp, error) {
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
			"Erro ao selecionar documento temporário MNI: id_ctxt=%s id_documento=%s - %v",
			idCtxt,
			idDocumento,
			err,
		)
		return nil, err
	}

	return row, nil
}

func (obj *MniDocumentoTmpService) SelectByContexto(
	idCtxt string,
) ([]consts.ResponseMniDocumentoTmp, error) {
	if err := obj.checkService(); err != nil {
		return nil, err
	}

	idCtxt = strings.TrimSpace(idCtxt)
	if idCtxt == "" {
		return nil, fmt.Errorf("id_ctxt vazio")
	}

	rows, err := obj.idx.ConsultaByIdCtxt(idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar documentos temporários MNI do contexto ID=%s - %v", idCtxt, err)
		return nil, err
	}

	return rows, nil
}

func (obj *MniDocumentoTmpService) SelectByNumeroProcesso(
	numeroProcesso string,
) ([]consts.ResponseMniDocumentoTmp, error) {
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
			"Erro ao selecionar documentos temporários MNI do processo %s - %v",
			numeroProcesso,
			err,
		)
		return nil, err
	}

	return rows, nil
}

func (obj *MniDocumentoTmpService) SelectByStatus(
	status string,
) ([]consts.ResponseMniDocumentoTmp, error) {
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
			"Erro ao selecionar documentos temporários MNI por status=%s - %v",
			status,
			err,
		)
		return nil, err
	}

	return rows, nil
}

func (obj *MniDocumentoTmpService) GetDocumentosByContexto(
	idCtxt string,
) ([]consts.ResponseMniDocumentoTmp, error) {
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
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Nenhum documento temporário MNI encontrado no contexto.", idCtxt)
		return rows, nil
	}

	mslogger.LoggerGlobal.Infof("[id_ctxt=%s] Recuperados %d documentos temporários MNI.", idCtxt, len(rows))

	return rows, nil
}

func (obj *MniDocumentoTmpService) IsExiste(
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
			"Documento temporário MNI não encontrado ou erro na consulta: id_ctxt=%s id_documento=%s - %v",
			idCtxt,
			idDocumento,
			err,
		)
		return false, err
	}

	return existe, nil
}

func (obj *MniDocumentoTmpService) Upsert(
	data consts.MniDocumentoTmp,
) (*consts.ResponseMniDocumentoTmp, error) {
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

	id := opensearch.MakeMniDocumentoTmpID(data.IdCtxt, data.IdPje)
	if id == "" {
		return nil, fmt.Errorf("não foi possível montar o id do documento temporário")
	}

	return obj.idx.Indexa(data, id)
}

func (obj *MniDocumentoTmpService) LimparExpirados() error {
	if err := obj.checkService(); err != nil {
		return err
	}

	err := obj.idx.DeleteExpirados()
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao limpar documentos temporários MNI expirados.", err)
		return err
	}

	mslogger.LoggerGlobal.Info("Limpeza de documentos temporários MNI expirados executada com sucesso.")

	return nil
}
