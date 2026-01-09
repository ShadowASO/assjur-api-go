package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ocrserver/internal/config"
	"ocrserver/internal/types"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

const ExpectedRagVectorSize = 3072

type BaseIndexType struct {
	osCli     *opensearchapi.Client
	indexName string
	timeout   time.Duration
}

var BaseIndexGlobal *BaseIndexType
var onceInitBaseIndex sync.Once

// Inicializa global
func InitBaseIndex() {
	onceInitBaseIndex.Do(func() {
		BaseIndexGlobal = NewBaseIndex()
		logger.Log.Info("Global RagService configurado com sucesso.")
	})
}

func NewBaseIndex() *BaseIndexType {
	osClient, err := OpenSearchGlobal.GetClient()
	if err != nil {
		msg := fmt.Sprintf("Erro ao obter uma instância do cliente OpenSearch: %v", err)
		logger.Log.Error(msg)
		return nil
	}
	return &BaseIndexType{
		osCli:     osClient,
		indexName: config.GlobalConfig.OpenSearchRagName,
		timeout:   10 * time.Second,
	}
}

type BaseRow struct {
	IdCtxt      string    `json:"id_ctxt,omitempty"`      // keyword
	IdPje       string    `json:"id_pje,omitempty"`       // keyword
	HashTexto   string    `json:"hash_texto,omitempty"`   // keyword (sha256/semelhante do texto)
	UsernameInc string    `json:"username_inc,omitempty"` // keyword
	DtInc       time.Time `json:"dt_inc,omitempty"`       // date
	Status      string    `json:"status,omitempty"`       // keyword (ex.: "S", "N", "A", etc.)

	Classe   string `json:"classe,omitempty"`   // text (analyzer=brazilian) + subfield classe.kw
	Assunto  string `json:"assunto,omitempty"`  // text + assunto.kw
	Natureza string `json:"natureza,omitempty"` // text + natureza.kw
	Tipo     string `json:"tipo,omitempty"`     // text + tipo.kw
	Tema     string `json:"tema,omitempty"`     // text + tema.kw

	Fonte string `json:"fonte,omitempty"` // keyword + fonte.text (se precisar)
	Texto string `json:"texto,omitempty"` // text (conteúdo do chunk)

	TextoEmbedding []float32 `json:"texto_embedding,omitempty"` // knn_vector dimension=3072

}

type ResponseBaseRow struct {
	Id          string    `json:"id"`
	IdCtxt      string    `json:"id_ctxt,omitempty"`      // keyword
	IdPje       string    `json:"id_pje,omitempty"`       // keyword
	HashTexto   string    `json:"hash_texto,omitempty"`   // keyword (sha256/semelhante do texto)
	UsernameInc string    `json:"username_inc,omitempty"` // keyword
	DtInc       time.Time `json:"dt_inc,omitempty"`       // date
	Status      string    `json:"status,omitempty"`       // keyword (ex.: "S", "N", "A", etc.)

	Classe   string `json:"classe,omitempty"`   // text (analyzer=brazilian) + subfield classe.kw
	Assunto  string `json:"assunto,omitempty"`  // text + assunto.kw
	Natureza string `json:"natureza,omitempty"` // text + natureza.kw
	Tipo     string `json:"tipo,omitempty"`     // text + tipo.kw
	Tema     string `json:"tema,omitempty"`     // text + tema.kw

	Fonte string `json:"fonte,omitempty"` // keyword + fonte.text (se precisar)
	Texto string `json:"texto,omitempty"` // text (conteúdo do chunk)

	//TextoEmbedding []float32 `json:"texto_embedding,omitempty"` // knn_vector dimension=3072

}

