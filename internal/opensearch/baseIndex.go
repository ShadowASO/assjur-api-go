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
	"sync"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

const ExpectedRagVectorSize = 3072

type BaseIndexType struct {
	osCli     *opensearchapi.Client
	indexName string
}

var BaseIndexGlobal *BaseIndexType
var onceInitBaseIndex sync.Once

// Inicializa global
func InitBaseIndex() {
	onceInitBaseIndex.Do(func() {
		BaseIndexGlobal = NewBaseIndex()
		logger.Log.Info("Global RagService configurado com sucesso.")
	})
}

func NewBaseIndex() *BaseIndexType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}
	return &BaseIndexType{
		osCli:     osClient,
		indexName: config.GlobalConfig.OpenSearchRagName,
	}
}

// Estruturas do RAG
// type RagText struct {
// 	DataTexto string `json:"data_texto"`
// }

type BaseRow struct {
	Id            string    `json:"id"`
	IdPje         string    `json:"id_pje"`
	Classe        string    `json:"classe"`
	Assunto       string    `json:"assunto"`
	Natureza      string    `json:"natureza"`
	Tipo          string    `json:"tipo"`
	Tema          string    `json:"tema"`
	Fonte         string    `json:"fonte"`
	DataTexto     string    `json:"data_texto"`
	DataEmbedding []float32 `json:"data_embedding"`
}

//	type BodyRagUpdate struct {
//		Doc RagText `json:"doc"`
//	}
type BaseUpdate struct {
	DataTexto     string    `json:"data_texto,omitempty"`
	DataEmbedding []float32 `json:"data_embedding,omitempty"`
}

type BodyBaseUpdate struct {
	Doc BaseUpdate `json:"doc"`
}

type ResponseBase struct {
	Id        string `json:"id"`
	IdPje     string `json:"id_pje"`
	Classe    string `json:"classe"`
	Assunto   string `json:"assunto"`
	Natureza  string `json:"natureza"`
	Tipo      string `json:"tipo"`
	Tema      string `json:"tema"`
	Fonte     string `json:"fonte"`
	DataTexto string `json:"data_texto"`
}

type ParamsBaseInsert struct {
	IdPje         string    `json:"id_pje"`
	Classe        string    `json:"classe"`
	Assunto       string    `json:"assunto"`
	Natureza      string    `json:"natureza"`
	Tipo          string    `json:"tipo"`
	Tema          string    `json:"tema"`
	Fonte         string    `json:"fonte"`
	DataTexto     string    `json:"data_texto"`
	DataEmbedding []float32 `json:"data_embedding"`
}

// Indexar documento
func (idx *BaseIndexType) IndexaDocumento(params ParamsBaseInsert) (*opensearchapi.IndexResp, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar JSON: %w", err)
	}
	req, err := idx.osCli.Index(context.Background(),
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: "",
			Body:       bytes.NewReader(data),
		})
	if err != nil {
		return nil, fmt.Errorf("erro ao indexar documento: %w", err)
	}
	defer req.Inspect().Response.Body.Close()
	return req, nil
}

// Atualizar documento
func (idx *BaseIndexType) UpdateDocumento(id string, params BaseUpdate) (*opensearchapi.UpdateResp, error) {
	data, err := json.Marshal(BodyBaseUpdate{Doc: params})
	if err != nil {
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
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()
	return res, nil
}

// Deletar documento
func (idx *BaseIndexType) DeleteDocumento(id string) (*opensearchapi.DocumentDeleteResp, error) {
	res, err := idx.osCli.Document.Delete(context.Background(),
		opensearchapi.DocumentDeleteReq{
			Index:      idx.indexName,
			DocumentID: id,
		})
	if err != nil {
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

// Consulta por ID
func (idx *BaseIndexType) ConsultaDocumentoById(id string) (*ResponseBase, error) {
	res, err := idx.osCli.Document.Get(context.Background(),
		opensearchapi.DocumentGetReq{
			Index:      idx.indexName,
			DocumentID: id,
		})
	if err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var result struct {
		Source ResponseBase `json:"_source"`
	}
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		return nil, err
	}

	doc := result.Source
	doc.Id = id
	return &doc, nil
}

// Busca semântica
func (idx *BaseIndexType) ConsultaSemantica(vector []float32, natureza string) ([]ResponseBase, error) {
	if idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	if len(vector) != ExpectedRagVectorSize {
		return nil, erros.CreateError(fmt.Sprintf("vetor tem %d dimensões, esperado %d", len(vector), ExpectedRagVectorSize))
	}

	knnQuery := map[string]interface{}{
		"knn": map[string]interface{}{
			"data_embedding": map[string]interface{}{
				"vector": vector,
				"k":      20,
			},
		},
	}
	if natureza != "" {
		knnQuery = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": knnQuery,
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{"natureza": natureza},
					},
				},
			},
		}
	}

	query := map[string]interface{}{
		"size": 10,
		"_source": map[string]interface{}{
			"excludes": []string{"data_embedding"},
		},
		"query": knnQuery,
	}

	queryJSON, _ := json.Marshal(query)
	res, err := idx.osCli.Search(context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryJSON),
		})
	if err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string       `json:"_id"`
				Source ResponseBase `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		return nil, err
	}

	var docs []ResponseBase
	for _, h := range result.Hits.Hits {
		doc := h.Source
		doc.Id = h.ID
		docs = append(docs, doc)
	}
	return docs, nil
}
