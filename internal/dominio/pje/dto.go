package pje

import "time"

type PjeDocumentoTmp struct {
	IdCtxt                string     `json:"id_ctxt,omitempty"`
	NumeroProcesso        string     `json:"numero_processo,omitempty"`
	IdPje                 string     `json:"id_pje,omitempty"`
	TipoDocumento         string     `json:"tipo_documento,omitempty"`
	Descricao             string     `json:"descricao,omitempty"`
	StorageID             string     `json:"storage_id,omitempty"`
	Mimetype              string     `json:"mimetype,omitempty"`
	FormatoEntrega        string     `json:"formato_entrega,omitempty"`
	ConteudoTexto         string     `json:"conteudo_texto,omitempty"`
	ConteudoHTML          string     `json:"conteudo_html,omitempty"`
	ConteudoBase64        string     `json:"conteudo_base64,omitempty"`
	Status                string     `json:"status,omitempty"`
	Erro                  string     `json:"erro,omitempty"`
	DataJuntadaMillis     int64      `json:"data_juntada_millis,omitempty"`
	UsuarioJuntadaArquivo string     `json:"usuario_juntada_arquivo,omitempty"`
	CriadoEm              *time.Time `json:"criado_em,omitempty"`
	AtualizadoEm          *time.Time `json:"atualizado_em,omitempty"`
	ExpiraEm              *time.Time `json:"expira_em,omitempty"`
}

type ResponsePjeDocumentoTmp struct {
	Id string `json:"id,omitempty"`

	IdCtxt                string     `json:"id_ctxt,omitempty"`
	NumeroProcesso        string     `json:"numero_processo,omitempty"`
	IdPje                 string     `json:"id_pje,omitempty"`
	TipoDocumento         string     `json:"tipo_documento,omitempty"`
	Descricao             string     `json:"descricao,omitempty"`
	StorageID             string     `json:"storage_id,omitempty"`
	Mimetype              string     `json:"mimetype,omitempty"`
	FormatoEntrega        string     `json:"formato_entrega,omitempty"`
	ConteudoTexto         string     `json:"conteudo_texto,omitempty"`
	ConteudoHTML          string     `json:"conteudo_html,omitempty"`
	ConteudoBase64        string     `json:"conteudo_base64,omitempty"`
	Status                string     `json:"status,omitempty"`
	Erro                  string     `json:"erro,omitempty"`
	DataJuntadaMillis     int64      `json:"data_juntada_millis,omitempty"`
	UsuarioJuntadaArquivo string     `json:"usuario_juntada_arquivo,omitempty"`
	CriadoEm              *time.Time `json:"criado_em,omitempty"`
	AtualizadoEm          *time.Time `json:"atualizado_em,omitempty"`
	ExpiraEm              *time.Time `json:"expira_em,omitempty"`
}

type PjeDocumentoTmpPatch struct {
	IdCtxt                *string    `json:"id_ctxt,omitempty"`
	NumeroProcesso        *string    `json:"numero_processo,omitempty"`
	IdPje                 *string    `json:"id_pje,omitempty"`
	TipoDocumento         *string    `json:"tipo_documento,omitempty"`
	Descricao             *string    `json:"descricao,omitempty"`
	StorageID             *string    `json:"storage_id,omitempty"`
	Mimetype              *string    `json:"mimetype,omitempty"`
	FormatoEntrega        *string    `json:"formato_entrega,omitempty"`
	ConteudoTexto         *string    `json:"conteudo_texto,omitempty"`
	ConteudoHTML          *string    `json:"conteudo_html,omitempty"`
	ConteudoBase64        *string    `json:"conteudo_base64,omitempty"`
	Status                *string    `json:"status,omitempty"`
	Erro                  *string    `json:"erro,omitempty"`
	DataJuntadaMillis     *int64     `json:"data_juntada_millis,omitempty"`
	UsuarioJuntadaArquivo *string    `json:"usuario_juntada_arquivo,omitempty"`
	CriadoEm              *time.Time `json:"criado_em,omitempty"`
	AtualizadoEm          *time.Time `json:"atualizado_em,omitempty"`
	ExpiraEm              *time.Time `json:"expira_em,omitempty"`
}
