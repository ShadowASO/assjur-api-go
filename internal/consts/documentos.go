package consts

import (
	"regexp"
	"strings"
	"unicode"
)

// ============================================================================
// Códigos das naturezas de documentos processuais
// ============================================================================
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

	NATU_DOC_AUTOS        = 1000
	NATU_DOC_OUTROS       = 1001
	NATU_DOC_CERTIDOES    = 1002
	NATU_DOC_MOVIMENTACAO = 1003

	NATU_DOC_IA_PROMPT     = 100
	NATU_DOC_IA_PREANALISE = 101
	NATU_DOC_IA_ANALISE    = 102
	NATU_DOC_IA_SENTENCA   = 103
)

// ============================================================================
// Estruturas de apoio
// ============================================================================
type Item struct {
	Key          int
	Descriptions []string // denominações possíveis do tipo documental
}

// ============================================================================
// Lista de naturezas reconhecidas
// ============================================================================
var itemsDocumento = []Item{
	{Key: 0, Descriptions: []string{"selecione o documento"}},
	{Key: NATU_DOC_INICIAL, Descriptions: []string{"Petição Inicial", "peticao inicial"}},
	{Key: NATU_DOC_CONTESTACAO, Descriptions: []string{"Contestação", "contestacao"}},
	{Key: NATU_DOC_REPLICA, Descriptions: []string{"Réplica", "replica"}},
	{Key: NATU_DOC_DESPACHO, Descriptions: []string{"Despacho", "despacho ordinatório", "despacho ordinatorio"}},

	{Key: NATU_DOC_PETICAO, Descriptions: []string{"Petição", "alegações", "pedido", "proposta de acordo", "razões", "informações"}},

	{Key: NATU_DOC_DECISAO, Descriptions: []string{"Decisão", "decisao", "interlocutória", "interlocutoria"}},
	{Key: NATU_DOC_SENTENCA, Descriptions: []string{"Sentença", "sentenca"}},
	{Key: NATU_DOC_EMBARGOS, Descriptions: []string{"Embargos de Declaração", "embargos de declaracao"}},
	{Key: NATU_DOC_CONTRA_RAZOES, Descriptions: []string{"Contra-razões", "contrarazoes"}},
	{Key: NATU_DOC_APELACAO, Descriptions: []string{"Recurso de Apelação", "recurso de apelacao", "apelação", "apelacao", "recurso"}},
	{Key: NATU_DOC_PROCURACAO, Descriptions: []string{"Procuração", "procuracao"}},
	{Key: NATU_DOC_ROL_TESTEMUNHAS, Descriptions: []string{"rol de testemunhas"}},
	{Key: NATU_DOC_CONTRATO, Descriptions: []string{"contrato"}},
	{Key: NATU_DOC_LAUDO_PERICIAL, Descriptions: []string{

		"Laudo",
		"Laudo de Perícia",
		"Laudo Perícia Médica",
		"Laudo Médico",
		"Laudo Psicológico",
		"Perícia",
	},
	},

	{Key: NATU_DOC_TERMO_AUDIENCIA, Descriptions: []string{

		"Ata de Audiência",
		"ata de julgamento",
		"ata de audiência de conciliacao",
		"ata de audiência de instrucao",
		"ata de audiência de instrucao e julgamento",
		"ata de audiência de julgamento",
		"ata de audiência de mediacao",
		"termo de audiencia",
		"termo de audiencia - com acordo",
		"termo de audiencia - sem acordo",
	},
	},

	{Key: NATU_DOC_PARECER_MP, Descriptions: []string{"manifestação do ministério público", "manifestacao do ministerio publico"}},
	{Key: NATU_DOC_AUTOS, Descriptions: []string{"autos processuais", "autos"}},
	{Key: NATU_DOC_OUTROS, Descriptions: []string{"outros documentos"}},
	{Key: NATU_DOC_CERTIDOES, Descriptions: []string{"certidões", "certidoes"}},
	{Key: NATU_DOC_MOVIMENTACAO, Descriptions: []string{"movimentação", "movimentacao", "processo"}},

	// IA
	{Key: NATU_DOC_IA_PROMPT, Descriptions: []string{"prompt de ia"}},
	{Key: NATU_DOC_IA_PREANALISE, Descriptions: []string{"pré-análise jurídica", "pre-analise juridica"}},
	{Key: NATU_DOC_IA_ANALISE, Descriptions: []string{"análise jurídica", "analise juridica"}},
	{Key: NATU_DOC_IA_SENTENCA, Descriptions: []string{"minuta de sentença", "minuta de sentenca"}},
}

// ============================================================================
// Mapas de acesso rápido
// ============================================================================
var (
	descricaoParaKey  map[string]int
	keyParaDescricao  map[int]string
	regexComplementos = regexp.MustCompile(`\s*\([^()]*\)$`)
)

// ============================================================================
// Inicialização
// ============================================================================
func init() {
	descricaoParaKey = make(map[string]int)
	keyParaDescricao = make(map[int]string)

	for _, item := range itemsDocumento {
		//Copio todos os registros de itensDocumento para keyParaDescricao
		if len(item.Descriptions) > 0 {
			keyParaDescricao[item.Key] = item.Descriptions[0]
		}

		for _, desc := range item.Descriptions {
			descNorm := normalizeText(desc)
			descricaoParaKey[descNorm] = item.Key
		}
	}
}

// ============================================================================
// Funções utilitárias
// ============================================================================

// normalizeText converte para minúsculas, remove acentos e espaços
func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Map(func(r rune) rune {
		switch r {
		case 'á', 'à', 'ã', 'â', 'ä':
			return 'a'
		case 'é', 'è', 'ê', 'ë':
			return 'e'
		case 'í', 'ì', 'î', 'ï':
			return 'i'
		case 'ó', 'ò', 'õ', 'ô', 'ö':
			return 'o'
		case 'ú', 'ù', 'û', 'ü':
			return 'u'
		case 'ç':
			return 'c'
		default:
			return unicode.ToLower(r)
		}
	}, s)
}

// removeComplemento remove o texto entre parênteses no final da string
func removeComplemento(texto string) string {
	return regexComplementos.ReplaceAllString(texto, "")
}

// GetNaturezaDocumento retorna a descrição principal da natureza pelo código
func GetNaturezaDocumento(key int) string {
	if desc, ok := keyParaDescricao[key]; ok {
		return desc
	}
	return "não identificado"
}

// GetCodigoNatureza retorna o código da natureza a partir da descrição
func GetCodigoNatureza(nmTipo string) int {
	tipoLimpo := removeComplemento(nmTipo)
	tipoNorm := normalizeText(tipoLimpo)

	if key, ok := descricaoParaKey[tipoNorm]; ok {
		return key
	}
	return -1 // indica “não encontrado”
}
