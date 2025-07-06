package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type IndexDecisoesType struct {
	osCli     *opensearchapi.Client
	indexName string
}

// Cria um novo cliente para o índice decisões
func NewIndexDecisoes() *IndexDecisoesType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}

	return &IndexDecisoesType{
		osCli:     osClient,
		indexName: "decisoes", // Use o nome fixo do índice decisões ou configure externamente
	}
}

// Estrutura do documento no índice decisões
type DecisoesDoc struct {
	IdCtxt int    `json:"id_ctxt"`
	IdNatu int    `json:"id_natu"`
	IdPje  string `json:"id_pje"`
	Doc    string `json:"doc"`
}

type DecisoesEmbedding struct {
	IdCtxt       int       `json:"id_ctxt"`
	IdNatu       int       `json:"id_natu"`
	IdPje        string    `json:"id_pje"`
	Doc          string    `json:"doc"`
	DocEmbedding []float32 `json:"doc_embedding"`
}

type BodyDecisoesUpdate struct {
	Doc DecisoesDoc `json:"doc"`
}

type ResponseDecisoes struct {
	Id     string `json:"id"`
	IdCtxt int    `json:"id_ctxt"`
	IdNatu int    `json:"id_natu"`
	IdPje  string `json:"id_pje"`
	Doc    string `json:"doc"`
}

type searchResponseDecisoes struct {
	Hits struct {
		Hits []struct {
			ID     string           `json:"_id"`
			Source ResponseDecisoes `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// Indexar um novo documento no índice decisões
func (idx *IndexDecisoesType) IndexaDocumento(paramsData DecisoesEmbedding) (*opensearchapi.IndexResp, error) {
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

// Atualizar documento no índice decisões
func (idx *IndexDecisoesType) UpdateDocumento(id string, paramsData DecisoesDoc) (*opensearchapi.UpdateResp, error) {
	updateData := BodyDecisoesUpdate{Doc: paramsData}

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

// Deletar documento pelo ID no índice decisões
func (idx *IndexDecisoesType) DeleteDocumento(id string) (*opensearchapi.DocumentDeleteResp, error) {

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

// Consultar documento pelo ID no índice decisões
func (idx *IndexDecisoesType) ConsultaDocumentoById(id string) (*ResponseDecisoes, error) {

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

	doc := &ResponseDecisoes{Id: id}
	source := result["_source"].(map[string]interface{})

	// Conversão segura
	if v, ok := source["id_ctxt"].(float64); ok {
		doc.IdCtxt = int(v)
	}
	if v, ok := source["id_natu"].(float64); ok {
		doc.IdNatu = int(v)
	}
	if v, ok := source["doc"].(string); ok {
		doc.Doc = v
	}

	return doc, nil
}

func (idx *IndexDecisoesType) ConsultaDocumentoByIdCtxt(idCtxt int) ([]ResponseDecisoes, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	query := map[string]interface{}{
		"size": 10, // aumentar o número de documentos retornados
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"id_ctxt": idCtxt,
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	res, err := idx.osCli.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryJSON),
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao executar busca no OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound || res.Inspect().Response.StatusCode == http.StatusNoContent {
		msg := fmt.Sprintf("Documento com id_ctxt %d não encontrado no índice %s", idCtxt, idx.indexName)
		logger.Log.Info(msg)
		return nil, nil
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string           `json:"_id"`
				Source ResponseDecisoes `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]ResponseDecisoes, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		doc.Id = hit.ID
		docs = append(docs, doc)
	}

	return docs, nil
}

func (idx *IndexDecisoesType) ConsultaDocumentosByIdNatu(idNatu int) ([]ResponseDecisoes, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	// Monta query para buscar documentos com id_natu igual ao parâmetro
	query := map[string]interface{}{
		"size": 100, // limite arbitrário, ajuste conforme necessidade
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"id_natu": idNatu,
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	res, err := idx.osCli.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryJSON),
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao executar busca no OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound || res.Inspect().Response.StatusCode == http.StatusNoContent {
		logger.Log.Info(fmt.Sprintf("Nenhum documento encontrado para id_natu %d", idNatu))
		return nil, nil
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string           `json:"_id"`
				Source ResponseDecisoes `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	docs := make([]ResponseDecisoes, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		doc.Id = hit.ID
		docs = append(docs, doc)
	}

	return docs, nil
}

/*
*
Faz uma busca semântica, utilizando embedding passado em vector,
limitando a resposta a 5 registros no máximo
*/
func (idx *IndexDecisoesType) ConsultaSemantica(vector []float32, idNatuFilter int) ([]ResponseDecisoes, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	if len(vector) != ExpectedVectorSize {
		msg := fmt.Sprintf("Erro: o vetor enviado tem dimensão %d, mas o índice espera %d dimensões.", len(vector), ExpectedVectorSize)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg)
	}

	// Monta a query KNN com filtro por id_natu (se passado)
	boolQuery := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []interface{}{
				map[string]interface{}{
					"knn": map[string]interface{}{
						"doc_embedding": map[string]interface{}{
							"vector": vector,
							"k":      10,
						},
					},
				},
			},
		},
	}

	// Se filtro id_natu for maior que zero, adiciona filtro
	if idNatuFilter > 0 {
		boolQuery["bool"].(map[string]interface{})["filter"] = []interface{}{
			map[string]interface{}{
				"term": map[string]interface{}{
					"id_natu": idNatuFilter,
				},
			},
		}
	}

	query := map[string]interface{}{
		"size": 10,
		"_source": map[string]interface{}{
			"excludes": []string{"doc_embedding"},
		},
		"query": boolQuery,
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	res, err := idx.osCli.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryJSON),
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar o OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}
	defer res.Inspect().Response.Body.Close()

	var result searchResponseDecisoes
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}

	var documentos []ResponseDecisoes
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		doc.Id = hit.ID

		// Aplica filtro local de id_natu, caso desejado (redundante se usado no filtro ES)
		if idNatuFilter > 0 && doc.IdNatu != idNatuFilter {
			continue
		}

		documentos = append(documentos, doc)
		if len(documentos) >= 5 {
			break
		}
	}

	return documentos, nil
}
