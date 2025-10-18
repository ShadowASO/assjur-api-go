package pipeline

import (
	"context"
	"encoding/json"

	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"

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

func (obj *IngestorType) StartAddSentencaBase(ctx context.Context, sentencas []consts.ResponseAutosRow) error {

	//jsonObj := sentencas[0].DocJsonRaw
	for _, sentenca := range sentencas {

		//*** Converte objeto JSON para um objeto GO(tipoResponse)
		jsonObj := sentenca.DocJsonRaw

		var objSentenca SentencaAutos
		err := json.Unmarshal([]byte(jsonObj), &objSentenca)
		if err != nil {
			logger.Log.Errorf("Erro ao realizar unmarshal resposta da análise: %v", err)
			return erros.CreateError("Erro ao unmarshal resposta da análise")
		}
		//Metadados da sentença
		idPje := objSentenca.IdPje

		//Verifica se já existe algum registro com o id_pje
		isExist, err := services.BaseServiceGlobal.IsExist(idPje)
		if err != nil {
			logger.Log.Errorf("Erro ao verificar se sentença já foi adicionada à base de conhecimento: id_pje=%s.", idPje)
			return err
		}
		if isExist {
			logger.Log.Errorf("Documento já foi adicionada à base de conhecimento: id_pje=%s.", idPje)
			continue
		}

		classe := objSentenca.Metadados.Classe
		assunto := objSentenca.Metadados.Assunto
		natureza := "sentenca"
		fonte := objSentenca.Processo

		regs := objSentenca.Questoes
		for _, item := range regs {
			obj.salvaRegistro(idPje, classe, assunto, natureza, item.Tipo, item.Tema, fonte, item.Paragrafos)
		}
		//Deleta o registro da sentença
		err = services.AutosServiceGlobal.DeletaAutos(sentenca.Id)
		if err != nil {
			logger.Log.Errorf("Erro ao deletar sentença nos autos: %v", err)
			//return err
		}

	}
	return nil
}

func (obj *IngestorType) salvaRegistro(idPje, classe, assunto, natureza, tipo, tema, fonte string, texto []string) error {

	// Concatenar o vetor de textos com quebra de linha
	raw := strings.Join(texto, "\n")

	vector, err := ialib.GetDocumentoEmbeddings(raw)
	if err != nil {
		logger.Log.Errorf("Erro ao extrair os embeddings do documento: %v", err)
		return erros.CreateErrorf("Erro ao extrair o embedding: Contexto: %s - IdDoc: %s", idPje, &raw)
	}
	//logger.Log.Info(raw)

	//*************************************
	doc := opensearch.ParamsBaseInsert{
		IdPje:         idPje,
		Classe:        classe,
		Assunto:       assunto,
		Natureza:      natureza,
		Tipo:          tipo,
		Tema:          tema,
		Fonte:         fonte,
		DataTexto:     raw,
		DataEmbedding: vector,
	}
	//_, err = opensearch.RagServiceGlobal.IndexaDocumento(rag)
	services.BaseServiceGlobal.InserirDocumento(doc)

	if err != nil {
		logger.Log.Error("Erro ao inserir documento no índice 'autos'")
		return erros.CreateError("Erro ao inserir documento no índice 'autos'")
	}

	msg := "Concluído com sucesso!"
	logger.Log.Info(msg)
	return nil
}
