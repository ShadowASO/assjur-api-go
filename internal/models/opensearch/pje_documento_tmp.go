package opensearch

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ocrserver/internal/dominio/pje"
	"ocrserver/internal/types"
	"ocrserver/internal/utils/mslogger"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

type PjeDocumentoTmpIndex struct {
	osCli     *opensearchapi.Client
	indexName string
	timeout   time.Duration
}

func NewPjeDocumentoTmpIndex() *PjeDocumentoTmpIndex {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		mslogger.LoggerGlobal.Error(msg)
		return nil
	}

	return &PjeDocumentoTmpIndex{
		osCli:     osClient,
		indexName: "pje_documentos_tmp",
		timeout:   10 * time.Second,
	}
}

// MakePjeDocumentoTmpID monta um _id estável para o índice temporário.
// Como cada registro corresponde a um documento PJe dentro de um contexto,
// o ideal é usar id_ctxt + "_" + id_pje.
func MakePjeDocumentoTmpID(idCtxt string, idPje string) string {
	idCtxt = strings.TrimSpace(idCtxt)
	idPje = strings.TrimSpace(idPje)

	if idCtxt == "" || idPje == "" {
		return ""
	}

	return idCtxt + "_" + idPje
}

func pjeDocumentoTmpToResponse(id string, src pje.PjeDocumentoTmp) *pje.ResponsePjeDocumentoTmp {
	return &pje.ResponsePjeDocumentoTmp{
		Id:                    id,
		IdCtxt:                src.IdCtxt,
		NumeroProcesso:        src.NumeroProcesso,
		IdPje:                 src.IdPje,
		TipoDocumento:         src.TipoDocumento,
		Descricao:             src.Descricao,
		StorageID:             src.StorageID,
		Mimetype:              src.Mimetype,
		FormatoEntrega:        src.FormatoEntrega,
		ConteudoTexto:         src.ConteudoTexto,
		ConteudoHTML:          src.ConteudoHTML,
		ConteudoBase64:        src.ConteudoBase64,
		Status:                src.Status,
		Erro:                  src.Erro,
		DataJuntadaMillis:     src.DataJuntadaMillis,
		UsuarioJuntadaArquivo: src.UsuarioJuntadaArquivo,
		CriadoEm:              src.CriadoEm,
		AtualizadoEm:          src.AtualizadoEm,
		ExpiraEm:              src.ExpiraEm,
	}
}

func (idx *PjeDocumentoTmpIndex) Indexa(
	row pje.PjeDocumentoTmp,
	idOptional string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	row.IdCtxt = strings.TrimSpace(row.IdCtxt)
	row.NumeroProcesso = strings.TrimSpace(row.NumeroProcesso)
	row.IdPje = strings.TrimSpace(row.IdPje)
	row.TipoDocumento = strings.TrimSpace(row.TipoDocumento)
	row.Descricao = strings.TrimSpace(row.Descricao)
	row.StorageID = strings.TrimSpace(row.StorageID)
	row.Mimetype = strings.TrimSpace(row.Mimetype)
	row.FormatoEntrega = strings.TrimSpace(row.FormatoEntrega)
	row.Status = strings.TrimSpace(row.Status)

	if row.IdCtxt == "" {
		return nil, fmt.Errorf("id_ctxt vazio")
	}

	if row.IdPje == "" {
		return nil, fmt.Errorf("id_pje vazio")
	}

	idOptional = strings.TrimSpace(idOptional)
	if idOptional == "" {
		idOptional = MakePjeDocumentoTmpID(row.IdCtxt, row.IdPje)
	}

	if idOptional == "" {
		return nil, fmt.Errorf("não foi possível montar o _id do documento temporário")
	}

	now := time.Now()

	if row.CriadoEm == nil {
		row.CriadoEm = &now
	}

	if row.AtualizadoEm == nil {
		row.AtualizadoEm = &now
	}

	if row.Status == "" {
		row.Status = "pendente"
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Index(
		ctx,
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: idOptional,
			Body:       opensearchutil.NewJSONReader(row),
			Params: opensearchapi.IndexParams{
				Refresh: "true",
			},
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao indexar documento temporário PJe: %v", err)
		mslogger.LoggerGlobal.Error(msg)
		return nil, err
	}

	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	return pjeDocumentoTmpToResponse(res.ID, row), nil
}

func (idx *PjeDocumentoTmpIndex) Update(
	id string,
	patch pje.PjeDocumentoTmpPatch,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	doc := types.JsonMap{}

	if patch.IdCtxt != nil {
		doc["id_ctxt"] = strings.TrimSpace(*patch.IdCtxt)
	}

	if patch.NumeroProcesso != nil {
		doc["numero_processo"] = strings.TrimSpace(*patch.NumeroProcesso)
	}

	if patch.IdPje != nil {
		doc["id_pje"] = strings.TrimSpace(*patch.IdPje)
	}

	if patch.TipoDocumento != nil {
		doc["tipo_documento"] = strings.TrimSpace(*patch.TipoDocumento)
	}

	if patch.Descricao != nil {
		doc["descricao"] = strings.TrimSpace(*patch.Descricao)
	}

	if patch.StorageID != nil {
		doc["storage_id"] = strings.TrimSpace(*patch.StorageID)
	}

	if patch.Mimetype != nil {
		doc["mimetype"] = strings.TrimSpace(*patch.Mimetype)
	}

	if patch.FormatoEntrega != nil {
		doc["formato_entrega"] = strings.TrimSpace(*patch.FormatoEntrega)
	}

	if patch.ConteudoTexto != nil {
		doc["conteudo_texto"] = *patch.ConteudoTexto
	}

	if patch.ConteudoHTML != nil {
		doc["conteudo_html"] = *patch.ConteudoHTML
	}

	if patch.ConteudoBase64 != nil {
		doc["conteudo_base64"] = *patch.ConteudoBase64
	}

	if patch.Status != nil {
		doc["status"] = strings.TrimSpace(*patch.Status)
	}

	if patch.Erro != nil {
		doc["erro"] = *patch.Erro
	}

	if patch.DataJuntadaMillis != nil {
		doc["data_juntada_millis"] = *patch.DataJuntadaMillis
	}

	if patch.UsuarioJuntadaArquivo != nil {
		doc["usuario_juntada_arquivo"] = strings.TrimSpace(*patch.UsuarioJuntadaArquivo)
	}

	if patch.CriadoEm != nil {
		doc["criado_em"] = patch.CriadoEm
	}

	if patch.ExpiraEm != nil {
		doc["expira_em"] = patch.ExpiraEm
	}

	now := time.Now()
	if patch.AtualizadoEm != nil {
		doc["atualizado_em"] = patch.AtualizadoEm
	} else {
		doc["atualizado_em"] = &now
	}

	if len(doc) == 0 {
		return nil, fmt.Errorf("nenhum campo informado para atualização")
	}

	body := types.JsonMap{
		"doc":     doc,
		"_source": true,
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

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
		msg := fmt.Sprintf("Erro ao atualizar documento temporário PJe: %v", err)
		mslogger.LoggerGlobal.Error(msg)
		return nil, err
	}

	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result UpdateResponseGeneric[pje.PjeDocumentoTmp]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao decodificar resposta do update PJe tmp: %v", err)
		return nil, err
	}

	return pjeDocumentoTmpToResponse(res.ID, result.Get.Source), nil
}

