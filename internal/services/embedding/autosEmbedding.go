/*
---------------------------------------------------------------------------------------
File: autosEmbedding.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 04-07-2025
---------------------------------------------------------------------------------------
*/
package embedding

import (
	"context"

	"fmt"
	consts "ocrserver/internal/constants"
	"ocrserver/internal/models"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"
	"strconv"

	"sync"
)

type AutosEmbeddingType struct {
	IndexAutos    *opensearch.IndexAutosType
	IndexDecisoes *opensearch.IndexDecisoesType
}

var IndexAutosEmbedding *AutosEmbeddingType
var onceAutosEmbedding sync.Once

// InitAutosEmbedding inicializa o serviço global para o índice autos_embedding
func InitAutosEmbedding() {
	onceAutosEmbedding.Do(func() {
		IndexAutosEmbedding = &AutosEmbeddingType{
			IndexAutos:    opensearch.NewIndexAutos(),
			IndexDecisoes: opensearch.NewIndexDecisoes(),
		}

		logger.Log.Info("Global IndexAutosEmbedding configurado com sucesso.")
	})
}

func NewAutosEmbedding() *AutosEmbeddingType {
	return &AutosEmbeddingType{
		IndexAutos:    opensearch.NewIndexAutos(),
		IndexDecisoes: opensearch.NewIndexDecisoes(),
	}
}

// Busca documento pelo ID usando o modelo autos_embedding
func (obj *AutosEmbeddingType) IncluirAutosByContexto(idCtxt int) ([]models.AutosRow, []models.AutosRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar IndexAutosService global sem inicializá-la.")
		return nil, nil, fmt.Errorf("IndexAutosService global não configurada")
	}
	autos, err := services.AutosService.GetAutosByContexto(idCtxt)
	if err != nil {
		logger.Log.Error("Erro ao selecionar autos!")
		return nil, nil, err
	}
	// Rastreamento de resultados
	var docsInseridos []models.AutosRow
	var docsRejeitados []models.AutosRow

	for _, doc := range autos {

		row, err := obj.IncluirDocumento(idCtxt, doc.IdNat, doc.IdPje, string(doc.AutosJson))
		if err != nil {
			logger.Log.Errorf("Documento %s não incluído: %v ", row, err)
			docsRejeitados = append(docsRejeitados, doc)
			continue
		}
		docsInseridos = append(docsInseridos, doc)
		logger.Log.Infof("Documento %d inserido nos embeddings!", doc.IdAutos)

	}

	return docsInseridos, docsRejeitados, nil

}

// Inclui um novo documento no índice autos_embedding
func (obj *AutosEmbeddingType) IncluirDocumento(idCtxt int, idNatu int, idPje string, doc string) (string, error) {
	ctx := context.Background()

	if obj == nil {
		logger.Log.Error("Tentativa de utilizar AutosEmbeddingType global sem inicializá-la.")
		return "", fmt.Errorf("AutosEmbeddingType global não configurada")
	}
	//Verifica se o documento já está incluído
	inserido, err := obj.IsEmbedding(idCtxt, idPje)
	if err != nil {
		logger.Log.Errorf("Erro ao verificar se documentos isEmbedding %v", err)
		return "", fmt.Errorf("Erro ao verificar se documentos isEmbedding %v", err)
	}
	if inserido {
		// logger.Log.Info("Documento já transformado em modelo embedding.")
		logger.Log.Error("O documento de ID_PJE=%v já inserido nos embeddings.", idPje)
		return "", nil
	}
	//***********  FAZER AQUI A FORMATAÇÃO DO DOCUMENTO, QUE DEVE TER VINDO NO FORMATO JSON
	//jsonFormatado, err:=ParseJsonToEmbedding(idNatu int, doc string)
	// if err != nil {
	// 	logger.Log.Errorf("Erro ao formatar json para embedding %v", err)
	// 	return "", fmt.Errorf("Erro ao verificar se documentos isEmbedding %v", err)
	// }

	// Gera o embedding do documento
	embeddingResp, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar embedding do texto: %w", err)
	}
	//Converte o vetor para 32
	vector32 := services.OpenaiServiceGlobal.Float64ToFloat32Slice(embeddingResp)

	//Cria o objeto para inclusão
	docObj := opensearch.IndexAutos{IdCtxt: idCtxt, IdNatu: idNatu, IdPje: idPje, DocEmbedding: vector32}

	//(*** CONTINUAR AQUI )

	//Insere efetivamente o documento
	resp, err := obj.IndexAutos.IndexaDocumento(docObj)
	if err != nil {
		logger.Log.Errorf("Erro ao indexar documento: %v", err)
		return "", err
	}
	logger.Log.Infof("Documento inserido em %v: %v", resp.Index, resp.ID)

	//Salva os documentos ****   DECISÕES   **********

	//Insere efetivamente o documento
	//indexDecisoes := opensearch.NewIndexDecisoes()
	decObj := opensearch.DecisoesEmbedding{IdCtxt: idCtxt, IdNatu: idNatu, Doc: doc, DocEmbedding: vector32}

	rspDec, err := obj.IndexDecisoes.IndexaDocumento(decObj)
	if err != nil {
		logger.Log.Errorf("Erro ao indexar documento: %v", err)
		return "", err
	}
	logger.Log.Infof("Documento inserido em %v: %v", rspDec.Index, rspDec.ID)

	return resp.ID, nil
}

