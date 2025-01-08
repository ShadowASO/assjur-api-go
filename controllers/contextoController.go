package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"net/http"
	"ocrserver/models"

	"strconv"

	"github.com/gin-gonic/gin"
)

type ContextoControllerType struct {
	contextoModel *models.ContextoModelType
}

// var PrompService PromptControllerType
// var contextoModel *models.ContextoModelType

func NewContextoController() *ContextoControllerType {
	model := models.NewContextoModel()
	return &ContextoControllerType{contextoModel: model}
}

/**
 * Insere um novo registro de contexto
 * Rota: "/contexto"
 * Método: POST
 * Body: {
	NrProc: string
	Juizo: string
	Classe: string
	Assunto: string
	}
*/

func (service *ContextoControllerType) InsertHandler(c *gin.Context) {
	log.Printf("Entrei")
	//var requestData models.ContextoRow
	bodyParams := models.BodyParamsContextoInsert{}

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {
		log.Printf("Entrei primewiro erro")
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if bodyParams.NrProc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O campo numeroProcesso é obrigatório"})
		return
	}

	/* Verificamos se o processo já existe*/
	log.Printf("bodyParams.NrProc=%v", bodyParams.NrProc)

	isExiste, err := service.contextoModel.RowExists(bodyParams.NrProc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na verificação da existência!"})
		return
	}

	if isExiste {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Processo já existe!"})
		return
	}

	ret, err := service.contextoModel.InsertRow(bodyParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na inclusão do contexto!"})
		response := gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "Erro na inclusão do contexto!",
			"rows":       ret,
		}

		c.JSON(http.StatusCreated, response)
		return
	}

	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusCreated,
		"message":    "Contexto inserido com sucesso!",
		"rows":       ret,
	}
	log.Printf("Contexto inserido com sucesso!")

	c.JSON(http.StatusCreated, response)
}

/*
* Insere um novo registro de contexto
  - Rota: "/contexto"
  - Método: POST
  - Body: {
    IdCtxt           int
    NrProc           string
    Juizo            string
    Classe           string
    Assunto          string
    PromptTokens     int
    CompletionTokens int
    }
*/
func (service *ContextoControllerType) UpdateHandler(c *gin.Context) {
	bodyParams := models.BodyParamsContextoUpdate{}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if bodyParams.NrProc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Número do processo é obrigatório!"})
		return
	}

	//ret, err := models.PromptModel.UpdateReg(requestData)
	ret, err := service.contextoModel.UpdateRow(bodyParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na alteração do registro!"})
		return
	}
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusCreated,
		"message":    "Record successfully updated!",
		"rows":       ret,
	}

	c.JSON(http.StatusOK, response)
}

/**
 * Devolve os dados dos usuários cadastrados na tabela 'users'
 * Rota: "/contexto"
 * Método: DELETE
 */
func (service *ContextoControllerType) DeleteHandler(c *gin.Context) {
	paramID := c.Param("id")
	if paramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID da sessão não informado!"})
		return
	}
	id, err := strconv.Atoi(paramID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID inválido!"})
		return
	}

	//ret, err := models.PromptModel.DeleteReg(id)
	ret, err := service.contextoModel.DeleteReg(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na deleção do registro!"})
		return
	}

	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "registro deletado com sucesso!",
		"rows":       ret,
	}

	c.JSON(http.StatusOK, response)
}

/**
 * Devolve os dados do contexto indicado
 * Rota: "/contexto/:id"
 * Método: GET
 */
func (service *ContextoControllerType) SelectByIDHandler(c *gin.Context) {
	paramID := c.Param("id")
	if paramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID da sessão não informado!"})
		return
	}
	id, err := strconv.Atoi(paramID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID inválido!"})
		return
	}

	//ret, err := models.PromptModel.SelectById(id)
	ret, err := service.contextoModel.SelectContextoById(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Registro nçao encontrado!"})
		return
	}
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "registro selecionado com sucesso!",
		"rows":       ret,
	}

	c.JSON(http.StatusOK, response)
}

/**
 * Devolve os dados do contexto indicado pelo processo
 * Rota: "/contexto/processo/:id"
 * Método: GET
 */
func (service *ContextoControllerType) SelectByProcessoHandler(c *gin.Context) {
	// Obtém o parâmetro "id" da rota
	paramID := c.Param("id")
	if paramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"mensagem":   "ID do processo não informado!",
		})
		return
	}

	//rows, err := models.PromptModel.SelectById(id)
	rows, err := service.contextoModel.SelectContextoByProcesso(paramID)
	if err != nil {
		// Verifica se o erro é de "registro não encontrado"
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"ok":         false,
				"statusCode": http.StatusNotFound,
				"mensagem":   "Nenhum registro encontrado para o processo informado.",
			})
			return
		}

		// Caso contrário, erro interno do servidor
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":         false,
			"statusCode": http.StatusInternalServerError,
			"mensagem":   "Erro ao buscar o registro no banco de dados.",
		})
		return
	}
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "registro selecionado com sucesso!",
		"rows":       rows,
	}

	c.JSON(http.StatusOK, response)
}

/**
 * Devolve os dados de todos os contextos
 * Rota: "/contexto"
 * Método: GET
 */

func (service *ContextoControllerType) SelectAllHandler(c *gin.Context) {
	// Simulate fetching all records
	//ret, err := models.PromptModel.SelectRegs()
	ret, err := service.contextoModel.SelectContextos()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na deleção do registro!"})
		return
	}
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "All records successfully retrieved!",
		"rows":       ret,
	}

	c.JSON(http.StatusOK, response)
}
