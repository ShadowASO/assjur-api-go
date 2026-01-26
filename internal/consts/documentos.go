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
	NATU_DOC_CERTIDAO        = 17

	NATU_DOC_AUTOS        = 1000
	NATU_DOC_OUTROS       = 1001
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
	{Key: NATU_DOC_INICIAL, Descriptions: []string{
		"Petição Inicial",
		"Emenda à Inicial",
	}},
	{Key: NATU_DOC_CONTESTACAO, Descriptions: []string{"Contestação"}},
	{Key: NATU_DOC_REPLICA, Descriptions: []string{"Réplica"}},
	{Key: NATU_DOC_DESPACHO, Descriptions: []string{
		"Despacho",
		"Despacho Ordinatório",
	}},

	{Key: NATU_DOC_PETICAO, Descriptions: []string{
		"Petição",
		"Alegações",
		"Alegações Finais",
		"Memoriais",
		"Manifestação",
		"Manifestação da Defensoria Pública",
		"Manifestação do Ministério Público",
		"Exceção de Pré-Executividade",
		"Informações - Agravo de Instrumento",

		"Pedido",

		"Informações",
		"Petição de Habilitação",
		"Petição intercorrente",
		"Petição intermediária",
		"Petição Requerendo",
		"Petição Simples de Terceiro Interessado",
		"Proposta de acordo",
		"Razões",
		"Reconvenção",
	}},

	{Key: NATU_DOC_DECISAO, Descriptions: []string{
		"Decisão",
		"interlocutória",
	}},
	{Key: NATU_DOC_SENTENCA, Descriptions: []string{"Sentença"}},
	{Key: NATU_DOC_EMBARGOS, Descriptions: []string{"Embargos de Declaração"}},
	{Key: NATU_DOC_CONTRA_RAZOES, Descriptions: []string{
		"Contra-razões",
		"Contrarazões",
	}},
	{Key: NATU_DOC_APELACAO, Descriptions: []string{
		"Recurso de Apelação",
		"Apelação",
		"Recurso",
	}},
	{Key: NATU_DOC_PROCURACAO, Descriptions: []string{"Procuração"}},
	{Key: NATU_DOC_ROL_TESTEMUNHAS, Descriptions: []string{"Rol de Testemunhas"}},
	{Key: NATU_DOC_CONTRATO, Descriptions: []string{"Contrato"}},
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
		"Ata de Julgamento",
		"Ata de Audiência de Conciliacão",
		"Ata de Audiência de Instrucão",
		"Ata de Audiência de Instrucão e Julgamento",
		"Ata de Audiência de Julgamento",
		"Ata de Audiência de Mediacão",
		"Termo de Audiencia",
		"Termo de Audiência - com acordo",
		"Termo de Audiência - sem acordo",
	},
	},

	{Key: NATU_DOC_PARECER_MP, Descriptions: []string{
		"Manifestação do MP",
		"Manifestacão do Ministério Público",
	}},
	{Key: NATU_DOC_AUTOS, Descriptions: []string{"Autos Processuais", "autos"}},
	//{Key: NATU_DOC_OUTROS, Descriptions: []string{"Outros Documentos"}},
	{Key: NATU_DOC_OUTROS, Descriptions: []string{"Outros Documentos"}},
	{Key: NATU_DOC_CERTIDAO, Descriptions: []string{"Certidões", "Certidão"}},
	{Key: NATU_DOC_MOVIMENTACAO, Descriptions: []string{"Movimentação", "Processo"}},

	// IA
	{Key: NATU_DOC_IA_PROMPT, Descriptions: []string{"Prompt de ia"}},
	{Key: NATU_DOC_IA_PREANALISE, Descriptions: []string{"Pré-análise jurídica"}},
	{Key: NATU_DOC_IA_ANALISE, Descriptions: []string{"Análise Jurídica"}},
	{Key: NATU_DOC_IA_SENTENCA, Descriptions: []string{"Minuta de Sentença"}},
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
// func GetCodigoNatureza(nmTipo string) int {
// 	tipoLimpo := removeComplemento(nmTipo)
// 	tipoNorm := normalizeText(tipoLimpo)

// 	if key, ok := descricaoParaKey[tipoNorm]; ok {
// 		return key
// 	}
// 	return -1 // indica “não encontrado”
// }

func GetCodigoNatureza(nmTipo string) int {
	code, _ := ClassificarDocumento(nmTipo)
	return code
}

// ClassificarDocumento analisa o nome do tipo documental e retorna:
// - o código da natureza (conhecido ou -1 se não identificado);
// - um booleano indicando se o documento é válido (não excluído).
func ClassificarDocumento(nmTipo string) (int, bool) {
	tipoLimpo := removeComplemento(nmTipo)
	tipoNorm := normalizeText(tipoLimpo)

	// 1️⃣ Exclusões: documentos que não devem ser aceitos
	// if strings.Contains(tipoNorm, "documentos diversos") ||
	// 	strings.Contains(tipoNorm, "documento diverso") ||
	// 	strings.Contains(tipoNorm, "certidao") {
	// 	return -1, false // tipo excluído
	// }

	// 2️⃣ Verificação de tipos conhecidos
	if key, ok := descricaoParaKey[tipoNorm]; ok {
		return key, true // tipo identificado e válido
	}

	// 3️⃣ Caso não se enquadre em nenhum tipo conhecido
	//return -1, true // "não identificado", mas válido
	return NATU_DOC_OUTROS, true
}
