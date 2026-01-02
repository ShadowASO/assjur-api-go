package opensearch

import (
	"bytes"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"ocrserver/internal/consts"
	"ocrserver/internal/types"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

type AutosIndexType struct {
	osCli     *opensearchapi.Client
	indexName string
	timeout   time.Duration
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
		timeout:   10 * time.Second,
	}
}

func (idx *AutosIndexType) Indexa(
	IdCtxt string,
	IdNatu int,
	IdPje string,
	Doc string,
	DocJsonRaw string,
	DocEmbedding []float32,
	idOptional string,
) (*consts.ResponseAutosRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	doc := consts.AutosRow{
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJsonRaw:   DocJsonRaw,
		DocEmbedding: DocEmbedding,
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
	row := &consts.ResponseAutosRow{
		Id:           res.ID, // Você não tem esse campo ainda, pode deixar zero ou tratar fora
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJsonRaw:   DocJsonRaw,
		DocEmbedding: DocEmbedding,
	}

	return row, nil
}

// Atualizar documento parcial no índice autos pelo ID
func (idx *AutosIndexType) Update(
	id string, // ID do documento a atualizar
	idCtxt string,
	IdNatu int,
	IdPje string,
	Doc string,
	DocJson string,
	DocEmbedding []float32,
) (*consts.ResponseAutosRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	if strings.TrimSpace(idCtxt) == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	// Monta o documento com os campos que deseja atualizar

	doc := consts.AutosRow{
		IdCtxt:       idCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJsonRaw:   DocJson,
		DocEmbedding: DocEmbedding,
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
	row := &consts.ResponseAutosRow{
		Id:           id,
		IdCtxt:       idCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		Doc:          Doc,
		DocJsonRaw:   DocJson,
		DocEmbedding: DocEmbedding,
	}

	return row, nil
}

// Deletar documento pelo ID no índice autos
func (idx *AutosIndexType) Delete(id string) error {
	if idx == nil || idx.osCli == nil {
		err := fmt.Errorf("OpenSearch não conectado")
		logger.Log.Error(err.Error())
		return err
	}
	//ctx, cancel := idx.ctx()
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Document.Delete(
		ctx,
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
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return err
	}

	if res.Inspect().Response.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Inspect().Response.Body)
		log.Printf("Erro na resposta do OpenSearch: %s", body)
		return fmt.Errorf("erro ao deletar documento: %s", res.Inspect().Response.Status())
	}

	ctx2, cancel2 := NewCtx(idx.timeout)
	defer cancel2()
	// Refresh manual do índice para garantir que a deleção esteja visível nas buscas
	refreshRes, err := idx.osCli.Indices.Refresh(
		ctx2,
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
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return err
	}

	if refreshRes.Inspect().Response.StatusCode >= 400 {
		body, _ := io.ReadAll(refreshRes.Inspect().Response.Body)
		log.Printf("Erro na resposta do refresh: %s", body)
		return fmt.Errorf("erro ao fazer refresh do índice: %s", refreshRes.Inspect().Response.Status())
	}

	return nil
}

