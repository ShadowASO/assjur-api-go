package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"ocrserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
)

type SessionsControllerType struct {
	sessionModel *models.SessionsModelType
}

func NewSessionController() *SessionsControllerType {
	model := models.NewSessionsModel()
	return &SessionsControllerType{sessionModel: model}
}

func (service *SessionsControllerType) SelectAllHandler(c *gin.Context) {

	rows, err := service.sessionModel.SelectSessions()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na seleção de sessões!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": rows})
}

func (service *SessionsControllerType) SelectHandler(c *gin.Context) {
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

	//rows, err := models.SessionsModel.SelectSession(id)
	rows, err := service.sessionModel.SelectSession(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na seleção de sessões!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": rows})
}

func (service *SessionsControllerType) InsertHandler(c *gin.Context) {
	var requestData models.SessionsRow
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	log.Printf("user_id=%v", requestData.Model)

	ret, err := service.sessionModel.InsertSession(requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na seleção de sessões!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": ret})
}
func (service *SessionsControllerType) UpdateTokensUso(retSubmit *openai.ChatCompletion) error {
	/* Calcula os valores de tokesn */
	var sessionData models.SessionsRow
	sessionData.SessionID = 1
	sessionData.UserID = 1

	currentTokens, err := service.sessionModel.SelectSession(sessionData.SessionID)
	if err != nil {
		log.Printf("erro ao buscar sessão para atualização")
		return err
	}
	sessionData.PromptTokens = retSubmit.Usage.PromptTokens + currentTokens.PromptTokens
	sessionData.CompletionTokens = retSubmit.Usage.CompletionTokens + currentTokens.CompletionTokens
	sessionData.TotalTokens = retSubmit.Usage.TotalTokens + currentTokens.TotalTokens

	_, err = service.sessionModel.UpdateSession(sessionData)
	if err != nil {
		log.Printf("UpdateTokensUso: Erro na atualização do uso de tokens!")
	}
	log.Printf("Tokens atualizados:\n")
	log.Printf("PromptTokens: %d", sessionData.PromptTokens)
	log.Printf("CompletionTokens: %d", sessionData.CompletionTokens)
	log.Printf("TotalTokens: %d", sessionData.TotalTokens)
	return err
}
