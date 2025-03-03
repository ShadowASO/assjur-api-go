package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type ElasticClienteType struct {
	esCli *elasticsearch.Client
}

// Função para criar um novo cliente Elasticsearch
func NewElasticCliente() *ElasticClienteType {
	es, err := EServer.GetClienteElastic()
	if err != nil {
		log.Printf("Erro ao obter uma instância do cliente Elasticsearch: %v", err)
		return nil
	}

	return &ElasticClienteType{esCli: es}
}

// Estrutura do documento no Elasticsearch
type ModelosDoc struct {
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
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

// Função para obter informações do cluster
func (cliente *ElasticClienteType) Info() (*esapi.Response, error) {
	res, err := cliente.esCli.Cluster.Health()
	if err != nil {
		log.Printf("Erro ao extrair informações do cluster Elasticsearch: %v", err)
	}
	return res, err
}

// Verifica se o índice existe
func (cliente *ElasticClienteType) IndicesExists(indexStr string) (bool, error) {
	res, err := cliente.esCli.Indices.Exists([]string{indexStr})
	if err != nil {
		log.Printf("Erro ao verificar a existência do índice %s no Elasticsearch: %v", indexStr, err)
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, nil
}

// Indexa um novo documento
func (cliente *ElasticClienteType) IndexDocumento(indexName string, paramsData ModelosDoc) (*esapi.Response, error) {
	log.Println(paramsData)
	data, err := json.Marshal(paramsData)
	if err != nil {
		log.Printf("Erro ao serializar JSON: %v", err)
		return nil, err
	}
	log.Println(string(data))

	res, err := cliente.esCli.Index(
		indexName,
		bytes.NewReader(data),
		cliente.esCli.Index.WithRefresh("true"), // Refresh
	)
	if err != nil {
		log.Printf("Erro ao indexar documento no Elasticsearch: %v", err)
		return nil, err
	}

	if res != nil {
		defer res.Body.Close()
	}

	return res, nil
}

// Atualiza um documento existente
func (cliente *ElasticClienteType) UpdateDocumento(indexName, id string, paramsData ModelosDoc) (*esapi.Response, error) {
	data, err := json.Marshal(paramsData)
	if err != nil {
		log.Printf("Erro ao serializar JSON: %v", err)
		return nil, err
	}

	res, err := cliente.esCli.Update(
		indexName,
		id,
		bytes.NewReader(data),
		cliente.esCli.Update.WithRefresh("true"), // Refresh
	)

	if err != nil {
		log.Printf("Erro ao atualizar documento no Elasticsearch: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	return res, nil
}

// Deleta um documento pelo ID
func (cliente *ElasticClienteType) DeleteDocumento(indexName, id string) (*esapi.Response, error) {
	res, err := cliente.esCli.Delete(
		indexName,
		id,
		cliente.esCli.Delete.WithRefresh("true"),
	)
	if err != nil {
		log.Printf("Erro ao deletar documento no Elasticsearch: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		log.Printf("Erro na resposta do Elasticsearch: %s", body)
		return res, fmt.Errorf("erro ao deletar documento: %s", res.Status())
	}
	return res, nil
}

// Consulta um documento pelo ID
func (cliente *ElasticClienteType) ConsultaDocumento(indexName, id string) (*ModelosResponse, error) {
	res, err := cliente.esCli.Get(indexName, id)
	if err != nil {
		log.Printf("Erro ao consultar documento %s no índice %s: %v", id, indexName, err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		log.Printf("Documento %s não encontrado no índice %s", id, indexName)
		return nil, nil
	}

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		log.Printf("Erro na resposta do Elasticsearch: %s", body)
		return nil, fmt.Errorf("erro ao buscar documento: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Printf("Erro ao decodificar resposta JSON: %v", err)
		return nil, err
	}

	source, ok := result["_source"].(map[string]interface{})
	if !ok {
		log.Println("Erro: resposta do Elasticsearch não contém _source")
		return nil, fmt.Errorf("erro ao extrair _source")
	}

	doc := &ModelosResponse{
		Id: id, // Adiciona o ID do documento
	}

	if natureza, exists := source["natureza"].(string); exists {
		doc.Natureza = natureza
	}
	if ementa, exists := source["ementa"].(string); exists {
		doc.Ementa = ementa
	}
	if inteiro_teor, exists := source["inteiro_teor"].(string); exists {
		doc.Inteiro_teor = inteiro_teor
	}

	return doc, nil
}

// Consulta por conteúdo no campo "inteiro_teor"
func (cliente *ElasticClienteType) ConsultaPorConteudo(indexName, search_texto string) ([]ModelosResponse, error) {
	if cliente.esCli == nil {
		log.Printf("Erro: Elasticsearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao Elasticsearch")
	}

	// Construção da query com multi_match para os campos "ementa" e "inteiro_teor"
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": search_texto, // Texto de busca
				"fields": []string{
					"ementa",       // Campo "ementa"
					"inteiro_teor", // Campo "inteiro_teor"
				},
			},
		},
	}

	// Serialização da query para JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		log.Printf("Erro ao serializar query JSON: %v", err)
		return nil, err
	}

	// Executa a consulta no Elasticsearch
	res, err := cliente.esCli.Search(
		cliente.esCli.Search.WithContext(context.Background()),
		cliente.esCli.Search.WithIndex(indexName),
		cliente.esCli.Search.WithBody(bytes.NewReader(queryJSON)),
		cliente.esCli.Search.WithTrackTotalHits(true), // Retorna o total de hits corretamente
		cliente.esCli.Search.WithPretty(),
	)

	if err != nil {
		log.Printf("Erro ao consultar o Elasticsearch: %v", err)
		return nil, err
	}
	// Verifica o status da resposta antes de chamar o defer
	if res.IsError() {
		log.Printf("Erro na resposta do Elasticsearch: %s", res.String())
		defer res.Body.Close() // Fecha o corpo da resposta apenas se for um erro
		return nil, fmt.Errorf("erro na consulta ao Elasticsearch")
	}

	// Certifique-se de chamar defer res.Body.Close() após garantir que não há erro
	defer res.Body.Close()

	// Decodifica a resposta do Elasticsearch
	var result searchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Printf("Erro ao decodificar resposta JSON: %v", err)
		return nil, err
	}

	// Preenche os documentos com os dados encontrados
	var documentos []ModelosResponse
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		doc.Id = hit.ID // Adiciona o _id do documento à estrutura

		documentos = append(documentos, doc)
	}

	// Retorna os documentos encontrados
	return documentos, nil
}