// Busca documento pelo ID usando o modelo autos_embedding
func (obj *AutosEmbeddingType) GetDocumentoById(id string) (*opensearch.ResponseAutosEmbedding, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar IndexAutosService global sem inicializá-la.")
		return nil, fmt.Errorf("IndexAutosService global não configurada")
	}
	doc, err := obj.IndexAutos.ConsultaDocumentoById(id)
	if err != nil {
		logger.Log.Info("Erro ao selecionar documentos do índice autos_embedding!")
		return nil, err
	}
	return doc, nil
}

// Busca documento pelo ID usando o modelo autos_embedding
func (obj *AutosEmbeddingType) GetDocumentoByCtxt(idCtxt string) ([]opensearch.ResponseAutosEmbedding, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar IndexAutosService global sem inicializá-la.")
		return nil, fmt.Errorf("IndexAutosService global não configurada")
	}

	id, err := strconv.Atoi(idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao converte idNatu: %v para int.", idCtxt)
		return nil, fmt.Errorf("Erro ao converte idNatu: %v para int.", idCtxt)
	}

	doc, err := obj.IndexAutos.ConsultaDocumentoByIdCtxt(id)
	if err != nil {
		logger.Log.Info("Erro ao selecionar documentos do índice autos_embedding!")
		return nil, err
	}
	return doc, nil
}

// Busca documento pelo ID usando o modelo autos_embedding
func (obj *AutosEmbeddingType) GetDocumentoByNatureza(idNatu string) ([]opensearch.ResponseAutosEmbedding, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar IndexAutosService global sem inicializá-la.")
		return nil, fmt.Errorf("IndexAutosService global não configurada")
	}
	id, err := strconv.Atoi(idNatu)
	if err != nil {
		logger.Log.Errorf("Erro ao converte idNatu: %v para int.", idNatu)
		return nil, fmt.Errorf("Erro ao converte idNatu: %v para int.", idNatu)
	}

	doc, err := obj.IndexAutos.ConsultaDocumentosByIdNatu(id)
	if err != nil {
		logger.Log.Info("Erro ao selecionar documentos do índice autos_embedding!")
		return nil, err
	}
	return doc, nil
}

// Busca documento pelo ID usando o modelo autos_embedding
func (obj *AutosEmbeddingType) GetDecisaoById(id string) (*opensearch.ResponseDecisoes, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar IndexAutosService global sem inicializá-la.")
		return nil, fmt.Errorf("IndexAutosService global não configurada")
	}
	doc, err := obj.IndexDecisoes.ConsultaDocumentoById(id)
	if err != nil {
		logger.Log.Info("Erro ao selecionar documentos do índice decisoes!")
		return nil, err
	}
	return doc, nil
}

