package handlers

import (
	"database/sql"

	"errors"

	"net/http"

	"ocrserver/internal/handlers/response"

	"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

type ContextoHandlerType struct {
	service *services.ContextoServiceType
}

func NewContextoHandlers(service *services.ContextoServiceType) *ContextoHandlerType {
	return &ContextoHandlerType{service: service}
}

type BodyParamsContextoInsert struct {
	NrProc  string
	Juizo   string
	Classe  string
	Assunto string
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

func (obj *ContextoHandlerType) InsertHandler(c *gin.Context) {
	userName := c.GetString("userName")

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	bodyParams := BodyParamsContextoInsert{}

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

	//isExiste, err := service.contextoModel.RowExists(bodyParams.NrProc)
	isExiste, err := obj.service.ContextoExiste(bodyParams.NrProc)
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

	row, err := obj.service.InsertContexto(
		bodyParams.NrProc,
		bodyParams.Juizo,
		bodyParams.Classe,
		bodyParams.Assunto,
		userName)
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
* aLTERA um registro de contexto
  - Rota: "/contexto"
  - Método: PUT
  - Body: {
    IdCtxt           int
    Juizo            string
    Classe           string
    Assunto          string
    }
*/
type BodyParamsContextoUpdate struct {
	Id      string
	Juizo   string
	Classe  string
	Assunto string
}

func (obj *ContextoHandlerType) UpdateHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	bodyParams := BodyParamsContextoUpdate{}
	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Errorf("Parâmetros inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	if bodyParams.Id == "" {

		logger.Log.Error("O campo IdCtxt é obrigatório")
		response.HandleError(c, http.StatusBadRequest, "O campo IdCtxt é obrigatório", "", requestID)
		return

	}

	row, err := obj.service.UpdateContexto(
		bodyParams.Id,
		bodyParams.Juizo,
		bodyParams.Classe,
		bodyParams.Assunto)
	if err != nil {

		logger.Log.Errorf("Erro na alteração do registro!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no servidor ao altear o registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro alterado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/**
 * Devolve os dados dos usuários cadastrados na tabela 'users'
 * Rota: "/contexto/:id"
 * Método: DELETE
 */
func (obj *ContextoHandlerType) DeleteHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID da sessão não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID da sessão não informado!", "", requestID)
		return
	}

	//Verifica se o contexto possui registros  cadastrados nos autos
	autos, err := services.AutosJsonServiceGlobal.SelectByContexto(paramID)
	if err != nil {

		logger.Log.Errorf("Erro ao selecionar os autos do contexto!: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao selecionar os autos do contexto!", "", requestID)
		return
	}
	if len(autos) > 0 {

		logger.Log.Errorf("Os autos não estão vazios! Contexto não pode ser excluído!: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Os autos não estão vazios! Contexto não pode ser excluído!", "", requestID)
		return
	}

	err = obj.service.DeletaContexto(paramID)
	if err != nil {

		logger.Log.Errorf("Erro na deleção do registro!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na deleção do registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"ok":      true,
		"message": "Registro deletado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/**
 * Devolve os dados do contexto indicado
 * Rota: "/contexto/:id"
 * Método: GET
 */
func (obj *ContextoHandlerType) SelectByIDHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	paramID := c.Param("id")
	if paramID == "" {

		logger.Log.Error("ID da sessão não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID da sessão não informado!", "", requestID)
		return
	}
	// id, err := strconv.Atoi(paramID)
	// if err != nil {

	// 	logger.Log.Errorf("ID inválido!: %v", err)
	// 	response.HandleError(c, http.StatusBadRequest, "ID inválido!", "", requestID)
	// 	return
	// }

	row, err := obj.service.SelectContextoById(paramID)
	//row, err := obj.service.SelectContextoByIdCtxt(paramID)
	if err != nil {

		logger.Log.Errorf("Registro não encontrado!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Registro não encontrado!", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/**
 * Devolve o uso de tokens por contexto
 * Rota: "/tokens/:id"
 * Método: GET
 */
func (obj *ContextoHandlerType) SelectByIdCtxtHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	paramID := c.Param("id")
	if paramID == "" {

		logger.Log.Error("ID_CTXT da sessão não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID_CTXT da sessão não informado!", "", requestID)
		return
	}

	row, err := obj.service.SelectContextoByIdCtxt(paramID)
	if err != nil {

		logger.Log.Errorf("Registro não encontrado!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Registro não encontrado!", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/**
 * Devolve os dados do contexto indicado pelo número do processo
 * Rota: "/contexto/processo/:id"
 * Método: GET
 */
func (obj *ContextoHandlerType) SelectByProcessoHandler(c *gin.Context) {

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

	row, err := obj.service.SelectContextoByProcesso(paramID)
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

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/**
 * Devolve o uso de tokens por contexto
 * Rota: "/tokens/:id"
 * Método: GET
 */
func (obj *ContextoHandlerType) SelectTokenUsoHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	paramID := c.Param("id")
	if paramID == "" {

		logger.Log.Error("ID da sessão não informado!")
		response.HandleError(c, http.StatusBadRequest, "ID da sessão não informado!", "", requestID)
		return
	}

	row, err := obj.service.SelectContextoByIdCtxt(paramID)
	if err != nil {

		logger.Log.Errorf("Registro não encontrado!: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Registro não encontrado!", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/**
 * Devolve os registros que possuam o número do processo semelhante ao valor informado
 * Rota: "/contexto/processo/:id"
 * Método: GET
 */
type BodySearchContexto struct {
	SearchProcesso string `json:"search_processo"`
}

func (obj *ContextoHandlerType) SearchByProcessoHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	// Obtém o parâmetro "id" da rota
	bodyParams := BodySearchContexto{}
	if err := c.ShouldBindJSON(&bodyParams); err != nil {

		logger.Log.Errorf("Formato inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}

	if bodyParams.SearchProcesso == "" {

		logger.Log.Error("index_name, natureza e search_texto são obrigatórios")
		response.HandleError(c, http.StatusBadRequest, "index_name, natureza e search_texto são obrigatórios", "", requestID)
		return
	}

	rows, err := obj.service.SelectContextoByProcessoLike(bodyParams.SearchProcesso)
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
		"rows":    rows,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/**
 * Devolve os dados de todos os contextos
 * Rota: "/contexto"
 * Método: GET
 */

func (obj *ContextoHandlerType) SelectAllHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	rows, err := obj.service.SelectContextos(5, 0)
	if err != nil {

		logger.Log.Errorf("Erro na deleção do registro!: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro na deleção do registro!", "", requestID)
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
