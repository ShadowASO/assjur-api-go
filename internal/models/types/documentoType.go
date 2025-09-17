package types

import (
	"ocrserver/internal/consts"
)

// Constantes para cada tipo de documento
// const (
// 	ItemDocumentoSelecionar     = 0
// 	ItemDocumentoPeticaoInicial = 1
// 	ItemDocumentoContestacao    = 2
// 	ItemDocumentoReplica        = 3
// 	ItemDocumentoDespacho       = 4
// 	//ItemDocumentoDespachoOrdinatorio           = 5
// 	ItemDocumentoPeticao = 5
// 	//ItemDocumentoDecisaoInterlocutoria         = 7
// 	ItemDocumentoDecisao                       = 6
// 	ItemDocumentoSentenca                      = 7
// 	ItemDocumentoEmbargosDeclaracao            = 8
// 	ItemDocumentoContraRazoes                  = 10
// 	ItemDocumentoRecursoApelacao               = 11
// 	ItemDocumentoProcuracao                    = 12
// 	ItemDocumentoRolTestemunhas                = 13
// 	ItemDocumentoContrato                      = 14
// 	ItemDocumentoLaudoPericial                 = 15
// 	ItemDocumentoAtaAudiencia                  = 16
// 	ItemDocumentoManifestacaoMinisterioPublico = 17
// 	ItemDocumentoAutosProcessuais              = 1000
// )

// Estrutura opcional (se precisar iterar ou exportar como lista)
type ItemDocumento struct {
	Key         int
	Description string
}

// Lista de todos os documentos
var ItemsDocumento = []ItemDocumento{
	//{Key: ItemDocumentoSelecionar, Description: "Selecione o documento"},
	{Key: 0, Description: "Selecione o documento"},
	//{Key: ItemDocumentoPeticaoInicial, Description: "Petição inicial"},
	{Key: consts.NATU_DOC_INICIAL, Description: "Petição inicial"},
	{Key: consts.NATU_DOC_CONTESTACAO, Description: "Contestação"},
	{Key: consts.NATU_DOC_REPLICA, Description: "Réplica"},
	{Key: consts.NATU_DOC_DESPACHO, Description: "Despacho"},

	{Key: consts.NATU_DOC_PETICAO, Description: "Petição"},
	{Key: consts.NATU_DOC_DECISAO, Description: "Decisão"},
	{Key: consts.NATU_DOC_SENTENCA, Description: "Sentença"},
	{Key: consts.NATU_DOC_EMBARGOS, Description: "Embargos de declaração"},
	{Key: consts.NATU_DOC_CONTRA_RAZOES, Description: "Contra-razões"},
	{Key: consts.NATU_DOC_APELACAO, Description: "Recurso de Apelação"},
	{Key: consts.NATU_DOC_PROCURACAO, Description: "Procuração"},
	{Key: consts.NATU_DOC_ROL_TESTEMUNHAS, Description: "Rol de Testemunhas"},
	{Key: consts.NATU_DOC_CONTRATO, Description: "Contrato"},
	{Key: consts.NATU_DOC_LAUDO_PERICIAL, Description: "Laudo Pericial"},
	{Key: consts.NATU_DOC_TERMO_AUDIENCIA, Description: "Termo de Audiência"},
	{Key: consts.NATU_DOC_PARECER_MP, Description: "Manifestação do Ministério Público"},
	{Key: consts.NATU_DOC_AUTOS, Description: "Autos Processuais"},
}

// Mapa para lookup rápido
var documentoDescriptionMap map[int]string

// Inicialização automática do map
func init() {
	documentoDescriptionMap = make(map[int]string, len(ItemsDocumento))
	for _, item := range ItemsDocumento {
		documentoDescriptionMap[item.Key] = item.Description
	}
}

// Função de busca rápida
func GetDocumentoDescriptionByKey(key int) string {
	if desc, found := documentoDescriptionMap[key]; found {
		return desc
	}
	return "Documento desconhecido"
}
