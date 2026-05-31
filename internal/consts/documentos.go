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
	Descriptions []string
}

type regraDocumento struct {
	Key      int
	Termos   []string
	Exclui   []string
	MinScore int
}

// ============================================================================
// Lista de naturezas reconhecidas
// ============================================================================
var itemsDocumento = []Item{
	{Key: 0, Descriptions: []string{"selecione o documento"}},

	{Key: NATU_DOC_INICIAL, Descriptions: []string{
		"Petição Inicial",
		"Inicial",
		"Emenda à Inicial",
		"Emenda da Inicial",
	}},

	{Key: NATU_DOC_CONTESTACAO, Descriptions: []string{
		"Contestação",
	}},

	{Key: NATU_DOC_REPLICA, Descriptions: []string{
		"Réplica",
	}},

	{Key: NATU_DOC_DESPACHO, Descriptions: []string{
		"Despacho",
		"Despacho Ordinatório",
		"Ato Ordinatório",
		"Atos Ordinatórios",
	}},

	{Key: NATU_DOC_PETICAO, Descriptions: []string{
		"Petição",
		"Petição Intermediária",
		"Petição Intercorrente",
		"Petição Requerendo",
		"Petição de Habilitação",
		"Petição de Citação",
		"Petição Simples de Terceiro Interessado",
		"Petições Intermediárias Diversas",
		"Alegações",
		"Alegações Finais",
		"Memoriais",
		"Manifestação",
		"Manifestação da Defensoria Pública",
		"Pedido",
		"Pedido de Juntada de Documento",
		"Proposta de acordo",
		"Razões",
		"Reconvenção",
		"Exceção de Pré-Executividade",
		"Informações",
		"Informações - Agravo de Instrumento",
	}},

	{Key: NATU_DOC_DECISAO, Descriptions: []string{
		"Decisão",
		"Decisão Interlocutória",
		"Decisões Interlocutórias",
		"Interlocutória",
	}},

	{Key: NATU_DOC_SENTENCA, Descriptions: []string{
		"Sentença",
	}},

	{Key: NATU_DOC_EMBARGOS, Descriptions: []string{
		"Embargos de Declaração",
		"Embargos Declaratórios",
	}},

	{Key: NATU_DOC_APELACAO, Descriptions: []string{
		"Recurso de Apelação",
		"Apelação",
		"Recurso",
	}},

	{Key: NATU_DOC_CONTRA_RAZOES, Descriptions: []string{
		"Contra-razões",
		"Contrarrazões",
		"Contrarrazões de Apelação",
		"Contrarrazões ao Recurso",
	}},

	{Key: NATU_DOC_PROCURACAO, Descriptions: []string{
		"Procuração",
		"Procuração/Substabelecimento",
		"Substabelecimento",
	}},

	{Key: NATU_DOC_ROL_TESTEMUNHAS, Descriptions: []string{
		"Rol de Testemunhas",
	}},

	{Key: NATU_DOC_CONTRATO, Descriptions: []string{
		"Contrato",
	}},

	{Key: NATU_DOC_LAUDO_PERICIAL, Descriptions: []string{
		"Laudo",
		"Laudo de Perícia",
		"Laudo Pericial",
		"Laudo Perícia Médica",
		"Laudo Médico",
		"Laudo Psicológico",
		"Perícia",
		"Perícia Médica",
	}},

	{Key: NATU_DOC_TERMO_AUDIENCIA, Descriptions: []string{
		"Ata de Audiência",
		"Ata de Julgamento",
		"Ata de Audiência de Conciliação",
		"Ata de Audiência de Instrução",
		"Ata de Audiência de Instrução e Julgamento",
		"Ata de Audiência de Julgamento",
		"Ata de Audiência de Mediação",
		"Termo de Audiência",
		"Termo de Audiência - com acordo",
		"Termo de Audiência - sem acordo",
	}},

	{Key: NATU_DOC_PARECER_MP, Descriptions: []string{
		"Parecer do Ministério Público",
		"Manifestação do Ministério Público",
		"Manifestação do MP",
	}},

	{Key: NATU_DOC_CERTIDAO, Descriptions: []string{
		"Certidão",
		"Certidões",
		"Certidão de Publicação",
		"Certidão de Decurso de Prazo",
		"Certidão da Secretaria",
		"Certidões da Secretaria",
	}},

	{Key: NATU_DOC_AUTOS, Descriptions: []string{
		"Autos Processuais",
		"Autos",
	}},

	{Key: NATU_DOC_MOVIMENTACAO, Descriptions: []string{
		"Movimentação",
		"Anexo de Movimentação",
		"Processo",
	}},

	{Key: NATU_DOC_OUTROS, Descriptions: []string{
		"Outros Documentos",
		"Documentos Diversos",
		"Documento Diverso",
		"Documentação",
		"Documento de Comprovação",
		"Documento Pessoal",
		"Relatório Informativo",
		"Carta",
		"Cartas",
	}},

	// IA
	{Key: NATU_DOC_IA_PROMPT, Descriptions: []string{"Prompt de ia"}},
	{Key: NATU_DOC_IA_PREANALISE, Descriptions: []string{"Pré-análise jurídica"}},
	{Key: NATU_DOC_IA_ANALISE, Descriptions: []string{"Análise Jurídica"}},
	{Key: NATU_DOC_IA_SENTENCA, Descriptions: []string{"Minuta de Sentença"}},
}

