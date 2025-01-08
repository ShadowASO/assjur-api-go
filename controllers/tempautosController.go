package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"ocrserver/lib/tools"
	"ocrserver/models"
	"strconv"
)

type TempautosControllerType struct {
	tempautosModel *models.TempautosModelType
}

func NewTempautosController() *TempautosControllerType {
	return &TempautosControllerType{
		tempautosModel: models.NewTempautosModel(),
	}
}

func (service *TempautosControllerType) InsertHandler(c *gin.Context) {
	var requestData models.TempAutosRow

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {
		//c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		//return
		c.JSON(http.StatusBadRequest, tools.CreateResponse(false, http.StatusBadRequest, "Invalid data provided", nil))
		return
	}

	// Validação de campos obrigatórios
	if requestData.IdCtxt == 0 || requestData.NmFileNew == "" || requestData.NmFileOri == "" || requestData.TxtDoc == "" {
		c.JSON(http.StatusBadRequest, tools.CreateResponse(false, http.StatusBadRequest, "All fields are required", nil))
		return
	}

	// Insere o registro no banco de dados
	ret, err := service.tempautosModel.InsertRow(requestData)
	if err != nil {
		log.Printf("Insert error: %v", err)
		c.JSON(http.StatusInternalServerError, tools.CreateResponse(false, http.StatusInternalServerError, "Failed to insert record", nil))
		return
	}

	c.JSON(http.StatusCreated, tools.CreateResponse(true, http.StatusCreated, "Record successfully inserted", ret))

}

func (service *TempautosControllerType) DeleteHandler(c *gin.Context) {
	paramID := c.Param("id")
	if paramID == "" {
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID da sessão não informado!"})
		// return
		c.JSON(http.StatusBadRequest, tools.CreateResponse(false, http.StatusBadRequest, "ID not provided", nil))
		return
	}
	id, err := strconv.Atoi(paramID)
	if err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID inválido!"})
		// return
		c.JSON(http.StatusBadRequest, tools.CreateResponse(false, http.StatusBadRequest, "Invalid ID format", nil))
		return
	}

	err = service.tempautosModel.DeleteRow(id)
	if err != nil {
		//c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na deleção do registro!"})
		//return
		log.Printf("Delete error: %v", err)
		c.JSON(http.StatusInternalServerError, tools.CreateResponse(false, http.StatusInternalServerError, "Failed to delete record", nil))
		return
	}

	// response := gin.H{
	// 	"ok":         true,
	// 	"statusCode": http.StatusOK,
	// 	"message":    "registro deletado com sucesso!",
	// }

	// c.JSON(http.StatusOK, response)
	c.JSON(http.StatusOK, tools.CreateResponse(true, http.StatusOK, "Record successfully deleted", nil))

}

func (service *TempautosControllerType) SelectByIDHandler(c *gin.Context) {
	paramID := c.Param("id")
	if paramID == "" {
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID da sessão não informado!"})
		// return
		c.JSON(http.StatusBadRequest, tools.CreateResponse(false, http.StatusBadRequest, "ID not provided", nil))
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID inválido!"})
		// return
		c.JSON(http.StatusBadRequest, tools.CreateResponse(false, http.StatusBadRequest, "Invalid ID format", nil))
		return
	}

	ret, err := service.tempautosModel.SelectByIdDoc(id)
	if err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Registro nçao encontrado!"})
		// return
		log.Printf("Select by ID error: %v", err)
		c.JSON(http.StatusNotFound, tools.CreateResponse(false, http.StatusNotFound, "Record not found", nil))
		return
	}
	// response := gin.H{
	// 	"ok":         true,
	// 	"statusCode": http.StatusOK,
	// 	"message":    "registro selecionado com sucesso!",
	// 	"rows":       ret,
	// }

	// c.JSON(http.StatusOK, response)
	c.JSON(http.StatusOK, tools.CreateResponse(true, http.StatusOK, "Record successfully retrieved", ret))

}

/**
 * Devolve os registros da tabela 'temp_autos' para um determinado contexto'
 * Rota: "/contexto/documentos/:id"
 * Params: ID do Contexto
 * Método: GET
 */
func (service *TempautosControllerType) SelectAllHandler(c *gin.Context) {
	ctxtID := c.Param("id")

	if ctxtID == "" {
		c.JSON(http.StatusBadRequest, tools.CreateResponse(false, http.StatusBadRequest, "Context ID not provided", nil))
		return
	}

	idKey, err := strconv.Atoi(ctxtID)
	if err != nil {
		c.JSON(http.StatusBadRequest, tools.CreateResponse(false, http.StatusBadRequest, "Invalid context ID format", nil))
		return
	}

	rows, err := service.tempautosModel.SelectByContexto(idKey)
	if err != nil {
		log.Printf("Select by context error: %v", err)
		c.JSON(http.StatusInternalServerError, tools.CreateResponse(false, http.StatusInternalServerError, "Failed to retrieve records", nil))
		return
	}
	// Verifica se nenhum registro foi encontrado
	if len(rows) == 0 {
		c.JSON(http.StatusOK, tools.CreateResponse(true, http.StatusOK, "No records found for the provided context", nil))
		return
	}
	c.JSON(http.StatusOK, tools.CreateResponse(true, http.StatusOK, "Records successfully retrieved", rows))

}
