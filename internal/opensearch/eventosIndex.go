package opensearch

import (
	"bytes"
	"net/http"

	"encoding/json"
	"fmt"

	"strings"
	"time"

	"ocrserver/internal/types"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

// =========================================================
// Estrutura principal
// =========================================================

type EventosIndex struct {
	osCli     *opensearchapi.Client
	indexName string
	timeout   time.Duration
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
		timeout:   10 * time.Second,
	}
}

// =========================================================
// Operações CRUD
// =========================================================

// Indexar um novo documento
func (idx *EventosIndex) Indexa(
	IdCtxt string,
	IdNatu int,
	IdPje string,
	Doc string,
	DocJsonRaw string,
	DocEmbedding []float32,
	idOptional string,
	userName string,
) (*ResponseEventosRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	nowTime := time.Now()

	body := EventosRow{
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		UsernameInc:  userName,
		DtInc:        nowTime,
		Doc:          Doc,
		DocJsonRaw:   DocJsonRaw,
		DocEmbedding: DocEmbedding,
	}

	res, err := idx.osCli.Index(
		ctx,
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: idOptional,
			Body:       opensearchutil.NewJSONReader(body),
			Params: opensearchapi.IndexParams{
				Refresh: "true",
			},
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao indexar: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	row := &ResponseEventosRow{
		Id:           res.ID,
		IdCtxt:       IdCtxt,
		IdNatu:       IdNatu,
		IdPje:        IdPje,
		UsernameInc:  userName,
		DtInc:        nowTime,
		Doc:          Doc,
		DocJsonRaw:   DocJsonRaw,
		DocEmbedding: DocEmbedding,
	}

	return row, nil
}

// Atualizar documento existente
func (idx *EventosIndex) Update(
	id string,
	idCtxt string,
	IdNatu int,
	IdPje string,
	Doc string,
	DocJson string,
	DocEmbedding []float32,
) (*ResponseEventosRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	if strings.TrimSpace(idCtxt) == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	//ATENÇÃO: Não podemos usar a estrutura genérica do registro, salvo se todos os campos
	//estiverem sendo alterados. Os campos não preenchidos importam no preenchimento do
	//registro com valores zerados/vazios.
	//Ou seja, todos os campos presentes em uma estrutura são considerados como campos a
	//serem modificado no registro do OpenSearch, resultado em campos zerados se não fo-
	//rem passados valores.
	//Se o update é parcial, precisamos criar uma estrutura sob medida contendo apenas os
	//campos a alterar. Além disso, O json deve NESESSARIAMENTE conter o field "doc":
	//**
	//Exemplo: types.JsonMap{doc:types.JsonMap{fields}}

	body := types.JsonMap{
		"doc": ResponseEventosRow{
			Id:           id,
			IdCtxt:       idCtxt,
			IdNatu:       IdNatu,
			IdPje:        IdPje,
			Doc:          Doc,
			DocJsonRaw:   DocJson,
			DocEmbedding: DocEmbedding,
		},
		"_source": true, // tenta devolver o source atualizado
	}

	res, err := idx.osCli.Update(
		ctx,
		opensearchapi.UpdateReq{
			Index:      idx.indexName,
			DocumentID: id,
			Body:       opensearchutil.NewJSONReader(body),
			Params: opensearchapi.UpdateParams{
				Refresh: "true",
			},
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao realizar indexação: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	row := &ResponseEventosRow{
		Id:         id,
		IdCtxt:     idCtxt,
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
	if idx == nil || idx.osCli == nil {
		err := fmt.Errorf("OpenSearch não conectado")
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
		},
	)

	if err != nil {
		msg := fmt.Sprintf("Erro realizar delete: %v", err)
		logger.Log.Error(msg)
		return err
	}
	if err = ReadOSErr(res.Inspect().Response); err != nil {
		return err
	}
	defer res.Inspect().Response.Body.Close()

	return nil
}

// Consultar documento pelo ID
func (idx *EventosIndex) ConsultaById(id string) (*ResponseEventosRow, int, error) {
	if idx == nil || idx.osCli == nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("OpenSearch não conectado")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("id vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	//Cria o objeto da requisição
	req := opensearchapi.DocumentGetReq{
		Index:      idx.indexName,
		DocumentID: id,
	}
	//Executa passando o objeto da requisição
	res, _ := idx.osCli.Document.Get(
		ctx,
		req,
	)
	/*
		Os erros do opensearchapi precisam ser verificos na resposta, pois quando não
		são encontrados registros, ele informa como um erro e não podemos definir o
		motivo, se erro interno ou registro não encontrado.
	*/
	if res == nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("resposta nula do OpenSearch")
	}
	defer res.Inspect().Response.Body.Close()

	statusCode := res.Inspect().Response.StatusCode
	if statusCode == http.StatusNotFound {
		return nil, statusCode, nil
	}

	var result DocumentGetResponse[EventosRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, http.StatusInternalServerError, err
	}

	if !result.Found {
		logger.Log.Infof("id=%s não encontrado (found=false)", id)
		return nil, http.StatusNotFound, nil
	}

	src := result.Source

	return &ResponseEventosRow{
		Id:          result.ID,
		IdCtxt:      src.IdCtxt,
		IdNatu:      src.IdNatu,
		IdPje:       src.IdPje,
		UsernameInc: src.UsernameInc,
		DtInc:       src.DtInc,
		Doc:         src.Doc,
		DocJsonRaw:  src.DocJsonRaw,
		//DocEmbedding: src.DocEmbedding,
	}, statusCode, nil
}

// Consultar documentos por id_ctxt
func (idx *EventosIndex) ConsultaByIdCtxt(idCtxt string) ([]ResponseEventosRow, error) {
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
			{"id_natu": types.JsonMap{"order": "asc"}},
		},
	}

	// queryJSON, err := json.Marshal(query)
	// if err != nil {
	// 	msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
	// 	logger.Log.Error(msg)
	// 	return nil, err
	// }

	res, err := idx.osCli.Search(
		ctx,
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			//Body:    bytes.NewReader(queryJSON),
			Body: opensearchutil.NewJSONReader(query),
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro realizar consulta by query: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result SearchResponseGeneric[EventosRow]

	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]ResponseEventosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		docs = append(docs, ResponseEventosRow{
			Id:          hit.ID,
			IdCtxt:      doc.IdCtxt,
			IdNatu:      doc.IdNatu,
			IdPje:       doc.IdPje,
			UsernameInc: doc.UsernameInc,
			DtInc:       doc.DtInc,
			Doc:         doc.Doc,
			DocJsonRaw:  doc.DocJsonRaw,
			//DocEmbedding: doc.DocEmbedding,
		})
	}
	return docs, nil
}

