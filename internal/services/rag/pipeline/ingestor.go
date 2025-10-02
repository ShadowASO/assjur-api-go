package pipeline

import (
	"context"
	"encoding/json"
	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services/ialib"
	"strings"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

type IngestorType struct {
}

func NewIngestorType() *IngestorType {
	return &IngestorType{}
}

type docSentenca struct {
	Tipo *struct {
		Key         int    `json:"key"`
		Description string `json:"description"`
	} `json:"tipo"`

	Processo  string     `json:"processo"`
	IdPje     string     `json:"id_pje"`
	Metadados *metadados `json:"metadados"`

	Questoes    []questao   `json:"questoes"`
	Dispositivo dispositivo `json:"dispositivo"`
}

type metadados struct {
	Classe  string  `json:"classe"`
	Assunto string  `json:"assunto"`
	Juizo   string  `json:"juizo"`
	Partes  *partes `json:"partes"`
}

type partes struct {
	Autor string `json:"autor"`
	Reu   string `json:"reu"`
}

type questao struct {
	Tipo       string   `json:"tipo"` // "preliminar" ou "mérito"
	Tema       string   `json:"tema"`
	Paragrafos []string `json:"paragrafos"`
	Decisao    string   `json:"decisao"`
}

type dispositivo struct {
	Paragrafos []string `json:"paragrafos"`
}

func (service *IngestorType) StartAddSentencaBase(ctx context.Context, sentenca []consts.ResponseAutosRow) (bool, error) {

	jsonObj := sentenca[0].DocJsonRaw

	//*** Converte objeto JSON para um objeto GO(tipoResponse)

	var objSentenca docSentenca
	err := json.Unmarshal([]byte(jsonObj), &objSentenca)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar unmarshal resposta da análise: %v", err)
		return false, erros.CreateError("Erro ao unmarshal resposta da análise")
	}
	//Metadados da sentença
	//processo := objSentenca.Processo
	idPje := objSentenca.IdPje
	classe := objSentenca.Metadados.Classe
	assunto := objSentenca.Metadados.Assunto
	natureza := "sentenca"
	fonte := objSentenca.Processo

	//Cadas registro
	// var tipo string
	// var tema string

	regs := objSentenca.Questoes
	for _, item := range regs {
		service.salvaRegistro(ctx, idPje, classe, assunto, natureza, item.Tipo, item.Tema, fonte, item.Paragrafos)
	}

	return true, nil
}

func (service *IngestorType) salvaRegistro(ctx context.Context, idPje, classe, assunto, natureza, tipo, tema, fonte string, texto []string) error {

	// Concatenar com quebra de linha
	raw := strings.Join(texto, "\n")

	embVector, err := ialib.GetDocumentoEmbeddings(raw)
	if err != nil {
		logger.Log.Errorf("Erro ao extrair os embeddings do documento: %v", err)
		return erros.CreateErrorf("Erro ao extrair o embedding: Contexto: %d - IdDoc: %s", idPje, &raw)
	}

	//*************************************
	rag := opensearch.DocumentoRag{idPje, classe, assunto, natureza, tipo, tema, fonte, raw, embVector}
	_, err = opensearch.RagServiceGlobal.IndexaDocumento(rag)
	if err != nil {
		logger.Log.Error("Erro ao inserir documento no índice 'autos'")
		return erros.CreateError("Erro ao inserir documento no índice 'autos'")
	}

	msg := "Concluído com sucesso!"
	logger.Log.Info(msg)
	return nil
}
