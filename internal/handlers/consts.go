package handlers

// Estrutura base para o JSON (mantida como referência genérica)
type DocumentoBase struct {
	Tipo *struct {
		Key         int    `json:"key"`
		Description string `json:"description"`
	} `json:"tipo"`
	Processo string `json:"processo"`
	IdEvento string `json:"id_evento"`
}
