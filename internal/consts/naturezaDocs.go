package consts

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
	NATU_DOC_CONTRA_RAZ      = 10
	NATU_DOC_APELACAO        = 11
	NATU_DOC_PROCURACAO      = 12
	NATU_DOC_ROL_TESTEMUNHAS = 13
	NATU_DOC_CONTRATO        = 14
	NATU_DOC_LAUDO_PERICIA   = 15
	NATU_DOC_ATA_AUDIENCIA   = 16
	NATU_DOC_PARECER_MP      = 17
)

// Definição do tipo Item
type Item struct {
	Key         int
	Description string
}

var itemsDocumento = []Item{
	{Key: 0, Description: "Selecione o documento"},
	{Key: 1, Description: "Petição inicial"},
	{Key: 2, Description: "Contestação"},
	{Key: 3, Description: "Réplica"},
	{Key: 4, Description: "Despacho inicial"},
	{Key: 5, Description: "Despacho ordinatório"},
	{Key: 6, Description: "Petição diversa"},
	{Key: 7, Description: "Decisão interlocutória"},
	{Key: 8, Description: "Sentença"},
	{Key: 9, Description: "Embargos de declaração"},
	{Key: 10, Description: "Contra-razões"},
	{Key: 11, Description: "Recurso de Apelação"},
	{Key: 12, Description: "Procuração"},
	{Key: 13, Description: "Rol de Testemunhas"},
	{Key: 14, Description: "Contrato"},
	{Key: 15, Description: "Laudo Pericial"},
	{Key: 16, Description: "Ata de Audiência"},
	{Key: 17, Description: "Manifestação do Ministério Público"},
	{Key: 1000, Description: "Autos Processuais"},
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
