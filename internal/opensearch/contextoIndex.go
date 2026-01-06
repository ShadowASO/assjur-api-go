package opensearch

import (
	"bytes"

	"encoding/json"
	"errors"
	"fmt"
	"io"

	"net/http"
	"strings"
	"time"

	"ocrserver/internal/types"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/google/uuid"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

// =========================================================
// Estrutura principal
// =========================================================

type ContextoIndexType struct {
	osCli     *opensearchapi.Client
	indexName string
	timeout   time.Duration
}

// =========================================================
// Construtor e inicialização
// =========================================================
// Novo cliente para o índice contexto
func NewContextoIndex() *ContextoIndexType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}

	return &ContextoIndexType{
		osCli:     osClient,
		indexName: "contexto",
		timeout:   10 * time.Second,
	}
}

// =========================================================
// Estruturas auxiliares internas
// =========================================================
type ContextoRow struct {
	IdCtxt           string    `json:"id_ctxt"`
	NrProc           string    `json:"nr_proc"`
	Juizo            string    `json:"juizo"`
	Classe           string    `json:"classe"`
	Assunto          string    `json:"assunto"`
	PromptTokens     int       `json:"prompt_tokens,omitempty"`
	CompletionTokens int       `json:"completion_tokens,omitempty"`
	DtInc            time.Time `json:"dt_inc"`
	UsernameInc      string    `json:"username_inc"`
	Status           string    `json:"status"`
}

type ResponseContextoRow struct {
	Id               string    `json:"id"`      // _id do OpenSearch
	IdCtxt           string    `json:"id_ctxt"` // campo do documento
	NrProc           string    `json:"nr_proc"`
	Juizo            string    `json:"juizo"`
	Classe           string    `json:"classe"`
	Assunto          string    `json:"assunto"`
	PromptTokens     int       `json:"prompt_tokens,omitempty"`
	CompletionTokens int       `json:"completion_tokens,omitempty"`
	DtInc            time.Time `json:"dt_inc"`
	UsernameInc      string    `json:"username_inc"`
	Status           string    `json:"status"`
}