// Consultar documento pelo ID no índice autos
func (idx *AutosIndexType) ConsultaById(id string) (*consts.ResponseAutosRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	//ctx, cancel := idx.ctx()
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	//Cria a requisição
	req := opensearchapi.DocumentSourceReq{
		Index:      idx.indexName,
		DocumentID: id,
	}
	//Executa a requisição, passando a requisição
	res, err := idx.osCli.Document.Source(
		ctx,
		req,
	)
	//------------------------------------------------------

	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar documento %s no índice %s: %v", id, idx.indexName, err)
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

	var result SearchResponseGeneric[consts.AutosRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	// ✅ Correção do panic
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	hit := result.Hits.Hits[0]
	src := hit.Source

	doc := consts.ResponseAutosRow{
		Id:           hit.ID,
		IdCtxt:       src.IdCtxt,
		IdNatu:       src.IdNatu,
		IdPje:        src.IdPje,
		Doc:          src.Doc,
		DocJsonRaw:   src.DocJsonRaw,
		DocEmbedding: src.DocEmbedding,
	}

	return &doc, nil
}

func (idx *AutosIndexType) ConsultaByIdCtxt(idCtxt string) ([]consts.ResponseAutosRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	idCtxt = strings.TrimSpace(idCtxt)
	if idCtxt == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"id_ctxt": idCtxt,
			},
		},
		"sort": []types.JsonMap{
			{
				"id_natu": types.JsonMap{
					"order": "asc",
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	//Crio a SearchReq
	req := opensearchapi.SearchReq{
		Indices: []string{idx.indexName},
		Body:    bytes.NewReader(queryJSON),
	}

	//Executo a chamada da busca
	res, err := idx.osCli.Search(
		ctx,
		&req,
	)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar search: %s : %s = %v", idx.indexName, idCtxt, err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	var result SearchResponseGeneric[consts.AutosRow]

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]consts.ResponseAutosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source

		docAdd := consts.ResponseAutosRow{
			Id:           hit.ID,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			IdPje:        doc.IdPje,
			Doc:          doc.Doc,
			DocJsonRaw:   doc.DocJsonRaw,
			DocEmbedding: doc.DocEmbedding,
		}

		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Consultar documentos pelo campo id_natu
func (idx *AutosIndexType) ConsultaByIdNatu(idNatu int) ([]consts.ResponseAutosRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	if idNatu == 0 {
		return nil, fmt.Errorf("idNatu zerado")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
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

	//Crio a SearchReq
	req := opensearchapi.SearchReq{
		Indices: []string{idx.indexName},
		Body:    bytes.NewReader(queryJSON),
	}

	//Executo a chamada da busca
	res, err := idx.osCli.Search(
		ctx,
		&req,
	)
	if err != nil {
		logger.Log.Errorf("Erro ao executar search: %s : %d = %v", idx.indexName, idNatu, err)
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

	var result SearchResponseGeneric[consts.AutosRow]

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	// ✅ Correção do panic
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]consts.ResponseAutosRow, 0, len(result.Hits.Hits))

	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		docAdd := consts.ResponseAutosRow{
			Id:           hit.ID,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			DocEmbedding: doc.DocEmbedding,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Busca semântica pelo embedding no campo doc_embedding, filtrando por id_natu opcionalmente
func (idx *AutosIndexType) ConsultaSemantica(vector []float32, idNatuFilter int) ([]consts.ResponseAutosRow, error) {
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
		"size": QUERY_MAX_SIZE,
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

	//Crio a SearchReq
	req := opensearchapi.SearchReq{
		Indices: []string{idx.indexName},
		Body:    bytes.NewReader(queryJSON),
	}

	//Executo a chamada da busca
	res, err := idx.osCli.Search(
		ctx,
		&req,
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar o OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}
	defer res.Inspect().Response.Body.Close()

	//var result searchResponseAutos
	var result SearchResponseGeneric[consts.AutosRow]

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg, err.Error())
	}

	var docs []consts.ResponseAutosRow
	for _, hit := range result.Hits.Hits {
		doc := hit.Source

		if idNatuFilter > 0 && doc.IdNatu != idNatuFilter {
			continue
		}

		if len(docs) >= 5 {
			break
		}
		docAdd := consts.ResponseAutosRow{
			Id:           hit.ID,
			IdCtxt:       doc.IdCtxt,
			IdNatu:       doc.IdNatu,
			DocEmbedding: doc.DocEmbedding,
		}
		docs = append(docs, docAdd)
	}

	return docs, nil
}

// Verificar se documento com id_ctxt e id_pje já existe
func (idx *AutosIndexType) IsExiste(idCtxt string, idPje string) (bool, error) {
	if idCtxt == "" || idPje == "" {
		return false, fmt.Errorf("parâmetros inválidos: idCtxt=%d, idPje=%q", idCtxt, idPje)
	}
	if idx.osCli == nil {
		logger.Log.Error("Erro: OpenSearch não conectado.")
		return false, fmt.Errorf("erro ao conectar ao OpenSearch")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

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

	queryJSON, err := json.Marshal(query)
	if err != nil {
		msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
		logger.Log.Error(msg)
		return false, err
	}

	//Crio a SearchReq
	req := opensearchapi.SearchReq{
		Indices: []string{idx.indexName},
		Body:    bytes.NewReader(queryJSON),
	}

	//Executo a chamada da busca
	res, err := idx.osCli.Search(
		ctx,
		&req,
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
