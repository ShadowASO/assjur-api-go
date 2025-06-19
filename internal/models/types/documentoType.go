package types

// Constantes para cada tipo de documento
const (
	ItemDocumentoSelecionar                    = 0
	ItemDocumentoPeticaoInicial                = 1
	ItemDocumentoContestacao                   = 2
	ItemDocumentoReplica                       = 3
	ItemDocumentoDespachoInicial               = 4
	ItemDocumentoDespachoOrdinatorio           = 5
	ItemDocumentoPeticaoDiversa                = 6
	ItemDocumentoDecisaoInterlocutoria         = 7
	ItemDocumentoSentenca                      = 8
	ItemDocumentoEmbargosDeclaracao            = 9
	ItemDocumentoContraRazoes                  = 10
	ItemDocumentoRecursoApelacao               = 11
	ItemDocumentoProcuracao                    = 12
	ItemDocumentoRolTestemunhas                = 13
	ItemDocumentoContrato                      = 14
	ItemDocumentoLaudoPericial                 = 15
	ItemDocumentoAtaAudiencia                  = 16
	ItemDocumentoManifestacaoMinisterioPublico = 17
	ItemDocumentoAutosProcessuais              = 1000
)

// Estrutura opcional (se precisar iterar ou exportar como lista)
type ItemDocumento struct {
	Key         int
	Description string
}

// Lista de todos os documentos
var ItemsDocumento = []ItemDocumento{
	{Key: ItemDocumentoSelecionar, Description: "Selecione o documento"},
	{Key: ItemDocumentoPeticaoInicial, Description: "Petição inicial"},
	{Key: ItemDocumentoContestacao, Description: "Contestação"},
	{Key: ItemDocumentoReplica, Description: "Réplica"},
	{Key: ItemDocumentoDespachoInicial, Description: "Despacho inicial"},
	{Key: ItemDocumentoDespachoOrdinatorio, Description: "Despacho ordinatório"},
	{Key: ItemDocumentoPeticaoDiversa, Description: "Petição diversa"},
	{Key: ItemDocumentoDecisaoInterlocutoria, Description: "Decisão interlocutória"},
	{Key: ItemDocumentoSentenca, Description: "Sentença"},
	{Key: ItemDocumentoEmbargosDeclaracao, Description: "Embargos de declaração"},
	{Key: ItemDocumentoContraRazoes, Description: "Contra-razões"},
	{Key: ItemDocumentoRecursoApelacao, Description: "Recurso de Apelação"},
	{Key: ItemDocumentoProcuracao, Description: "Procuração"},
	{Key: ItemDocumentoRolTestemunhas, Description: "Rol de Testemunhas"},
	{Key: ItemDocumentoContrato, Description: "Contrato"},
	{Key: ItemDocumentoLaudoPericial, Description: "Laudo Pericial"},
	{Key: ItemDocumentoAtaAudiencia, Description: "Ata de Audiência"},
	{Key: ItemDocumentoManifestacaoMinisterioPublico, Description: "Manifestação do Ministério Público"},
	{Key: ItemDocumentoAutosProcessuais, Description: "Autos Processuais"},
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
