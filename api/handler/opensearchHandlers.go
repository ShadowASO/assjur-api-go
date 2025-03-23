package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"ocrserver/internal/config"
	"ocrserver/internal/opensearch" // Atualizado para refletir a mudança para OpenSearch
	"ocrserver/internal/utils/msgs"

	"github.com/gin-gonic/gin"
)

// Estrutura do Handler
type OpenSearchHandlerType struct {
	cliente *opensearch.OpenSearchClienteType
}

//const NM_INDEX_MODELOS = "ml-modelos-msmarco"

// Construtor do Handler
func NewOpenSearchHandlers() *OpenSearchHandlerType {
	cli := opensearch.NewOpenSearchCliente()
	return &OpenSearchHandlerType{cliente: cli}
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
func (handler *OpenSearchHandlerType) InsertHandler(c *gin.Context) {
	var bodyParams opensearch.ModelosDoc

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos", "erro": err.Error()})
		return
	}

	if bodyParams.Natureza == "" || bodyParams.Ementa == "" || bodyParams.Inteiro_teor == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Todos os campos são obrigatórios: Natureza, Ementa, Inteiro_teor"})
		return
	}

	res, err := handler.cliente.IndexDocumento(config.OpenSearchIndexName, bodyParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao inserir documento!", "erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ok":         true,
		"statusCode": http.StatusCreated,
		"message":    "Documento inserido com sucesso!",
		"response":   res,
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
func (handler *OpenSearchHandlerType) UpdateHandler(c *gin.Context) {
	idDoc := c.Param("id")
	var bodyParams opensearch.ModelosDoc

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos", "erro": err.Error()})
		return
	}

	if idDoc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Id do documento é obrigatório!"})
		return
	}

	res, err := handler.cliente.UpdateDocumento(config.OpenSearchIndexName, idDoc, bodyParams)
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
func (handler *OpenSearchHandlerType) DeleteHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID do documento não informado!"})
		return
	}

	res, err := handler.cliente.DeleteDocumento(config.OpenSearchIndexName, id)
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
func (handler *OpenSearchHandlerType) SelectByIDHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID do documento não informado!"})
		return
	}

	documento, err := handler.cliente.ConsultaDocumento(config.OpenSearchIndexName, id)
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
    *Rota: "/tabelas/elastic/search"
  - Método: POST
  - Body: {
		Index_name   string `json:"index_name"`
		Natureza     string `json:"natureza"`
		Search_texto string `json:"search_texto"`
    }
*/
// Estrutura para busca no OpenSearch
type BodyOpenSearch struct {
	Index_name   string `json:"index_name"`
	Natureza     string `json:"natureza"`
	Search_texto string `json:"search_texto"`
}

// Busca documentos pelo conteúdo no OpenSearch
func (handler *OpenSearchHandlerType) SearchByContentHandler(c *gin.Context) {
	bodyParams := BodyOpenSearch{}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {
		log.Printf("Dados inválidos!")
		response := msgs.CreateResponseMessage("Dados inválidos!" + err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if bodyParams.Index_name == "" || bodyParams.Natureza == "" || bodyParams.Search_texto == "" {
		log.Printf("Todos os campos são obrigatórios no corpo da mensagem!")
		response := msgs.CreateResponseMessage("IndexName, Natureza e SearchText são obrigatórios no corpo da mensagem!")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	var documentos []opensearch.ModelosResponse
	var err error
	if config.ApplicationMode == "development" {
		documentos, err = handler.cliente.ConsultaSemantica(config.OpenSearchIndexName, bodyParams.Search_texto, bodyParams.Natureza)
	} else {
		documentos, err = handler.cliente.ConsultaPorConteudo(config.OpenSearchIndexName, bodyParams.Search_texto, bodyParams.Natureza)
	}

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

// Busca todos os documentos no OpenSearch
func (handler *OpenSearchHandlerType) SelectAllHandler(c *gin.Context) {
	documentos, err := handler.cliente.ConsultaPorConteudo("sentenca", "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao buscar documentos!", "erro": err.Error()})
		return
	}

	if len(documentos) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"mensagem": "Nenhum documento encontrado!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Todos os documentos recuperados com sucesso!",
		"documentos": documentos,
	})
}
