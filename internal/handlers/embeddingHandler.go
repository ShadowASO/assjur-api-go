package handlers

import (
	"net/http"
	"strconv"

	"ocrserver/internal/handlers/response"

	"ocrserver/internal/services/embedding"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

// Estrutura do Handler
type EmbeddingHandlerType struct {
	service *embedding.AutosEmbeddingType
}

// Construtor do Handler
func NewEmbeddingHandlers(service *embedding.AutosEmbeddingType) *EmbeddingHandlerType {

	return &EmbeddingHandlerType{service: service}
}

type BodyAutosInsert struct {
	IdCtxt  int    `json:"id_ctxt"`
	IdNatu  int    `json:"id_natu"`
	IdPje   string `json:"id_pje"`
	DocText string `json:"doc_text"`
}

/*
  - Insere um novo documento no banco vetorial mantido no openSearch, nos índices
  autos_embedding e decisoes.

  * Rota: "/tabelas/modelos/autos/:id"
  - Método: POST

  - Body: {

    }
*/

func (handler *EmbeddingHandlerType) InsertHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID do contexto não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID do contexto não informado!", "", requestID)
		return
	}

	idCtxt, err := strconv.Atoi(paramID)
	if err != nil {
		logger.Log.Errorf("ID inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "ID inválido!", "", requestID)
		return
	}

	rspSuc, rspFal, err := handler.service.IncluirAutosByContexto(idCtxt)
	if err != nil {
		logger.Log.Errorf("Erro ao inserir documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao inserir documento!", "", requestID)
		return
	}

	rsp := gin.H{
		"sucesso": rspSuc,
		"falha":   rspFal,
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/*
  - Insere um novo documento no banco vetorial mantido no openSearch, nos índices
  autos_embedding e decisoes.

  *Rota: "/tabelas/modelos/autos/doc"
- Método: POST

  - Body: {
		IdCtxt int    `json:"id_ctxt"`
		IdNatu     string `json:"id_natu"`
		IdPje        string    `json:"id_pje"`
		DocText string `json:"doc_text"`
    }
*/

func (handler *EmbeddingHandlerType) InsertDocumentoHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	var bodyParams BodyAutosInsert

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Dados do body de requisição inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body da requisição inválidos", "", requestID)
		return
	}

	if bodyParams.IdCtxt == 0 || bodyParams.IdNatu == 0 || bodyParams.DocText == "" {
		logger.Log.Error("Um dos campos obrigatórios está ausente: id_ctxt, id_natu e doc_text")
		response.HandleError(c, http.StatusBadRequest, "Todos os campos são obrigatórios: id_ctxt, id_natu e doc_text", "", requestID)
		return
	}

	resp, err := handler.service.IncluirDocumento(bodyParams.IdCtxt, bodyParams.IdNatu, bodyParams.IdPje, bodyParams.DocText)
	if err != nil {
		logger.Log.Errorf("Erro ao inserir documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao inserir documento!", "", requestID)
		return
	}

	rsp := gin.H{
		"id":      resp,
		"message": "Registro inserido com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/*
  - Modifica  um documento existente no Elasticsearch
    *Rota: "/tabelas/modelos/{id}"
  - Método: PUT
  - Body: {
    Natureza     string `json:"natureza"`
    Ementa       string `json:"ementa"`
    Inteiro_teor string `json:"inteiro_teor"`
    }
*/
// Atualiza um documento existente no OpenSearch
func (handler *EmbeddingHandlerType) UpdateHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	idDoc := c.Param("id")
	var bodyParams BodyAutosInsert

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Dados inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	if idDoc == "" {
		logger.Log.Error("Id do documento é obrigatório!")
		response.HandleError(c, http.StatusBadRequest, "Id do documento é obrigatório!", "", requestID)
		return
	}

	// doc, err := handler.service.IndexAutos.UpdateDocumento(idDoc, bodyParams)
	// if err != nil {
	// 	logger.Log.Errorf("Erro ao atualizar documento: %v", err)
	// 	response.HandleError(c, http.StatusInternalServerError, "Erro ao atualizar documento!", "", requestID)
	// 	return
	// }

	rsp := gin.H{
		//"doc":     doc,
		"message": "Registro alterado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/*
  - Deleta  um documento existente no Elasticsearch
    *Rota: "/tabelas/modelos/:{id}"
  - Método: DELETE
  - Body: {
    }
*/
// Deleta um documento existente no OpenSearch
func (handler *EmbeddingHandlerType) DeleteHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	id := c.Param("id")
	if id == "" {

		logger.Log.Error("ID do documento não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID do documento não informado!", "", requestID)
		return
	}

	err := handler.service.AutosIndex.Delete(id)
	if err != nil {

		logger.Log.Errorf("Erro ao deletar documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao deletar documento!", "", requestID)
		return
	}

	rsp := gin.H{
		"ok":      true,
		"message": "Registro excluído com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

/*
  - Deleta  um documento existente no Elasticsearch
    *Rota: "/tabelas/elastic/{id}"
  - Método: GET
  - Body: {
    }
*/
// Busca um documento pelo ID no OpenSearch
func (handler *EmbeddingHandlerType) SelectByIdHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	id := c.Param("id")
	if id == "" {

		logger.Log.Error("ID do documento não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID do documento não informado!", "", requestID)
		return
	}

	documento, err := handler.service.GetDocumentoById(id)
	if err != nil {

		logger.Log.Errorf("Erro ao buscar documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao buscar documento!", "", requestID)

		return
	}

	if documento == nil {

		logger.Log.Error("Documento não encontrado!")
		response.HandleError(c, http.StatusBadRequest, "Documento não encontrado!", "", requestID)
		return
	}

	rsp := gin.H{
		"doc":     documento,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/*
  - Seleciona documentos que sejam da "Natureza" apontada e contenham o conteúdo "Search_texto"
    *Rota: "/tabelas/modelos/search"
  - Método: POST
  - Body: {
		Index_name   string `json:"index_name"`
		Natureza     string `json:"natureza"`
		Search_texto string `json:"search_texto"`
    }
*/
// Estrutura para o corpo da requisição
type BodySearchEmbedding struct {
	IdCtxt      int    `json:"id_ctxt"`
	IdNatu      int    `json:"id_natu"`
	SearchTexto string `json:"search_texto"`
}

// Busca documentos pelo conteúdo no OpenSearch
func (handler *EmbeddingHandlerType) SearchEmbeddingHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	bodyParams := BodySearchEmbedding{}
	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Errorf("Formato inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	if bodyParams.IdCtxt == 0 || bodyParams.IdNatu == 0 || bodyParams.SearchTexto == "" {

		logger.Log.Error("contexto, natureza e searchtexto são obrigatórios")
		response.HandleError(c, http.StatusBadRequest, "index_name, natureza e search_texto são obrigatórios", "", requestID)
		return
	}

	//Converte a string de busca num embedding
	// rspEmbeddings, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(c.Request.Context(), bodyParams.SearchTexto)
	// if err != nil {

	// 	logger.Log.Errorf("Erro ao converter a string de busca em embeddings: %v", err)
	// 	response.HandleError(c, http.StatusInternalServerError, "Erro ao converter a string de busca em embeddings!", "", requestID)
	// 	return
	// }

	//Converte os embeddings de float64 para float32, reconhecido pelo OpenSearch
	//vector32 := services.OpenaiServiceGlobal.Float64ToFloat32Slice(rspEmbeddings)
	// documentos, err := handler.idx.ConsultaSemantica(vector32, bodyParams.Natureza)
	// if err != nil {

	// 	logger.Log.Errorf("Erro ao buscar documentos: %v", err)
	// 	response.HandleError(c, http.StatusInternalServerError, "Erro ao buscar documentos!", "", requestID)
	// 	return
	// }

	// msg := "Consulta realizada com sucesso"
	// if len(documentos) == 0 {
	// 	msg = msg + ": nenhum documento retornado!"
	// }

	rsp := gin.H{
		//"docs":    documentos,
		//"message": msg,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
