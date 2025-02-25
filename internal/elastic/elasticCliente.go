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
		log.Println("Erro ao obter uma instância do cliente Elasticsearch:", err)
		return nil
	}

	return &ElasticClienteType{esCli: es}
}

// Estrutura do documento no Elasticsearch
type ModelosRow struct {
	Id           string `json:"id"`
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
}

// Estrutura para decodificar resposta da consulta
type searchResponse struct {
	Hits struct {
		Hits []struct {
			ID     string     `json:"_id"`
			Source ModelosRow `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// Obtém informações sobre o cluster Elasticsearch
func (cliente *ElasticClienteType) Info() (*esapi.Response, error) {
	res, err := cliente.esCli.Cluster.Health()
	if err != nil {
		log.Printf("Erro ao extrair informações do cluster Elasticsearch: %v", err)
	}
	return res, err
}

// Verifica se um índice existe
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
func (cliente *ElasticClienteType) IndexDocumento(indexName string, paramsData ModelosRow) (*esapi.Response, error) {
	data, err := json.Marshal(paramsData)
	if err != nil {
		log.Println("Erro ao serializar JSON:", err)
		return nil, err
	}

	res, err := cliente.esCli.Index(
		indexName,
		bytes.NewReader(data),
		//cliente.esCli.Index.WithDocumentID(""),  // Document ID
		cliente.esCli.Index.WithRefresh("true"), // Refresh
	)
	if err != nil {
		log.Printf("Erro ao indexar documento no Elasticsearch: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	return res, nil
}

// Atualiza um documento existente
func (cliente *ElasticClienteType) UpdateDocumento(indexName, id string, paramsData ModelosRow) (*esapi.Response, error) {
	data, err := json.Marshal(paramsData)
	if err != nil {
		log.Println("Erro ao serializar JSON:", err)
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
		log.Printf("Erro ao atualizar documento no Elasticsearch: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	// Verificar se a resposta contém erro
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		log.Printf("Erro na resposta do Elasticsearch: %s", body)
		return res, fmt.Errorf("erro ao deletar documento: %s", res.Status())
	}
	return res, nil
}

// 🔹 **Consulta um documento pelo ID**
func (cliente *ElasticClienteType) ConsultaDocumento(indexName, id string) (*ModelosRow, error) {
	res, err := cliente.esCli.Get(
		indexName,
		id,
	)

	if err != nil {
		log.Printf("Erro ao consultar documento %s no índice %s: %v", id, indexName, err)
		return nil, err
	}
	defer res.Body.Close()

	// Verifica se o documento não foi encontrado
	if res.StatusCode == http.StatusNotFound {
		log.Printf("Documento %s não encontrado no índice %s", id, indexName)
		return nil, nil
	}

	// Captura e exibe erro caso a resposta seja inválida
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		log.Printf("Erro na resposta do Elasticsearch: %s", body)
		return nil, fmt.Errorf("erro ao buscar documento: %s", res.Status())
	}

	// Decodifica a resposta JSON
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Println("Erro ao decodificar resposta JSON:", err)
		return nil, err
	}

	// Extrai os dados do documento
	source, ok := result["_source"].(map[string]interface{})
	if !ok {
		log.Println("Erro: resposta do Elasticsearch não contém _source")
		return nil, fmt.Errorf("erro ao extrair _source")
	}

	// Preenchendo a estrutura com os valores do documento
	doc := &ModelosRow{
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

// 🔹 **Consulta por conteúdo no campo "inteiro_teor"**
func (cliente *ElasticClienteType) ConsultaPorConteudo(indexName, search_texto string) ([]ModelosRow, error) {
	if cliente.esCli == nil {
		log.Println("Erro: Elasticsearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao Elasticsearch")
	}

	// Construindo a query JSON para busca
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"inteiro_teor": search_texto, // Busca no campo "inteiro_teor"
			},
		},
	}

	// Convertendo a query para JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		log.Println("Erro ao serializar query JSON:", err)
		return nil, err
	}

	// Fazendo a busca diretamente com a função Search
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
	defer res.Body.Close()

	// Verificando status da resposta
	if res.IsError() {
		log.Printf("Erro na resposta do Elasticsearch: %s", res.String())
		return nil, fmt.Errorf("erro na consulta ao Elasticsearch")
	}

	// Decodificando resposta
	var result searchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Println("Erro ao decodificar resposta JSON:", err)
		return nil, err
	}

	// Extraindo documentos encontrados e incluindo o ID do documento
	var documentos []ModelosRow
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		doc.Id = hit.ID // Adiciona o _id do documento

		documentos = append(documentos, doc)
	}

	// Retornando resultados
	return documentos, nil
}
