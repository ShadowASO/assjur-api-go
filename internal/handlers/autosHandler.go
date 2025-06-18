package handlers

import (
	"encoding/json"
	"fmt"

	"net/http"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/msgs"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AutosHandlerType struct {
	service *services.AutosServiceType
}

// Estrutura base para o JSON
type DocumentoBase struct {
	Tipo *struct {
		Key         int    `json:"key"`
		Description string `json:"description"`
	} `json:"tipo"`
	Processo string `json:"processo"`
	IdPje    string `json:"id_pje"`
}

func NewAutosHandlers(service *services.AutosServiceType) *AutosHandlerType {
	return &AutosHandlerType{
		service: service,
	}
}

/*
 * Deleta os registros da tabela 'temp_uploadfiles' e respectivos arquivos da pasta 'upload'.
 *
 * - **Rota**: "/contexto/documentos/upload"
 * - **Params**:
 * - **Método**: POST
 * - **Body:
 *		{
 *			IdAutos   int
 *			IdCtxt    int
 *			IdNat     int
 *			IdPje     string
 *			DtPje     time.Time
 *			AutosJson string
 *			DtInc     time.Time
 *			Status    string
 *		}
 * - **Resposta**:
 *		{
 *			IdAutos   int
 *			IdCtxt    int
 *			IdNat     int
 *			IdPje     string
 *			DtPje     time.Time
 *			AutosJson string
 *			DtInc     time.Time
 *			Status    string
 *		}
 */
func (obj *AutosHandlerType) InsertHandler(c *gin.Context) {
	requestID := uuid.New().String()
	var requestData models.AutosRow

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {
		logger.Log.Error("Dados do request.body inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Dados inválidos", "", requestID)
		return
	}

	if requestData.IdCtxt == 0 || requestData.IdNat == 0 || requestData.IdPje == "" {
		logger.Log.Error("Campos obrigatórios ausentes!")
		response.HandleError(c, http.StatusBadRequest, "Campos obrigatórios ausentes!", "", requestID)
		return
	}

	//row, err := service.service.autosModel.InsertRow(requestData)
	row, err := obj.service.InserirAutos(requestData)
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro na inclusão do registro", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro inserido com sucesso!",
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

func (obj *AutosHandlerType) UpdateHandler(c *gin.Context) {
	requestID := uuid.New().String()
	var requestData models.AutosRow
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {
		logger.Log.Error("Dados do request.body inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Formato inválidos", "", requestID)
		return
	}

	if requestData.IdAutos == 0 {
		logger.Log.Error("Campos IdAutos inválidos")
		response.HandleError(c, http.StatusBadRequest, "Campos IdAutos com valor zero", "", requestID)
		return
	}

	//row, err := service.autosModel.UpdateRow(requestData)
	row, err := obj.service.UpdateAutos(requestData)
	if err != nil {
		logger.Log.Error("Erro no update do registro!", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro durante o update", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro alterado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

func (obj *AutosHandlerType) DeleteHandler(c *gin.Context) {

	requestID := uuid.New().String()
	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID ausente")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}
	id, err := strconv.Atoi(paramID)
	if err != nil {
		logger.Log.Error("ID inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "ID inválidos", "", requestID)
		return
	}

	//err = service.autosModel.DeleteRow(id)
	err = obj.service.DeletaAutos(id)
	if err != nil {
		logger.Log.Error("Erro ao deletar o registro", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro na deleção do registro", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     nil,
		"message": "Registro deletado com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

func (obj *AutosHandlerType) SelectByIdHandler(c *gin.Context) {
	requestID := uuid.New().String()
	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID ausente na requisição")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}
	id, err := strconv.Atoi(paramID)
	if err != nil {
		logger.Log.Error("ID inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "ID inválidos", "", requestID)
		return
	}

	//row, err := service.autosModel.SelectById(id)
	row, err := obj.service.SelectById(id)

	if err != nil {
		logger.Log.Error("Registro não localizado pelo ID", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Registro não localizado pelo ID", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro selecionado com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/**
 * Devolve os registros da tabela 'autos' para um determinado contexto'
 * Rota: "/contexto/documentos/:id"
 * Params: ID do Contexto
 * Método: GET
 */
func (obj *AutosHandlerType) SelectAllHandler(c *gin.Context) {
	requestID := uuid.New().String()
	ctxtID := c.Param("id")
	if ctxtID == "" {
		logger.Log.Error("ID não informado")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}
	idKey, err := strconv.Atoi(ctxtID)
	if err != nil {
		logger.Log.Error("ID inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "ID inválidos", "", requestID)
		return
	}

	//rows, err := service.autosModel.SelectByContexto(idKey)
	rows, err := obj.service.SelectByContexto(idKey)
	if err != nil {
		logger.Log.Error("Erro ao realizar busca pelo contexto", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao realizar busca pelo contexto", "", requestID)
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Registro selecionado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/*
*
  - Executa uma análise do texto constante no registro de 'temp_autos',
  - indicado pelo 'idDoc', e salva o resultado no formato JSON, que é salvo
  - na tabela 'autos'. Em seguida, deleta o registro na tabela 'temp_autos'.
  - Rota: "/contexto/documentos/analise" *
  - Body: regKeys: [ {
    idContexto: number,
    idDoc: number,
    }]
  - Método: POST
*/
// type regKeys struct {
// 	IdContexto int
// 	IdDoc      int
// }

func (service *AutosHandlerType) AutuarDocumentos(c *gin.Context) {

	requestID := uuid.New().String()
	var autuaFiles []services.RegKeys
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&autuaFiles); err != nil {
		logger.Log.Error("Formato inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Formado do request.body inválidos", "", requestID)
		return
	}
	if len(autuaFiles) == 0 {
		logger.Log.Error("Nenhum documento informado")
		response.HandleError(c, http.StatusBadRequest, "Nenhum documento informado", "", requestID)
		return
	}

	msgs.CreateLogTimeMessage("Iniciando processamento")

	for _, reg := range autuaFiles {
		//if err := service.processarDocumento(reg); err != nil {
		if err := services.TempautosServiceGlobal.ProcessarDocumento(reg); err != nil {
			msg := fmt.Sprintf("Erro ao processar documento IdDoc=%d - Contexto=%d: %v", reg.IdDoc, reg.IdContexto, err)
			logger.Log.Error(msg, err.Error())
			continue
		}
	}

	msgs.CreateLogTimeMessage("Processamento concluído")

	rsp := gin.H{
		"rows":    nil,
		"message": "Documento(s) autuados(s) com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}
