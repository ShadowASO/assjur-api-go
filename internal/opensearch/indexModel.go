package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ocrserver/internal/config"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

const ExpectedVectorSize = 3072

type IndexModelosType struct {
	osCli     *opensearchapi.Client
	indexName string
	//openAi    *services.OpenaiServiceType
}

// Função para criar um novo cliente OpenSearch
// func NewIndexModelos(serviceOpenAi *services.OpenaiServiceType) *IndexModelosType {
func NewIndexModelos() *IndexModelosType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		//log.Printf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		//return nil
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}

	return &IndexModelosType{
		osCli:     osClient,
		indexName: config.GlobalConfig.OpenSearchIndexName,
	}
}

// Estrutura do documento no OpenSearch
type ModelosText struct {
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
}

type ModelosEmbedding struct {
	Natureza             string    `json:"natureza"`
	Ementa               string    `json:"ementa"`
	Inteiro_teor         string    `json:"inteiro_teor"`
	EmentaEmbedding      []float32 `json:"ementa_embedding"`
	InteiroTeorEmbedding []float32 `json:"inteiro_teor_embedding"`
}

type BodyModelosInsert struct {
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
}

type BodyModelosUpdate struct {
	Doc ModelosText `json:"doc"`
}
type BodyModelosSearch struct {
	Index_name   string `json:"index_name"`
	Natureza     string `json:"natureza"`
	Search_texto string `json:"search_texto"`
}

type ResponseModelos struct {
	Id           string `json:"id"`
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
}

type searchResponse struct {
	Hits struct {
		Hits []struct {
			ID     string          `json:"_id"`
			Source ResponseModelos `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// Indexar um novo documento

func (idx *IndexModelosType) IndexaDocumento(paramsData ModelosEmbedding) (*opensearchapi.IndexResp, error) {
	data, err := json.Marshal(paramsData)
	if err != nil {

		msg := fmt.Sprintf("Erro ao serializar JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	req, err := idx.osCli.Index(context.Background(),
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: "",
			Body:       bytes.NewReader(data),
		})

	if err != nil {

		msg := fmt.Sprintf("Erro ao indexar documento no OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer req.Inspect().Response.Body.Close()

	return req, nil
}

// Atualizar documento
func (idx *IndexModelosType) UpdateDocumento(id string, paramsData ModelosText) (*opensearchapi.UpdateResp, error) {
	updateData := BodyModelosUpdate{Doc: paramsData}

	data, err := json.Marshal(updateData)
	if err != nil {

		msg := fmt.Sprintf("Erro ao serializar JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	res, err := idx.osCli.Update(
		context.Background(),
		opensearchapi.UpdateReq{
			Index:      idx.indexName,
			DocumentID: id,
			Body:       bytes.NewReader(data),
		})

	if err != nil {

		msg := fmt.Sprintf("Erro ao atualizar documento no OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	return res, nil
}

// Deletar documento identificado pelo ID
func (idx *IndexModelosType) DeleteDocumento(id string) (*opensearchapi.DocumentDeleteResp, error) {

	res, err := idx.osCli.Document.Delete(
		context.Background(),
		opensearchapi.DocumentDeleteReq{
			Index:      idx.indexName,
			DocumentID: id,
		})

	if err != nil {

		msg := fmt.Sprintf("Erro ao deletar documento no OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Inspect().Response.Body)
		log.Printf("Erro na resposta do OpenSearch: %s", body)
		return res, fmt.Errorf("erro ao deletar documento: %s", res.Inspect().Response.Status())
	}
	return res, nil
}

// Devolve o documento identificado pelo ID
func (idx *IndexModelosType) ConsultaDocumentoById(id string) (*ResponseModelos, error) {

	res, err := idx.osCli.Document.Get(context.Background(),
		opensearchapi.DocumentGetReq{
			Index:      idx.indexName,
			DocumentID: id,
		})

	if err != nil {

		msg := fmt.Sprintf("Erro ao consultar documento %s no índice %s: %v", id, idx.indexName, err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound {

		msg := fmt.Sprintf("Documento %s não encontrado no índice %s", id, idx.indexName)
		logger.Log.Error(msg)
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {

		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	doc := &ResponseModelos{Id: id}
	source := result["_source"].(map[string]interface{})
	doc.Natureza = source["natureza"].(string)
	doc.Ementa = source["ementa"].(string)
	doc.Inteiro_teor = source["inteiro_teor"].(string)

	return doc, nil
}

/*
*
Faz uma busca semântica, utilizando os embeddings passados em vector e filtra por natureza,
limitando a resposta a 5 registros no máximo
*/

func (idx *IndexModelosType) ConsultaSemantica(vector []float32, natureza string) ([]ResponseModelos, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}
	// Validação de dimensão
	if len(vector) != ExpectedVectorSize {

		msg := fmt.Sprintf("Erro: o vetor enviado tem dimensão %d, mas o índice espera %d dimensões.", len(vector), ExpectedVectorSize)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg)
	}

	// Monta a query principal com knn (sem filtro de natureza)
	query := map[string]interface{}{
		"size": 10,
		"_source": map[string]interface{}{
			"excludes": []string{"ementa_embedding", "inteiro_teor_embedding"},
		},
		"query": map[string]interface{}{
			"knn": map[string]interface{}{
				"inteiro_teor_embedding": map[string]interface{}{
					"vector": vector,
					"k":      10,
				},
			},
		},
	}

	// Serializa a query para JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		//log.Printf("Erro ao serializar query JSON: %v", err)
		msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	// Envia a requisição
	res, err := idx.osCli.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryJSON),
		},
	)
	if err != nil {
		//log.Printf("Erro ao consultar o OpenSearch: %v", err)
		msg := fmt.Sprintf("Erro ao consultar o OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}
	defer res.Inspect().Response.Body.Close()

	// Decodifica a resposta
	var result searchResponse
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		//log.Printf("Erro ao decodificar resposta JSON: %v", err)
		//return nil, err
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}

	// Monta a lista de documentos retornados, aplicando filtro manual de natureza (se necessário)
	var documentos []ResponseModelos
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		doc.Id = hit.ID

		// Aplica filtro local de natureza (caso informado)
		if natureza != "" && doc.Natureza != natureza {
			continue
		}

		documentos = append(documentos, doc)
		if len(documentos) >= 5 {
			break
		}
	}

	return documentos, nil
}
