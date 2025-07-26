package consts

import (
	"regexp"
	"strings"
)

/* Naturezas reconhecidas de documentos que compõem os autos processuais. */
const (
	NATU_DOC_INICIAL         = 1
	NATU_DOC_CONTESTACAO     = 2
	NATU_DOC_REPLICA         = 3
	NATU_DOC_DESPACHO        = 4
	NATU_DOC_PETICAO         = 5
	NATU_DOC_DECISAO         = 6
	NATU_DOC_SENTENCA        = 7
	NATU_DOC_EMBARGOS        = 8
	NATU_DOC_APELACAO        = 9
	NATU_DOC_CONTRA_RAZOES   = 10
	NATU_DOC_PROCURACAO      = 11
	NATU_DOC_ROL_TESTEMUNHAS = 12
	NATU_DOC_CONTRATO        = 13
	NATU_DOC_LAUDO_PERICIAL  = 14
	NATU_DOC_TERMO_AUDIENCIA = 15
	NATU_DOC_PARECER_MP      = 16
	NATU_DOC_AUTOS           = 1000
	NATU_DOC_OUTROS          = 1001
	NATU_DOC_CERTIDOES       = 1002
	NATU_DOC_MOVIMENTACAO    = 1003
	NATU_DOC_ANALISE_IA      = 2000
)

// Item com múltiplas descrições (sinônimos)
type Item struct {
	Key          int
	Descriptions []string // várias denominações possíveis para o tipo
}

// Lista as descrições das naturezas(tipos) de documentos e seus sinônimos como aparecem no PJe.
var itemsDocumento = []Item{
	{Key: 0, Descriptions: []string{"selecione o documento"}},
	{Key: NATU_DOC_INICIAL, Descriptions: []string{"petição inicial", "peticao inicial"}},
	{Key: NATU_DOC_CONTESTACAO, Descriptions: []string{"contestação", "contestacao"}},
	{Key: NATU_DOC_REPLICA, Descriptions: []string{"réplica", "replica"}},
	{Key: NATU_DOC_DESPACHO, Descriptions: []string{"despacho", "despacho ordinatório", "despacho ordinatorio"}},
	{Key: NATU_DOC_PETICAO, Descriptions: []string{"petição", "peticao"}},
	{Key: NATU_DOC_DECISAO, Descriptions: []string{"decisão", "decisao", "interlocutória", "interlocutoria"}},
	{Key: NATU_DOC_SENTENCA, Descriptions: []string{"sentença", "sentenca"}},
	{Key: NATU_DOC_EMBARGOS, Descriptions: []string{"embargos de declaração", "embargos de declaracao"}},
	{Key: NATU_DOC_CONTRA_RAZOES, Descriptions: []string{"contra-razões", "contrarazoes"}},
	{Key: NATU_DOC_APELACAO, Descriptions: []string{"recurso de apelação", "recurso de apelacao", "apelação", "apelacao"}},
	{Key: NATU_DOC_PROCURACAO, Descriptions: []string{"procuração", "procuracao"}},
	{Key: NATU_DOC_ROL_TESTEMUNHAS, Descriptions: []string{"rol de testemunhas"}},
	{Key: NATU_DOC_CONTRATO, Descriptions: []string{"contrato"}},
	{Key: NATU_DOC_LAUDO_PERICIAL, Descriptions: []string{"laudo pericial"}},
	{Key: NATU_DOC_TERMO_AUDIENCIA, Descriptions: []string{"termo de audiência", "termo de audiencia"}},
	{Key: NATU_DOC_PARECER_MP, Descriptions: []string{"manifestação do ministério público", "manifestacao do ministerio publico"}},
	{Key: NATU_DOC_AUTOS, Descriptions: []string{"autos processuais", "autos"}},
	{Key: NATU_DOC_OUTROS, Descriptions: []string{"outros documentos"}},
	{Key: NATU_DOC_CERTIDOES, Descriptions: []string{"certidões", "certidoes"}},
	{Key: NATU_DOC_MOVIMENTACAO, Descriptions: []string{"movimentação", "movimentacao", "processo"}},
	{Key: NATU_DOC_ANALISE_IA, Descriptions: []string{"análise pela ia", "analise pela ia"}},
}

// Mapa para consulta rápida: descrição -> key
var descricaoParaKey map[string]int

// Mapa para consulta rápida: key -> descrição principal (primeira da lista)
var keyParaDescricao map[int]string

func init() {
	descricaoParaKey = make(map[string]int)
	keyParaDescricao = make(map[int]string)

	for _, item := range itemsDocumento {
		// Usa a primeira descrição como principal para key->descrição
		if len(item.Descriptions) > 0 {
			keyParaDescricao[item.Key] = item.Descriptions[0]
		} else {
			keyParaDescricao[item.Key] = ""
		}
		//Salva todas as subdescrições no mapa[desc]=key
		for _, desc := range item.Descriptions {
			// normaliza para lowercase e trim
			descNorm := strings.ToLower(strings.TrimSpace(desc))
			descricaoParaKey[descNorm] = item.Key
		}
	}
}

// Remove complemento entre parênteses e espaço antes deles
func removeComplemento(texto string) string {
	// Regex para remover " (qualquer coisa)" no final da string
	re := regexp.MustCompile(`\s*\(.*\)$`)
	return re.ReplaceAllString(texto, "")
}

// Retorna a descrição principal do documento pelo código
func GetNaturezaDocumento(key int) string {
	if desc, ok := keyParaDescricao[key]; ok {
		return desc
	}
	return "Não identificado"
}

// Retorna o código da natureza a partir da sua descrição
func GetCodigoNatureza(nmNatureza string) int {
	tipoLimpo := removeComplemento(nmNatureza)
	tipoNorm := strings.ToLower(strings.TrimSpace(tipoLimpo))
	return descricaoParaKey[tipoNorm]
}
