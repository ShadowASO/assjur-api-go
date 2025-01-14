package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"ocrserver/internal/utils/msgs"
	"ocrserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
)

type SessionsHandlerType struct {
	sessionsModel *models.SessionsModelType
}

func NewSessionsHandlers() *SessionsHandlerType {
	model := models.NewSessionsModel()
	return &SessionsHandlerType{sessionsModel: model}
}

/*
 * Verifica se o refreshToken é valido e caso positivo, gera um novo acessToken.
 *
 * - **Rota**: "/sessions"
 * - **Params**:
 * - **Método**: POST
 * - **Body:
 *		{
*			"UserID": int
*			"Model":  string
*			"PromptTokens": int64
*			"CompletionTokens": int64
*			"TotalTokens":  int64
* 		}
 * - **Resposta**:
 *  	{
 * 			"ok": true,
 * 			"statusCode": 200/401/500,
 * 			"message": string,
 * 			"sessionID": string
 *		}
*/
func (service *SessionsHandlerType) InsertHandler(c *gin.Context) {
	var requestData models.SessionsRow
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	log.Printf("user_id=%v", requestData.Model)

	sessionID, err := service.sessionsModel.InsertSession(requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na inclusão em sessions!"})
		return
	}

	response := msgs.CreateResponseSessionsInsert(true, http.StatusCreated, "Sessão incluída com sucesso", sessionID)
	c.JSON(http.StatusCreated, response)
}

/*
 * Lista todas as sessions cadastradas
 *
 * - **Rota**: "/sessions"
 * - **Params**:
 * - **Método**: GET
 * - **Body:{}
 * - **Resposta**:{
 * 			"message": string,
 * 			"ok": bool,
 *    		"statusCode": 200/400/500,
 * 			"rows": [
 *   			{
 *     				"SessionID": int,
 *     				"UserID": int,
 *     				"Model": "gpt-4o-mini-2024-07-18",
 *     				"PromptTokens": int64,
 *     				"CompletionTokens": int64,
 *     				"TotalTokens": int64,
 *     				"SessionStart": Date,
 *     				"SessionEnd": null
 *   			},]
 *			}
 */
func (service *SessionsHandlerType) SelectAllHandler(c *gin.Context) {

	rows, err := service.sessionsModel.SelectSessions()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na seleção de sessões!"})
		return
	}

	response := msgs.CreateResponseSelectRows(true, http.StatusOK, "Consulta incluído com sucesso", rows)
	c.JSON(http.StatusOK, response)
}

/*
 * Lista todas as sessions cadastradas
 *
 * - **Rota**: "/sessions/:id"
 * - **Params**: int
 * - **Método**: GET
 * - **Body:{}
 * - **Resposta**:{
 * 			"message": string,
 * 			"ok": bool,
 *			"statusCode": 200/400/500,
 * 			"data":
 *   			{
 *     				"SessionID": int,
 *     				"UserID": int,
 *     				"Model": "gpt-4o-mini-2024-07-18",
 *     				"PromptTokens": int64,
 *     				"CompletionTokens": int64,
 *     				"TotalTokens": int64,
 *     				"SessionStart": Date,
 *     				"SessionEnd": null
 *   			},
 *			}
 */
func (service *SessionsHandlerType) SelectHandler(c *gin.Context) {
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

	singleRow, err := service.sessionsModel.SelectSession(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na seleção de sessões!"})
		return
	}

	response := msgs.CreateResponseSelectSingle(true, http.StatusOK, "Consulta incluída com sucesso", singleRow)
	c.JSON(http.StatusOK, response)
}

/*
Atualiza os campos relativos ao uso de tokens
*/

func (service *SessionsHandlerType) UpdateTokensUso(retSubmit *openai.ChatCompletion) error {
	/* Calcula os valores de tokesn */
	var sessionData models.SessionsRow
	sessionData.SessionID = 1
	sessionData.UserID = 1

	currentTokens, err := service.sessionsModel.SelectSession(sessionData.SessionID)
	if err != nil {
		log.Printf("erro ao buscar sessão para atualização")
		return err
	}
	sessionData.PromptTokens = retSubmit.Usage.PromptTokens + currentTokens.PromptTokens
	sessionData.CompletionTokens = retSubmit.Usage.CompletionTokens + currentTokens.CompletionTokens
	sessionData.TotalTokens = retSubmit.Usage.TotalTokens + currentTokens.TotalTokens

	_, err = service.sessionsModel.UpdateSession(sessionData)
	if err != nil {
		log.Printf("UpdateTokensUso: Erro na atualização do uso de tokens!")
	}

	return err
}

/*
 * Devolve os totais de tokens usados
 *
 * - **Rota**: "/sessions/uso"
 * - **Params**:
 * - **Método**: GET
 * - **Body**:
 * - **Resposta**:
 *  	{
 * 			"ok": true,
 * 			"statusCode": 200
 *			"message": "Consulta incluída com sucesso",
 * 			"data": {
 *   			"CompletionTokens": 39354,
 *   			"PromptTokens": 229774,
 *   			"TotalTokens": 269127
 * 		},
 *
 */
func (service *SessionsHandlerType) GetTokenUsoHandler(c *gin.Context) {
	rows, err := service.sessionsModel.SelectSessions()
	if err != nil {
		//c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na seleção de sessões!"})
		msgs.CreateResponseErrorMessage(c, http.StatusBadRequest, "Erro na seleção de sessões!")
		return
	}
	// Inicializa os contadores de tokens
	var pTokens, cTokens, tTokens int64

	for _, row := range rows {
		pTokens += row.PromptTokens
		cTokens += row.CompletionTokens
		tTokens += row.TotalTokens
	}

	// Cria a estrutura de resposta
	respTokens := gin.H{
		"PromptTokens":     pTokens,
		"CompletionTokens": cTokens,
		"TotalTokens":      tTokens,
	}
	response := msgs.CreateResponseSelectSingle(true, http.StatusOK, "Consulta incluída com sucesso", respTokens)
	c.JSON(http.StatusOK, response)
}