func (idx *PjeDocumentoTmpIndex) UpdateStatus(
	id string,
	status string,
	erroMsg string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return nil, fmt.Errorf("status vazio")
	}

	now := time.Now()

	patch := pje.PjeDocumentoTmpPatch{
		Status:       &status,
		Erro:         &erroMsg,
		AtualizadoEm: &now,
	}

	return idx.Update(id, patch)
}

func (idx *PjeDocumentoTmpIndex) Delete(id string) error {
	if idx == nil || idx.osCli == nil {
		err := fmt.Errorf("OpenSearch não conectado")
		mslogger.LoggerGlobal.Error(err.Error())
		return err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("id vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Document.Delete(
		ctx,
		opensearchapi.DocumentDeleteReq{
			Index:      idx.indexName,
			DocumentID: id,
			Params: opensearchapi.DocumentDeleteParams{
				Refresh: "true",
			},
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao deletar documento temporário PJe: %v", err)
		mslogger.LoggerGlobal.Error(msg)
		return err
	}

	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return err
	}
	defer res.Inspect().Response.Body.Close()

	return nil
}

func (idx *PjeDocumentoTmpIndex) ConsultaById(
	id string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	query := types.JsonMap{
		"size": 1,
		"query": types.JsonMap{
			"ids": types.JsonMap{
				"values": []string{id},
			},
		},
	}

	req := opensearchapi.SearchReq{
		Indices: []string{idx.indexName},
		Body:    opensearchutil.NewJSONReader(query),
	}

	res, err := idx.osCli.Search(ctx, &req)
	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar documento temporário PJe por ID: %v", err)
		mslogger.LoggerGlobal.Error(msg)
		return nil, err
	}

	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result SearchResponseGeneric[pje.PjeDocumentoTmp]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		mslogger.LoggerGlobal.Error(msg)
		return nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	hit := result.Hits.Hits[0]

	return pjeDocumentoTmpToResponse(hit.ID, hit.Source), nil
}

func (idx *PjeDocumentoTmpIndex) ConsultaByIdPje(
	idCtxt string,
	idPje string,
) (*pje.ResponsePjeDocumentoTmp, error) {
	id := MakePjeDocumentoTmpID(idCtxt, idPje)
	if id == "" {
		return nil, fmt.Errorf("parâmetros inválidos: id_ctxt=%q, id_pje=%q", idCtxt, idPje)
	}

	return idx.ConsultaById(id)
}

