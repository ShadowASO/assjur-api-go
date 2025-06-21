/*
---------------------------------------------------------------------------------------
File: promptHandler.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 17-05-2025
---------------------------------------------------------------------------------------
*/
package handlers

import (
	"net/http"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"strconv"

	"github.com/gin-gonic/gin"
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

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	bodyParams := models.BodyParamsPromptInsert{}

	err := c.ShouldBindJSON(&bodyParams)
	if err != nil {
		logger.Log.Errorf("JSON com Formato inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	if bodyParams.IdNat == 0 || bodyParams.IdDoc == 0 || bodyParams.IdClasse == 0 || bodyParams.IdAssunto == 0 {
		logger.Log.Error("Faltam campos obrigatórios")
		response.HandleError(c, http.StatusBadRequest, "Faltam campos obrigatórios", "", requestID)
		return
	}

	row, err := obj.service.InsertPrompt(bodyParams)
	if err != nil {
		logger.Log.Errorf("Erro na inserção do registro: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na inserção do registro", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro inserido com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
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
	// Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	bodyParams := models.BodyParamsPromptUpdate{}
	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Errorf("Dados inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	if bodyParams.IdPrompt == 0 {

		logger.Log.Error("O campo IdPrompt é obrigatório")
		response.HandleError(c, http.StatusBadRequest, "O campo IdPrompt é obrigatório", "", requestID)
		return
	}

	ret, err := obj.service.UpdatePrompt(bodyParams)
	if err != nil {

		logger.Log.Errorf("Erro na alteração do registro!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na alteração do registro!", "", requestID)
		return
	}
	rsp := gin.H{
		"message": "Record successfully updated!",
		"rows":    ret,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

func (obj *PromptHandlerType) DeleteHandler(c *gin.Context) {
	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		logger.Log.Errorf("ID inválido ou não informado: %v", err)
		response.HandleError(c, http.StatusBadRequest, "ID inválido ou não informado", "", requestID)
		return
	}

	ret, err := obj.service.DeletaPrompt(id)
	if err != nil {

		logger.Log.Errorf("Erro na deleção do registro!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na deleção do registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"message": "registro deletado com sucesso!",
		"rows":    ret,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

func (obj *PromptHandlerType) SelectByIDHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	paramID := c.Param("id")
	if paramID == "" {

		logger.Log.Error("ID da sessão não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID da sessão não informado!", "", requestID)
		return
	}
	id, err := strconv.Atoi(paramID)
	if err != nil {

		logger.Log.Errorf("Dados inválido: %s", err)
		response.HandleError(c, http.StatusBadRequest, "ID inválido!", "", requestID)
		return
	}

	row, err := obj.service.SelectById(id)
	if err != nil {

		logger.Log.Errorf("Erro ao selecionar o registro: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao selecionar o registro", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

func (obj *PromptHandlerType) SelectAllHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	rows, err := obj.service.SelectAll()
	if err != nil {

		logger.Log.Errorf("Erro na deleção do registro!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na deleção do registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Todos os registros retornados com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
