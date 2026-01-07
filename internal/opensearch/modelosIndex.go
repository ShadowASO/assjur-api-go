package opensearch

import (
	"bytes"

	"encoding/json"
	"fmt"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

const ExpectedVectorSize = 3072

type ModelosIndexType struct {
	osCli     *opensearchapi.Client
	indexName string
	timeout   time.Duration
}

var ModelosServiceGlobal *ModelosIndexType
var onceInitModelosService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitModelosService() {
	onceInitModelosService.Do(func() {
		ModelosServiceGlobal = NewIndexModelos()
		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

// Função para criar um novo cliente OpenSearch
// func NewIndexModelos(serviceOpenAi *services.OpenaiServiceType) *IndexModelosType {
func NewIndexModelos() *ModelosIndexType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		//log.Printf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		//return nil
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}

	return &ModelosIndexType{
		osCli:     osClient,
		indexName: "modelos",
		timeout:   10 * time.Second,
	}
}

// Estrutura do documento no OpenSearch
type ModelosText struct {
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
}

type ModelosEmbedding struct {
	Natureza             string    `json:"natureza"`
	Ementa               string    `json:"ementa"`
	Inteiro_teor         string    `json:"inteiro_teor"`
	EmentaEmbedding      []float32 `json:"ementa_embedding"`
	InteiroTeorEmbedding []float32 `json:"inteiro_teor_embedding"`
}
type ModelosRow struct {
	Natureza             string    `json:"natureza"`
	Ementa               string    `json:"ementa"`
	Inteiro_teor         string    `json:"inteiro_teor"`
	EmentaEmbedding      []float32 `json:"ementa_embedding"`
	InteiroTeorEmbedding []float32 `json:"inteiro_teor_embedding"`
}

type BodyModelosInsert struct {
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
}

type BodyModelosUpdate struct {
	Doc ModelosText `json:"doc"`
}
type BodyModelosSearch struct {
	Index_name   string `json:"index_name"`
	Natureza     string `json:"natureza"`
	Search_texto string `json:"search_texto"`
}

type ResponseModelos struct {
	Id           string `json:"id"`
	Natureza     string `json:"natureza"`
	Ementa       string `json:"ementa"`
	Inteiro_teor string `json:"inteiro_teor"`
}

// Indexar um novo documento

// func (idx *ModelosIndexType) Indexa(paramsData ModelosEmbedding) (*opensearchapi.IndexResp, error) {
func (idx *ModelosIndexType) Indexa(
	natureza string,
	ementa string,
	inteiro_teor string,
	ementaEmbedding []float32,
	inteiroTeorEmbedding []float32) (*opensearchapi.IndexResp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	// data, err := json.Marshal(paramsData)
	// if err != nil {

	// 	msg := fmt.Sprintf("Erro ao serializar JSON: %v", err)
	// 	logger.Log.Error(msg)
	// 	return nil, err
	// }

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	body := ModelosRow{
		Natureza:             natureza,
		Ementa:               ementa,
		Inteiro_teor:         inteiro_teor,
		EmentaEmbedding:      ementaEmbedding,
		InteiroTeorEmbedding: inteiroTeorEmbedding,
	}

	res, err := idx.osCli.Index(
		ctx,
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: "",
			//Body:       bytes.NewReader(data),
			Body: opensearchutil.NewJSONReader(body),
			Params: opensearchapi.IndexParams{
				Refresh: "true",
			},
		})

	if err != nil {
		msg := fmt.Sprintf("Erro ao realizar indexação: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	return res, nil
}

// Atualizar documento
func (idx *ModelosIndexType) Update(id string, paramsData ModelosText) (*opensearchapi.UpdateResp, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	updateData := BodyModelosUpdate{Doc: paramsData}

	data, err := json.Marshal(updateData)
	if err != nil {

		msg := fmt.Sprintf("Erro ao serializar JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}

	res, err := idx.osCli.Update(
		ctx,
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
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}

	return res, nil
}

// Deletar documento identificado pelo ID
// func (idx *ModelosIndexType) Delete(id string) (*opensearchapi.DocumentDeleteResp, error) {
func (idx *ModelosIndexType) Delete(id string) error {
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

func (idx *ModelosIndexType) ConsultaById(id string) (*ResponseModelos, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	//Cria o objeto da requisição
	req := opensearchapi.DocumentGetReq{
		Index:      idx.indexName,
		DocumentID: id,
	}
	//Executa passando o objeto da requisição
	res, err := idx.osCli.Document.Get(
		ctx,
		req,
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

	var result DocumentGetResponse[ModelosEmbedding]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	if !result.Found {
		logger.Log.Infof("id=%s não encontrado (found=false)", id)
		return nil, nil
	}

	src := result.Source

	return &ResponseModelos{
		Id:           result.ID,
		Natureza:     src.Natureza,
		Ementa:       src.Ementa,
		Inteiro_teor: src.Inteiro_teor,
	}, nil
}

// ==========================
// Busca semântica (KNN)
// ==========================

/*
ConsultaSemantica:
- faz busca KNN separada para ementa_embedding e inteiro_teor_embedding
- aplica filtro por natureza (term) quando informado
- mescla resultados por ID, preservando o maior score
- ordena por score desc e limita retorno
*/
func (idx *ModelosIndexType) ConsultaSemantica(vector []float32, natureza string) ([]ResponseModelos, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	if len(vector) != ExpectedVectorSize {
		msg := fmt.Sprintf("Erro: o vetor enviado tem dimensão %d, mas o índice espera %d dimensões.", len(vector), ExpectedVectorSize)
		logger.Log.Error(msg)
		return nil, erros.CreateError(msg)
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	natureza = strings.TrimSpace(natureza)

	// Função auxiliar para construir query KNN para um campo com filtro opcional de natureza
	buildKnnQuery := func(field string) map[string]any {
		knn := map[string]any{
			"knn": map[string]any{
				field: map[string]any{
					"vector": vector,
					"k":      20,
				},
			},
		}

		if natureza == "" {
			return knn
		}

		return map[string]any{
			"bool": map[string]any{
				"must": knn,
				"filter": []any{
					map[string]any{
						"term": map[string]any{
							"natureza": natureza,
						},
					},
				},
			},
		}
	}

	queries := []map[string]any{
		buildKnnQuery("ementa_embedding"),
		buildKnnQuery("inteiro_teor_embedding"),
	}

	type searchResultItem struct {
		Id    string
		Doc   ResponseModelos
		Score float64
	}

	// Evita duplicados por ID (mantém o maior score)
	resultMap := make(map[string]searchResultItem)

	for _, q := range queries {
		queryBody := map[string]any{
			"size": 20,
			"_source": map[string]any{
				"excludes": []string{"ementa_embedding", "inteiro_teor_embedding"},
			},
			"query": q,
		}

		queryJSON, err := json.Marshal(queryBody)
		if err != nil {
			msg := fmt.Sprintf("Erro ao serializar query JSON: %v", err)
			logger.Log.Error(msg)
			return nil, err
		}

		res, err := idx.osCli.Search(ctx, &opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			Body:    bytes.NewReader(queryJSON),
		})
		if err != nil {
			msg := fmt.Sprintf("Erro ao consultar o OpenSearch: %v", err)
			logger.Log.Error(msg)
			return nil, erros.CreateError(msg, err.Error())
		}

		// Fecha por iteração (não use defer no loop)
		httpRes := res.Inspect().Response

		// Decodifica em memória (garante que body será consumido uma vez e fechado)
		var result SearchResponseGeneric[ModelosText]
		err = func() error {
			defer httpRes.Body.Close()
			return DecodeJSONHTTP(httpRes, &result)
		}()
		if err != nil {
			logger.Log.Error(err.Error())
			return nil, err
		}

		for _, hit := range result.Hits.Hits {
			src := hit.Source

			doc := ResponseModelos{
				Id:       hit.ID,
				Natureza: src.Natureza,
				Ementa:   src.Ementa,
				//InteiroTeor: src.InteiroTeor,
				Inteiro_teor: src.Inteiro_teor,
			}

			// redundante (já filtrou), mas mantém segurança
			if natureza != "" && doc.Natureza != natureza {
				continue
			}

			score := 0.0
			if hit.Score != nil {
				score = *hit.Score
			}

			existing, found := resultMap[doc.Id]
			if !found || score > existing.Score {
				resultMap[doc.Id] = searchResultItem{
					Id:    doc.Id,
					Doc:   doc,
					Score: score,
				}
			}
		}
	}

	// Ordena por score desc
	results := make([]searchResultItem, 0, len(resultMap))
	for _, v := range resultMap {
		results = append(results, v)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limite final
	limit := 10
	if len(results) < limit {
		limit = len(results)
	}

	out := make([]ResponseModelos, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, results[i].Doc)
	}

	return out, nil
}
