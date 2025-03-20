package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type OpenSearchClienteType struct {
	osCli *opensearchapi.Client
}

// Função para criar um novo cliente OpenSearch
func NewOpenSearchCliente() *OpenSearchClienteType {
	osClient, err := OServer.GetClient()
	if err != nil {
		log.Printf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		return nil
	}

	return &OpenSearchClienteType{osCli: osClient}
}

// Estrutura do documento no OpenSearch
type ModelosDoc struct {
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
}

type UpdateBody struct {
	Doc ModelosDoc `json:"doc"`
}

type ModelosResponse struct {
	Id           string `json:"id"`
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
}

type searchResponse struct {
	Hits struct {
		Hits []struct {
			ID     string          `json:"_id"`
			Source ModelosResponse `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// Obter informações do cluster
func (cliente *OpenSearchClienteType) Info() (*opensearchapi.ClusterHealthResp, error) {

	res, err := cliente.osCli.Cluster.Health(context.Background(), &opensearchapi.ClusterHealthReq{})
	if err != nil {
		log.Printf("Erro ao extrair informações do cluster OpenSearch: %v", err)
		return nil, err
	}
	return res, nil
}

// Verifica se o índice existe
func (cliente *OpenSearchClienteType) IndicesExists(indexStr string) (bool, error) {

	res, err := cliente.osCli.Indices.Exists(
		context.Background(),
		opensearchapi.IndicesExistsReq{
			Indices: []string{indexStr},
		})
	if err != nil {
		log.Printf("Erro ao verificar a existência do índice %s no OpenSearch: %v", indexStr, err)
		return false, err
	}
	defer res.Body.Close()

	return res.StatusCode == http.StatusOK, nil
}

// Indexar um novo documento
func (cliente *OpenSearchClienteType) IndexDocumento(indexName string, paramsData ModelosDoc) (*opensearchapi.IndexResp, error) {
	data, err := json.Marshal(paramsData)
	if err != nil {
		log.Printf("Erro ao serializar JSON: %v", err)
		return nil, err
	}

	req, err := cliente.osCli.Index(context.Background(),
		opensearchapi.IndexReq{
			Index:      indexName,
			DocumentID: "",
			Body:       bytes.NewReader(data),
		})

	if err != nil {
		log.Printf("Erro ao indexar documento no OpenSearch: %v", err)
		return nil, err
	}
	defer req.Inspect().Response.Body.Close()

	return req, nil
}

// Atualizar documento
func (cliente *OpenSearchClienteType) UpdateDocumento(indexName, id string, paramsData ModelosDoc) (*opensearchapi.UpdateResp, error) {
	updateData := UpdateBody{Doc: paramsData}
	data, err := json.Marshal(updateData)
	if err != nil {
		log.Printf("Erro ao serializar JSON: %v", err)
		return nil, err
	}

	res, err := cliente.osCli.Update(
		context.Background(),
		opensearchapi.UpdateReq{
			Index:      indexName,
			DocumentID: id,
			Body:       bytes.NewReader(data),
		})

	// m, e := json.MarshalIndent(updateData, "", "    ")
	// if e != nil {
	// 	log.Fatalf("JSON marshaling failhou!: %s", m)
	// }
	// log.Printf("%s\n", m)

	if err != nil {
		log.Printf("Erro ao atualizar documento no OpenSearch: %v", err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	return res, nil
}

// Deletar documento
func (cliente *OpenSearchClienteType) DeleteDocumento(indexName, id string) (*opensearchapi.DocumentDeleteResp, error) {

	res, err := cliente.osCli.Document.Delete(
		context.Background(),
		opensearchapi.DocumentDeleteReq{
			Index:      indexName,
			DocumentID: id,
		})

	if err != nil {
		log.Printf("Erro ao deletar documento no OpenSearch: %v", err)
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

// Consulta um documento pelo ID
func (cliente *OpenSearchClienteType) ConsultaDocumento(indexName, id string) (*ModelosResponse, error) {

	res, err := cliente.osCli.Document.Get(context.Background(),
		opensearchapi.DocumentGetReq{
			Index:      indexName,
			DocumentID: id,
		})

	if err != nil {
		log.Printf("Erro ao consultar documento %s no índice %s: %v", id, indexName, err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound {
		log.Printf("Documento %s não encontrado no índice %s", id, indexName)
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		log.Printf("Erro ao decodificar resposta JSON: %v", err)
		return nil, err
	}

	doc := &ModelosResponse{Id: id}
	source := result["_source"].(map[string]interface{})
	doc.Natureza = source["natureza"].(string)
	doc.Ementa = source["ementa"].(string)
	doc.Inteiro_teor = source["inteiro_teor"].(string)

	return doc, nil
}

// Consulta por conteúdo no OpenSearch
func (cliente *OpenSearchClienteType) ConsultaPorConteudo(indexName, search_texto, natureza string) ([]ModelosResponse, error) {
	if cliente.osCli == nil {
		log.Printf("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":  search_texto,
							"fields": []string{"ementa", "inteiro_teor"},
						},
					},
				},
				"filter": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"natureza": natureza,
						},
					},
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		log.Printf("Erro ao serializar query JSON: %v", err)
		return nil, err
	}

	res, err := cliente.osCli.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{indexName},
			Body:    bytes.NewReader(queryJSON)},
	)

	if err != nil {
		log.Printf("Erro ao consultar o OpenSearch: %v", err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result searchResponse
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		log.Printf("Erro ao decodificar resposta JSON: %v", err)
		return nil, err
	}

	var documentos []ModelosResponse
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		doc.Id = hit.ID

		documentos = append(documentos, doc)
	}

	return documentos, nil
}
