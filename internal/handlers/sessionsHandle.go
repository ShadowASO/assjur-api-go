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
	"strconv"
	"strings"

	"ocrserver/internal/models/postgres"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type SessionsHandlerType struct {
	Model *postgres.SessionsModelType
}

func NewSessionsHandlers(service *services.SessionServiceType) *SessionsHandlerType {
	modelo, err := service.GetSessionModel()
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao obter SessionsModel: %v", err)
		return nil
	}

	return &SessionsHandlerType{Model: modelo}
}

/*
 * Verifica se o refreshToken é válido e, caso positivo, gera um novo accessToken.
 *
 * - **Rota**: "/sessions"
 * - **Método**: POST
 * - **Body:
 *		{
 *			"UserID": int
 *			"Model":  string
 *			"PromptTokens": int64
 *			"CompletionTokens": int64
 *			"TotalTokens":  int64
 * 		}
 */
func (service *SessionsHandlerType) InsertHandler(c *gin.Context) {
	var requestData postgres.SessionsRow

	if err := c.ShouldBindJSON(&requestData); err != nil {
		mslogger.LoggerGlobal.Errorf("Dados inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	sessionID, err := service.Model.InsertSession(requestData)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na inclusão em sessions: %s", err.Error())

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro na inclusão em sessions",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"session_id": sessionID,
	}

	msresponse.OK(c, http.StatusCreated, "Sessão incluída com sucesso", rsp)
}

/*
 * Lista todas as sessions cadastradas
 *
 * - **Rota**: "/sessions"
 * - **Método**: GET
 */
func (service *SessionsHandlerType) SelectAllHandler(c *gin.Context) {
	rows, err := service.Model.SelectSessions()
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na seleção de sessões: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro na seleção de sessões",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"rows": rows,
	}

	msresponse.OK(c, http.StatusOK, "Sessões selecionadas com sucesso", rsp)
}

/*
 * Lista uma session pelo ID
 *
 * - **Rota**: "/sessions/:id"
 * - **Método**: GET
 */
func (service *SessionsHandlerType) SelectHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID da sessão não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID da sessão não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("ID inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	singleRow, err := service.Model.SelectSession(id)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na seleção de sessão: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro na seleção de sessão",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"row": singleRow,
	}

	msresponse.OK(c, http.StatusOK, "Sessão selecionada com sucesso", rsp)
}

/*
Atualiza os campos relativos ao uso de tokens
*/

// func (service *SessionsHandlerType) UpdateTokensUso(retSubmit *openai.ChatCompletion) error {
// 	/* Calcula os valores de tokens */
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
 * - **Método**: GET
 */
func (service *SessionsHandlerType) GetTokenUsoHandler(c *gin.Context) {
	rows, err := service.Model.SelectSessions()
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na seleção de sessões: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro na seleção de sessões",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

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

	msresponse.OK(c, http.StatusOK, "Uso de tokens selecionado com sucesso", rsp)
}
