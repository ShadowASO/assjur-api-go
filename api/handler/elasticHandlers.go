package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"ocrserver/internal/elastic"
	"ocrserver/internal/utils/msgs"

	"github.com/gin-gonic/gin"
)

// Estrutura do Handler
type ElasticHandlerType struct {
	cliente *elastic.ElasticClienteType
}

// Construtor do Handler
func NewElasticHandlers() *ElasticHandlerType {
	cli := elastic.NewElasticCliente()
	return &ElasticHandlerType{cliente: cli}
}

/*
  - Insere um novo documento no Elasticsearch
    *Rota: "/tabelas/elastic"
- Método: POST
  - Body: {
    Natureza string `json:"natureza"`
    Ementa     string `json:"ementa"`
    Inteiro_teor string `json:"inteiro_teor"`

    }
*/

func (handler *ElasticHandlerType) InsertHandler(c *gin.Context) {
	var bodyParams elastic.ModelosDoc

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos", "erro": err.Error()})
		return
	}

	if bodyParams.Natureza == "" || bodyParams.Ementa == "" || bodyParams.Inteiro_teor == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Todos os campos são obrigatórios"})
		return
	}
	// log.Println(bodyParams)
	res, err := handler.cliente.IndexDocumento("modelos", bodyParams)
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
    *Rota: "/tabelas/elastic/{id}"
  - Método: PUT
  - Body: {
    Natureza     string `json:"natureza"`
    Ementa       string `json:"ementa"`
    Inteiro_teor string `json:"inteiro_teor"`
    }
*/
func (handler *ElasticHandlerType) UpdateHandler(c *gin.Context) {
	idDoc := c.Param("id")
	var bodyParams elastic.ModelosDoc

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos", "erro": err.Error()})
		return
	}

	if idDoc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Id do documento é obrigatório!"})
		return
	}

	res, err := handler.cliente.UpdateDocumento("modelos", idDoc, bodyParams)
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
    *Rota: "/tabelas/elastic/:{id}"
  - Método: DELETE
  - Body: {
    }
*/

func (handler *ElasticHandlerType) DeleteHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("%s", id)
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID do documento não informado!"})
		return
	}

	res, err := handler.cliente.DeleteDocumento("modelos", id)
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

// Handler para buscar um documento pelo ID no Elasticsearch
func (handler *ElasticHandlerType) SelectByIDHandler(c *gin.Context) {
	id := c.Param("id")
	log.Printf("ID=%s", id)
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID do documento não informado!"})
		return
	}

	documento, err := handler.cliente.ConsultaDocumento("modelos", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao buscar documento!", "erro": err.Error()})
		return
	}

	if documento == nil {
		c.JSON(http.StatusNotFound, gin.H{"mensagem": "Documento não encontrado!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Documento encontrado!",
		"documento":  documento,
	})
}

/*
  - Seleciona documentos que contenham o conteúdo Search_texto
    *Rota: "/tabelas/elastic/search"
  - Método: POST
  - Body: {
		Index_name   string `json:"index_name"`
		Search_texto string `json:"search_texto"`
    }
*/
// Estruturas para inserção e atualização
type BodyElasticSearch struct {
	Index_name string `json:"index_name"`
	//Ementa       string `json:"ementa"`
	Search_texto string `json:"search_texto"`
}

// Handler para buscar documentos pelo conteúdo no Elasticsearch
func (handler *ElasticHandlerType) SearchByContentHandler(c *gin.Context) {
	bodyParams := BodyElasticSearch{}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {

		log.Printf("Dados inválidos!")
		response := msgs.CreateResponseMessage("Dados inválidos!" + err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if bodyParams.Index_name == "" {
		log.Printf("Index_name é obrigatório no corpo da mensagtem!")
		response := msgs.CreateResponseMessage("IndexName é obrigatório no corpo da mensagtem!")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	if bodyParams.Search_texto == "" {
		log.Printf("Search_texto é obrigatório no corpo da mensagtem!")
		response := msgs.CreateResponseMessage("SearchText é obrigatório no corpo da mensagtem!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	//documentos, err := handler.cliente.ConsultaPorConteudo(bodyParams.Index_name, bodyParams.Search_texto, bodyParams.Ementa)
	documentos, err := handler.cliente.ConsultaPorConteudo(bodyParams.Index_name, bodyParams.Search_texto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro ao buscar documentos!", "erro": err.Error()})
		return
	}

	if len(documentos) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"mensagem": "Nenhum documento encontrado!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"docs": documentos,
	})

}

// Handler para buscar todos os documentos no Elasticsearch
func (handler *ElasticHandlerType) SelectAllHandler(c *gin.Context) {
	documentos, err := handler.cliente.ConsultaPorConteudo("sentenca", "")
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
