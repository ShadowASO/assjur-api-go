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
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type IndexAutosType struct {
	osCli     *opensearchapi.Client
	indexName string
}

// Novo cliente para o índice autos_embedding
func NewIndexAutos() *IndexAutosType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}

	return &IndexAutosType{
		osCli:     osClient,
		indexName: "autos_embedding",
	}
}

// Documento do índice autos_embedding (sem campo doc)
type IndexAutosDoc struct {
	IdCtxt int `json:"id_ctxt"`
	IdNatu int `json:"id_natu"`
}

// type IndexAutos struct {
// 	IdCtxt       int       `json:"id_ctxt"`
// 	IdNatu       int       `json:"id_natu"`
// 	IdPje        string    `json:"id_pje"`
// 	DocEmbedding []float32 `json:"doc_embedding"`
// }

type BodyAutosUpdate struct {
	Doc consts.AutosRow `json:"doc"`
}

// type ResponseAutosEmbedding struct {
// 	Id     string `json:"id"`
// 	IdCtxt int    `json:"id_ctxt"`
// 	IdNatu int    `json:"id_natu"`
// 	IdPje  string `json:"id_pje"`
// }

type searchResponseAutos struct {
	Hits struct {
		Hits []struct {
			ID     string                  `json:"_id"`
			Source consts.ResponseAutosRow `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// Indexar um novo documento no índice autos_embedding
func (idx *IndexAutosType) IndexaDocumento(paramsData consts.AutosRow) (*opensearchapi.IndexResp, error) {
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

// Atualizar documento no índice autos_embedding
func (idx *IndexAutosType) UpdateDocumento(id string, paramsData consts.AutosRow) (*opensearchapi.UpdateResp, error) {
	updateData := BodyAutosUpdate{Doc: paramsData}

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

// Deletar documento pelo ID no índice autos_embedding
func (idx *IndexAutosType) DeleteDocumento(id string) (*opensearchapi.DocumentDeleteResp, error) {

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

// Consultar documento pelo ID no índice autos_embedding
func (idx *IndexAutosType) ConsultaDocumentoById(id string) (*consts.ResponseAutosRow, error) {

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

	doc := &consts.ResponseAutosRow{Id: id}
	source := result["_source"].(map[string]interface{})

	if v, ok := source["id_ctxt"].(float64); ok {
		doc.IdCtxt = int(v)
	}
	if v, ok := source["id_natu"].(float64); ok {
		doc.IdNatu = int(v)
	}

	return doc, nil
}

func (idx *IndexAutosType) ConsultaDocumentoByIdCtxt(idCtxt int) ([]consts.ResponseAutosRow, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	// Ajustei size para 10 para trazer múltiplos documentos
	query := map[string]interface{}{
		"size": 10,
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
	//defer res.Body.Close()
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound || res.Inspect().Response.StatusCode == http.StatusNoContent {
		msg := fmt.Sprintf("Documento com id_ctxt %d não encontrado no índice %s", idCtxt, idx.indexName)
		logger.Log.Info(msg)
		return nil, nil
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source consts.AutosRow `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	body := res.Inspect().Response.Body

	//if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	docs := make([]consts.ResponseAutosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		//doc.Id = hit.ID
		//docs = append(docs, doc)
		docAdd := consts.ResponseAutosRow{
			Id: hit.ID,
			//IdDoc:        doc.IdDoc,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			DocEmbedding: doc.DocEmbedding,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

func (idx *IndexAutosType) ConsultaDocumentosByIdNatu(idNatu int) ([]consts.ResponseAutosRow, error) {
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
				ID     string          `json:"_id"`
				Source consts.AutosRow `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	docs := make([]consts.ResponseAutosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		//doc.Id = hit.ID
		//docs = append(docs, doc)
		docAdd := consts.ResponseAutosRow{
			Id: hit.ID,
			//IdDoc:        doc.IdDoc,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			DocEmbedding: doc.DocEmbedding,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

/*
*
Faz uma busca semântica, utilizando embedding passado em vector,
limitando a resposta a 5 registros no máximo
*/
func (idx *IndexAutosType) ConsultaSemantica(vector []float32, idNatuFilter int) ([]consts.ResponseAutosRow, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	if len(vector) != ExpectedVectorSize {
		msg := fmt.Sprintf("Erro: o vetor enviado tem dimensão %d, mas o índice espera %d dimensões.", len(vector), ExpectedVectorSize)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg)
	}

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

	var result searchResponseAutosIndex
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}

	var docs []consts.ResponseAutosRow
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		//doc.Id = hit.ID

		if idNatuFilter > 0 && doc.IdNatu != idNatuFilter {
			continue
		}

		// documentos = append(documentos, doc)
		if len(docs) >= 5 {
			break
		}
		docAdd := consts.ResponseAutosRow{
			Id: hit.ID,
			//IdDoc:        doc.IdDoc,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			DocEmbedding: doc.DocEmbedding,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

func (idx *IndexAutosType) IsDocumentoEmbedding(idCtxt int, idPje string) (bool, error) {
	// Validação dos parâmetros
	if idCtxt <= 0 || idPje == "" {
		return false, fmt.Errorf("parâmetros inválidos: idCtxt=%d, idPje=%q", idCtxt, idPje)
	}
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return false, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	// Monta query para buscar documento com id_ctxt e id_pje correspondentes
	query := map[string]interface{}{
		"size": 1,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"id_ctxt": idCtxt,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
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

	//logger.Log.Info(string(queryBody))

	// Executa a busca no índice autos_embedding
	res, err := idx.osCli.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{"autos_embedding"},
			Body:    bytes.NewReader(queryBody),
		},
	)

	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar o OpenSearch: %v", err)
		logger.Log.Error(msg)
		return false, erros.CreateError(msg, err.Error())
	}
	defer res.Inspect().Response.Body.Close()

	// Verifica status HTTP da resposta
	if res.Errors {
		msg := fmt.Sprintf("Resposta inválida do OpenSearch: %s", res.Inspect().Response.Status())
		logger.Log.Error(msg)
		return false, erros.CreateError(msg)
	}

	//logger.Log.Infof("valor de MaxScore=%v", res.Hits.MaxScore)
	//logger.Log.Infof("valor de Total=%v", res.Hits.Total.Value)
	if res.Hits.Total.Value > 0 {
		msg := fmt.Sprintf("Documento ID_PJE=%v já inserido", idPje)
		logger.Log.Error(msg)
		return true, nil
	}

	return false, nil
}
