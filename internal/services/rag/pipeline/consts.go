package pipeline

/* Eventos do usuário. */
const (
	EVENTO_PREANALISE = 200
	EVENTO_ANALISE    = 201
	EVENTO_SENTENCA   = 202
	EVENTO_DECISAO    = 203
	EVENTO_DESPACHO   = 204
	EVENTO_CONCEITOS  = 205
	//-----  Comp

	EVENTO_CONFIRMACAO = 300
	EVENTO_COMPLEMENTO = 301
	EVENTO_ADD_BASE    = 302

	EVENTO_OUTROS = 999
)

// Tamanho máximo, em tokens de cada documentos a ser inserido em uma mensagem para o modelo.
const MAX_DOC_TOKENS = 4000
