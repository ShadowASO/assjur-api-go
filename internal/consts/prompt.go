package consts

import ()

/* Naturezas dos prompts cadastrados. */
const (
	PROMPT_ANALISE_AUTUACAO   = 1
	PROMPT_ANALISE_CONTEXTO   = 2
	PROMPT_ANALISE_JULGAMENTO = 3
	PROMPT_ANALISE_DOCUMENTO  = 4
	//----------------
	PROMPT_ANALISE_ANONIMIZA = 100
	PROMPT_ANALISE_FORMATA   = 101
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
	{Key: PROMPT_ANALISE_FORMATA, Descriptions: "Análise de Formatação"},
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