// Busca documento pelo ID usando o modelo autos_embedding
func (obj *AutosEmbeddingType) GetDecisoesByCtxt(idCtxt string) ([]opensearch.ResponseDecisoes, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar IndexAutosService global sem inicializá-la.")
		return nil, fmt.Errorf("IndexAutosService global não configurada")
	}

	id, err := strconv.Atoi(idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao converte idNatu: %v para int.", idCtxt)
		return nil, fmt.Errorf("Erro ao converte idNatu: %v para int.", idCtxt)
	}

	doc, err := obj.IndexDecisoes.ConsultaDocumentoByIdCtxt(id)
	if err != nil {
		logger.Log.Info("Erro ao selecionar documentos do índice decisoes!")
		return nil, err
	}
	return doc, nil
}

// Busca documento pelo ID usando o modelo autos_embedding
func (obj *AutosEmbeddingType) GetDecisoesByNatureza(idNatu string) ([]opensearch.ResponseDecisoes, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar IndexAutosService global sem inicializá-la.")
		return nil, fmt.Errorf("IndexAutosService global não configurada")
	}
	id, err := strconv.Atoi(idNatu)
	if err != nil {
		logger.Log.Errorf("Erro ao converte idNatu: %v para int.", idNatu)
		return nil, fmt.Errorf("Erro ao converte idNatu: %v para int.", idNatu)
	}

	doc, err := obj.IndexDecisoes.ConsultaDocumentosByIdNatu(id)
	if err != nil {
		logger.Log.Info("Erro ao selecionar documentos do índice autos_embedding!")
		return nil, err
	}
	return doc, nil
}

func (obj *AutosEmbeddingType) IsEmbedding(idCtxt int, idPje string) (bool, error) {
	// Validação dos parâmetros
	if idCtxt <= 0 || idPje == "" {
		return false, fmt.Errorf("parâmetros inválidos: idCtxt=%d, idPje=%q", idCtxt, idPje)
	}
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar IndexAutosService global sem inicializá-la.")
		return false, fmt.Errorf("IndexAutosService global não configurada")
	}

	// Se total hits > 0, existe documento correspondente
	return obj.IndexAutos.IsDocumentoEmbedding(idCtxt, idPje)
}

func (obj *AutosEmbeddingType) ParseJsonToEmbedding(idNatu int, doc string) (string, error) {

	if obj == nil {
		logger.Log.Error("Tentativa de utilizar AutosEmbeddingType global sem inicializá-la.")
		return "", fmt.Errorf("AutosEmbeddingType global não configurada")
	}

	switch idNatu {
	case consts.NATU_DOC_INICIAL:
		return "Petição inicial", nil
	case consts.NATU_DOC_CONTESTACAO:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_REPLICA:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_DESP_INI:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_DESP_ORD:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_PETICAO:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_DECISAO:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_SENTENCA:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_EMBARGOS:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_CONTRA_RAZ:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_APELACAO:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_PROCURACAO:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_ROL_TESTEMUNHAS:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_CONTRATO:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_LAUDO_PERICIA:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_ATA_AUDIENCIA:
		return consts.GetNaturezaDocumento(idNatu), nil
	case consts.NATU_DOC_PARECER_MP:
		return consts.GetNaturezaDocumento(idNatu), nil
	case 1000:
		return consts.GetNaturezaDocumento(idNatu), nil
	default:
		return "Não identificado", nil
	}

}

/*
*
Gera o embedding para o índice autos_embedding.
Como não há campos de texto fixos, essa função recebe o texto completo do documento (doc string)
e retorna a estrutura com embedding pronta para indexação.
*/
// func (obj *AutosEmbeddingType) GetDocumentoEmbeddingFromText(docText string, idCtxt int, idNatu int) ([]float32, error) {
// 	ctx := context.Background()

// 	// Gera o embedding do texto inteiro
// 	embeddingResp, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, docText)
// 	if err != nil {
// 		return nil, fmt.Errorf("erro ao gerar embedding do texto: %w", err)
// 	}

// 	vector32 := services.OpenaiServiceGlobal.Float64ToFloat32Slice(embeddingResp)

// 	return vector32, nil
// }
