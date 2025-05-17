package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/utils/files"
	"ocrserver/internal/utils/msgs"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DocsocrHandlerType struct {
	tempautosModel *models.TempautosModelType
}

func NewDocsocrHandlers(model *models.TempautosModelType) *DocsocrHandlerType {
	return &DocsocrHandlerType{
		tempautosModel: model,
	}
}

func (service *DocsocrHandlerType) InsertHandler(c *gin.Context) {
	var requestData models.TempAutosRow

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {

		c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "Invalid data provided", nil))
		return
	}

	// Validação de campos obrigatórios
	if requestData.IdCtxt == 0 || requestData.NmFileNew == "" || requestData.NmFileOri == "" || requestData.TxtDoc == "" {
		c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "All fields are required", nil))
		return
	}

	// Insere o registro no banco de dados
	ret, err := service.tempautosModel.InsertRow(requestData)
	if err != nil {
		log.Printf("Insert error: %v", err)
		c.JSON(http.StatusInternalServerError, msgs.CreateResponse(false, http.StatusInternalServerError, "Failed to insert record", nil))
		return
	}

	c.JSON(http.StatusCreated, msgs.CreateResponse(true, http.StatusCreated, "Record successfully inserted", ret))

}

type paramsBodyTempAutosDelete struct {
	IdContexto int
	IdDoc      int
}

func (service *DocsocrHandlerType) DeleteHandler(c *gin.Context) {
	var deleteFiles []paramsBodyTempAutosDelete

	// Decodifica o corpo da requisição
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&deleteFiles); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "Dados inválidos",
		})
		return
	}

	// Validação inicial
	if len(deleteFiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "Nenhum arquivo para deletar",
		})
		return
	}

	// Rastreamento de resultados
	var deletedFiles []int
	var failedFiles []int

	// Processa os arquivos para deleção
	for _, reg := range deleteFiles {
		// Busca o registro no banco
		row, err := service.tempautosModel.SelectByIdDoc(reg.IdDoc)
		if err != nil {
			log.Printf("Arquivo não encontrado - id_doc=%d - contexto=%d", reg.IdDoc, reg.IdContexto)
			failedFiles = append(failedFiles, reg.IdDoc)
			continue
		}

		// Deleta o registro do banco
		err = service.tempautosModel.DeleteRow(reg.IdDoc)
		if err != nil {
			log.Printf("Erro ao deletar o registro no banco - id_doc=%d", reg.IdDoc)
			failedFiles = append(failedFiles, reg.IdDoc)
			continue
		}

		// Deleta o arquivo do sistema de arquivos
		fullFileName := filepath.Join("uploads", row.NmFileNew)
		if files.FileExist(fullFileName) {
			err = files.DeletarFile(fullFileName)
			if err != nil {
				log.Printf("Erro ao deletar o arquivo físico - %s", fullFileName)
				failedFiles = append(failedFiles, reg.IdDoc)
				continue
			}
		}

		// Adiciona ao rastreamento de sucessos
		deletedFiles = append(deletedFiles, reg.IdDoc)
	}

	// Monta a resposta
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Processamento concluído",
		"deleted":    deletedFiles,
		"errors":     failedFiles,
	}

	// Retorna a resposta padronizada
	c.JSON(http.StatusOK, response)

}

func (service *DocsocrHandlerType) SelectByIDHandler(c *gin.Context) {
	paramID := c.Param("id")
	if paramID == "" {

		c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "ID not provided", nil))
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {

		c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "Invalid ID format", nil))
		return
	}

	ret, err := service.tempautosModel.SelectByIdDoc(id)
	if err != nil {

		log.Printf("Select by ID error: %v", err)
		c.JSON(http.StatusNotFound, msgs.CreateResponse(false, http.StatusNotFound, "Record not found", nil))
		return
	}
	// response := gin.H{
	// 	"ok":         true,
	// 	"statusCode": http.StatusOK,
	// 	"message":    "registro selecionado com sucesso!",
	// 	"rows":       ret,
	// }

	// c.JSON(http.StatusOK, response)
	c.JSON(http.StatusOK, msgs.CreateResponse(true, http.StatusOK, "Record successfully retrieved", ret))

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
		c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "Context ID not provided", nil))
		return
	}

	idKey, err := strconv.Atoi(ctxtID)
	if err != nil {
		c.JSON(http.StatusBadRequest, msgs.CreateResponse(false, http.StatusBadRequest, "Invalid context ID format", nil))
		return
	}

	rows, err := service.tempautosModel.SelectByContexto(idKey)
	if err != nil {
		log.Printf("Select by context error: %v", err)
		c.JSON(http.StatusInternalServerError, msgs.CreateResponse(false, http.StatusInternalServerError, "Failed to retrieve records", nil))
		return
	}
	// Verifica se nenhum registro foi encontrado
	// if len(rows) == 0 {
	// 	c.JSON(http.StatusOK, msgs.CreateResponse(true, http.StatusOK, "No records found for the provided context", nil))
	// 	return
	// }
	//c.JSON(http.StatusOK, msgs.CreateResponse(true, http.StatusOK, "Records successfully retrieved", rows))
	rsp := gin.H{
		"rows":    rows,
		"message": "Todos os registros retornados com sucesso!",
	}

	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))

}

/**
 * Devolve os registros da tabela 'temp_autos' para um determinado contexto'
 * Rota: "/contexto/documentos/:id"
 * Params: ID do documento
 * Método: GET
 */
func (service *DocsocrHandlerType) SelectHandler(c *gin.Context) {
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

	row, err := service.tempautosModel.SelectByIdDoc(idKey)
	if err != nil {
		log.Printf("Select by id doc error: %v", err)
		c.JSON(http.StatusInternalServerError, msgs.CreateResponse(false, http.StatusInternalServerError, "Failed to retrieve records", nil))
		return
	}

	c.JSON(http.StatusOK, msgs.CreateResponse(true, http.StatusOK, "Records successfully retrieved", row))

}
