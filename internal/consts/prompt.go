package consts

/* Naturezas dos prompts cadastrados. */
const (
	PROMPT_ANALISE_AUTUACAO   = 1
	PROMPT_ANALISE_CONTEXTO   = 2
	PROMPT_ANALISE_JULGAMENTO = 3
	PROMPT_ANALISE_DOCUMENTO  = 4
	//----------------
	PROMPT_ANALISE_ANONIMIZA = 100
	//------ prompts RAG
	PROMPT_RAG_IDENTIFICA             = 101
	PROMPT_RAG_ANALISE                = 102
	PROMPT_RAG_JULGAMENTO             = 103
	PROMPT_RAG_DECISAO                = 104
	PROMPT_RAG_DESPACHO               = 105
	PROMPT_RAG_FORMATA_SENTENCA       = 300
	PROMPT_RAG_COMPLEMENTA_JULGAMENTO = 301
	PROMPT_RAG_OUTROS                 = 999
)

// // Item com múltiplas descrições (sinônimos)
type item struct {
	Key          int
	Descriptions string // várias denominações possíveis para o tipo
}

// Lista as descrições das naturezas(tipos) de documentos e seus sinônimos como aparecem no PJe.
var itemsPrompt = []item{
	{Key: 0, Descriptions: "selecione o documento"},
	{Key: PROMPT_ANALISE_AUTUACAO, Descriptions: "Análise de Autuação"},
	{Key: PROMPT_ANALISE_CONTEXTO, Descriptions: "Análise de Contexto"},
	{Key: PROMPT_ANALISE_JULGAMENTO, Descriptions: "Análise de Julgamento"},
	{Key: PROMPT_ANALISE_DOCUMENTO, Descriptions: "Análise de Documento"},

	{Key: PROMPT_ANALISE_ANONIMIZA, Descriptions: "Análise de Anonimização"},
	//------  RAG
	{Key: PROMPT_RAG_IDENTIFICA, Descriptions: "Identifica finalidade RAG"},
	{Key: PROMPT_RAG_ANALISE, Descriptions: "Anaĺise Jurídica(RAG)"},
	{Key: PROMPT_RAG_JULGAMENTO, Descriptions: "Análise Julgamento(RAG)"},
	{Key: PROMPT_RAG_DECISAO, Descriptions: "Análise Decisão(RAG)"},
	{Key: PROMPT_RAG_DESPACHO, Descriptions: "Análise Despacho(RAG)"},
	{
		Key:          PROMPT_RAG_FORMATA_SENTENCA,
		Descriptions: "Formatar sentença para RAG",
	},
	{
		Key:          PROMPT_RAG_COMPLEMENTA_JULGAMENTO,
		Descriptions: "Complementa Julgamento(RAG)",
	},
	{
		Key:          PROMPT_RAG_OUTROS,
		Descriptions: "Outras naturezas(RAG)",
	},
}

// Mapa para consulta rápida: key -> descrição principal (primeira da lista)
var keyPromptDescricao map[int]string

func init() {

	keyPromptDescricao = make(map[int]string)

	for _, item := range itemsPrompt {

		keyPromptDescricao[item.Key] = item.Descriptions

	}
}

// Retorna a descrição principal do documento pelo código
func GetPromptNatureza(key int) string {
	if desc, ok := keyParaDescricao[key]; ok {
		return desc
	}
	return "Não identificado"
}
