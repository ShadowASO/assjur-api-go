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

	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type AutosTempIndexType struct {
	osCli     *opensearchapi.Client
	indexName string
}

// Novo cliente para o índice autos
func NewAutos_tempIndex() *AutosTempIndexType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}

	return &AutosTempIndexType{
		osCli:     osClient,
		indexName: "autos_temp",
	}
}

// Documento do índice autos
type BodyAutosTempIndex struct {
	IdCtxt int    `json:"id_ctxt"`
	IdNatu int    `json:"id_natu"`
	IdPje  string `json:"id_pje"`
	Doc    string `json:"doc"` // texto analisado com analyzer brazilian
}

// Estrutura para update parcial (usa o mesmo IndexAutosDoc para atualizar qualquer campo)

type ResponseAutosTempIndex struct {
	Id     string `json:"id"`
	IdCtxt int    `json:"id_ctxt"`
	IdNatu int    `json:"id_natu"`
	IdPje  string `json:"id_pje"`
	Doc    string `json:"doc"`
}

type searchResponseAutosTempIndex struct {
	Hits struct {
		Hits []struct {
			ID     string                 `json:"_id"`
			Source ResponseAutosTempIndex `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (idx *AutosTempIndexType) Indexa(
	IdCtxt int,
	IdNatu int,
	IdPje string,
	Doc string,
	idOptional string,
) (*consts.ResponseAutosTempRow, error) {

	// Monta o documento para indexar
	doc := BodyAutosTempIndex{
		IdCtxt: IdCtxt,
		IdNatu: IdNatu,
		IdPje:  IdPje,
		Doc:    Doc,
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
	row := &consts.ResponseAutosTempRow{
		Id:     res.ID, // Você não tem esse campo ainda, pode deixar zero ou tratar fora
		IdCtxt: IdCtxt,
		IdNatu: IdNatu,
		IdPje:  IdPje,
	}

	return row, nil
}

// Atualizar documento parcial no índice autos pelo ID
func (idx *AutosTempIndexType) Update(
	id string, // ID do documento a atualizar
	IdCtxt int,
	IdNatu int,
	IdPje string,
	Doc string,
	// DocJSON map[string]interface{},
	// DocEmbedding []float32,
) (*consts.ResponseAutosTempRow, error) {

	// Monta o documento com os campos que deseja atualizar
	doc := BodyAutosTempIndex{
		IdCtxt: IdCtxt,
		IdNatu: IdNatu,
		IdPje:  IdPje,
		Doc:    Doc,
	}

	// Monta o corpo do update com o campo "doc"
	// updateBody := map[string]interface{}{
	// 	"doc": doc,
	// }

	//bodyReader := opensearchutil.NewJSONReader(updateBody)

	res, err := idx.osCli.Update(context.Background(),
		opensearchapi.UpdateReq{
			Index:      idx.indexName,
			DocumentID: id,
			//Body:       bodyReader,
			Body: opensearchutil.NewJSONReader(&doc),
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

	// Converte DocJSON para json.RawMessage para AutosJson
	// var rawJson json.RawMessage
	// if DocJSON != nil {
	// 	bDocJson, err := json.Marshal(DocJSON)
	// 	if err == nil {
	// 		rawJson = json.RawMessage(bDocJson)
	// 	} else {
	// 		logger.Log.Warning(fmt.Sprintf("Erro ao serializar DocJSON para AutosJson: %v", err))
	// 	}
	// }

	// Monta o objeto AutosRow para retorno
	row := &consts.ResponseAutosTempRow{
		Id:     id,
		IdCtxt: IdCtxt,
		IdNatu: IdNatu,
		IdPje:  IdPje,
	}

	return row, nil
}

// Deletar documento pelo ID no índice autos
// func (idx *Autos_tempIndexType) Delete(id string) error {
// 	res, err := idx.osCli.Document.Delete(
// 		context.Background(),
// 		opensearchapi.DocumentDeleteReq{
// 			Index:      idx.indexName,
// 			DocumentID: id,
// 		})

// 	if err != nil {
// 		msg := fmt.Sprintf("Erro ao deletar documento no OpenSearch: %v", err)
// 		logger.Log.Error(msg)
// 		return err
// 	}
// 	defer res.Inspect().Response.Body.Close()

// 	if res.Inspect().Response.StatusCode >= 400 {
// 		body, _ := io.ReadAll(res.Inspect().Response.Body)
// 		log.Printf("Erro na resposta do OpenSearch: %s", body)
// 		return fmt.Errorf("erro ao deletar documento: %s", res.Inspect().Response.Status())
// 	}

// 	return nil
// }

// Deletar documento pelo ID no índice autos e fazer refresh manual
func (idx *AutosTempIndexType) Delete(id string) error {
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
	//defer refreshRes.Body.Close()
	defer refreshRes.Inspect().Response.Body.Close()

	if refreshRes.Inspect().Response.StatusCode >= 400 {
		body, _ := io.ReadAll(refreshRes.Inspect().Response.Body)
		log.Printf("Erro na resposta do refresh: %s", body)
		return fmt.Errorf("erro ao fazer refresh do índice: %s", refreshRes.Inspect().Response.Status())
	}

	return nil
}

// Consultar documento pelo ID no índice autos
func (idx *AutosTempIndexType) ConsultaById(id string) (*consts.ResponseAutosTempRow, error) {

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

	//var result map[string]interface{}
	if err := json.NewDecoder(body).Decode(&docResp); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	// Agora convertemos as strings de data para time.Time
	// var dtPje, dtInc time.Time
	// if docResp.Source.DtPje.String() != "" {
	// 	if t, err := time.Parse(time.RFC3339, docResp.Source.DtPje.GoString()); err == nil {
	// 		dtPje = t
	// 	} else {
	// 		logger.Log.Warningf("Falha ao parsear dt_pje: %v", err)
	// 	}
	// }
	// var dtInc time.Time
	// if docResp.Source.DtInc.GoString() != "" {
	// 	if t, err := time.Parse(time.RFC3339, docResp.Source.DtInc.GoString()); err == nil {
	// 		dtInc = t
	// 	} else {
	// 		logger.Log.Warningf("Falha ao parsear dt_inc: %v", err)
	// 	}
	// }

	return &consts.ResponseAutosTempRow{
		Id:     id,
		IdCtxt: docResp.Source.IdCtxt,
		IdNatu: docResp.Source.IdNatu,
		IdPje:  docResp.Source.IdPje,
		Doc:    docResp.Source.Doc,
		//DtInc:  dtInc,
	}, nil
}

func (idx *AutosTempIndexType) ConsultaByIdCtxt(idCtxt int) ([]consts.ResponseAutosTempRow, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	query := map[string]interface{}{
		"size": 20,
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"id_ctxt": idCtxt,
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		msg := fmt.Sprintf("Erro ao serializar query   JSON: %v", err)
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

	docs := make([]consts.ResponseAutosTempRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		var row consts.ResponseAutosTempRow
		if err := json.Unmarshal(hit.Source, &row); err != nil {
			logger.Log.Warning(fmt.Sprintf("Erro ao deserializar documento %s: %v", hit.ID, err))
			continue
		}
		row.Id = hit.ID

		// Caso DtPje esteja em formato string, converta para time.Time aqui (se necessário)
		// Se DtPje já vem como time.Time do Unmarshal, não precisa converter

		docs = append(docs, row)
	}

	return docs, nil
}

// Consultar documentos pelo campo id_natu
func (idx *AutosTempIndexType) ConsultaByIdNatu(idNatu int) ([]consts.ResponseAutosTempRow, error) {
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	query := map[string]interface{}{
		"size": 10,
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

	docs := make([]consts.ResponseAutosTempRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		var row consts.ResponseAutosTempRow
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
func (idx *AutosTempIndexType) ConsultaSemantica(vector []float32, idNatuFilter int) ([]consts.ResponseAutosTempRow, error) {
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

	var docs []consts.ResponseAutosTempRow
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		//doc.Id = hit.ID

		if idNatuFilter > 0 && doc.IdNatu != idNatuFilter {
			continue
		}

		//documentos = append(documentos, doc)
		if len(docs) >= 5 {
			break
		}
		docAdd := consts.ResponseAutosTempRow{
			Id: hit.ID,
			//IdDoc:        doc.IdDoc,
			IdCtxt: doc.IdCtxt,
			IdNatu: doc.IdNatu,
			IdPje:  doc.IdPje,
			Doc:    doc.Doc,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Verificar se documento com id_ctxt e id_pje já existe
func (idx *AutosTempIndexType) IsExiste(idCtxt int, idPje string) (bool, error) {
	if idCtxt <= 0 || idPje == "" {
		return false, fmt.Errorf("parâmetros inválidos: idCtxt=%d, idPje=%q", idCtxt, idPje)
	}
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return false, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

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
