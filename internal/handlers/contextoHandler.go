package handlers

import (
	"database/sql"

	"errors"

	"net/http"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"strconv"

	"github.com/gin-gonic/gin"
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

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	bodyParams := models.BodyParamsContextoInsert{}

	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Errorf("Parâmetros inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros inválidos", "", requestID)
		return
	}

	if bodyParams.NrProc == "" {

		logger.Log.Error("O campo nrProc é obrigatório")
		response.HandleError(c, http.StatusBadRequest, "O campo nrProc é obrigatório", "", requestID)
		return
	}

	isExiste, err := service.contextoModel.RowExists(bodyParams.NrProc)
	if err != nil {

		logger.Log.Errorf("Erro na verificação existência!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no servidor ao verificar existência!", "", requestID)
		return
	}

	if isExiste {

		logger.Log.Error("Processo já existe!")
		response.HandleError(c, http.StatusBadRequest, "Processo já existe!", "", requestID)
		return
	}

	row, err := service.contextoModel.InsertRow(bodyParams)
	if err != nil {

		logger.Log.Errorf("Erro ao inserir contexto: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no servidor ao inserir contexto!", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro inserido com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
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

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	bodyParams := models.BodyParamsContextoUpdate{}
	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Errorf("Parâmetros inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	if bodyParams.NrProc == "" {

		logger.Log.Error("O campo NrProc é obrigatório")
		response.HandleError(c, http.StatusBadRequest, "O campo NrProc é obrigatório", "", requestID)
		return

	}

	row, err := service.contextoModel.UpdateRow(bodyParams)
	if err != nil {

		logger.Log.Errorf("Erro na alteração do registro!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no servidor ao altear o registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro alterado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/**
 * Devolve os dados dos usuários cadastrados na tabela 'users'
 * Rota: "/contexto"
 * Método: DELETE
 */
func (service *ContextoHandlerType) DeleteHandler(c *gin.Context) {

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

		logger.Log.Errorf("ID inválido!: %v", err)
		response.HandleError(c, http.StatusBadRequest, "ID inválido!", "", requestID)
		return
	}

	_, err = service.contextoModel.DeleteReg(id)
	if err != nil {

		logger.Log.Errorf("Erro na deleção do registro!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na deleção do registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"ok":      true,
		"message": "Registro deletado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/**
 * Devolve os dados do contexto indicado
 * Rota: "/contexto/:id"
 * Método: GET
 */
func (service *ContextoHandlerType) SelectByIDHandler(c *gin.Context) {

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

		logger.Log.Errorf("ID inválido!: %v", err)
		response.HandleError(c, http.StatusBadRequest, "ID inválido!", "", requestID)
		return
	}

	row, err := service.contextoModel.SelectContextoById(id)
	if err != nil {

		logger.Log.Errorf("Registro não encontrado!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Registro não encontrado!", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/**
 * Devolve os dados do contexto indicado pelo processo
 * Rota: "/contexto/processo/:id"
 * Método: GET
 */
func (service *ContextoHandlerType) SelectByProcessoHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	// Obtém o parâmetro "id" da rota
	paramID := c.Param("id")
	if paramID == "" {

		logger.Log.Error("ID do processo não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID do processo não informado!", "", requestID)
		return
	}

	row, err := service.contextoModel.SelectContextoByProcesso(paramID)
	if err != nil {
		// Verifica se o erro é de "registro não encontrado"
		if errors.Is(err, sql.ErrNoRows) {

			response.HandleError(c, http.StatusNotFound, "Nenhum registro encontrado para o processo informado", "", requestID)
			logger.Log.Errorf("Nenhum registro encontrado para o processo informado: %v", err)
			return
		}

		response.HandleError(c, http.StatusInternalServerError, "Erro ao buscar o registro no banco de dados", "", requestID)
		logger.Log.Errorf("Erro ao buscar o registro no banco de dados: %v", err)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/**
 * Devolve os dados de todos os contextos
 * Rota: "/contexto"
 * Método: GET
 */

func (service *ContextoHandlerType) SelectAllHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	rows, err := service.contextoModel.SelectContextos()
	if err != nil {

		logger.Log.Errorf("Erro na deleção do registro!: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro na deleção do registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}