// ============================================================================
// Mapas e regex de apoio
// ============================================================================
var (
	descricaoParaKey map[string]int
	keyParaDescricao map[int]string

	regexComplementosFinais = regexp.MustCompile(`\s*\([^()]*\)\s*$`)
	regexSeparadores        = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	regexEspacos            = regexp.MustCompile(`\s+`)
)

// ============================================================================
// Regras fortes de classificação
//
// A ordem importa:
// 1. tipos mais específicos primeiro;
// 2. tipos recursais antes de petição genérica;
// 3. "contrarrazões" antes de "razões";
// 4. "embargos de declaração" antes de "recurso".
// ============================================================================
var regrasDocumento = []regraDocumento{
	{
		Key: NATU_DOC_EMBARGOS,
		Termos: []string{
			"embargos de declaracao",
			"embargo de declaracao",
			"embargos declaratorios",
			"embargo declaratorio",
		},
	},
	{
		Key: NATU_DOC_CONTRA_RAZOES,
		Termos: []string{
			"contrarrazoes",
			"contra razoes",
			"contra razao",
			"contraminuta",
		},
	},
	{
		Key: NATU_DOC_APELACAO,
		Termos: []string{
			"recurso de apelacao",
			"apelacao",
			"apelante",
			"apelado",
			"recurso inominado",
			"recurso",
		},
		Exclui: []string{
			"embargos de declaracao",
			"agravo de instrumento",
			"contrarrazoes",
			"contra razoes",
		},
	},
	{
		Key: NATU_DOC_INICIAL,
		Termos: []string{
			"peticao inicial",
			"inicial",
			"emenda a inicial",
			"emenda da inicial",
		},
	},
	{
		Key: NATU_DOC_CONTESTACAO,
		Termos: []string{
			"contestacao",
		},
	},
	{
		Key: NATU_DOC_REPLICA,
		Termos: []string{
			"replica",
		},
	},
	{
		Key: NATU_DOC_SENTENCA,
		Termos: []string{
			"sentenca",
		},
	},
	{
		Key: NATU_DOC_DECISAO,
		Termos: []string{
			"decisao",
			"decisoes",
			"interlocutoria",
			"interlocutorias",
		},
	},
	{
		Key: NATU_DOC_DESPACHO,
		Termos: []string{
			"despacho",
			"despacho ordinatorio",
			"ato ordinatorio",
			"atos ordinatorios",
		},
	},
	{
		Key: NATU_DOC_CERTIDAO,
		Termos: []string{
			"certidao",
			"certidoes",
		},
	},
	{
		Key: NATU_DOC_TERMO_AUDIENCIA,
		Termos: []string{
			"ata de audiencia",
			"termo de audiencia",
			"audiencia",
			"conciliacao",
			"instrucao e julgamento",
		},
	},
	{
		Key: NATU_DOC_LAUDO_PERICIAL,
		Termos: []string{
			"laudo pericial",
			"laudo de pericia",
			"laudo",
			"pericia",
			"pericial",
		},
	},
	{
		Key: NATU_DOC_PROCURACAO,
		Termos: []string{
			"procuracao",
			"substabelecimento",
		},
	},
	{
		Key: NATU_DOC_ROL_TESTEMUNHAS,
		Termos: []string{
			"rol de testemunhas",
			"testemunhas",
		},
	},
	{
		Key: NATU_DOC_CONTRATO,
		Termos: []string{
			"contrato",
		},
	},
	{
		Key: NATU_DOC_PARECER_MP,
		Termos: []string{
			"parecer do ministerio publico",
			"manifestacao do ministerio publico",
			"manifestacao do mp",
			"parecer ministerial",
		},
	},
	{
		Key: NATU_DOC_PETICAO,
		Termos: []string{
			"peticao",
			"peticoes",
			"pedido",
			"manifestacao",
			"alegacoes",
			"memoriais",
			"reconvecao",
			"reconvencao",
			"excecao de pre executividade",
			"proposta de acordo",
			"informacoes",
		},
		Exclui: []string{
			"peticao inicial",
			"embargos de declaracao",
			"recurso de apelacao",
			"contrarrazoes",
			"contra razoes",
		},
	},
	{
		Key: NATU_DOC_MOVIMENTACAO,
		Termos: []string{
			"anexo de movimentacao",
			"movimentacao",
		},
	},
}

