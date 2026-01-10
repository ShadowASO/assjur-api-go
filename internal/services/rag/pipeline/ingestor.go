package pipeline

import (
	"context"
	"encoding/json"

	"ocrserver/internal/consts"

	"ocrserver/internal/services"

	"strings"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

type IngestorType struct {
}

func NewIngestorType() *IngestorType {
	return &IngestorType{}
}

func (obj *IngestorType) StartAddSentencaBase(
	ctx context.Context,
	sentencas []consts.ResponseAutosRow,
	id_ctxt string,
	userName string) error {

	for _, sentenca := range sentencas {

		//*** Converte objeto JSON para um objeto GO(tipoResponse)
		jsonObj := sentenca.DocJsonRaw

		var objSentenca SentencaAutos
		err := json.Unmarshal([]byte(jsonObj), &objSentenca)
		if err != nil {
			logger.Log.Errorf("Erro ao realizar unmarshal da sentença: %v", err)
			return erros.CreateError("Erro ao unmarshal sentença")
		}
		// //Metadados da sentença
		// idPje := objSentenca.IdPje
		// hash_texto:=GetHashFromTexto()

		// //Verifica se já existe algum registro com o id_pje
		// isExist, err := services.BaseServiceGlobal.IsExist(id_ctxt,idPje)
		// if err != nil {
		// 	logger.Log.Errorf("Erro ao verificar se sentença já foi adicionada à base de conhecimento: id_pje=%s.", idPje)
		// 	return err
		// }
		// if isExist {
		// 	logger.Log.Errorf("Documento já foi adicionada à base de conhecimento: id_pje=%s.", idPje)
		// 	continue
		// }

		classe := objSentenca.Metadados.Classe
		assunto := objSentenca.Metadados.Assunto
		natureza := "sentenca"
		fonte := objSentenca.Processo

		regs := objSentenca.Questoes
		for _, item := range regs {

			//Metadados da sentença
			idPje := objSentenca.IdPje
			// Concatenar o vetor de textos com quebra de linha
			chunk := strings.Join(item.Paragrafos, "\n")
			hash_texto := GetHashFromTexto(chunk)

			//Verifica se já existe algum registro com o id_pje
			isExist, err := services.BaseServiceGlobal.IsExist(id_ctxt, idPje, hash_texto)
			if err != nil {
				logger.Log.Errorf("Erro ao verificar se sentença já foi adicionada à base de conhecimento: id_pje=%s.", idPje)
				return err
			}
			if isExist {
				logger.Log.Errorf("Documento já foi adicionada à base de conhecimento: id_pje=%s.", idPje)
				continue
			}

			obj.salvaRegistro(idPje, classe, assunto, natureza, item.Tipo, item.Tema, fonte, item.Paragrafos, id_ctxt,
				userName, hash_texto)
		}
		//ATENÇÃO: Deleta o registro da sentença
		// err = services.AutosServiceGlobal.DeletaAutos(sentenca.Id)
		// if err != nil {
		// 	logger.Log.Errorf("Erro ao deletar sentença nos autos: %v", err)

		// }

	}
	return nil
}

func (obj *IngestorType) salvaRegistro(idPje, classe, assunto, natureza, tipo, tema, fonte string, texto []string, id_ctxt string,
	userName string, hash_texto string) error {

	// Concatenar o vetor de textos com quebra de linha
	raw := strings.Join(texto, "\n")

	// vector, err := ialib.GetDocumentoEmbeddings(raw)
	// if err != nil {
	// 	logger.Log.Errorf("Erro ao extrair os embeddings do documento: %v", err)
	// 	return erros.CreateErrorf("Erro ao extrair o embedding: Contexto: %s - IdDoc: %s", idPje, &raw)
	// }
	//logger.Log.Info(raw)

	//*************************************
	// doc := opensearch.ParamsBaseInsert{
	// doc := opensearch.BaseRow{
	// 	IdPje:          idPje,
	// 	Classe:         classe,
	// 	Assunto:        assunto,
	// 	Natureza:       natureza,
	// 	Tipo:           tipo,
	// 	Tema:           tema,
	// 	Fonte:          fonte,
	// 	Texto:          raw,
	// 	TextoEmbedding: vector,
	// }
	//_, err = opensearch.RagServiceGlobal.IndexaDocumento(rag)
	doc, err := services.BaseServiceGlobal.InserirDocumento(
		id_ctxt,
		idPje,
		userName,
		classe,
		assunto,
		natureza,
		tipo,
		tema,
		fonte,
		raw,
		hash_texto,
	)

	if err != nil {
		logger.Log.Error("Erro ao inserir documento no índice 'autos'")
		return erros.CreateError("Erro ao inserir documento no índice 'autos'")
	}

	//msg := "Concluído com sucesso!"
	logger.Log.Infof("Concluído com sucesso: %s", doc.Id)
	return nil
}
