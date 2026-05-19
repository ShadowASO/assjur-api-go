package handlers

import (
	"net/http"
	"strings"

	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/services/rag/pipeline"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

// Estrutura do Handler RAG
type BaseHandlerType struct {
	Service *services.BaseServiceType
	idx     *opensearch.BaseIndexType
}

// Construtor do Handler
//
//	func NewRagHandlers(index *opensearch.BaseIndexType) *BaseHandlerType {
//		return &BaseHandlerType{idx: index}
//	}
func NewBaseHandlers(service *services.BaseServiceType) *BaseHandlerType {
	return &BaseHandlerType{
		Service: service,
	}
}

type bodyParamsBaseUpdate struct {
	Id    string `json:"id"`
	Tema  string `json:"tema"`
	Texto string `json:"texto"`
}

type bodyParamsBaseInsert struct {
	IdCtxt   string `json:"id_ctxt"`
	IdPje    string `json:"id_pje"`
	Classe   string `json:"classe"`
	Assunto  string `json:"assunto"`
	Natureza string `json:"natureza"`
	Tipo     string `json:"tipo"`
	Tema     string `json:"tema"`
	Fonte    string `json:"fonte"`
	Texto    string `json:"texto"`
}

/*
  - Insere um novo documento no índice RAG
    *Rota: "/rag"
    *Método: POST
*/
func (obj *BaseHandlerType) InsertHandler(c *gin.Context) {
	userName := c.GetString("userName")

	var bodyParams bodyParamsBaseInsert
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		mslogger.LoggerGlobal.Errorf("Dados inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Parâmetros do body inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	bodyParams.Natureza = strings.TrimSpace(bodyParams.Natureza)
	bodyParams.Texto = strings.TrimSpace(bodyParams.Texto)

	if bodyParams.Texto == "" || bodyParams.Natureza == "" {
		mslogger.LoggerGlobal.Error("Campos obrigatórios: natureza e texto")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Os campos natureza e texto são obrigatórios.",
		)
		return
	}

	hashTexto := pipeline.GetHashFromTexto(bodyParams.Texto)
	mslogger.LoggerGlobal.Infof("\nhash_texto: %s", hashTexto)

	row, err := obj.Service.InserirDocumento(
		bodyParams.IdCtxt,
		bodyParams.IdPje,
		userName,
		bodyParams.Classe,
		bodyParams.Assunto,
		bodyParams.Natureza,
		bodyParams.Tipo,
		bodyParams.Tema,
		bodyParams.Fonte,
		bodyParams.Texto,
		hashTexto,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao inserir contexto: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno no servidor ao inserir contexto",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"row": row,
	}

	msresponse.OK(c, http.StatusCreated, "Documento inserido com sucesso em RAG", rsp)
}

/*
  - Atualiza documento no RAG (somente o campo texto, por enquanto)
    *Rota: "/rag/:id"
    *Método: PUT
*/
func (obj *BaseHandlerType) UpdateHandler(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	if id == "" {
		mslogger.LoggerGlobal.Error("ID do documento não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do documento não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	var bodyParams bodyParamsBaseUpdate
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		mslogger.LoggerGlobal.Errorf("Body inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Body inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	bodyParams.Tema = strings.TrimSpace(bodyParams.Tema)
	bodyParams.Texto = strings.TrimSpace(bodyParams.Texto)

	if bodyParams.Texto == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Texto não informado",
			msresponse.ErrorValidacao,
			"O campo texto é obrigatório.",
		)
		return
	}

	vector, err := services.GetDocumentoEmbeddings(bodyParams.Texto)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar embeddings: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao gerar embeddings",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	row, err := obj.Service.UpdateDocumento(
		id,
		bodyParams.Tema,
		bodyParams.Texto,
		vector,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na alteração do registro!: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno no servidor ao alterar o registro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"row": row,
	}

	msresponse.OK(c, http.StatusOK, "Registro alterado com sucesso", rsp)
}

/*
  - Deleta documento do RAG
    *Rota: "/rag/:id"
    *Método: DELETE
*/
func (obj *BaseHandlerType) DeleteHandler(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	if id == "" {
		mslogger.LoggerGlobal.Error("ID do documento não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do documento não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	if err := obj.Service.DeletaDocumento(id); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na deleção do registro!: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro na deleção do registro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Registro deletado com sucesso")
}

/*
  - Busca documento pelo ID no RAG
    *Rota: "/rag/:id"
    *Método: GET
*/
func (obj *BaseHandlerType) SelectByIdHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID do documento não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do documento não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	row, err := obj.Service.SelectById(paramID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Registro não encontrado!: %v", err)

		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Registro não encontrado",
			msresponse.ErrorNaoEncontrado,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"doc": row,
	}

	msresponse.OK(c, http.StatusOK, "Registro selecionado com sucesso", rsp)
}

/*
  - Busca semântica no RAG
    *Rota: "/rag/search"
    *Método: POST
*/
type BodySearchRag struct {
	Natureza    string `json:"natureza"`
	SearchTexto string `json:"search_texto"`
}

func (obj *BaseHandlerType) SearchHandler(c *gin.Context) {
	var bodyParams BodySearchRag

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		mslogger.LoggerGlobal.Errorf("Body inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Body inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	bodyParams.SearchTexto = strings.TrimSpace(bodyParams.SearchTexto)
	bodyParams.Natureza = strings.TrimSpace(bodyParams.Natureza)

	if bodyParams.SearchTexto == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"search_texto é obrigatório",
			msresponse.ErrorValidacao,
			"O campo search_texto é obrigatório.",
		)
		return
	}

	if len(bodyParams.SearchTexto) > 8000 {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"search_texto muito grande",
			msresponse.ErrorValidacao,
			"O campo search_texto excede o limite máximo de 8000 caracteres.",
		)
		return
	}

	// Se o cliente abortou, não trata como erro interno.
	if err := c.Request.Context().Err(); err != nil {
		c.Status(499)
		return
	}

	docs, err := obj.Service.ConsultaSemantica(
		bodyParams.SearchTexto,
		bodyParams.Natureza,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar documentos: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro na consulta",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	message := "Consulta realizada com sucesso"
	if len(docs) == 0 {
		message = "Consulta realizada com sucesso: nenhum documento retornado"
	}

	rsp := gin.H{
		"docs": docs,
	}

	msresponse.OK(c, http.StatusOK, message, rsp)
}
