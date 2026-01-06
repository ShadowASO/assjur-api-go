package opensearch

import (
	"bytes"

	"encoding/json"
	"fmt"

	"net/http"
	"strings"
	"time"

	"ocrserver/internal/consts"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

type AutosTempIndexType struct {
	osCli     *opensearchapi.Client
	indexName string
	timeout   time.Duration
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
		timeout:   10 * time.Second,
	}
}

// Documento do índice autos
type BodyAutosTempIndex struct {
	IdCtxt string    `json:"id_ctxt"`
	IdNatu int       `json:"id_natu"`
	IdPje  string    `json:"id_pje"`
	DtInc  time.Time `json:"dt_inc"` // data/hora da inclusão
	Doc    string    `json:"doc"`    // texto analisado com analyzer brazilian
}

// Estrutura para update parcial (usa o mesmo IndexAutosDoc para atualizar qualquer campo)

type ResponseAutosTempIndex struct {
	Id     string    `json:"id"`
	IdCtxt string    `json:"id_ctxt"`
	IdNatu int       `json:"id_natu"`
	IdPje  string    `json:"id_pje"`
	DtInc  time.Time `json:"dt_inc"` // data/hora da inclusão
	Doc    string    `json:"doc"`
}

func (idx *AutosTempIndexType) Indexa(
	IdCtxt string,
	IdNatu int,
	IdPje string,
	Doc string,
	idOptional string,
) (*consts.ResponseAutosTempRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	dt_inc := time.Now()
	// Monta o documento para indexar
	doc := BodyAutosTempIndex{
		IdCtxt: IdCtxt,
		IdNatu: IdNatu,
		IdPje:  IdPje,
		DtInc:  dt_inc,
		Doc:    Doc,
	}

	res, err := idx.osCli.Index(
		ctx,
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
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	// Monta o objeto AutosRow para retorno
	row := &consts.ResponseAutosTempRow{
		Id:     res.ID, // Você não tem esse campo ainda, pode deixar zero ou tratar fora
		IdCtxt: IdCtxt,
		IdNatu: IdNatu,
		IdPje:  IdPje,
		DtInc:  dt_inc,
	}

	return row, nil
}

// Atualizar documento parcial no índice autos pelo ID
func (idx *AutosTempIndexType) Update(
	id string, // ID do documento a atualizar
	idCtxt string,
	IdNatu int,
	IdPje string,
	Doc string,

) (*consts.ResponseAutosTempRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	if strings.TrimSpace(idCtxt) == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	dt_inc := time.Now()
	// Monta o documento com os campos que deseja atualizar
	doc := BodyAutosTempIndex{
		IdCtxt: idCtxt,
		IdNatu: IdNatu,
		IdPje:  IdPje,
		DtInc:  dt_inc,
		Doc:    Doc,
	}

	res, err := idx.osCli.Update(
		ctx,
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
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	// Monta o objeto AutosRow para retorno
	row := &consts.ResponseAutosTempRow{
		Id:     id,
		IdCtxt: idCtxt,
		IdNatu: IdNatu,
		IdPje:  IdPje,
		DtInc:  dt_inc,
	}

	return row, nil
}

// Deletar documento pelo ID no índice autos e fazer refresh manual
func (idx *AutosTempIndexType) Delete(id string) error {
	if idx == nil || idx.osCli == nil {
		err := fmt.Errorf("OpenSearch não conectado")
		logger.Log.Error(err.Error())
		return err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		err := fmt.Errorf("id vazio")
		logger.Log.Error(err.Error())
		return err
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Document.Delete(
		ctx,
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
func (idx *AutosTempIndexType) ConsultaById(id string) (*consts.ResponseAutosTempRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}
	logger.Log.Infof("id=%s", id)
	logger.Log.Infof("idx.indexName=%s", idx.indexName)

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Document.Get(
		ctx,
		opensearchapi.DocumentGetReq{
			Index:      idx.indexName,
			DocumentID: id,
		})

	//------------------------------------------------------

	// if err != nil {
	// 	msg := fmt.Sprintf("Erro ao consultar documento %s no índice %s: %v", id, idx.indexName, err)
	// 	logger.Log.Error(msg)
	// 	return nil, err
	// }
	// defer res.Inspect().Response.Body.Close()

	// if res.Inspect().Response.StatusCode == http.StatusNotFound {
	// 	logger.Log.Infof("id=%s não encontrado!", id)
	// 	return nil, nil
	// }
	// if err := ReadOSErr(res.Inspect().Response); err != nil {
	// 	return nil, err
	// }
	err = ReadOSErr(res.Inspect().Response)
	if err != nil {
		msg := fmt.Sprintf("Erro ao deletar documento: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result DocumentGetResponse[consts.AutosTempRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	if !result.Found {
		logger.Log.Infof("id=%s não encontrado (found=false)", id)
		return nil, nil
	}

	src := result.Source

	return &consts.ResponseAutosTempRow{
		Id:     id,
		IdCtxt: src.IdCtxt,
		IdNatu: src.IdNatu,
		IdPje:  src.IdPje,
		DtInc:  src.DtInc,
		Doc:    src.Doc,
	}, nil
}

func (idx *AutosTempIndexType) ConsultaByIdCtxt(idCtxt string) ([]consts.ResponseAutosTempRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	idCtxt = strings.TrimSpace(idCtxt)
	if idCtxt == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	query := map[string]interface{}{
		"size": 50,
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
		ctx,
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
	if res.Inspect().Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	statusCode := res.Inspect().Response.StatusCode
	if statusCode == http.StatusNotFound || statusCode == http.StatusNoContent {
		msg := fmt.Sprintf("Documento com id_ctxt %d não encontrado no índice %s", idCtxt, idx.indexName)
		logger.Log.Info(msg)
		return nil, nil
	}

	var result SearchResponseGeneric[consts.AutosTempRow]

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]consts.ResponseAutosTempRow, 0, len(result.Hits.Hits))

	for _, hit := range result.Hits.Hits {
		doc := hit.Source

		docAdd := consts.ResponseAutosTempRow{
			Id:     hit.ID,
			IdCtxt: doc.IdCtxt,
			IdNatu: doc.IdNatu,
			IdPje:  doc.IdPje,
			Doc:    doc.Doc,
		}

		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Consultar documentos pelo campo id_natu
func (idx *AutosTempIndexType) ConsultaByIdNatu(idNatu int) ([]consts.ResponseAutosTempRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	if idNatu == 0 {
		return nil, fmt.Errorf("idNatu zerado")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

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
		ctx,
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
	if res.Inspect().Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	statusCode := res.Inspect().Response.StatusCode
	if statusCode == http.StatusNotFound || statusCode == http.StatusNoContent {
		msg := fmt.Sprintf("Documento com id_natu %d não encontrado no índice %s", idNatu, idx.indexName)
		logger.Log.Info(msg)
		return nil, nil
	}

	var result SearchResponseGeneric[consts.AutosTempRow]

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	// ✅ Correção do panic
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]consts.ResponseAutosTempRow, 0, len(result.Hits.Hits))

	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		docAdd := consts.ResponseAutosTempRow{
			Id:     hit.ID,
			IdCtxt: doc.IdCtxt,
			IdNatu: doc.IdNatu,
			IdPje:  doc.IdPje,
			DtInc:  doc.DtInc,
			Doc:    doc.Doc,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Busca semântica pelo embedding no campo doc_embedding, filtrando por id_natu opcionalmente
func (idx *AutosTempIndexType) ConsultaSemantica(vector []float32, idNatuFilter int) ([]consts.ResponseAutosTempRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

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
		ctx,
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

	//var result searchResponseAutosTempIndex
	var result SearchResponseGeneric[consts.AutosTempRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}

	var docs []consts.ResponseAutosTempRow
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		if idNatuFilter > 0 && doc.IdNatu != idNatuFilter {
			continue
		}

		if len(docs) >= 5 {
			break
		}
		docAdd := consts.ResponseAutosTempRow{
			Id:     hit.ID,
			IdCtxt: doc.IdCtxt,
			IdNatu: doc.IdNatu,
			IdPje:  doc.IdPje,
			DtInc:  doc.DtInc,
			Doc:    doc.Doc,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Verificar se documento com id_ctxt e id_pje já existe
func (idx *AutosTempIndexType) IsExiste(idCtxt string, idPje string) (bool, error) {
	if idCtxt == "" || idPje == "" {
		return false, fmt.Errorf("parâmetros inválidos: idCtxt=%d, idPje=%q", idCtxt, idPje)
	}
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return false, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

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
		ctx,
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
