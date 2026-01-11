package handlers

import (
	"net/http"
	"strings"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/services/rag/pipeline"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

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
	//Status   string `json:"status"`
}

/*
  - Insere um novo documento no índice RAG
    *Rota: "/rag"
    *Método: POST
*/
func (obj *BaseHandlerType) InsertHandler(c *gin.Context) {
	userName := c.GetString("userName")
	requestID := middleware.GetRequestID(c)

	var bodyParams bodyParamsBaseInsert
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Dados inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	if bodyParams.Texto == "" || bodyParams.Natureza == "" {
		logger.Log.Error("Campos obrigatórios: natureza e texto")
		response.HandleError(c, http.StatusBadRequest, "Campos obrigatórios: natureza e data_texto", "", requestID)
		return
	}
	hash_texto := pipeline.GetHashFromTexto(bodyParams.Texto)
	logger.Log.Infof("\nhash_texto: %s", hash_texto)

	resp, err := obj.Service.InserirDocumento(
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
		hash_texto,
	)
	if err != nil {
		logger.Log.Errorf("Erro ao inserir contexto: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no servidor ao inserir contexto!", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     resp,
		"message": "Documento inserido com sucesso em RAG!",
	}
	response.HandleSucesso(c, http.StatusCreated, rsp, requestID)
}

/*
  - Atualiza documento no RAG (somente o campo texto, por enquanto)
    *Rota: "/rag/:id"
    *Método: PUT
*/
func (obj *BaseHandlerType) UpdateHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	id := c.Param("id")

	var bodyParams bodyParamsBaseUpdate
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Body inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Body inválido", "", requestID)
		return
	}
	vector, err := services.GetDocumentoEmbeddings(bodyParams.Texto)
	if err != nil {
		logger.Log.Errorf("Erro ao gerar embeddings: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar embeddings", "", requestID)
		return
	}

	row, err := obj.Service.UpdateDocumento(
		id,
		bodyParams.Tema,
		bodyParams.Texto,
		vector,
	)
	if err != nil {

		logger.Log.Errorf("Erro na alteração do registro!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no servidor ao altear o registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro alterado com sucesso!",
	}

	response.HandleSucesso(c, http.StatusOK, rsp, requestID)
}

/*
  - Deleta documento do RAG
    *Rota: "/rag/:id"
    *Método: DELETE
*/
func (obj *BaseHandlerType) DeleteHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	id := c.Param("id")
	if id == "" {
		logger.Log.Error("ID da sessão não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID da sessão não informado!", "", requestID)
		return
	}

	err := obj.Service.DeletaDocumento(id)
	if err != nil {
		logger.Log.Errorf("Erro na deleção do registro!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na deleção do registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"ok":      true,
		"message": "Registro deletado com sucesso!",
	}

	response.HandleSucesso(c, http.StatusOK, rsp, requestID)
}

/*
  - Busca documento pelo ID no RAG
    *Rota: "/rag/:id"
    *Método: GET
*/
func (obj *BaseHandlerType) SelectByIdHandler(c *gin.Context) {
	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	paramID := c.Param("id")
	if paramID == "" {

		logger.Log.Error("ID da sessão não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID da sessão não informado!", "", requestID)
		return
	}

	row, err := obj.Service.SelectById(paramID)

	if err != nil {

		logger.Log.Errorf("Registro não encontrado!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Registro não encontrado!", "", requestID)
		return
	}

	rsp := gin.H{
		"doc":     row,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSucesso(c, http.StatusOK, rsp, requestID)
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
	requestID := middleware.GetRequestID(c)

	var bodyParams BodySearchRag
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Body inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Body inválido", "", requestID)
		return
	}

	// ✅ NORMALIZA
	bodyParams.SearchTexto = strings.TrimSpace(bodyParams.SearchTexto)
	bodyParams.Natureza = strings.TrimSpace(bodyParams.Natureza)

	if bodyParams.SearchTexto == "" {
		response.HandleError(c, http.StatusBadRequest, "search_texto é obrigatório", "", requestID)
		return
	}
	// (opcional) limite defensivo
	if len(bodyParams.SearchTexto) > 8000 {
		response.HandleError(c, http.StatusBadRequest, "search_texto muito grande", "", requestID)
		return
	}

	// ✅ se o cliente abortou, não trate como erro interno
	if err := c.Request.Context().Err(); err != nil {
		// 499 é comum (nginx), mas como não existe constante no net/http:
		c.Status(499)
		return
	}

	docs, err := obj.Service.ConsultaSemantica(bodyParams.SearchTexto, bodyParams.Natureza)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar documentos: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na consulta", "", requestID)
		return
	}

	msg := "Consulta realizada com sucesso"
	if len(docs) == 0 {
		msg += ": nenhum documento retornado"
	}

	rsp := gin.H{"docs": docs, "message": msg}
	response.HandleSucesso(c, http.StatusOK, rsp, requestID)
}
