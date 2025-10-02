package handlers

import (
	"net/http"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

// Estrutura do Handler RAG
type RagHandlerType struct {
	idx *opensearch.IndexRagType
}

// Construtor do Handler
func NewRagHandlers(index *opensearch.IndexRagType) *RagHandlerType {
	return &RagHandlerType{idx: index}
}

/*
  - Insere um novo documento no índice RAG
    *Rota: "/rag"
    *Método: POST
*/
func (handler *RagHandlerType) InsertHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var bodyParams opensearch.DocumentoRag
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Dados inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	if bodyParams.DataTexto == "" || bodyParams.Natureza == "" {
		logger.Log.Error("Campos obrigatórios: natureza e data_texto")
		response.HandleError(c, http.StatusBadRequest, "Campos obrigatórios: natureza e data_texto", "", requestID)
		return
	}

	vector, err := services.GetDocumentoEmbeddings(bodyParams.DataTexto)
	if err != nil {
		logger.Log.Errorf("Erro ao gerar embeddings: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar embeddings", "", requestID)
		return
	}

	bodyParams.DataEmbedding = vector

	resp, err := handler.idx.IndexaDocumento(bodyParams)
	if err != nil {
		logger.Log.Errorf("Erro ao inserir documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao inserir documento", "", requestID)
		return
	}

	rsp := gin.H{
		"id":      resp.ID,
		"message": "Documento inserido com sucesso em RAG!",
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/*
  - Atualiza documento no RAG (somente o campo texto, por enquanto)
    *Rota: "/rag/:id"
    *Método: PUT
*/
func (handler *RagHandlerType) UpdateHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	idDoc := c.Param("id")

	var bodyParams opensearch.RagText
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Body inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Body inválido", "", requestID)
		return
	}

	doc, err := handler.idx.UpdateDocumento(idDoc, bodyParams)
	if err != nil {
		logger.Log.Errorf("Erro ao atualizar documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao atualizar documento", "", requestID)
		return
	}

	rsp := gin.H{
		"doc":     doc,
		"message": "Documento atualizado com sucesso em RAG!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/*
  - Deleta documento do RAG
    *Rota: "/rag/:id"
    *Método: DELETE
*/
func (handler *RagHandlerType) DeleteHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	id := c.Param("id")

	_, err := handler.idx.DeleteDocumento(id)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao deletar documento", "", requestID)
		return
	}

	rsp := gin.H{"ok": true, "message": "Documento excluído de RAG!"}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/*
  - Busca documento pelo ID no RAG
    *Rota: "/rag/:id"
    *Método: GET
*/
func (handler *RagHandlerType) SelectByIdHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	id := c.Param("id")

	doc, err := handler.idx.ConsultaDocumentoById(id)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao buscar documento", "", requestID)
		return
	}

	if doc == nil {
		response.HandleError(c, http.StatusNotFound, "Documento não encontrado em RAG!", "", requestID)
		return
	}

	rsp := gin.H{"doc": doc, "message": "Documento encontrado em RAG!"}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
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

func (handler *RagHandlerType) SearchHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var bodyParams BodySearchRag
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Body inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Body inválido", "", requestID)
		return
	}

	if bodyParams.SearchTexto == "" {
		response.HandleError(c, http.StatusBadRequest, "search_texto é obrigatório", "", requestID)
		return
	}

	vec, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(c.Request.Context(), bodyParams.SearchTexto)
	if err != nil {
		logger.Log.Errorf("Erro ao gerar embedding da busca: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao gerar embedding", "", requestID)
		return
	}

	docs, err := handler.idx.ConsultaSemantica(vec, bodyParams.Natureza)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar documentos RAG: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na consulta", "", requestID)
		return
	}

	msg := "Consulta RAG realizada com sucesso"
	if len(docs) == 0 {
		msg += ": nenhum documento retornado"
	}

	rsp := gin.H{"docs": docs, "message": msg}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
