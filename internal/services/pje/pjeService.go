/*
---------------------------------------------------------------------------------------
File: pjeService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
---------------------------------------------------------------------------------------
*/
package pje

import (
	"context"
	"fmt"

	"ocrserver/internal/services/rest_services/mnicnj"

	"ocrserver/internal/utils/mslogger"
	"sync"
)

type PjeService struct {
	doc_tmp_srv *MniDocumentoTmpService
	mnicnj      *mnicnj.ClientMNI
}

var PjeServiceGlobal *PjeService
var onceInitPjeService sync.Once

func InitPjeService(doc_tmp_srv *MniDocumentoTmpService, mnicnj *mnicnj.ClientMNI) {
	onceInitPjeService.Do(func() {

		PjeServiceGlobal = &PjeService{
			doc_tmp_srv: doc_tmp_srv,
			mnicnj:      mnicnj,
		}

		mslogger.LoggerGlobal.Info("Global PjeService configurado com sucesso.")
	})
}

func NewPjeService(doc_tmp_srv *MniDocumentoTmpService, mnicnj *mnicnj.ClientMNI) *PjeService {
	return &PjeService{
		doc_tmp_srv: doc_tmp_srv,
		mnicnj:      mnicnj,
	}
}

func (obj *PjeService) ListaDocumentos(
	ctx context.Context,
	numero_processo string,
	usuario_cpf string,
	usuario_senha string,
	documentos []string,
	formato string,
) (*mnicnj.ConsultaProcessoPJEResponse, error) {
	if obj == nil {
		mslogger.LoggerGlobal.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}
	req := mnicnj.ConsultaProcessoPJERequest{
		NumeroProcesso: numero_processo,
		UsuarioCpf:     usuario_cpf,
		UsuarioSenha:   usuario_senha,
		DataReferencia: "",
		Documentos:     documentos,
		Formato:        formato,
	}

	// Indexa diretamente a string JSON
	rsp, err := obj.mnicnj.ConsultarListaDocumentos(ctx, req)
	if err != nil {

		mslogger.LoggerGlobal.Errorf("Erro na consulta do processo: %s - %v", numero_processo, err)
		return nil, err
	}
	return rsp, nil
}