func (idx *PjeDocumentoTmpIndex) ConsultaByIdCtxt(
	idCtxt string,
) ([]pje.ResponsePjeDocumentoTmp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	idCtxt = strings.TrimSpace(idCtxt)
	if idCtxt == "" {
		return nil, fmt.Errorf("id_ctxt vazio")
	}

	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"id_ctxt": idCtxt,
			},
		},
		"sort": []types.JsonMap{
			{
				"data_juntada_millis": types.JsonMap{
					"order":   "asc",
					"missing": "_last",
				},
			},
			{
				"criado_em": types.JsonMap{
					"order": "asc",
				},
			},
		},
	}

	return idx.searchMany(query)
}

func (idx *PjeDocumentoTmpIndex) ConsultaByNumeroProcesso(
	numeroProcesso string,
) ([]pje.ResponsePjeDocumentoTmp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	numeroProcesso = strings.TrimSpace(numeroProcesso)
	if numeroProcesso == "" {
		return nil, fmt.Errorf("numero_processo vazio")
	}

	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"numero_processo": numeroProcesso,
			},
		},
		"sort": []types.JsonMap{
			{
				"data_juntada_millis": types.JsonMap{
					"order":   "asc",
					"missing": "_last",
				},
			},
			{
				"criado_em": types.JsonMap{
					"order": "asc",
				},
			},
		},
	}

	return idx.searchMany(query)
}

func (idx *PjeDocumentoTmpIndex) ConsultaByStatus(
	status string,
) ([]pje.ResponsePjeDocumentoTmp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	status = strings.TrimSpace(status)
	if status == "" {
		return nil, fmt.Errorf("status vazio")
	}

	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"status": status,
			},
		},
		"sort": []types.JsonMap{
			{
				"criado_em": types.JsonMap{
					"order": "asc",
				},
			},
		},
	}

	return idx.searchMany(query)
}

func (idx *PjeDocumentoTmpIndex) ConsultaByTipoDocumento(
	tipoDocumento string,
) ([]pje.ResponsePjeDocumentoTmp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	tipoDocumento = strings.TrimSpace(tipoDocumento)
	if tipoDocumento == "" {
		return nil, fmt.Errorf("tipo_documento vazio")
	}

	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"tipo_documento": tipoDocumento,
			},
		},
		"sort": []types.JsonMap{
			{
				"data_juntada_millis": types.JsonMap{
					"order":   "asc",
					"missing": "_last",
				},
			},
			{
				"criado_em": types.JsonMap{
					"order": "asc",
				},
			},
		},
	}

	return idx.searchMany(query)
}

func (idx *PjeDocumentoTmpIndex) IsExiste(
	idCtxt string,
	idPje string,
) (bool, error) {
	if idx == nil || idx.osCli == nil {
		return false, fmt.Errorf("OpenSearch não conectado")
	}

	id := MakePjeDocumentoTmpID(idCtxt, idPje)
	if id == "" {
		return false, fmt.Errorf("parâmetros inválidos: id_ctxt=%q, id_pje=%q", idCtxt, idPje)
	}

	doc, err := idx.ConsultaById(id)
	if err != nil {
		return false, err
	}

	return doc != nil, nil
}

func (idx *PjeDocumentoTmpIndex) DeleteExpirados() error {
	if idx == nil || idx.osCli == nil {
		return fmt.Errorf("OpenSearch não conectado")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	query := types.JsonMap{
		"query": types.JsonMap{
			"range": types.JsonMap{
				"expira_em": types.JsonMap{
					"lt": "now",
				},
			},
		},
	}

	refresh := true
	waitForCompletion := true

	res, err := idx.osCli.Document.DeleteByQuery(
		ctx,
		opensearchapi.DocumentDeleteByQueryReq{
			Indices: []string{idx.indexName},
			Body:    opensearchutil.NewJSONReader(query),
			Params: opensearchapi.DocumentDeleteByQueryParams{
				Refresh:           &refresh,
				Conflicts:         "proceed",
				WaitForCompletion: &waitForCompletion,
			},
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro ao deletar documentos temporários expirados: %v", err)
		mslogger.LoggerGlobal.Error(msg)
		return err
	}

	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return err
	}
	defer res.Inspect().Response.Body.Close()

	return nil
}

func (idx *PjeDocumentoTmpIndex) searchMany(
	query types.JsonMap,
) ([]pje.ResponsePjeDocumentoTmp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	req := opensearchapi.SearchReq{
		Indices: []string{idx.indexName},
		Body:    opensearchutil.NewJSONReader(query),
	}

	res, err := idx.osCli.Search(ctx, &req)
	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar documentos temporários PJe: %v", err)
		mslogger.LoggerGlobal.Error(msg)
		return nil, err
	}

	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result SearchResponseGeneric[pje.PjeDocumentoTmp]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		mslogger.LoggerGlobal.Error(msg)
		return nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]pje.ResponsePjeDocumentoTmp, 0, len(result.Hits.Hits))

	for _, hit := range result.Hits.Hits {
		row := pjeDocumentoTmpToResponse(hit.ID, hit.Source)
		docs = append(docs, *row)
	}

	return docs, nil
}
