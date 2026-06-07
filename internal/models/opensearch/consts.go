package opensearch

import "time"

const QUERY_MAX_SIZE = 30

/* Naturezas dos modelos. */
const (
	MODELO_NATUREZA_DESPACHO     = 1
	MODELO_NATUREZA_DECISAO      = 2
	MODELO_NATUREZA_SENTENCA     = 3
	MODELO_NATUREZA_EXPEDIENTE   = 4
	MODELO_NATUREZA_ACORDAO      = 5
	MODELO_NATUREZA_SUMULA       = 6
	MODELO_NATUREZA_CONSTITUICAO = 7
	MODELO_NATUREZA_LEI          = 8
	MODELO_NATUREZA_DECRETO      = 9
	MODELO_NATUREZA_REGULAMENTO  = 10
	MODELO_NATUREZA_DOUTRINA     = 11
	MODELO_NATUREZA_AUDIENCIA    = 12
	MODELO_NATUREZA_CERTIDOES    = 13
)

// // Item com múltiplas descrições (sinônimos)
type item struct {
	Key          int
	Descriptions string // várias denominações possíveis para o tipo
}

// Lista as descrições das naturezas(tipos) de documentos e seus sinônimos como aparecem no PJe.
var itemsPrompt = []item{
	{Key: 0, Descriptions: "selecione o documento"},
	{Key: MODELO_NATUREZA_DESPACHO, Descriptions: "Despacho"},
	{Key: MODELO_NATUREZA_DECISAO, Descriptions: "Decisão"},
	{Key: MODELO_NATUREZA_SENTENCA, Descriptions: "Sentença"},
	{Key: MODELO_NATUREZA_EXPEDIENTE, Descriptions: "Expediente"},
	{Key: MODELO_NATUREZA_ACORDAO, Descriptions: "Acordão"},
	{Key: MODELO_NATUREZA_SUMULA, Descriptions: "Súmula"},
	{Key: MODELO_NATUREZA_CONSTITUICAO, Descriptions: "Constituição"},
	{Key: MODELO_NATUREZA_LEI, Descriptions: "Lei"},
	{Key: MODELO_NATUREZA_DECRETO, Descriptions: "Decreto"},
	{Key: MODELO_NATUREZA_REGULAMENTO, Descriptions: "Regulamento"},
	{Key: MODELO_NATUREZA_DOUTRINA, Descriptions: "Doutrina"},
	{Key: MODELO_NATUREZA_AUDIENCIA, Descriptions: "Audiência"},
	{Key: MODELO_NATUREZA_CERTIDOES, Descriptions: "Cerdiões"},
}

// Mapa para consulta rápida: key -> descrição principal (primeira da lista)
var keyNaturezaModeloDescricao map[int]string

func init() {

	keyNaturezaModeloDescricao = make(map[int]string)

	for _, item := range itemsPrompt {

		keyNaturezaModeloDescricao[item.Key] = item.Descriptions

	}
}

// Retorna a descrição principal do documento pelo código
func GetNaturezaModelo(key int) string {
	if desc, ok := keyNaturezaModeloDescricao[key]; ok {
		return desc
	}
	return "Não identificado"
}

//-- INDEXEVENTOS
// ========================================
// Estruturas para o índice eventos_embedding
// ========================================

type EventosRow struct {
	IdCtxt       string    `json:"id_ctxt"`
	IdNatu       int       `json:"id_natu"`
	IdPje        string    `json:"id_pje"`
	UsernameInc  string    `json:"username_inc,omitempty"` // keyword
	DtInc        time.Time `json:"dt_inc,omitempty"`       // date
	Doc          string    `json:"doc"`
	DocJsonRaw   string    `json:"doc_json_raw"`
	DocEmbedding []float32 `json:"doc_embedding"`
}

type ResponseEventosRow struct {
	Id           string    `json:"id"`
	IdCtxt       string    `json:"id_ctxt"`
	IdNatu       int       `json:"id_natu"`
	IdPje        string    `json:"id_pje"`
	UsernameInc  string    `json:"username_inc,omitempty"` // keyword
	DtInc        time.Time `json:"dt_inc,omitempty"`       // date
	Doc          string    `json:"doc"`
	DocJsonRaw   string    `json:"doc_json_raw"`
	DocEmbedding []float32 `json:"doc_embedding"`
}
