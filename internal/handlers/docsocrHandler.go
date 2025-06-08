package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/utils/files"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/msgs"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DocsocrHandlerType struct {
	Model *models.TempautosModelType
}

func NewDocsocrHandlers(model *models.TempautosModelType) *DocsocrHandlerType {
	return &DocsocrHandlerType{
		Model: model,
	}
}

func (service *DocsocrHandlerType) InsertHandler(c *gin.Context) {
	requestID := uuid.New().String()
	var requestData models.TempAutosRow

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {

		// c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "Invalid data provided", nil))
		// return
		logger.Log.Error("Dados no body inválidos: ", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Dados no body inválidos", err.Error(), requestID)
		return
	}

	// Validação de campos obrigatórios
	if requestData.IdCtxt == 0 || requestData.NmFileNew == "" || requestData.NmFileOri == "" || requestData.TxtDoc == "" {
		// c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "All fields are required", nil))
		// return
		logger.Log.Error("Campos do body ausentes: ")
		response.HandleError(c, http.StatusBadRequest, "Campos do body ausentes", "", requestID)
		return
	}

	// Insere o registro no banco de dados
	row, err := service.Model.InsertRow(requestData)
	if err != nil {
		// log.Printf("Insert error: %v", err)
		// c.JSON(http.StatusInternalServerError, msgs.CreateResponse(false, http.StatusInternalServerError, "Failed to insert record", nil))
		// return
		logger.Log.Error("Erro ao inserir registro: ", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Erro ao inserir registro", err.Error(), requestID)
		return
	}

	//c.JSON(http.StatusCreated, msgs.CreateResponse(true, http.StatusCreated, "Record successfully inserted", ret))
	rsp := gin.H{
		"row":     row,
		"message": "Inserido com sucesso!",
	}

	//c.JSON(http.StatusCreated, response.NewSuccess(rsp, requestID))
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)

}

type paramsBodyTempAutosDelete struct {
	IdContexto int
	IdDoc      int
}

func (service *DocsocrHandlerType) DeleteHandler(c *gin.Context) {
	requestID := uuid.New().String()

	var deleteFiles []paramsBodyTempAutosDelete

	// Decodifica o corpo da requisição
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&deleteFiles); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{
		// 	"ok":         false,
		// 	"statusCode": http.StatusBadRequest,
		// 	"message":    "Dados inválidos",
		// })
		// return
		logger.Log.Error("Dados no body inválidos: ", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Dados no body inválidos", err.Error(), requestID)
		return
	}

	// Validação inicial
	if len(deleteFiles) == 0 {
		// c.JSON(http.StatusBadRequest, gin.H{
		// 	"ok":         false,
		// 	"statusCode": http.StatusBadRequest,
		// 	"message":    "Nenhum arquivo para deletar",
		// })
		// return
		logger.Log.Error("Ausentes arquivos para deleção: ")
		response.HandleError(c, http.StatusBadRequest, "Ausentes arquivos para deleção", "", requestID)
		return
	}

	// Rastreamento de resultados
	var deletedFiles []int
	var failedFiles []int

	// Processa os arquivos para deleção
	for _, reg := range deleteFiles {
		// Busca o registro no banco
		row, err := service.Model.SelectByIdDoc(reg.IdDoc)
		if err != nil {
			//log.Printf("Arquivo não encontrado - id_doc=%d - contexto=%d", reg.IdDoc, reg.IdContexto)
			logger.Log.Error("Arquivo não encontrado ")
			failedFiles = append(failedFiles, reg.IdDoc)
			continue
		}

		// Deleta o registro do banco
		err = service.Model.DeleteRow(reg.IdDoc)
		if err != nil {
			//log.Printf("Erro ao deletar o registro no banco - id_doc=%d", reg.IdDoc)
			logger.Log.Error("Erro ao deletar registro ")
			failedFiles = append(failedFiles, reg.IdDoc)
			continue
		}

		// Deleta o arquivo do sistema de arquivos
		fullFileName := filepath.Join("uploads", row.NmFileNew)
		if files.FileExist(fullFileName) {
			err = files.DeletarFile(fullFileName)
			if err != nil {
				//log.Printf("Erro ao deletar o arquivo físico - %s", fullFileName)
				logger.Log.Error("Erro ao deletar arquivo: " + fullFileName)
				failedFiles = append(failedFiles, reg.IdDoc)
				continue
			}
		}

		// Adiciona ao rastreamento de sucessos
		deletedFiles = append(deletedFiles, reg.IdDoc)
	}

	rsp := gin.H{
		"message": "Processamento concluído",
		"deleted": deletedFiles,
		"errors":  failedFiles,
	}

	//c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

func (service *DocsocrHandlerType) DeleteHandlerByIdDoc(c *gin.Context) {

	requestID := uuid.New().String()
	idDoc, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Log.Error("IdDoc inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Formado do IdDoc inválidos", "", requestID)
		return
	}

	// Busca o registro no banco
	row, err := service.Model.SelectByIdDoc(idDoc)
	if err != nil {
		// log.Printf("Arquivo não encontrado - id_doc=%d - contexto=%d", reg.IdDoc, reg.IdContexto)
		// failedFiles = append(failedFiles, reg.IdDoc)
		// continue
		logger.Log.Error("Registro não encontrado", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Registro não encontrado", "", requestID)
		return
	}

	// Deleta o registro do banco
	err = service.Model.DeleteRow(idDoc)
	if err != nil {
		// log.Printf("Erro ao deletar o registro no banco - id_doc=%d", reg.IdDoc)
		// failedFiles = append(failedFiles, reg.IdDoc)
		// continue
		logger.Log.Error("Erro ao deletar Registro: ", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao deletar Registro", "", requestID)
		return
	}

	// Deleta o arquivo do sistema de arquivos
	fullFileName := filepath.Join("uploads", row.NmFileNew)
	if files.FileExist(fullFileName) {
		err = files.DeletarFile(fullFileName)
		if err != nil {
			// log.Printf("Erro ao deletar o arquivo físico - %s", fullFileName)
			// failedFiles = append(failedFiles, reg.IdDoc)
			// continue
			logger.Log.Error("Erro ao deletar arquivo físico: ", err.Error())
			response.HandleError(c, http.StatusBadRequest, "Erro ao deletar arquivo físico", "", requestID)
			return
		}
	}

	// Adiciona ao rastreamento de sucessos
	//deletedFiles = append(deletedFiles, reg.IdDoc)
	//}

	// rsp := gin.H{
	// 	"message": "Processamento concluído",
	// 	"deleted": deletedFiles,
	// 	"errors":  failedFiles,
	// }

	// c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
	rsp := gin.H{
		"rows":    nil,
		"message": "Documento(s) deletado(s) com sucesso!",
	}

	response.HandleSuccess(c, http.StatusNoContent, rsp, requestID)
	//response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

func (service *DocsocrHandlerType) SelectByIDHandler(c *gin.Context) {
	requestID := uuid.New().String()
	paramID := c.Param("id")
	if paramID == "" {

		// c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "ID not provided", nil))
		// return
		logger.Log.Error("ID não informado")
		response.HandleError(c, http.StatusBadRequest, "ID não informado", "", requestID)
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {

		// c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "Invalid ID format", nil))
		// return
		logger.Log.Error("ID inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "ID inválido", "", requestID)
		return
	}

	row, err := service.Model.SelectByIdDoc(id)
	if err != nil {

		// log.Printf("Select by ID error: %v", err)
		// c.JSON(http.StatusNotFound, msgs.CreateResponse(false, http.StatusNotFound, "Record not found", nil))
		// return
		logger.Log.Error("Erro ao selecionar documentos", err.Error())
		response.HandleError(c, http.StatusNotFound, "Erro ao selecionar documentos", "", requestID)
		return
	}

	rsp := gin.H{
		"rows":    row,
		"message": "Selecionado com sucesso!",
	}

	//c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

/**
 * Devolve os registros da tabela 'temp_autos' para um determinado contexto'
 * Rota: "/contexto/documentos/:id"
 * Params: ID do Contexto
 * Método: GET
 */
func (service *DocsocrHandlerType) SelectAllHandler(c *gin.Context) {
	//Generate request ID for tracing
	requestID := uuid.New().String()
	ctxtID := c.Param("id")

	if ctxtID == "" {
		// c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "Context ID not provided", nil))
		// return
		logger.Log.Error("ID não informado")
		response.HandleError(c, http.StatusBadRequest, "ID não informado", "", requestID)
		return
	}

	idKey, err := strconv.Atoi(ctxtID)
	if err != nil {
		// c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "Invalid context ID format", nil))
		// return
		logger.Log.Error("ID inválido")
		response.HandleError(c, http.StatusBadRequest, "ID inválido", "", requestID)
		return
	}

	rows, err := service.Model.SelectByContexto(idKey)
	if err != nil {
		// log.Printf("Select by context error: %v", err)
		// c.JSON(http.StatusInternalServerError, msgs.CreateResponse(false, http.StatusInternalServerError, "Failed to retrieve records", nil))
		// return
		logger.Log.Error("ID não informado")
		response.HandleError(c, http.StatusInternalServerError, "Erro ao selecionar registro", "", requestID)
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Registros selecionados com sucesso!",
	}

	//c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

/**
 * Devolve os registros da tabela 'temp_autos' para um determinado contexto'
 * Rota: "/contexto/documentos/:id"
 * Params: ID do documento
 * Método: GET
 */
func (service *DocsocrHandlerType) SelectHandler(c *gin.Context) {
	requestID := uuid.New().String()

	docID := c.Param("id")

	if docID == "" {
		c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "ID do documento não informado", nil))
		return
	}

	idKey, err := strconv.Atoi(docID)
	if err != nil {
		c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "Invalid formato do ID do documento", nil))
		return
	}

	row, err := service.Model.SelectByIdDoc(idKey)
	if err != nil {
		log.Printf("Select by id doc error: %v", err)
		c.JSON(http.StatusInternalServerError, msgs.CreateResponse(false, http.StatusInternalServerError, "Failed to retrieve records", nil))
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}

	//c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}