// Consultar documentos por id_natu
func (idx *EventosIndex) ConsultaByIdNatu(idNatu int) ([]ResponseEventosRow, error) {
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
	// queryJSON, err := json.Marshal(query)
	// if err != nil {
	// 	msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
	// 	logger.Log.Error(msg)
	// 	return nil, err
	// }

	res, err := idx.osCli.Search(
		ctx,
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			//Body:    bytes.NewReader(queryJSON),
			Body: opensearchutil.NewJSONReader(query),
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro realizar consulta by query: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result SearchResponseGeneric[EventosRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	// ✅ Correção do panic
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]ResponseEventosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		docs = append(docs, ResponseEventosRow{
			Id:          hit.ID,
			IdCtxt:      doc.IdCtxt,
			IdNatu:      doc.IdNatu,
			IdPje:       doc.IdPje,
			UsernameInc: doc.UsernameInc,
			DtInc:       doc.DtInc,
			Doc:         doc.Doc,
			DocJsonRaw:  doc.DocJsonRaw,
			//DocEmbedding: doc.DocEmbedding,
		})
	}
	return docs, nil
}

// Busca semântica por embedding
func (idx *EventosIndex) ConsultaSemantica(vector []float32, idNatuFilter int) ([]ResponseEventosRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

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
		logger.Log.Errorf("Erro ao consultar OpenSearch: %v", err)
		return nil, erros.CreateError(err.Error())
	}
	defer res.Inspect().Response.Body.Close()
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	// var result searchResponseEventos
	// if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
	// 	logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
	// 	return nil, erros.CreateError(err.Error())
	// }
	var result SearchResponseGeneric[EventosRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	//var docs []ResponseEventosRow
	docs := make([]ResponseEventosRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		doc := hit.Source
		if idNatuFilter > 0 && doc.IdNatu != idNatuFilter {
			continue
		}
		if len(docs) >= 5 {
			break
		}
		docs = append(docs, ResponseEventosRow{
			Id:          hit.ID,
			IdCtxt:      doc.IdCtxt,
			IdNatu:      doc.IdNatu,
			IdPje:       doc.IdPje,
			UsernameInc: doc.UsernameInc,
			DtInc:       doc.DtInc,
			Doc:         doc.Doc,
			DocJsonRaw:  doc.DocJsonRaw,
			//DocEmbedding: doc.DocEmbedding,
		})
	}
	return docs, nil
}

// Verificar existência de documento por id_ctxt + id_evento
func (idx *EventosIndex) IsExiste(idCtxt string, idEvento string) (bool, error) {
	if idCtxt == "" || idEvento == "" {
		return false, fmt.Errorf("parâmetros inválidos: idCtxt=%d, idPje=%q", idCtxt, idEvento)
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
					types.JsonMap{"term": types.JsonMap{"id_ctxt": idCtxt}},
					types.JsonMap{"term": types.JsonMap{"id_evento": idEvento}},
				},
			},
		},
	}

	queryBody, _ := json.Marshal(query)
	res, err := idx.osCli.Search(
		ctx,
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
