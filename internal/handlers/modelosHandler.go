package handlers

import (
	"log"
	"net/http"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/opensearch" // Atualizado para refletir a mudança para OpenSearch
	"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Estrutura do Handler
type ModelosHandlerType struct {
	idx *opensearch.IndexModelosType
}

// Construtor do Handler
func NewModelosHandlers(index *opensearch.IndexModelosType) *ModelosHandlerType {
	//index := opensearch.NewIndexModelos()
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
	requestID := uuid.New().String()
	var bodyParams opensearch.BodyModelosInsert

	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Error("Formato inválido", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao realizar busca pelo contexto", "", requestID)
		return
	}

	if bodyParams.Natureza == "" || bodyParams.Ementa == "" || bodyParams.Inteiro_teor == "" {

		logger.Log.Error("Todos os campos são obrigatórios: Natureza, Ementa, Inteiro_teor")
		response.HandleError(c, http.StatusBadRequest, "Todos os campos são obrigatórios: Natureza, Ementa, Inteiro_teor", "", requestID)
		return
	}
	log.Println(bodyParams)
	emb, err := handler.idx.GetDocumentoEmbeddings(opensearch.ModelosText(bodyParams))
	if err != nil {

		logger.Log.Error("Erro ao extrair os embeddings do documento!", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao extrair os embeddings do documento!", "", requestID)
		return
	}

	doc, err := handler.idx.IndexaDocumento(emb)

	if err != nil {

		logger.Log.Error("Erro ao inserir documento!")
		response.HandleError(c, http.StatusBadRequest, "Erro ao inserir documento!", "", requestID)
		return
	}

	rsp := gin.H{
		"doc":     doc,
		"message": "Registro selecionado com sucesso!",
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
	requestID := uuid.New().String()
	idDoc := c.Param("id")
	var bodyParams opensearch.ModelosText

	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Error("Formato inválido", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	if idDoc == "" {

		logger.Log.Error("Id do documento é obrigatório!")
		response.HandleError(c, http.StatusBadRequest, "Id do documento é obrigatório!", "", requestID)
		return
	}

	doc, err := handler.idx.UpdateDocumento(idDoc, bodyParams)
	if err != nil {

		logger.Log.Error("Erro ao atualizar documento!")
		response.HandleError(c, http.StatusBadRequest, "Erro ao atualizar documento!", "", requestID)
		return
	}

	rsp := gin.H{
		"doc":     doc,
		"message": "Registro selecionado com sucesso!",
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
	requestID := uuid.New().String()
	id := c.Param("id")
	if id == "" {

		logger.Log.Error("ID do documento não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID do documento não informado!", "", requestID)
		return
	}

	doc, err := handler.idx.DeleteDocumento(id)
	if err != nil {

		logger.Log.Error("Erro ao deletar documento!")
		response.HandleError(c, http.StatusBadRequest, "Erro ao deletar documento!", "", requestID)
		return
	}

	rsp := gin.H{
		"doc":     doc,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

/*
  - Deleta  um documento existente no Elasticsearch
    *Rota: "/tabelas/elastic/:{id}"
  - Método: GET
  - Body: {
    }
*/
// Busca um documento pelo ID no OpenSearch
func (handler *ModelosHandlerType) SelectByIdHandler(c *gin.Context) {
	requestID := uuid.New().String()
	id := c.Param("id")
	if id == "" {

		logger.Log.Error("ID do documento não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID do documento não informado!", "", requestID)
		return
	}

	documento, err := handler.idx.ConsultaDocumentoById(id)
	if err != nil {

		logger.Log.Error("Erro ao buscar documento!")
		response.HandleError(c, http.StatusBadRequest, "Erro ao buscar documento!", "", requestID)
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
	Index_name   string `json:"index_name"`
	Natureza     string `json:"natureza"`
	Search_texto string `json:"search_texto"`
}

// Busca documentos pelo conteúdo no OpenSearch
func (handler *ModelosHandlerType) SearchModelosHandler(c *gin.Context) {
	requestID := uuid.New().String()
	bodyParams := BodySearchModelos{}
	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Error("Formato inválido", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	if bodyParams.Index_name == "" || bodyParams.Natureza == "" || bodyParams.Search_texto == "" {

		logger.Log.Error("IndexName, Natureza e SearchText são obrigatórios no corpo da mensagem!")
		response.HandleError(c, http.StatusBadRequest, "IndexName, Natureza e SearchText são obrigatórios no corpo da mensagem!", "", requestID)
		return
	}

	//CONVERTE A STRING DE BUSCA EM EMBEDDINGS DA OPENAI
	//rspEmbeddings, err := openAI.OpenAIServiceGlobal.GetEmbeddingFromText(bodyParams.Search_texto)
	rspEmbeddings, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(bodyParams.Search_texto)
	if err != nil {

		logger.Log.Error("Erro ao converter a string de busca em embeddings!", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao converter a string de busca em embeddings!", "", requestID)
		return
	}
	//Converte os embeddings de float64 para float32, reconhecido pelo OpenSearch
	vector32 := services.Float64ToFloat32Slice(rspEmbeddings.Data[0].Embedding)
	//----------------------------------------------------------------------------

	documentos, err := handler.idx.ConsultaSemantica(vector32, bodyParams.Natureza)

	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao buscar documentos!", "erro": err.Error()})
		// return
		logger.Log.Error("Erro ao buscar documentos!", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao buscar documentos!", "", requestID)
		return
	}

	if len(documentos) == 0 {

		logger.Log.Error("Nenhum documento encontrado!", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Nenhum documento encontrado!", "", requestID)
		return
	}

	rsp := gin.H{
		"docs":    documentos,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
