package handlers

import (
	"log"
	"net/http"

	"ocrserver/internal/opensearch" // Atualizado para refletir a mudança para OpenSearch
	"ocrserver/internal/services/openAI"

	"ocrserver/internal/utils/msgs"

	"github.com/gin-gonic/gin"
)

// Estrutura do Handler
type ModelosHandlerType struct {
	idx *opensearch.IndexModelosType
}

// Construtor do Handler
func NewModelosHandlers() *ModelosHandlerType {
	index := opensearch.NewIndexModelos()
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
	var bodyParams opensearch.BodyModelosInsert

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos", "erro": err.Error()})
		return
	}

	if bodyParams.Natureza == "" || bodyParams.Ementa == "" || bodyParams.Inteiro_teor == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Todos os campos são obrigatórios: Natureza, Ementa, Inteiro_teor"})
		return
	}

	rsp, err := handler.idx.GetDocumentoEmbeddings(opensearch.ModelosText(bodyParams))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao extrair os embeddings do documento!", "erro": err.Error()})
		return
	}

	doc, err := handler.idx.IndexaDocumento(rsp)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao inserir documento!", "erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ok":         true,
		"statusCode": http.StatusCreated,
		"message":    "Documento inserido com sucesso!",
		//"response":   res,
		"response": doc,
	})
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
	idDoc := c.Param("id")
	var bodyParams opensearch.ModelosText

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos", "erro": err.Error()})
		return
	}

	if idDoc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Id do documento é obrigatório!"})
		return
	}

	res, err := handler.idx.UpdateDocumento(idDoc, bodyParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao atualizar documento!", "erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Documento atualizado com sucesso!",
		"response":   res,
	})
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
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID do documento não informado!"})
		return
	}

	res, err := handler.idx.DeleteDocumento(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao deletar documento!", "erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Documento deletado com sucesso!",
		"response":   res,
	})
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
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID do documento não informado!"})
		return
	}

	documento, err := handler.idx.ConsultaDocumentoById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao buscar documento!", "erro": err.Error()})
		return
	}

	if documento == nil {
		c.JSON(http.StatusNotFound, gin.H{"mensagem": "Documento não encontrado!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"docs": documento,
	})
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
	bodyParams := BodySearchModelos{}
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		log.Printf("Dados inválidos: %v", err)
		response := msgs.CreateResponseMessage("Dados inválidos: " + err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if bodyParams.Index_name == "" || bodyParams.Natureza == "" || bodyParams.Search_texto == "" {
		log.Printf("Todos os campos são obrigatórios no corpo da mensagem!")
		response := msgs.CreateResponseMessage("IndexName, Natureza e SearchText são obrigatórios no corpo da mensagem!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	//CONVERTE A STRING DE BUSCA EM EMBEDDINGS DA OPENAI
	rspEmbeddings, err := openAI.Service.GetEmbeddingFromText(bodyParams.Search_texto)
	if err != nil {
		response := msgs.CreateResponseMessage("Erro ao converter a string de busca em embeddings!")
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	//Converte os embeddings de float64 para float32, reconhecido pelo OpenSearch
	vector32 := openAI.Float64ToFloat32Slice(rspEmbeddings.Data[0].Embedding)
	//----------------------------------------------------------------------------

	//var documentos []opensearch.ResponseModelos

	documentos, err := handler.idx.ConsultaSemantica(vector32, bodyParams.Natureza)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao buscar documentos!", "erro": err.Error()})
		return
	}

	if len(documentos) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"mensagem": "Nenhum documento encontrado!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"docs": documentos,
	})
}
