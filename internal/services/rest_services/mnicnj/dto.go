package mnicnj

type APIError struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

type APIResponse[T any] struct {
	ID        string    `json:"id,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	OK        bool      `json:"ok"`
	Message   string    `json:"message,omitempty"`
	Data      T         `json:"data,omitempty"`
	Error     *APIError `json:"error,omitempty"`
}

type ConsultaProcessoPJERequest struct {
	NumeroProcesso string `json:"numero_processo"`

	UsuarioCpf   string `json:"usuario_cpf"`
	UsuarioSenha string `json:"usuario_senha"`

	DataReferencia string `json:"data_referencia,omitempty"`

	Documentos []string `json:"documentos,omitempty"`

	// auto | html | texto | base64 | metadados
	Formato string `json:"formato,omitempty"`
}

type MNIAnexoResponse struct {
	ContentID    string `json:"content_id"`
	ContentType  string `json:"content_type"`
	MediaType    string `json:"media_type"`
	TamanhoBytes int    `json:"tamanho_bytes"`

	ConteudoHTML   string `json:"conteudo_html,omitempty"`
	ConteudoTexto  string `json:"conteudo_texto,omitempty"`
	ConteudoBase64 string `json:"conteudo_base64,omitempty"`
}

type ConsultaProcessoPJEResponse struct {
	StatusCode int `json:"status_code"`

	Resultado any `json:"resultado,omitempty"`

	CIDsReferenciados []string           `json:"cids_referenciados,omitempty"`
	Anexos            []MNIAnexoResponse `json:"anexos,omitempty"`
}

type DocumentoConteudoResponse struct {
	IDDocumento          string `json:"id_documento"`
	IDDocumentoVinculado string `json:"id_documento_vinculado,omitempty"`

	Descricao    string `json:"descricao,omitempty"`
	Mimetype     string `json:"mimetype,omitempty"`
	ContentID    string `json:"content_id,omitempty"`
	ContentType  string `json:"content_type,omitempty"`
	MediaType    string `json:"media_type,omitempty"`
	TamanhoAnexo int    `json:"tamanho_anexo_bytes,omitempty"`

	StorageID          string `json:"storage_id,omitempty"`
	NomeDocumento      string `json:"nome_documento,omitempty"`
	NomeArquivo        string `json:"nome_arquivo,omitempty"`
	LinkValidacao      string `json:"link_validacao,omitempty"`
	TamanhoBytes       *int64 `json:"tamanho_bytes,omitempty"`
	DataInclusaoMillis *int64 `json:"data_inclusao_millis,omitempty"`
	DataJuntadaMillis  *int64 `json:"data_juntada_millis,omitempty"`

	FormatoEntrega string `json:"formato_entrega,omitempty"`

	ConteudoHTML   string `json:"conteudo_html,omitempty"`
	ConteudoTexto  string `json:"conteudo_texto,omitempty"`
	ConteudoBase64 string `json:"conteudo_base64,omitempty"`

	DocumentoEncontrado bool   `json:"documento_encontrado"`
	ConteudoEncontrado  bool   `json:"conteudo_encontrado"`
	ErroConteudo        string `json:"erro_conteudo,omitempty"`
}

type ConsultaDocumentoPJEResponse struct {
	StatusCode int `json:"status_code"`

	FormatoEntrega        string                      `json:"formato_entrega"`
	DocumentosSolicitados []string                    `json:"documentos_solicitados,omitempty"`
	Documentos            []DocumentoConteudoResponse `json:"documentos"`
	CIDsReferenciados     []string                    `json:"cids_referenciados,omitempty"`
	Anexos                []MNIAnexoResponse          `json:"anexos,omitempty"`
}
