package consts

import ()

/* Naturezas dos prompts cadastrados. */
const (
	RESPOSTA_RAG_CHAT     = 1
	RESPOSTA_RAG_ANALISE  = 2
	RESPOSTA_RAG_SENTENCA = 3
)

// // Item com múltiplas descrições (sinônimos)
type itemResposta struct {
	TipoResp int    `json:"tipo_resp"`
	Texto    string `json:"texto"`
}

// Lista as descrições das naturezas(tipos) de documentos e seus sinônimos como aparecem no PJe.
var itemsResposta = []itemResposta{
	{TipoResp: RESPOSTA_RAG_CHAT, Texto: "Chat"},
	{TipoResp: RESPOSTA_RAG_ANALISE, Texto: "Análise jurídica"},
	{TipoResp: RESPOSTA_RAG_SENTENCA, Texto: "Sentença"},
}

// Mapa para consulta rápida: key -> descrição principal (primeira da lista)
var keyRespostaDescricao map[int]string

func init() {

	keyRespostaDescricao = make(map[int]string)

	for _, item := range itemsResposta {

		keyRespostaDescricao[item.TipoResp] = item.Texto

	}
}

// Retorna a descrição principal do documento pelo código
func GetRespostaDescricao(key int) string {
	if desc, ok := keyRespostaDescricao[key]; ok {
		return desc
	}
	return "Não identificado"
}
