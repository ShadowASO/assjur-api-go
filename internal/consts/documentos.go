package consts

import (
	"strings"
)

const (
	NATU_DOC_INICIAL         = 1
	NATU_DOC_CONTESTACAO     = 2
	NATU_DOC_REPLICA         = 3
	NATU_DOC_DESP_INI        = 4
	NATU_DOC_DESP_ORD        = 5
	NATU_DOC_PETICAO         = 6
	NATU_DOC_DECISAO         = 7
	NATU_DOC_SENTENCA        = 8
	NATU_DOC_EMBARGOS        = 9
	NATU_DOC_APELACAO        = 10
	NATU_DOC_CONTRA_RAZ      = 11
	NATU_DOC_PROCURACAO      = 12
	NATU_DOC_ROL_TESTEMUNHAS = 13
	NATU_DOC_CONTRATO        = 14
	NATU_DOC_LAUDO_PERICIA   = 15
	NATU_DOC_TERMO_AUDIENCIA = 16
	NATU_DOC_PARECER_MP      = 17
	NATU_DOC_AUTOS           = 1000
	NATU_DOC_OUTROS          = 1001
	NATU_DOC_CERTIDOES       = 1002
	NATU_DOC_MOVIMENTACAO    = 1003
	NATU_DOC_ANALISE_IA      = 2000
)

// Definição do tipo Item
type Item struct {
	Key         int
	Description string
}

var itemsDocumento = []Item{
	{Key: 0, Description: "Selecione o documento"},
	{Key: NATU_DOC_INICIAL, Description: "Petição inicial"},
	{Key: NATU_DOC_CONTESTACAO, Description: "Contestação"},
	{Key: NATU_DOC_REPLICA, Description: "Réplica"},
	{Key: NATU_DOC_DESP_INI, Description: "Despacho inicial"},
	{Key: NATU_DOC_DESP_ORD, Description: "Despacho ordinatório"},
	{Key: NATU_DOC_PETICAO, Description: "Petição diversa"},
	{Key: NATU_DOC_DECISAO, Description: "Decisão interlocutória"},
	{Key: NATU_DOC_SENTENCA, Description: "Sentença"},
	{Key: NATU_DOC_EMBARGOS, Description: "Embargos de declaração"},
	{Key: NATU_DOC_CONTRA_RAZ, Description: "Contra-razões"},
	{Key: NATU_DOC_APELACAO, Description: "Recurso de Apelação"},
	{Key: NATU_DOC_PROCURACAO, Description: "Procuração"},
	{Key: NATU_DOC_ROL_TESTEMUNHAS, Description: "Rol de Testemunhas"},
	{Key: NATU_DOC_CONTRATO, Description: "Contrato"},
	{Key: NATU_DOC_LAUDO_PERICIA, Description: "Laudo Pericial"},
	{Key: NATU_DOC_TERMO_AUDIENCIA, Description: "Termo de Audiência"},
	{Key: NATU_DOC_PARECER_MP, Description: "Manifestação do Ministério Público"},
	{Key: NATU_DOC_AUTOS, Description: "Autos Processuais"},
	{Key: NATU_DOC_OUTROS, Description: "Outros documentos"},
	{Key: NATU_DOC_CERTIDOES, Description: "Certidões"},
	{Key: NATU_DOC_MOVIMENTACAO, Description: "Movimentação/processo"},
	{Key: NATU_DOC_ANALISE_IA, Description: "Análise pela IA"},
}

// Map para busca rápida por key
var itemsDocumentoMap map[int]string

// Inicializa o mapa (executar uma vez no init)
func init() {
	itemsDocumentoMap = make(map[int]string, len(itemsDocumento))
	for _, item := range itemsDocumento {
		itemsDocumentoMap[item.Key] = item.Description
	}
}

// GetNaturezaDocumento retorna a descrição do documento pelo código
func GetNaturezaDocumento(key int) string {
	if desc, ok := itemsDocumentoMap[key]; ok {
		return desc
	}
	return "Não identificado"
}

var naturezaDocsImportarPJE = []string{
	"petição",
	"contestação",
	"réplica",
	"despacho",
	"pedido",
	"interlocutória",
	"sentença",
	"recurso",
	"contrarrazões",
	"petição intermediária",
	"ata de audiência",
}

func GetNaturezaDocumentosImportarPJE() []string {
	return naturezaDocsImportarPJE
}

func GetTipoDocumento(tipo string) int {
	tipo = strings.ToLower(tipo)

	for i, ok := range naturezaDocsImportarPJE {
		if strings.Contains(tipo, ok) {
			return i + 1
		}
	}
	return 0
}
