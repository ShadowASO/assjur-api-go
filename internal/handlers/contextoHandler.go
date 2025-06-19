package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"

	"net/http"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/utils/logger"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ContextoHandlerType struct {
	contextoModel *models.ContextoModelType
}

func NewContextoHandlers(model *models.ContextoModelType) *ContextoHandlerType {
	return &ContextoHandlerType{contextoModel: model}
}

/**
 * Insere um novo registro de contexto
 * Rota: "/contexto"
 * Método: POST
 * Body: {
	NrProc: string
	Juizo: string
	Classe: string
	Assunto: string
	}
*/

func (service *ContextoHandlerType) InsertHandler(c *gin.Context) {
	requestID := uuid.New().String()

	bodyParams := models.BodyParamsContextoInsert{}

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if bodyParams.NrProc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O campo numeroProcesso é obrigatório"})
		return
	}

	isExiste, err := service.contextoModel.RowExists(bodyParams.NrProc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na verificação da existência!"})
		return
	}

	if isExiste {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Processo já existe!"})
		return
	}

	row, err := service.contextoModel.InsertRow(bodyParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na inclusão do contexto!"})
		response := gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "Erro na inclusão do contexto!",
			"rows":       row,
		}

		c.JSON(http.StatusCreated, response)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro deletado com sucesso!",
	}

	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
}

/*
* Insere um novo registro de contexto
  - Rota: "/contexto"
  - Método: POST
  - Body: {
    IdCtxt           int
    NrProc           string
    Juizo            string
    Classe           string
    Assunto          string
    PromptTokens     int
    CompletionTokens int
    }
*/
func (service *ContextoHandlerType) UpdateHandler(c *gin.Context) {
	requestID := uuid.New().String()
	bodyParams := models.BodyParamsContextoUpdate{}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if bodyParams.NrProc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Número do processo é obrigatório!"})
		return
	}

	row, err := service.contextoModel.UpdateRow(bodyParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na alteração do registro!"})
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro deletado com sucesso!",
	}

	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
}

/**
 * Devolve os dados dos usuários cadastrados na tabela 'users'
 * Rota: "/contexto"
 * Método: DELETE
 */
func (service *ContextoHandlerType) DeleteHandler(c *gin.Context) {
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

	row, err := service.contextoModel.DeleteReg(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na deleção do registro!"})
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro deletado com sucesso!",
	}

	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
}

/**
 * Devolve os dados do contexto indicado
 * Rota: "/contexto/:id"
 * Método: GET
 */
func (service *ContextoHandlerType) SelectByIDHandler(c *gin.Context) {
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

	row, err := service.contextoModel.SelectContextoById(id)
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

/**
 * Devolve os dados do contexto indicado pelo processo
 * Rota: "/contexto/processo/:id"
 * Método: GET
 */
func (service *ContextoHandlerType) SelectByProcessoHandler(c *gin.Context) {
	requestID := uuid.New().String()

	// Obtém o parâmetro "id" da rota
	paramID := c.Param("id")
	if paramID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"mensagem":   "ID do processo não informado!",
		})
		return
	}

	row, err := service.contextoModel.SelectContextoByProcesso(paramID)
	if err != nil {
		// Verifica se o erro é de "registro não encontrado"
		if errors.Is(err, sql.ErrNoRows) {

			response.HandleError(c, http.StatusNotFound, "Nenhum registro encontrado para o processo informado", "", requestID)
			logger.Log.Error("Nenhum registro encontrado para o processo informado", err.Error())
			return
		}

		response.HandleError(c, http.StatusInternalServerError, "Erro ao buscar o registro no banco de dados", "", requestID)
		logger.Log.Error("Erro ao buscar o registro no banco de dados", err.Error())
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
}

/**
 * Devolve os dados de todos os contextos
 * Rota: "/contexto"
 * Método: GET
 */

func (service *ContextoHandlerType) SelectAllHandler(c *gin.Context) {
	requestID := uuid.New().String()

	rows, err := service.contextoModel.SelectContextos()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na deleção do registro!"})
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Registro selecionado com sucesso!",
	}

	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
}
