package handlers

import (
	"net/http"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/opensearch" // Atualizado para refletir a mudança para OpenSearch
	"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

// Estrutura do Handler
type ModelosHandlerType struct {
	idx *opensearch.IndexModelosType
}

// Construtor do Handler
func NewModelosHandlers(index *opensearch.IndexModelosType) *ModelosHandlerType {

	return &ModelosHandlerType{idx: index}
}

/*
  - Insere um novo documento no Elasticsearch
    *Rota: "/tabelas/modelos"

- Método: POST

  - Body: {
    Natureza string `json:"natureza"`
    Ementa     string `json:"ementa"`
    Inteiro_teor string `json:"inteiro_teor"`

    }
*/

// Insere um novo documento no OpenSearch
func (handler *ModelosHandlerType) InsertHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	var bodyParams opensearch.BodyModelosInsert

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		logger.Log.Errorf("Dados inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	if bodyParams.Natureza == "" || bodyParams.Ementa == "" || bodyParams.Inteiro_teor == "" {

		logger.Log.Error("Todos os campos são obrigatórios: Natureza, Ementa, Inteiro_teor")
		response.HandleError(c, http.StatusBadRequest, "Todos os campos são obrigatórios: Natureza, Ementa, Inteiro_teor", "", requestID)
		return
	}
	//log.Println(bodyParams)
	ementaVector, err := services.GetDocumentoEmbeddings(bodyParams.Ementa)
	if err != nil {
		logger.Log.Errorf("Erro ao extrair os embeddings do documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao extrair os embeddings do documento!", "", requestID)
		return
	}
	teorVector, err := services.GetDocumentoEmbeddings(bodyParams.Inteiro_teor)
	if err != nil {
		logger.Log.Errorf("Erro ao extrair os embeddings do documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao extrair os embeddings do documento!", "", requestID)
		return
	}
	emb := opensearch.ModelosEmbedding{
		Natureza:             bodyParams.Natureza,
		Ementa:               bodyParams.Ementa,
		Inteiro_teor:         bodyParams.Inteiro_teor,
		EmentaEmbedding:      ementaVector,
		InteiroTeorEmbedding: teorVector,
	}

	resp, err := handler.idx.IndexaDocumento(emb)

	// if err != nil {

	// 	logger.Log.Errorf("Erro ao inserir documento: %v", err)
	// 	response.HandleError(c, http.StatusInternalServerError, "Erro ao inserir documento!", "", requestID)
	// 	return
	// }

	rsp := gin.H{
		"id":      resp.ID,
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
func (handler *ModelosHandlerType) UpdateHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	idDoc := c.Param("id")
	var bodyParams opensearch.ModelosText

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

	doc, err := handler.idx.UpdateDocumento(idDoc, bodyParams)
	if err != nil {
		logger.Log.Errorf("Erro ao atualizar documento: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao atualizar documento!", "", requestID)
		return
	}

	rsp := gin.H{
		"doc":     doc,
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
func (handler *ModelosHandlerType) DeleteHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	id := c.Param("id")
	if id == "" {

		logger.Log.Error("ID do documento não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID do documento não informado!", "", requestID)
		return
	}

	_, err := handler.idx.DeleteDocumento(id)
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
func (handler *ModelosHandlerType) SelectByIdHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	id := c.Param("id")
	if id == "" {

		logger.Log.Error("ID do documento não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID do documento não informado!", "", requestID)
		return
	}

	documento, err := handler.idx.ConsultaDocumentoById(id)
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
type BodySearchModelos struct {
	IndexName   string `json:"index_name"`
	Natureza    string `json:"natureza"`
	SearchTexto string `json:"search_texto"`
}

// Busca documentos pelo conteúdo no OpenSearch
func (handler *ModelosHandlerType) SearchModelosHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	bodyParams := BodySearchModelos{}
	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Errorf("Formato inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	if bodyParams.IndexName == "" || bodyParams.Natureza == "" || bodyParams.SearchTexto == "" {

		logger.Log.Error("index_name, natureza e search_texto são obrigatórios")
		response.HandleError(c, http.StatusBadRequest, "index_name, natureza e search_texto são obrigatórios", "", requestID)
		return
	}

	//Converte a string de busca num embedding
	rspEmbeddings, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(c.Request.Context(), bodyParams.SearchTexto)
	if err != nil {

		logger.Log.Errorf("Erro ao converter a string de busca em embeddings: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao converter a string de busca em embeddings!", "", requestID)
		return
	}

	//Converte os embeddings de float64 para float32, reconhecido pelo OpenSearch
	vector32 := services.OpenaiServiceGlobal.Float64ToFloat32Slice(rspEmbeddings)
	docs, err := handler.idx.ConsultaSemantica(vector32, bodyParams.Natureza)
	if err != nil {

		logger.Log.Errorf("Erro ao buscar documentos: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao buscar documentos!", "", requestID)
		return
	}

	msg := "Consulta realizada com sucesso"
	if len(docs) == 0 {
		msg = msg + ": nenhum documento retornado!"
	}

	rsp := gin.H{
		"docs":    docs,
		"message": msg,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
