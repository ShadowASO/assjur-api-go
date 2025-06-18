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
		log.Printf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		return nil
	}

	return &IndexModelosType{
		osCli:     osClient,
		indexName: "modelos_semantico",
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
		log.Printf("Erro ao serializar JSON: %v", err)
		return nil, err
	}

	req, err := idx.osCli.Index(context.Background(),
		opensearchapi.IndexReq{
			Index:      idx.indexName,
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
func (idx *IndexModelosType) UpdateDocumento(id string, paramsData ModelosText) (*opensearchapi.UpdateResp, error) {
	updateData := BodyModelosUpdate{Doc: paramsData}

	data, err := json.Marshal(updateData)
	if err != nil {
		log.Printf("Erro ao serializar JSON: %v", err)
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
		log.Printf("Erro ao atualizar documento no OpenSearch: %v", err)
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

// Devolve o documento identificado pelo ID
func (idx *IndexModelosType) ConsultaDocumentoById(id string) (*ResponseModelos, error) {

	res, err := idx.osCli.Document.Get(context.Background(),
		opensearchapi.DocumentGetReq{
			Index:      idx.indexName,
			DocumentID: id,
		})

	if err != nil {
		log.Printf("Erro ao consultar documento %s no índice %s: %v", id, idx.indexName, err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound {
		log.Printf("Documento %s não encontrado no índice %s", id, idx.indexName)
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		log.Printf("Erro ao decodificar resposta JSON: %v", err)
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
		log.Printf("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}
	// Validação de dimensão
	if len(vector) != ExpectedVectorSize {
		log.Fatalf("Erro: o vetor enviado tem dimensão %d, mas o índice espera %d dimensões.", len(vector), ExpectedVectorSize)
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
		log.Printf("Erro ao serializar query JSON: %v", err)
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
		log.Printf("Erro ao consultar o OpenSearch: %v", err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	// Decodifica a resposta
	var result searchResponse
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		log.Printf("Erro ao decodificar resposta JSON: %v", err)
		return nil, err
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

// /*
// *
// Obtem o embedding de cada campo texto do index Modelos e devolve uma strutura.
// */
// func (cliente *IndexModelosType) GetDocumentoEmbeddings(doc ModelosText) (ModelosEmbedding, error) {

// 	modelo := ModelosEmbedding{
// 		Natureza:     doc.Natureza,
// 		Ementa:       doc.Ementa,
// 		Inteiro_teor: doc.Inteiro_teor,
// 	}

// 	// Gera o embedding da ementa
// 	//ementaResp, err := openAI.Service.GetEmbeddingFromText(modelo.Ementa)
// 	ementaResp, err := cliente.openAi.GetEmbeddingFromText(modelo.Ementa)
// 	if err != nil {
// 		return modelo, fmt.Errorf("erro ao gerar embedding da ementa: %w", err)
// 	}
// 	//modelo.EmentaEmbedding = openAI.Float64ToFloat32Slice(ementaResp.Data[0].Embedding)
// 	modelo.EmentaEmbedding = services.Float64ToFloat32Slice(ementaResp.Data[0].Embedding)

// 	// Gera o embedding do inteiro teor
// 	//teorResp, err := openAI.Service.GetEmbeddingFromText(doc.Inteiro_teor)
// 	teorResp, err := cliente.openAi.GetEmbeddingFromText(doc.Inteiro_teor)
// 	if err != nil {
// 		return modelo, fmt.Errorf("erro ao gerar embedding do inteiro teor: %w", err)
// 	}
// 	modelo.InteiroTeorEmbedding = services.Float64ToFloat32Slice(teorResp.Data[0].Embedding)

// 	return modelo, nil
// }
