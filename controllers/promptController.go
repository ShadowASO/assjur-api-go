package controllers

import (
	"encoding/json"
	//"log"
	"net/http"
	"ocrserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PromptControllerType struct{}

// var PrompService PromptControllerType
var promptModel *models.PromptModelType

func NewPromptController() *PromptControllerType {
	promptModel = models.NewPromptModel()
	return &PromptControllerType{}
}

func (service *PromptControllerType) InsertHandler(c *gin.Context) {
	var requestData models.PromptRow
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if requestData.IdNat == 0 || requestData.IdDoc == 0 || requestData.IdClasse == 0 || requestData.IdAssunto == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	//ret, err := models.PromptModel.InsertReg(requestData)
	ret, err := promptModel.InsertReg(requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na seleção de sessões!"})
		return
	}
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusCreated,
		"message":    "Record successfully inserted!",
		"rows":       ret,
	}

	c.JSON(http.StatusCreated, response)
}

func (service *PromptControllerType) UpdateHandler(c *gin.Context) {
	var requestData models.PromptRow
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if requestData.IdPrompt == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IdPrompt is required"})
		return
	}

	//ret, err := models.PromptModel.UpdateReg(requestData)
	ret, err := promptModel.UpdateReg(requestData)
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

func (service *PromptControllerType) DeleteHandler(c *gin.Context) {
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
	ret, err := promptModel.DeleteReg(id)
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

func (service *PromptControllerType) SelectByIDHandler(c *gin.Context) {
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
	ret, err := promptModel.SelectById(id)
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

func (service *PromptControllerType) SelectAllHandler(c *gin.Context) {
	// Simulate fetching all records
	//ret, err := models.PromptModel.SelectRegs()
	ret, err := promptModel.SelectRegs()
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
