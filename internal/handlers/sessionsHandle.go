/*
---------------------------------------------------------------------------------------
File: sessionHandler.go
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

type SessionsHandlerType struct {
	Model *models.SessionsModelType
}

func NewSessionsHandlers(service *services.SessionServiceType) *SessionsHandlerType {
	modelo, err := service.GetSessionModel()
	if err != nil {
		logger.Log.Errorf("Erro ao ao obter usersModel: %v", err)
		return nil
	}
	return &SessionsHandlerType{Model: modelo}
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

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	var requestData models.SessionsRow
	if err := c.ShouldBindJSON(&requestData); err != nil {

		logger.Log.Errorf("Dados inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	sessionID, err := service.Model.InsertSession(requestData)
	if err != nil {

		logger.Log.Errorf("Erro na inclusão em sessions!", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na inclusão em sessions!", "", requestID)
		return
	}
	rsp := gin.H{
		"message":   "Usuário incluído com sucesso",
		"sessionID": sessionID,
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
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

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	rows, err := service.Model.SelectSessions()
	if err != nil {

		logger.Log.Errorf("Erro na seleção de sessões: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na seleção de sessões!", "", requestID)
		return
	}
	rsp := gin.H{
		"rows": rows,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
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

		logger.Log.Errorf("ID inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "ID inválido!", "", requestID)
		return
	}

	singleRow, err := service.Model.SelectSession(id)

	if err != nil {

		logger.Log.Errorf("Erro na seleção de sessão: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na seleção de sessões!", "", requestID)
		return
	}

	rsp := gin.H{
		"row": singleRow,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/*
Atualiza os campos relativos ao uso de tokens
*/

// func (service *SessionsHandlerType) UpdateTokensUso(retSubmit *openai.ChatCompletion) error {
// 	/* Calcula os valores de tokesn */
// 	var sessionData models.SessionsRow
// 	sessionData.SessionID = 1
// 	sessionData.UserID = 1

// 	currentTokens, err := service.sessionsModel.SelectSession(sessionData.SessionID)
// 	if err != nil {
// 		log.Printf("erro ao buscar sessão para atualização")
// 		return err
// 	}
// 	sessionData.PromptTokens = retSubmit.Usage.PromptTokens + currentTokens.PromptTokens
// 	sessionData.CompletionTokens = retSubmit.Usage.CompletionTokens + currentTokens.CompletionTokens
// 	sessionData.TotalTokens = retSubmit.Usage.TotalTokens + currentTokens.TotalTokens

// 	_, err = service.sessionsModel.UpdateSession(sessionData)
// 	if err != nil {
// 		log.Printf("UpdateTokensUso: Erro na atualização do uso de tokens!")
// 	}

// 	return err
// }

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

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	rows, err := service.Model.SelectSessions()
	if err != nil {
		logger.Log.Errorf("Erro na seleção de sessões: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na seleção de sessões!", "", requestID)
		return
	}
	// Inicializa os contadores de tokens
	var pTokens, cTokens, tTokens int64

	for _, row := range rows {
		pTokens += row.PromptTokens
		cTokens += row.CompletionTokens
		tTokens += row.TotalTokens
	}

	rsp := gin.H{
		"prompt_tokens":     pTokens,
		"completion_tokens": cTokens,
		"total_tokens":      tTokens,
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}
