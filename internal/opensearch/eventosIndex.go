package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"ocrserver/internal/types"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

// =========================================================
// Estrutura principal
// =========================================================

type EventosIndex struct {
	osCli     *opensearchapi.Client
	indexName string
}

// =========================================================
// Construtor e inicialização
// =========================================================

// Novo cliente para o índice "eventos"
func NewEventosIndex() *EventosIndex {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}

	return &EventosIndex{
		osCli:     osClient,
		indexName: "eventos",
	}
}

// =========================================================
// Estruturas auxiliares internas
// =========================================================

type searchResponseEventos struct {
	Hits struct {
		Hits []struct {
			ID     string     `json:"_id"`
			Source EventosRow `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// =========================================================
// Operações CRUD
// =========================================================

// Indexar um novo documento
func (idx *EventosIndex) Indexa(
	IdCtxt int,
	IdNatu int,
	IdPje string,
	Doc string,
	DocJsonRaw string,
	DocEmbedding []float32,
	idOptional string,
) (*ResponseEventosRow, error) {

	doc := EventosRow{
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJsonRaw:   DocJsonRaw,
		DocEmbedding: DocEmbedding,
	}

	res, err := idx.osCli.Index(
		context.Background(),
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: idOptional,
			Body:       opensearchutil.NewJSONReader(&doc),
			Params: opensearchapi.IndexParams{
				Refresh: "true",
			},
		},
	)
	if err != nil {
		logger.Log.Errorf("Erro ao indexar documento no OpenSearch: %v", err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	row := &ResponseEventosRow{
		Id:           res.ID,
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJsonRaw:   DocJsonRaw,
		DocEmbedding: DocEmbedding,
	}

	return row, nil
}

// Atualizar documento existente
func (idx *EventosIndex) Update(
	id string,
	IdCtxt int,
	IdNatu int,
	IdPje string,
	Doc string,
	DocJson string,
	DocEmbedding []float32,
) (*ResponseEventosRow, error) {

	doc := EventosRow{
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJsonRaw:   DocJson,
		DocEmbedding: DocEmbedding,
	}

	res, err := idx.osCli.Update(
		context.Background(),
		opensearchapi.UpdateReq{
			Index:      idx.indexName,
			DocumentID: id,
			Body:       opensearchutil.NewJSONReader(&doc),
			Params: opensearchapi.UpdateParams{
				Refresh: "true",
			},
		},
	)
	if err != nil {
		logger.Log.Errorf("Erro ao atualizar documento no OpenSearch: %v", err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	row := &ResponseEventosRow{
		Id:         id,
		IdCtxt:     IdCtxt,
		IdNatu:     IdNatu,
		IdPje:      IdPje,
		Doc:        Doc,
		DocJsonRaw: DocJson,
		//DocEmbedding: DocEmbedding,
	}

	return row, nil
}

// Deletar documento
func (idx *EventosIndex) Delete(id string) error {
	res, err := idx.osCli.Document.Delete(
		context.Background(),
		opensearchapi.DocumentDeleteReq{
			Index:      idx.indexName,
			DocumentID: id,
		},
	)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar documento: %v", err)
		return err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Inspect().Response.Body)
		log.Printf("Erro na resposta do OpenSearch: %s", body)
		return fmt.Errorf("erro ao deletar documento: %s", res.Inspect().Response.Status())
	}

	// Refresh para garantir visibilidade
	refreshRes, err := idx.osCli.Indices.Refresh(
		context.Background(),
		&opensearchapi.IndicesRefreshReq{
			Indices: []string{idx.indexName},
		},
	)
	if err != nil {
		logger.Log.Errorf("Erro ao fazer refresh do índice %s: %v", idx.indexName, err)
		return err
	}
	defer refreshRes.Inspect().Response.Body.Close()

	if refreshRes.Inspect().Response.StatusCode >= 400 {
		body, _ := io.ReadAll(refreshRes.Inspect().Response.Body)
		log.Printf("Erro no refresh: %s", body)
		return fmt.Errorf("erro ao fazer refresh do índice: %s", refreshRes.Inspect().Response.Status())
	}

	return nil
}

// Consultar documento pelo ID
func (idx *EventosIndex) ConsultaById(id string) (*ResponseEventosRow, error) {
	res, err := idx.osCli.Document.Get(
		context.Background(),
		opensearchapi.DocumentGetReq{
			Index:      idx.indexName,
			DocumentID: id,
		},
	)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar documento %s: %v", id, err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound {
		logger.Log.Infof("Documento %s não encontrado no índice %s", id, idx.indexName)
		return nil, nil
	}

	var docResp struct {
		Source EventosRow `json:"_source"`
	}
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&docResp); err != nil {
		logger.Log.Errorf("Erro ao decodificar resposta JSON: %v", err)
		return nil, err
	}

	return &ResponseEventosRow{
		Id:         id,
		IdCtxt:     docResp.Source.IdCtxt,
		IdNatu:     docResp.Source.IdNatu,
		IdPje:      docResp.Source.IdPje,
		Doc:        docResp.Source.Doc,
		DocJsonRaw: docResp.Source.DocJsonRaw,
		//DocEmbedding: docResp.Source.DocEmbedding,
	}, nil
}

// Consultar documentos por id_ctxt
func (idx *EventosIndex) ConsultaByIdCtxt(idCtxt int) ([]ResponseEventosRow, error) {
	if idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"id_ctxt": idCtxt,
			},
		},
		"sort": []types.JsonMap{
			{"id_natu": types.JsonMap{"order": "asc"}},
		},
	}

	queryJSON, _ := json.Marshal(query)
	res, err := idx.osCli.Search(context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryJSON),
		},
	)
	if err != nil {
		logger.Log.Errorf("Erro ao executar busca: %v", err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result searchResponseEventos
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	docs := make([]ResponseEventosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		docs = append(docs, ResponseEventosRow{
			Id:         hit.ID,
			IdCtxt:     doc.IdCtxt,
			IdNatu:     doc.IdNatu,
			IdPje:      doc.IdPje,
			Doc:        doc.Doc,
			DocJsonRaw: doc.DocJsonRaw,
			//DocEmbedding: doc.DocEmbedding,
		})
	}
	return docs, nil
}

// Consultar documentos por id_natu
func (idx *EventosIndex) ConsultaByIdNatu(idNatu int) ([]ResponseEventosRow, error) {
	if idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"id_natu": idNatu,
			},
		},
	}
	queryJSON, _ := json.Marshal(query)

	res, err := idx.osCli.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryJSON),
		},
	)
	if err != nil {
		logger.Log.Errorf("Erro ao executar busca: %v", err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result searchResponseEventos
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	docs := make([]ResponseEventosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		docs = append(docs, ResponseEventosRow{
			Id:     hit.ID,
			IdCtxt: doc.IdCtxt,
			IdNatu: doc.IdNatu,
			//DocEmbedding: doc.DocEmbedding,
		})
	}
	return docs, nil
}

// Busca semântica por embedding
func (idx *EventosIndex) ConsultaSemantica(vector []float32, idNatuFilter int) ([]ResponseEventosRow, error) {
	if idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	const ExpectedVectorSize = 3072
	if len(vector) != ExpectedVectorSize {
		msg := fmt.Sprintf("Vetor inválido: %d dimensões, esperado %d", len(vector), ExpectedVectorSize)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg)
	}

	boolQuery := types.JsonMap{
		"bool": types.JsonMap{
			"must": []interface{}{
				types.JsonMap{
					"knn": types.JsonMap{
						"doc_embedding": types.JsonMap{
							"vector": vector,
							"k":      10,
						},
					},
				},
			},
		},
	}

	if idNatuFilter > 0 {
		boolQuery["bool"].(types.JsonMap)["filter"] = []interface{}{
			types.JsonMap{"term": types.JsonMap{"id_natu": idNatuFilter}},
		}
	}

	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"_source": types.JsonMap{
			"excludes": []string{"doc_embedding"},
		},
		"query": boolQuery,
	}

	queryJSON, _ := json.Marshal(query)
	res, err := idx.osCli.Search(context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryJSON),
		},
	)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar OpenSearch: %v", err)
		return nil, erros.CreateError(err.Error())
	}
	defer res.Inspect().Response.Body.Close()

	var result searchResponseEventos
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, erros.CreateError(err.Error())
	}

	var docs []ResponseEventosRow
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		if idNatuFilter > 0 && doc.IdNatu != idNatuFilter {
			continue
		}
		if len(docs) >= 5 {
			break
		}
		docs = append(docs, ResponseEventosRow{
			Id:     hit.ID,
			IdCtxt: doc.IdCtxt,
			IdNatu: doc.IdNatu,
			//DocEmbedding: doc.DocEmbedding,
		})
	}
	return docs, nil
}

// Verificar existência de documento por id_ctxt + id_evento
func (idx *EventosIndex) IsExiste(idCtxt int, idEvento string) (bool, error) {
	if idCtxt <= 0 || idEvento == "" {
		return false, fmt.Errorf("parâmetros inválidos: idCtxt=%d, idEvento=%q", idCtxt, idEvento)
	}
	if idx.osCli == nil {
		return false, fmt.Errorf("OpenSearch não conectado")
	}

	query := types.JsonMap{
		"size": 1,
		"query": types.JsonMap{
			"bool": types.JsonMap{
				"must": []interface{}{
					types.JsonMap{"term": types.JsonMap{"id_ctxt": idCtxt}},
					types.JsonMap{"term": types.JsonMap{"id_evento": idEvento}},
				},
			},
		},
	}

	queryBody, _ := json.Marshal(query)
	res, err := idx.osCli.Search(context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryBody),
		},
	)
	if err != nil {
		logger.Log.Errorf("Erro ao consultar OpenSearch: %v", err)
		return false, erros.CreateError(err.Error())
	}
	defer res.Inspect().Response.Body.Close()

	if res.Errors {
		return false, erros.CreateError("Resposta inválida do OpenSearch")
	}

	if res.Hits.Total.Value > 0 {
		logger.Log.Infof("Documento com id_evento=%v já existe", idEvento)
		return true, nil
	}

	return false, nil
}
