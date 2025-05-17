package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/msgs"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PromptHandlerType struct {
	service *services.PromptServiceType
}

func NewPromptHandlers(service *services.PromptServiceType) *PromptHandlerType {

	return &PromptHandlerType{service: service}
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

func (obj *PromptHandlerType) InsertHandler(c *gin.Context) {
	requestID := uuid.New().String()
	bodyParams := models.BodyParamsPromptInsert{}

	// decoder := json.NewDecoder(c.Request.Body)
	// if err := decoder.Decode(&bodyParams); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
	// 	return
	// }
	err := c.ShouldBindJSON(&bodyParams)
	if err != nil {
		logger.Log.Error("JSON com Formato inválido", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Formato inválido", "", requestID)
		return
	}

	if bodyParams.IdNat == 0 || bodyParams.IdDoc == 0 || bodyParams.IdClasse == 0 || bodyParams.IdAssunto == 0 {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		// return
		logger.Log.Error("Campos ausentes")
		response.HandleError(c, http.StatusInternalServerError, "Faltam campos obrigatórios", "", requestID)
		return
	}

	row, err := obj.service.InsertPrompt(bodyParams)
	if err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na seleção de sessões!"})
		// return
		logger.Log.Error("Erro na seleção de sessões!", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Erro na seleção de sessões!", "", requestID)
		return
	}
	// response := gin.H{
	// 	"ok":         true,
	// 	"statusCode": http.StatusCreated,
	// 	"message":    "Record successfully inserted!",
	// 	"rows":       ret,
	// }

	// c.JSON(http.StatusCreated, response)
	rsp := gin.H{
		"row":     row,
		"message": "Registro inserido com sucesso!",
	}

	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
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
func (obj *PromptHandlerType) UpdateHandler(c *gin.Context) {

	bodyParams := models.BodyParamsPromptUpdate{}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {

		log.Printf("Dados inválidos!")
		response := msgs.CreateResponseMessage("Dados inválidos!" + err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if bodyParams.IdPrompt == 0 {
		log.Printf("IdPrompt is required!")
		response := msgs.CreateResponseMessage("IdPrompt is required!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	ret, err := obj.service.UpdatePrompt(bodyParams)
	if err != nil {
		log.Printf("Erro na alteração do registro!!")
		response := msgs.CreateResponseMessage("Erro na alteração do registro!")
		c.JSON(http.StatusBadRequest, response)
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

func (obj *PromptHandlerType) DeleteHandler(c *gin.Context) {
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

	ret, err := obj.service.DeletaPrompt(id)
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

func (obj *PromptHandlerType) SelectByIDHandler(c *gin.Context) {

	requestID := uuid.New().String()

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

	row, err := obj.service.SelectById(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Registro nçao encontrado!"})
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
}

func (obj *PromptHandlerType) SelectAllHandler(c *gin.Context) {
	//Generate request ID for tracing
	requestID := uuid.New().String()
	rows, err := obj.service.SelectAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensagem": "Erro na deleção do registro!"})
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Todos os registros retornados com sucesso!",
	}

	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
}