// Indexar documento
func (idx *BaseIndexType) Indexa(
	idCtxt string,
	idPje string,
	hashTexto string,
	usernameInc string,
	//dtInc time.Time,
	status string,

	classe string,
	assunto string,
	natureza string,
	tipo string,
	tema string,

	fonte string,
	texto string,

	textoEmbedding []float32,
	idOptional string,
) (*ResponseBaseRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	// ***** Criação do ID_CTXT  *************************
	idv7, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar uuidv7: %w", err)
	}

	if strings.TrimSpace(idCtxt) == "" {
		idUU := idv7.String()
		idCtxt = idUU
	}

	//****************************************************
	nowTime := time.Now()
	//*********************************
	body := BaseRow{
		IdCtxt:      idCtxt,
		IdPje:       idPje,
		HashTexto:   hashTexto,
		UsernameInc: usernameInc,
		DtInc:       nowTime,
		Status:      "S",

		Classe:         classe,
		Assunto:        assunto,
		Natureza:       natureza,
		Tipo:           tipo,
		Tema:           tema,
		Fonte:          fonte,
		Texto:          texto,
		TextoEmbedding: textoEmbedding,
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Index(
		ctx,
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: "", // Estou usando o id_ctxt como _id do documento
			Body:       opensearchutil.NewJSONReader(body),
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

	return &ResponseBaseRow{
		Id:          res.ID,
		IdCtxt:      idCtxt,
		IdPje:       idPje,
		HashTexto:   hashTexto,
		UsernameInc: usernameInc,
		DtInc:       nowTime,
		Status:      status,

		Classe:   classe,
		Assunto:  assunto,
		Natureza: natureza,
		Tipo:     tipo,
		Tema:     tema,
		Fonte:    fonte,
		Texto:    texto,
		//TextoEmbedding: textoEmbedding,
	}, nil
}

// Atualizar documento

func (idx *BaseIndexType) Update(
	id string,
	tema string,
	texto string,
	texto_embedding []float32,

) (*ResponseBaseRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id vazio")
	}

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
		"doc": types.JsonMap{
			"tema":            tema,
			"texto":           texto,
			"texto_embedding": texto_embedding,
		},
		"_source": true, // tenta devolver o source atualizado
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
		msg := fmt.Sprintf("Erro realizar update: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	//Pego o retorno do Update
	var result UpdateResponseGeneric[BaseRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	src := result.Get.Source

	return &ResponseBaseRow{
		Id:          res.ID,
		IdCtxt:      src.IdCtxt,
		IdPje:       src.IdPje,
		HashTexto:   src.HashTexto,
		UsernameInc: src.UsernameInc,
		DtInc:       src.DtInc,
		Status:      src.Status,

		Classe:   src.Classe,
		Assunto:  src.Assunto,
		Natureza: src.Natureza,
		Tipo:     src.Tipo,
		Tema:     src.Tema,
		Fonte:    src.Fonte,
		Texto:    src.Texto,
		//TextoEmbedding: src.TextoEmbedding,
	}, nil
}

// DeleteByID deleta um documento diretamente pelo _id do OpenSearch
func (idx *BaseIndexType) Delete(id string) error {
	if idx == nil || idx.osCli == nil {
		err := fmt.Errorf("OpenSearch não conectado")
		logger.Log.Error(err.Error())
		return err
	}
	if strings.TrimSpace(id) == "" {
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
				// ✅ Melhor opção para “sumir da lista” logo após o delete:
				Refresh: "true",
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

// Consulta por ID
func (idx *BaseIndexType) ConsultaById(id string) (*ResponseBaseRow, error) {
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
	res, _ := idx.osCli.Document.Get(
		ctx,
		req,
	)
	// if err != nil {
	// 	msg := fmt.Sprintf("Erro realizar consulta by query: %v", err)
	// 	logger.Log.Error(msg)
	// 	return nil, err
	// }
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result DocumentGetResponse[BaseRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	if !result.Found {
		logger.Log.Infof("id=%s não encontrado (found=false)", id)
		return nil, nil
	}

	src := result.Source

	return &ResponseBaseRow{
		Id:          res.ID,
		IdCtxt:      src.IdCtxt,
		IdPje:       src.IdPje,
		HashTexto:   src.HashTexto,
		UsernameInc: src.UsernameInc,
		DtInc:       src.DtInc,
		Status:      src.Status,

		Classe:   src.Classe,
		Assunto:  src.Assunto,
		Natureza: src.Natureza,
		Tipo:     src.Tipo,
		Tema:     src.Tema,
		Fonte:    src.Fonte,
		Texto:    src.Texto,
		//TextoEmbedding: src.TextoEmbedding,
	}, nil
}

// Busca semântica
func (idx *BaseIndexType) ConsultaSemantica(vector []float32, natureza string) ([]ResponseBaseRow, error) {
	if idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	if len(vector) != ExpectedRagVectorSize {
		return nil, erros.CreateError(fmt.Sprintf("vetor tem %d dimensões, esperado %d", len(vector), ExpectedRagVectorSize))
	}

	knnQuery := map[string]interface{}{
		"knn": map[string]interface{}{
			"texto_embedding": map[string]interface{}{
				"vector": vector,
				"k":      20,
			},
		},
	}
	if natureza != "" {
		knnQuery = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": knnQuery,
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{"natureza": natureza},
					},
				},
			},
		}
	}

	query := map[string]interface{}{
		"size": 10,
		"_source": map[string]interface{}{
			"excludes": []string{"texto_embedding"},
		},
		"query": knnQuery,
	}

	//queryJSON, _ := json.Marshal(query)
	res, err := idx.osCli.Search(context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			//Body:    bytes.NewReader(queryJSON),
			Body: opensearchutil.NewJSONReader(query),
		})
	if err != nil {
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result SearchResponseGeneric[BaseRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]ResponseBaseRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		src := hit.Source
		docs = append(docs, ResponseBaseRow{
			Id:          hit.ID,
			IdCtxt:      src.IdCtxt,
			IdPje:       src.IdPje,
			HashTexto:   src.HashTexto,
			UsernameInc: src.UsernameInc,
			DtInc:       src.DtInc,
			Status:      src.Status,

			Classe:   src.Classe,
			Assunto:  src.Assunto,
			Natureza: src.Natureza,
			Tipo:     src.Tipo,
			Tema:     src.Tema,
			Fonte:    src.Fonte,
			Texto:    src.Texto,
			//TextoEmbedding: src.TextoEmbedding,
		})
	}
	return docs, nil
}

// Verificar se documento com id_ctxt e id_pje já existe
func (idx *BaseIndexType) IsExiste(idPje string) (bool, error) {
	if idPje == "" {
		return false, fmt.Errorf("parâmetros inválidos:  idPje=%q", idPje)
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
							"id_pje": idPje,
						},
					},
				},
			},
		},
	}

	res, err := idx.osCli.Search(
		ctx,
		&opensearchapi.SearchReq{
			Indices: []string{idx.indexName},
			//Body:    bytes.NewReader(queryBody),
			Body: opensearchutil.NewJSONReader(query),
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro realizar consulta by query: %v", err)
		logger.Log.Error(msg)
		return false, err
	}

	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return false, err
	}
	defer res.Inspect().Response.Body.Close()

	var result SearchResponseGeneric[BaseRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		msg := fmt.Sprintf("Erro ao decodificar resposta JSON: %v", err)
		logger.Log.Error(msg)
		return false, err
	}
	if len(result.Hits.Hits) == 0 {
		return false, nil
	}

	return true, nil
}