// Indexa (cria/upsert) um contexto.
func (idx *ContextoIndexType) Indexa(
	nrProc string,
	juizo string,
	classe string,
	assunto string,
	usernameInc string,

) (*ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	// ***** Criação do ID_CTXT  *************************
	idv7, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar uuidv7: %w", err)
	}
	idCtxt := idv7.String()
	//****************************************************
	now := time.Now()
	//*********************************
	doc := ContextoRow{
		IdCtxt:           idCtxt,
		NrProc:           nrProc,
		Juizo:            juizo,
		Classe:           classe,
		Assunto:          assunto,
		PromptTokens:     0,
		CompletionTokens: 0,
		DtInc:            now,
		UsernameInc:      usernameInc,
		Status:           "S",
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Index(
		ctx,
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: idCtxt, // Estou usando o id_ctxt como _id do documento
			Body:       opensearchutil.NewJSONReader(&doc),
			Params: opensearchapi.IndexParams{
				Refresh: "true",
			},
		})
	if err != nil {
		msg := fmt.Sprintf("Erro ao indexar contexto: %w", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	return &ResponseContextoRow{
		Id:               idCtxt, // ✅ como você usou DocumentID=idCtxt, _id=idCtxt
		IdCtxt:           idCtxt,
		NrProc:           nrProc,
		Juizo:            juizo,
		Classe:           classe,
		Assunto:          assunto,
		PromptTokens:     0,
		CompletionTokens: 0,
		DtInc:            now,
		UsernameInc:      usernameInc,
		Status:           "S",
	}, nil
}

func (idx *ContextoIndexType) Update(
	idCtxt string,
	juizo string,
	classe string,
	assunto string,
) (*ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	if strings.TrimSpace(idCtxt) == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}

	body := types.JsonMap{
		"doc": types.JsonMap{
			"juizo":   juizo,
			"classe":  classe,
			"assunto": assunto,
		},
		"_source": true, // tenta devolver o source atualizado
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Update(
		ctx,
		opensearchapi.UpdateReq{
			Index:      idx.indexName,
			DocumentID: idCtxt, // ✅ se você adotou _id=id_ctxt
			Body:       opensearchutil.NewJSONReader(body),
			Params: opensearchapi.UpdateParams{
				Refresh: "true",
			},
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao atualizar contexto: %w", err)
		logger.Log.Error(msg)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	// Resposta do update pode trazer "get": {"_source": {...}} se _source=true
	var upd struct {
		Get struct {
			Source ContextoRow `json:"_source"`
		} `json:"get"`
	}
	_ = json.NewDecoder(res.Inspect().Response.Body).Decode(&upd) // se não vier, ok (depende do cluster/config)

	src := upd.Get.Source
	// fallback mínimo caso não venha _source
	if src.IdCtxt == "" {
		src.IdCtxt = idCtxt
		src.Juizo = juizo
		src.Classe = classe
		src.Assunto = assunto
	}

	return &ResponseContextoRow{
		Id:               idCtxt,
		IdCtxt:           src.IdCtxt,
		NrProc:           src.NrProc,
		Juizo:            src.Juizo,
		Classe:           src.Classe,
		Assunto:          src.Assunto,
		PromptTokens:     src.PromptTokens,
		CompletionTokens: src.CompletionTokens,
		DtInc:            src.DtInc,
		UsernameInc:      src.UsernameInc,
		Status:           src.Status,
	}, nil
}

// DeleteByID deleta um documento diretamente pelo _id do OpenSearch
func (idx *ContextoIndexType) Delete(id string) error {
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

// Consulta por _id
func (idx *ContextoIndexType) ConsultaById(id string) (*ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	req := opensearchapi.DocumentGetReq{
		Index:      idx.indexName,
		DocumentID: id,
	}
	//Executa, passando a requisição
	res, err := idx.osCli.Document.Get(
		ctx,
		req,
	)
	//------------------------------------------------------

	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar documento: %s : %s = %v", id, idx.indexName, err)
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

	//var result searchResponseContexto
	//var result SearchResponseGeneric[ContextoRow]
	var result DocumentGetResponse[ContextoRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	// if len(result.Hits.Hits) == 0 {
	// 	return nil, nil
	// }

	// hit := result.Hits.Hits[0]
	// src := hit.Source

	if !result.Found {
		logger.Log.Infof("id=%s não encontrado (found=false)", id)
		return nil, nil
	}

	src := result.Source

	doc := ResponseContextoRow{
		Id:               result.ID,
		IdCtxt:           src.IdCtxt,
		NrProc:           src.NrProc,
		Juizo:            src.Juizo,
		Classe:           src.Classe,
		Assunto:          src.Assunto,
		PromptTokens:     src.PromptTokens,
		CompletionTokens: src.CompletionTokens,
		DtInc:            src.DtInc,
		UsernameInc:      src.UsernameInc,
		Status:           src.Status,
	}
	return &doc, nil
}

// Consultar documentos por id_ctxt
func (idx *ContextoIndexType) ConsultaByIdCtxt(idCtxt string) ([]ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	idCtxt = strings.TrimSpace(idCtxt)
	if idCtxt == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	//----  Crio o body
	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"id_ctxt": idCtxt,
			},
		},
	}
	queryJSON, err := json.Marshal(query)
	if err != nil {
		msg := fmt.Sprintf("Erro ao serializar queryJSON: %v", err)
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
	//-----------------------------
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

	var result SearchResponseGeneric[ContextoRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	// ✅ Correção do panic
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]ResponseContextoRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		src := hit.Source
		docs = append(docs, ResponseContextoRow{

			Id:               hit.ID,
			IdCtxt:           src.IdCtxt,
			NrProc:           src.NrProc,
			Juizo:            src.Juizo,
			Classe:           src.Classe,
			Assunto:          src.Assunto,
			PromptTokens:     src.PromptTokens,
			CompletionTokens: src.CompletionTokens,
			DtInc:            src.DtInc,
			UsernameInc:      src.UsernameInc,
			Status:           src.Status,
		})
	}
	return docs, nil
}

var ErrContextoNotFound = errors.New("contexto não encontrado")

// Consulta por nr_proc (ex.: busca do “contexto” de um processo)
func (idx *ContextoIndexType) ConsultaByProcesso(nrProc string) (*ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	nrProc = strings.TrimSpace(nrProc)
	if nrProc == "" {
		return nil, fmt.Errorf("id vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	// Preferível para identificadores: keyword (exato).
	// Se seu mapping não tiver nr_proc.keyword, troque para "nr_proc" ou use match.
	query := types.JsonMap{
		"size": 1, // já que você retorna só um documento
		"query": types.JsonMap{
			"term": types.JsonMap{
				"nr_proc": nrProc,
			},
		},
	}
	queryJSON, err := json.Marshal(query)
	if err != nil {
		msg := fmt.Sprintf("Erro ao serializar queryJSON: %v", err)
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
		logger.Log.Errorf("Erro ao executar search: %s : %s = %v", idx.indexName, nrProc, err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	//var result searchResponseContexto
	var result SearchResponseGeneric[ContextoRow]
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

	doc := ResponseContextoRow{
		Id:               hit.ID,
		IdCtxt:           src.IdCtxt,
		NrProc:           src.NrProc,
		Juizo:            src.Juizo,
		Classe:           src.Classe,
		Assunto:          src.Assunto,
		PromptTokens:     src.PromptTokens,
		CompletionTokens: src.CompletionTokens,
		DtInc:            src.DtInc,
		UsernameInc:      src.UsernameInc,
		Status:           src.Status,
	}

	return &doc, nil
}

// Verifica se já existe um contexto para nr_proc
func (idx *ContextoIndexType) IsExistes(nrProc string) (bool, error) {
	if nrProc == "" {
		return false, fmt.Errorf("nr_proc vazio")
	}
	if idx.osCli == nil {
		return false, fmt.Errorf("OpenSearch não conectado")
	}

	docs, err := idx.ConsultaByProcesso(nrProc)
	if err != nil {
		return false, erros.CreateError("Erro ao consultar o OpenSearch", err.Error())
	}

	return docs != nil, nil
}
func (idx *ContextoIndexType) SelectContextoByProcessoStartsWith(nrProcPart string) ([]ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	if nrProcPart == "" {
		return []ResponseContextoRow{}, nil
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	//Crio o body/query
	query := types.JsonMap{
		"size": 100, // ajuste conforme necessidade
		"query": types.JsonMap{
			"prefix": types.JsonMap{
				"nr_proc.keyword": nrProcPart, // se nr_proc for keyword puro, use "nr_proc"
			},
		},
		"sort": types.JsonMap{
			"nr_proc.keyword": map[string]any{"order": "asc"},
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
		msg := fmt.Sprintf("Erro ao consultar documento: %s : %s = %v", idx.indexName, nrProcPart, err)
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

	var result SearchResponseGeneric[ContextoRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}
	// ✅ Correção do panic
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]ResponseContextoRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		src := hit.Source
		docs = append(docs, ResponseContextoRow{
			Id:               hit.ID,
			IdCtxt:           src.IdCtxt,
			NrProc:           src.NrProc,
			Juizo:            src.Juizo,
			Classe:           src.Classe,
			Assunto:          src.Assunto,
			PromptTokens:     src.PromptTokens,
			CompletionTokens: src.CompletionTokens,
			DtInc:            src.DtInc,
			UsernameInc:      src.UsernameInc,
			Status:           src.Status,
		})
	}

	return docs, nil
}

func (idx *ContextoIndexType) SelectContextos(limit, offset int) ([]ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	// saneamento básico
	if limit <= 0 {
		limit = QUERY_MAX_SIZE // ou um default seu (ex: 1000)
	}
	if offset < 0 {
		offset = 0
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	//Crio o body/query
	query := types.JsonMap{
		"size": limit,
		"from": offset,
		"query": types.JsonMap{
			"match_all": types.JsonMap{},
		},
		// Ordenação estável (evita “pulos” entre páginas quando há inserções concorrentes)
		"sort": []any{
			types.JsonMap{"dt_inc": types.JsonMap{"order": "desc"}},
			types.JsonMap{"_id": types.JsonMap{"order": "desc"}},
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
		logger.Log.Errorf("Erro ao executar busca: %v", err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	if res.Inspect().Response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	var result SearchResponseGeneric[ContextoRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}
	// ✅ Correção do panic
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]ResponseContextoRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		src := hit.Source
		docs = append(docs, ResponseContextoRow{
			Id:               hit.ID,
			IdCtxt:           src.IdCtxt,
			NrProc:           src.NrProc,
			Juizo:            src.Juizo,
			Classe:           src.Classe,
			Assunto:          src.Assunto,
			PromptTokens:     src.PromptTokens,
			CompletionTokens: src.CompletionTokens,
			DtInc:            src.DtInc,
			UsernameInc:      src.UsernameInc,
			Status:           src.Status,
		})
	}

	return docs, nil
}

func (idx *ContextoIndexType) IncrementTokensAtomic(
	idCtxt string,
	promptTokensInc int,
	completionTokensInc int,
) (*ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	idCtxt = strings.TrimSpace(idCtxt)
	if idCtxt == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}

	updateBody := types.JsonMap{
		"script": types.JsonMap{
			"lang": "painless",
			"source": `
				if (ctx._source.prompt_tokens == null) { ctx._source.prompt_tokens = 0; }
				if (ctx._source.completion_tokens == null) { ctx._source.completion_tokens = 0; }
				ctx._source.prompt_tokens += params.pt;
				ctx._source.completion_tokens += params.ct;
			`,
			"params": types.JsonMap{
				"pt": promptTokensInc,
				"ct": completionTokensInc,
			},
		},
		// faz o OpenSearch devolver o _source atualizado
		"_source": true,
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()
	//Crio a requisição
	req := opensearchapi.UpdateReq{
		Index:      idx.indexName,
		DocumentID: idCtxt, // _id
		Body:       opensearchutil.NewJSONReader(updateBody),
		Params: opensearchapi.UpdateParams{
			Refresh: "true",
		},
	}
	//Executo a chamada de update
	res, err := idx.osCli.Update(
		ctx,
		req,
	)

	if err != nil {
		logger.Log.Errorf("Erro no update: %s : %s = %v", idx.indexName, idCtxt, err)
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	// checa status HTTP
	if sc := res.Inspect().Response.StatusCode; sc < 200 || sc >= 300 {
		b, _ := io.ReadAll(res.Inspect().Response.Body)
		return nil, fmt.Errorf("update falhou (status=%d): %s", sc, string(b))
	}

	// Decodifica resposta do update para pegar _source atualizado
	// Estrutura mínima (adapte conforme seu client/response):
	var upd struct {
		Get struct {
			Source ContextoRow `json:"_source"`
		} `json:"get"`
	}
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&upd); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta do update: %w", err)
	}

	src := upd.Get.Source
	if src.IdCtxt == "" {
		src.IdCtxt = idCtxt
	}

	return &ResponseContextoRow{
		Id:               idCtxt,
		IdCtxt:           src.IdCtxt,
		NrProc:           src.NrProc,
		Juizo:            src.Juizo,
		Classe:           src.Classe,
		Assunto:          src.Assunto,
		PromptTokens:     src.PromptTokens,
		CompletionTokens: src.CompletionTokens,
		DtInc:            src.DtInc,
		UsernameInc:      src.UsernameInc,
		Status:           src.Status,
	}, nil
}
