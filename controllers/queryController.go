package controllers

import (
	"net/http"

	"ocrserver/models"
	"ocrserver/services/openAI"

	"github.com/gin-gonic/gin"
)

type QueryControllerType struct {
	sessionModel *models.SessionsModelType
}

func NewQueryController() *QueryControllerType {
	model := models.NewSessionsModel()
	return &QueryControllerType{sessionModel: model}
}

func (service *QueryControllerType) QueryHandler(c *gin.Context) {
	var messages openAI.MsgGpt
	// Extrai os dados do corpo da requisição
	if err := c.ShouldBindJSON(&messages); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Invalid request body"})
		return
	}

	if len(messages.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Messages array cannot be empty"})
		return
	}

	retSubmit, err := openAI.Service.SubmitPrompt(messages)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro no SubmitPrompt"})
		return
	}
	/* Atualiza o uso de tokens na tabela 'sessions' */

	sessionService := NewSessionController()
	err = sessionService.UpdateTokensUso(retSubmit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na atualização do uso de tokens!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": retSubmit})

}
