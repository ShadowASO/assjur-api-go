package controllers

import (
	"encoding/json"
	"net/http"
	"ocrserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PromptControllerType struct {
	promptModel *models.PromptModelType
}

func NewPromptController() *PromptControllerType {
	model := models.NewPromptModel()
	return &PromptControllerType{promptModel: model}
}

/*
  - Insere um novo prompt na tabela 'prompts'
    *Rota: "/tabelas/prompt"
  - Método: POST
  - Body: {
    "IdNat": int
    "IdDoc": int
    "IdClasse": int
    "IdAssunto": int
    "NmDesc": string
    "TxtPrompt": string
    }
*/

func (service *PromptControllerType) InsertHandler(c *gin.Context) {
	bodyParams := models.BodyParamsPromptInsert{}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if bodyParams.IdNat == 0 || bodyParams.IdDoc == 0 || bodyParams.IdClasse == 0 || bodyParams.IdAssunto == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	ret, err := service.promptModel.InsertReg(bodyParams)
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

/*
  - Modifica o registro na tabela 'prompts'
    *Rota: "/tabelas/prompt"
  - Método: PUT
  - Body: {
    "IdPrompt": int
    "NmDesc": string
    "TxtPrompt": string
    }
*/
func (service *PromptControllerType) UpdateHandler(c *gin.Context) {

	bodyParams := models.BodyParamsPromptUpdate{}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if bodyParams.IdPrompt == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IdPrompt is required"})
		return
	}

	ret, err := service.promptModel.UpdateReg(bodyParams)
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

	ret, err := service.promptModel.DeleteReg(id)
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

	ret, err := service.promptModel.SelectById(id)
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

	ret, err := service.promptModel.SelectRegs()
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
