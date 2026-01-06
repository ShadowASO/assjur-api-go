package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"net/http"
	"time"

	"ocrserver/internal/consts"
	"ocrserver/internal/types"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

type AutosJsonEmbeddingType struct {
	osCli     *opensearchapi.Client
	indexName string
	timeout   time.Duration
}

// Novo cliente para o índice autos
func NewAutosJsonEmbedding() *AutosJsonEmbeddingType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}

	return &AutosJsonEmbeddingType{
		osCli:     osClient,
		indexName: "autos_json_embedding",
		timeout:   10 * time.Second,
	}
}

type bodyRowIndex struct {
	IdDoc        string    `json:"id_doc"`
	IdCtxt       string    `json:"id_ctxt"`
	IdNatu       int       `json:"id_natu"`
	DocEmbedding []float32 `json:"doc_embedding"`
}

type searchResponseAutosJsonEmbedding struct {
	Hits struct {
		Hits []struct {
			ID     string                       `json:"_id"`
			Source consts.AutosJsonEmbeddingRow `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (idx *AutosJsonEmbeddingType) Indexa(
	idDoc string,
	idCtxt string,
	idNatu int,
	docEmbedding []float32,
) (*consts.ResponseAutosJsonEmbeddingRow, error) {

	// Monta o documento para indexar
	doc := bodyRowIndex{
		IdDoc:        idDoc,
		IdCtxt:       idCtxt,
		IdNatu:       idNatu,
		DocEmbedding: docEmbedding,
	}

	res, err := idx.osCli.Index(context.Background(),
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: "", // pode ser "" para id automático
			Body:       opensearchutil.NewJSONReader(&doc),
			Params: opensearchapi.IndexParams{
				Refresh: "true",
			},
		})
	if err != nil {
		msg := fmt.Sprintf("Erro ao indexar documento no OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	// Monta o objeto AutosRow para retorno
	row := &consts.ResponseAutosJsonEmbeddingRow{
		Id:           res.ID, // Você não tem esse campo ainda, pode deixar zero ou tratar fora
		IdDoc:        idDoc,
		IdCtxt:       idCtxt,
		IdNatu:       idNatu,
		DocEmbedding: docEmbedding,
	}

	return row, nil
}

// Atualizar documento parcial no índice autos pelo ID
func (idx *AutosJsonEmbeddingType) Update(
	id string, // ID do documento a atualizar
	idDoc string,
	idCtxt string,
	idNatu int,
	docEmbedding []float32,
) (*consts.ResponseAutosJsonEmbeddingRow, error) {

	// Monta o documento com os campos que deseja atualizar
	doc := bodyRowIndex{
		IdDoc:        idDoc,
		IdCtxt:       idCtxt,
		IdNatu:       idNatu,
		DocEmbedding: docEmbedding,
	}

	res, err := idx.osCli.Update(context.Background(),
		opensearchapi.UpdateReq{
			Index:      idx.indexName,
			DocumentID: id,
			Body:       opensearchutil.NewJSONReader(&doc),
			Params: opensearchapi.UpdateParams{
				Refresh: "true",
			},
		})
	if err != nil {
		msg := fmt.Sprintf("Erro ao atualizar documento no OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	row := &consts.ResponseAutosJsonEmbeddingRow{
		Id:           res.ID, // Você não tem esse campo ainda, pode deixar zero ou tratar fora
		IdDoc:        idDoc,
		IdCtxt:       idCtxt,
		IdNatu:       idNatu,
		DocEmbedding: docEmbedding,
	}

	return row, nil
}

// Deletar documento pelo ID no índice autos
func (idx *AutosJsonEmbeddingType) Delete(id string) error {

	res, err := idx.osCli.Document.Delete(
		context.Background(),
		opensearchapi.DocumentDeleteReq{
			Index:      idx.indexName,
			DocumentID: id,
			Params: opensearchapi.DocumentDeleteParams{
				// ✅ Melhor opção para “sumir da lista” logo após o delete:
				Refresh: "true", //"wait_for", ou "true"
			},
		})

	err = ReadOSErr(res.Inspect().Response)
	if err != nil {
		msg := fmt.Sprintf("Erro ao deletar documento: %v", err)
		logger.Log.Error(msg)
		return err
	}
	defer res.Inspect().Response.Body.Close()

	return nil
}

// Consultar documento pelo ID no índice autos
func (idx *AutosJsonEmbeddingType) ConsultaById(id string) (*consts.ResponseAutosJsonEmbeddingRow, error) {

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
		logger.Log.Warning(msg)
		return nil, nil
	}

	if res.Inspect().Response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Erro inesperado na consulta do documento %s: status %d", id, res.Inspect().Response.StatusCode)
		logger.Log.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	body := res.Inspect().Response.Body
	var docResp struct {
		Source consts.AutosRow `json:"_source"`
	}

	if err := json.NewDecoder(body).Decode(&docResp); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	return &consts.ResponseAutosJsonEmbeddingRow{
		Id:           id,
		IdDoc:        docResp.Source.Doc,
		IdCtxt:       docResp.Source.IdCtxt,
		IdNatu:       docResp.Source.IdNatu,
		DocEmbedding: docResp.Source.DocEmbedding,
	}, nil
}

func (idx *AutosJsonEmbeddingType) ConsultaByIdCtxt(idCtxt string) ([]consts.ResponseAutosJsonEmbeddingRow, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	query := types.JsonMap{
		"size": 10,
		"query": types.JsonMap{
			"term": types.JsonMap{
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

	statusCode := res.Inspect().Response.StatusCode
	if statusCode == http.StatusNotFound || statusCode == http.StatusNoContent {
		msg := fmt.Sprintf("Documento com id_ctxt %d não encontrado no índice %s", idCtxt, idx.indexName)
		logger.Log.Info(msg)
		return nil, nil
	}

	var result searchResponseAutosJsonEmbedding

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	docs := make([]consts.ResponseAutosJsonEmbeddingRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source

		docAdd := consts.ResponseAutosJsonEmbeddingRow{
			Id:           hit.ID,
			IdDoc:        doc.IdDoc,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			DocEmbedding: doc.DocEmbedding,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

func (idx *AutosJsonEmbeddingType) ConsultaByIdDoc(idDoc string) ([]consts.ResponseAutosJsonEmbeddingRow, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	query := types.JsonMap{
		"size": 10,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"id_doc": idDoc,
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

	statusCode := res.Inspect().Response.StatusCode
	if statusCode == http.StatusNotFound || statusCode == http.StatusNoContent {
		msg := fmt.Sprintf("Documento com id_doc %d não encontrado no índice %s", idDoc, idx.indexName)
		logger.Log.Info(msg)
		return nil, nil
	}

	var result searchResponseAutosJsonEmbedding

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	docs := make([]consts.ResponseAutosJsonEmbeddingRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source

		docAdd := consts.ResponseAutosJsonEmbeddingRow{
			Id:           hit.ID,
			IdDoc:        doc.IdDoc,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			DocEmbedding: doc.DocEmbedding,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Consultar documentos pelo campo id_natu
func (idx *AutosJsonEmbeddingType) ConsultaByIdNatu(idNatu int) ([]consts.ResponseAutosJsonEmbeddingRow, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	query := types.JsonMap{
		"size": 10,
		"query": types.JsonMap{
			"term": types.JsonMap{
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

	statusCode := res.Inspect().Response.StatusCode
	if statusCode == http.StatusNotFound || statusCode == http.StatusNoContent {
		msg := fmt.Sprintf("Documento com id_natu %d não encontrado no índice %s", idNatu, idx.indexName)
		logger.Log.Info(msg)
		return nil, nil
	}

	var result searchResponseAutosJsonEmbedding

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	docs := make([]consts.ResponseAutosJsonEmbeddingRow, 0, len(result.Hits.Hits))

	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		docAdd := consts.ResponseAutosJsonEmbeddingRow{
			Id:           hit.ID,
			IdDoc:        doc.IdDoc,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			DocEmbedding: doc.DocEmbedding,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Busca semântica pelo embedding no campo doc_embedding, filtrando por id_natu opcionalmente
func (idx *AutosJsonEmbeddingType) ConsultaSemantica(vector []float32, idNatuFilter int) ([]consts.ResponseAutosJsonEmbeddingRow, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	ExpectedVectorSize := 3072 // deve ser igual ao dimension do índice

	if len(vector) != ExpectedVectorSize {
		msg := fmt.Sprintf("Erro: vetor enviado tem dimensão %d, mas índice espera %d dimensões.", len(vector), ExpectedVectorSize)
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
			types.JsonMap{
				"term": types.JsonMap{
					"id_natu": idNatuFilter,
				},
			},
		}
	}

	query := types.JsonMap{
		"size": 10,
		"_source": types.JsonMap{
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

	var result searchResponseAutosJsonEmbedding
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}

	docs := make([]consts.ResponseAutosJsonEmbeddingRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source

		if idNatuFilter > 0 && doc.IdNatu != idNatuFilter {
			continue
		}
		if len(docs) >= 5 {
			break
		}
		docAdd := consts.ResponseAutosJsonEmbeddingRow{
			Id:           hit.ID,
			IdDoc:        doc.IdDoc,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			DocEmbedding: doc.DocEmbedding,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Verificar se documento com id_ctxt e id_pje já existe
func (idx *AutosJsonEmbeddingType) IsExiste(idCtxt int, idPje string) (bool, error) {
	if idCtxt <= 0 || idPje == "" {
		return false, fmt.Errorf("parâmetros inválidos: idCtxt=%d, idPje=%q", idCtxt, idPje)
	}
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return false, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	query := types.JsonMap{
		"size": 1,
		"query": types.JsonMap{
			"bool": types.JsonMap{
				"must": []interface{}{
					types.JsonMap{
						"term": types.JsonMap{
							"id_ctxt": idCtxt,
						},
					},
					types.JsonMap{
						"term": types.JsonMap{
							"id_pje": idPje,
						},
					},
				},
			},
		},
	}

	queryBody, err := json.Marshal(query)
	if err != nil {
		msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
		logger.Log.Error(msg)
		return false, err
	}

	res, err := idx.osCli.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryBody),
		},
	)

	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar o OpenSearch: %v", err)
		logger.Log.Error(msg)
		return false, erros.CreateError(msg, err.Error())
	}
	defer res.Inspect().Response.Body.Close()

	if res.Errors {
		msg := fmt.Sprintf("Resposta inválida do OpenSearch: %s", res.Inspect().Response.Status())
		logger.Log.Error(msg)
		return false, erros.CreateError(msg)
	}

	if res.Hits.Total.Value > 0 {
		msg := fmt.Sprintf("Documento com id_pje=%v já existe", idPje)
		logger.Log.Info(msg)
		return true, nil
	}

	return false, nil
}
