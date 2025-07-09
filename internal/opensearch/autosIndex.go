package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"ocrserver/internal/consts"
	"ocrserver/internal/types"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type AutosIndexType struct {
	osCli     *opensearchapi.Client
	indexName string
}

// Novo cliente para o índice autos
func NewAutosIndex() *AutosIndexType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}

	return &AutosIndexType{
		osCli:     osClient,
		indexName: "autos",
	}
}

// Documento do índice autos
type BodyAutosIndex struct {
	IdCtxt       int                    `json:"id_ctxt"`
	IdNatu       int                    `json:"id_natu"`
	IdPje        string                 `json:"id_pje"`
	Doc          string                 `json:"doc"`      // texto analisado com analyzer brazilian
	DocJSON      map[string]interface{} `json:"doc_json"` // campo objeto JSON livre
	DocEmbedding []float32              `json:"doc_embedding"`
}

type ResponseAutosIndex struct {
	Id           string                 `json:"id"`
	IdCtxt       int                    `json:"id_ctxt"`
	IdNatu       int                    `json:"id_natu"`
	IdPje        string                 `json:"id_pje"`
	Doc          string                 `json:"doc"`
	DocJSON      map[string]interface{} `json:"doc_json"`
	DocEmbedding []float32              `json:"doc_embedding,omitempty"`
}

type searchResponseAutosIndex struct {
	Hits struct {
		Hits []struct {
			ID     string             `json:"_id"`
			Source ResponseAutosIndex `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (idx *AutosIndexType) Indexa(
	IdCtxt int,
	IdNatu int,
	IdPje string,
	Doc string,
	DocJSON map[string]interface{},
	DocEmbedding []float32,
	idOptional string,
) (*consts.AutosRow, error) {

	// Monta o documento para indexar
	doc := BodyAutosIndex{
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJSON:      DocJSON,
		DocEmbedding: DocEmbedding,
	}

	res, err := idx.osCli.Index(context.Background(),
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: idOptional, // pode ser "" para id automático
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
	row := &consts.AutosRow{
		Id:           res.ID, // Você não tem esse campo ainda, pode deixar zero ou tratar fora
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJson:      DocJSON,
		DocEmbedding: DocEmbedding,
	}

	return row, nil
}

// Atualizar documento parcial no índice autos pelo ID
func (idx *AutosIndexType) Update(
	id string, // ID do documento a atualizar
	IdCtxt int,
	IdNatu int,
	IdPje string,
	Doc string,
	DocJSON map[string]interface{},
	DocEmbedding []float32,
) (*consts.AutosRow, error) {

	// Monta o documento com os campos que deseja atualizar
	doc := BodyAutosIndex{
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJSON:      DocJSON,
		DocEmbedding: DocEmbedding,
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

	// Monta o objeto AutosRow para retorno
	row := &consts.AutosRow{
		Id:           id,
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJson:      DocJSON,
		DocEmbedding: DocEmbedding,
	}

	return row, nil
}

// Deletar documento pelo ID no índice autos
func (idx *AutosIndexType) Delete(id string) error {

	res, err := idx.osCli.Document.Delete(
		context.Background(),
		opensearchapi.DocumentDeleteReq{
			Index:      idx.indexName,
			DocumentID: id,
		})

	if err != nil {
		msg := fmt.Sprintf("Erro ao deletar documento no OpenSearch: %v", err)
		logger.Log.Error(msg)
		return err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Inspect().Response.Body)
		log.Printf("Erro na resposta do OpenSearch: %s", body)
		return fmt.Errorf("erro ao deletar documento: %s", res.Inspect().Response.Status())
	}

	// Refresh manual do índice para garantir que a deleção esteja visível nas buscas
	refreshRes, err := idx.osCli.Indices.Refresh(
		context.Background(),
		&opensearchapi.IndicesRefreshReq{
			Indices: []string{idx.indexName},
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao fazer refresh do índice %s: %v", idx.indexName, err)
		logger.Log.Error(msg)
		return err
	}

	defer refreshRes.Inspect().Response.Body.Close()

	if refreshRes.Inspect().Response.StatusCode >= 400 {
		body, _ := io.ReadAll(refreshRes.Inspect().Response.Body)
		log.Printf("Erro na resposta do refresh: %s", body)
		return fmt.Errorf("erro ao fazer refresh do índice: %s", refreshRes.Inspect().Response.Status())
	}

	return nil
}

// Consultar documento pelo ID no índice autos
func (idx *AutosIndexType) ConsultaById(id string) (*consts.AutosRow, error) {

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

	return &consts.AutosRow{
		Id:           id,
		IdCtxt:       docResp.Source.IdCtxt,
		IdNatu:       docResp.Source.IdNatu,
		IdPje:        docResp.Source.IdPje,
		Doc:          docResp.Source.Doc,
		DocJson:      docResp.Source.DocJson,
		DocEmbedding: docResp.Source.DocEmbedding,
	}, nil
}

func (idx *AutosIndexType) ConsultaByIdCtxt(idCtxt int) ([]consts.AutosRow, error) {
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

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source json.RawMessage `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	docs := make([]consts.AutosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		var row consts.AutosRow
		if err := json.Unmarshal(hit.Source, &row); err != nil {
			logger.Log.Warning(fmt.Sprintf("Erro ao deserializar documento %s: %v", hit.ID, err))
			continue
		}
		row.Id = hit.ID

		docs = append(docs, row)
	}

	return docs, nil
}

// Consultar documentos pelo campo id_natu
func (idx *AutosIndexType) ConsultaByIdNatu(idNatu int) ([]consts.AutosRow, error) {
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

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source json.RawMessage `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	docs := make([]consts.AutosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		var row consts.AutosRow
		if err := json.Unmarshal(hit.Source, &row); err != nil {
			logger.Log.Warning(fmt.Sprintf("Erro ao deserializar documento %s: %v", hit.ID, err))
			continue
		}
		row.Id = hit.ID
		docs = append(docs, row)
	}

	return docs, nil
}

// Busca semântica pelo embedding no campo doc_embedding, filtrando por id_natu opcionalmente
func (idx *AutosIndexType) ConsultaSemantica(vector []float32, idNatuFilter int) ([]ResponseAutosIndex, error) {
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

	var result searchResponseAutosIndex
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}

	var documentos []ResponseAutosIndex
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		doc.Id = hit.ID

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

// Verificar se documento com id_ctxt e id_pje já existe
func (idx *AutosIndexType) IsExiste(idCtxt int, idPje string) (bool, error) {
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