// ============================================================================
// Inicialização
// ============================================================================
func init() {
	descricaoParaKey = make(map[string]int)
	keyParaDescricao = make(map[int]string)

	for _, item := range itemsDocumento {
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

// normalizeText converte para minúsculas, remove acentos, normaliza separadores
// e reduz múltiplos espaços.
func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))

	s = strings.Map(func(r rune) rune {
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

	s = regexSeparadores.ReplaceAllString(s, " ")
	s = regexEspacos.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}

// removeComplemento remove apenas complemento entre parênteses no final.
// Ex.: "Certidão (Outras)" -> "Certidão".
func removeComplemento(texto string) string {
	return strings.TrimSpace(regexComplementosFinais.ReplaceAllString(texto, ""))
}

func contemTermo(textoNorm, termoNorm string) bool {
	textoNorm = " " + textoNorm + " "
	termoNorm = " " + normalizeText(termoNorm) + " "
	return strings.Contains(textoNorm, termoNorm)
}

func contemAlgum(textoNorm string, termos []string) bool {
	for _, termo := range termos {
		if contemTermo(textoNorm, termo) {
			return true
		}
	}
	return false
}

func regraBloqueada(textoNorm string, exclui []string) bool {
	return len(exclui) > 0 && contemAlgum(textoNorm, exclui)
}

func classificarPorRegras(textoNorm string) (int, bool) {
	if textoNorm == "" {
		return NATU_DOC_OUTROS, true
	}

	for _, regra := range regrasDocumento {
		if regraBloqueada(textoNorm, regra.Exclui) {
			continue
		}

		if contemAlgum(textoNorm, regra.Termos) {
			return regra.Key, true
		}
	}

	return 0, false
}

// ============================================================================
// API pública
// ============================================================================

// GetNaturezaDocumento retorna a descrição principal da natureza pelo código.
func GetNaturezaDocumento(key int) string {
	if desc, ok := keyParaDescricao[key]; ok {
		return desc
	}
	return "não identificado"
}

// GetCodigoNatureza mantém compatibilidade com o código atual.
// Quando só houver o tipo, classifica apenas pelo tipo.
func GetCodigoNatureza(nmTipo string) int {
	code, _ := ClassificarDocumento(nmTipo)
	return code
}

// ClassificarDocumento mantém compatibilidade com o código atual.
// Preferir ClassificarDocumentoCampos quando houver também o nome do documento.
func ClassificarDocumento(nmTipo string) (int, bool) {
	return ClassificarDocumentoCampos("", nmTipo)
}

// ClassificarDocumentoCampos analisa conjuntamente o nome do documento e o tipo
// informado pelo sistema processual.
//
// Parâmetros:
// - nmDocumento: coluna "Documento", geralmente mais descritiva.
// - nmTipo: coluna "Tipo", às vezes genérica ou incorreta.
//
// Retorno:
// - código da natureza;
// - bool indicando se o documento é válido para processamento.
func ClassificarDocumentoCampos(nmDocumento, nmTipo string) (int, bool) {
	docLimpo := removeComplemento(nmDocumento)
	tipoLimpo := removeComplemento(nmTipo)

	docNorm := normalizeText(docLimpo)
	tipoNorm := normalizeText(tipoLimpo)

	// 1. Casos vazios.
	if docNorm == "" && tipoNorm == "" {
		return NATU_DOC_OUTROS, false
	}

	// 2. Primeiro tenta igualdade exata pelo tipo.
	// Isso preserva o comportamento anterior para tipos bem cadastrados.
	if key, ok := descricaoParaKey[tipoNorm]; ok && key != 0 {
		// Se o nome do documento indicar algo mais específico que o tipo genérico,
		// a regra por conteúdo abaixo poderá corrigir.
		if !tipoGenerico(tipoNorm) {
			return key, true
		}
	}

	// 3. Depois tenta igualdade exata pelo nome do documento.
	if key, ok := descricaoParaKey[docNorm]; ok && key != 0 {
		return key, true
	}

	// 4. Classificação por palavras fortes, analisando o conjunto.
	// O documento vem antes do tipo para permitir corrigir casos como:
	// Documento: "APELAÇÃO - ..."
	// Tipo: "Recurso"
	combinado := strings.TrimSpace(docNorm + " " + tipoNorm)

	if key, ok := classificarPorRegras(combinado); ok {
		return key, true
	}

	// 5. Se o tipo era exato, mas genérico, usa esse tipo.
	// Ex.: "Documentos Diversos" sem nenhum sinal no nome.
	if key, ok := descricaoParaKey[tipoNorm]; ok && key != 0 {
		return key, true
	}

	// 6. Fallback: documento válido, mas natureza não identificada.
	return NATU_DOC_OUTROS, true
}

func tipoGenerico(tipoNorm string) bool {
	switch tipoNorm {
	case "",
		"outros documentos",
		"documentos diversos",
		"documento diverso",
		"documentacao",
		"documento de comprovacao",
		"documento pessoal",
		"relatorio informativo",
		"carta",
		"cartas",
		"pedido outros",
		"peticao outras":
		return true
	default:
		return false
	}
}
