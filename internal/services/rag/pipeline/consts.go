package pipeline

/* Eventos do usuário. */
const (
	RAG_EVENTO_PREANALISE = 200
	RAG_EVENTO_ANALISE    = 201
	RAG_EVENTO_SENTENCA   = 202
	RAG_EVENTO_DECISAO    = 203
	RAG_EVENTO_DESPACHO   = 204
	RAG_EVENTO_CONCEITOS  = 205
	//-----  Comp

	RAG_EVENTO_CONFIRMACAO = 300
	RAG_EVENTO_COMPLEMENTO = 301
	RAG_EVENTO_ADD_BASE    = 302

	RAG_EVENTO_OUTROS = 999
)

// Tamanho máximo, em tokens de cada documentos a ser inserido em uma mensagem para o modelo.
const MAX_DOC_TOKENS = 3000
