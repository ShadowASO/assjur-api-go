package handlers

import (
	"net/http"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/files"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DocsocrHandlerType struct {
	//Model *models.DocsocrModelType
	Service *services.Autos_tempServiceType
}

// func NewDocsocrHandlers(model *models.DocsocrModelType) *DocsocrHandlerType {
// 	return &DocsocrHandlerType{
// 		Model: model,
// 	}
// }

func NewDocsocrHandlers(service *services.Autos_tempServiceType) *DocsocrHandlerType {
	return &DocsocrHandlerType{
		Service: service,
	}
}

func (service *DocsocrHandlerType) InsertHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	var body models.DocsocrRow

	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Errorf("Parâmetros inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	// Validação de campos obrigatórios
	if body.IdCtxt == 0 || body.NmFileNew == "" || body.NmFileOri == "" || body.TxtDoc == "" {

		logger.Log.Error("Campos do body ausentes: ")
		response.HandleError(c, http.StatusBadRequest, "Campos do body ausentes", "", requestID)
		return
	}

	// Insere o registro no banco de dados
	//row, err := service.Model.InsertRow(body)
	row, err := service.Service.InserirAutos(body.IdCtxt, 0, body.NmFileNew, body.TxtDoc)
	if err != nil {

		logger.Log.Errorf("Erro ao inserir registro: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao inserir registro", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Inserido com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)

}

type BodyAutos_tempDelete struct {
	Id         string
	IdContexto int
}

func (service *DocsocrHandlerType) DeleteHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	var body []BodyAutos_tempDelete

	// Decodifica o corpo da requisição
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Errorf("Parâmetros inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Parâmetros do body inválidos", "", requestID)
		return
	}

	// Validação inicial
	if len(body) == 0 {

		logger.Log.Error("Ausentes arquivos para deleção: ")
		response.HandleError(c, http.StatusBadRequest, "Ausentes arquivos para deleção", "", requestID)
		return
	}

	// Rastreamento de resultados
	var deletedFiles []string
	var failedFiles []string

	// Processa os arquivos para deleção
	for _, reg := range body {
		// Busca o registro no banco
		row, err := service.Service.SelectById(reg.Id)
		if err != nil {

			logger.Log.Errorf("Arquivo não encontrado: %v ", err)
			failedFiles = append(failedFiles, reg.Id)
			continue
		}

		// Deleta o registro do banco
		//err = service.Model.DeleteRow(reg.IdDoc)
		err = service.Service.DeletaAutos(reg.Id)
		if err != nil {

			logger.Log.Errorf("Erro ao deletar registro: %v ", err)
			failedFiles = append(failedFiles, reg.Id)
			continue
		}

		// Deleta o arquivo do sistema de arquivos
		fullFileName := filepath.Join("uploads", row.IdPje)
		if files.FileExist(fullFileName) {
			err = files.DeletarFile(fullFileName)
			if err != nil {

				logger.Log.Error("Erro ao deletar arquivo: " + fullFileName)
				failedFiles = append(failedFiles, reg.Id)
				continue
			}
		}

		// Adiciona ao rastreamento de sucessos
		deletedFiles = append(deletedFiles, reg.Id)
	}

	rsp := gin.H{
		"message": "Processamento concluído",
		"deleted": deletedFiles,
		"errors":  failedFiles,
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

func (service *DocsocrHandlerType) DeleteHandlerById(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	//idDoc, err := strconv.Atoi(c.Param("id"))
	idDoc := c.Param("id")

	// if err != nil {
	// 	logger.Log.Errorf("IdDoc inválidos: %v", err)
	// 	response.HandleError(c, http.StatusBadRequest, "Formado do IdDoc inválidos", "", requestID)
	// 	return
	// }

	// Busca o registro no banco
	//row, err := service.Model.SelectByIdDoc(idDoc)
	// row, err := service.Service.SelectById(idDoc)
	// if err != nil {

	// 	logger.Log.Errorf("Registro não encontrado: %v", err)
	// 	response.HandleError(c, http.StatusNotFound, "Registro não encontrado", "", requestID)
	// 	return
	// }

	// Deleta o registro do banco
	err := service.Service.DeletaAutos(idDoc)
	if err != nil {

		logger.Log.Errorf("Erro ao deletar Registro: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao deletar Registro", "", requestID)
		return
	}

	// Deleta o arquivo do sistema de arquivos
	// fullFileName := filepath.Join("uploads", row.IdPje)
	// if files.FileExist(fullFileName) {
	// 	err = files.DeletarFile(fullFileName)
	// 	if err != nil {

	// 		logger.Log.Errorf("Erro ao deletar arquivo físico: %v", err)
	// 		response.HandleError(c, http.StatusInternalServerError, "Erro ao deletar arquivo físico", "", requestID)
	// 		return
	// 	}
	// }

	rsp := gin.H{
		"ok":      true,
		"message": "Documento(s) deletado(s) com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

func (service *DocsocrHandlerType) SelectByIDHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	paramID := c.Param("id")
	if paramID == "" {

		logger.Log.Error("ID não informado")
		response.HandleError(c, http.StatusBadRequest, "ID não informado", "", requestID)
		return
	}

	// id, err := strconv.Atoi(paramID)
	// if err != nil {
	// 	logger.Log.Errorf("ID inválidos: %v", err)
	// 	response.HandleError(c, http.StatusBadRequest, "ID inválido", "", requestID)
	// 	return
	// }

	row, err := service.Service.SelectById(paramID)
	if err != nil {
		logger.Log.Errorf("Erro ao selecionar documentos: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao selecionar documentos", "", requestID)
		return
	}

	rsp := gin.H{
		"rows":    row,
		"message": "Selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

/**
 * Devolve os registros da tabela 'docsocr' para um determinado contexto'
 * Rota: "/contexto/documentos/:id"
 * Params: ID do Contexto
 * Método: GET
 */
func (service *DocsocrHandlerType) SelectAllHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	ctxtID := c.Param("id")

	if ctxtID == "" {

		logger.Log.Error("ID não informado")
		response.HandleError(c, http.StatusBadRequest, "ID não informado", "", requestID)
		return
	}

	idKey, err := strconv.Atoi(ctxtID)
	if err != nil {

		logger.Log.Errorf("ID inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "ID inválido", "", requestID)
		return
	}

	rows, err := service.Service.SelectByContexto(idKey)
	if err != nil {

		logger.Log.Errorf("ID não informado: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao selecionar registro", "", requestID)
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Registros selecionados com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

/**
 * Devolve os registros da tabela 'temp_autos' para um determinado contexto'
 * Rota: "/contexto/documentos/:id"
 * Params: ID do documento
 * Método: GET
 */
func (service *DocsocrHandlerType) SelectHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	docID := c.Param("id")

	if docID == "" {

		logger.Log.Error("ID do documento não informado")
		response.HandleError(c, http.StatusBadRequest, "ID do documento não informado", "", requestID)

		return
	}

	// idKey, err := strconv.Atoi(docID)
	// if err != nil {

	// 	logger.Log.Errorf("Invalid formato do ID do documento: %v", err)
	// 	response.HandleError(c, http.StatusBadRequest, "Invalid formato do ID do documento", "", requestID)
	// 	return
	// }

	row, err := service.Service.SelectById(docID)
	if err != nil {

		logger.Log.Errorf("Erro ao selecionar registro: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no servidor ao selecionar registro", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}
