package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"net/http"
	"ocrserver/api/handler/response"
	"ocrserver/internal/services/openAI"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/msgs"
	"ocrserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AutosHandlerType struct {
	autosModel     *models.AutosModelType
	promptModel    *models.PromptModelType
	tempautosModel *models.TempautosModelType
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

func NewAutosHandlers() *AutosHandlerType {
	return &AutosHandlerType{
		promptModel:    models.NewPromptModel(),
		autosModel:     models.NewAutosModel(),
		tempautosModel: models.NewTempautosModel(),
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
func (service *AutosHandlerType) InsertHandler(c *gin.Context) {
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

	row, err := service.autosModel.InsertRow(requestData)
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

func (service *AutosHandlerType) UpdateHandler(c *gin.Context) {
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

	row, err := service.autosModel.UpdateRow(requestData)
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

func (service *AutosHandlerType) DeleteHandler(c *gin.Context) {

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

	err = service.autosModel.DeleteRow(id)
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

func (service *AutosHandlerType) SelectByIdHandler(c *gin.Context) {
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

	row, err := service.autosModel.SelectById(id)

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
func (service *AutosHandlerType) SelectAllHandler(c *gin.Context) {
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

	rows, err := service.autosModel.SelectByContexto(idKey)
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
type regKeys struct {
	IdContexto int
	IdDoc      int
}

func (service *AutosHandlerType) AutuarDocumentos(c *gin.Context) {

	requestID := uuid.New().String()
	var autuaFiles []regKeys
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
		if err := service.processarDocumento(reg); err != nil {
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

func (service *AutosHandlerType) processarDocumento(reg regKeys) error {

	msg := fmt.Sprintf("Processando documento: IdDoc=%d, IdContexto=%d", reg.IdDoc, reg.IdContexto)
	logger.Log.Info(msg)

	//REcupero o registro da tabela temp_autos
	dataTempautos, err := service.tempautosModel.SelectByIdDoc(reg.IdDoc)
	if err != nil {

		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%d - IdContexto=%d", reg.IdDoc, reg.IdContexto)
	}

	/* Recupero o prompt da tabela promptsModel*/
	dataPrompt, err := service.promptModel.SelectByNatureza(models.PROMPT_NATUREZA_IDENTIFICA)
	if err != nil {

		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%d - IdContexto=%d", reg.IdDoc, reg.IdContexto)
	}
	var messages openAI.MsgGpt
	messages.CreateMessage("user", dataTempautos.TxtDoc)
	messages.CreateMessage("user", dataPrompt.TxtPrompt)

	retSubmit, err := openAI.Service.SubmitPrompt(messages)
	if err != nil {
		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%d - IdContexto=%d", reg.IdDoc, reg.IdContexto)
	}

	/* Atualiza o uso de tokens na tabela 'sessions' */
	// sessionService := NewSessionsHandlers()
	// err = sessionService.UpdateTokensUso(retSubmit)
	// if err != nil {
	// 	return fmt.Errorf("ERROR: Erro na atualização do uso de tokens")
	// }

	/* Verifico se a resposta é um json válido*/
	rspJson := retSubmit.Choices[0].Message.Content

	var objJson DocumentoBase
	err = json.Unmarshal([]byte(rspJson), &objJson)
	if err != nil {

		return fmt.Errorf("ERROR: Erro ao fazer o parse do JSON")
	}

	isAutuado, err := service.autosModel.IsDocAutuado(context.Background(), reg.IdContexto, objJson.IdPje)
	if err != nil {
		return fmt.Errorf("ERROR: Erro ao verificar se documento já existe")

	}
	if isAutuado {
		return fmt.Errorf("ERROR: Documento já existe na tabela autosModel")
	}

	//Faz a inclusão do documentos na tabela autos
	autosParams := models.AutosRow{}
	autosParams.IdCtxt = reg.IdContexto
	autosParams.IdNat = objJson.Tipo.Key
	autosParams.IdPje = objJson.IdPje

	autosParams.AutosJson = json.RawMessage(rspJson) // Suporte para JSON nativo no Go
	autosParams.DtInc = time.Now()
	autosParams.Status = "S"

	_, err = service.autosModel.InsertRow(autosParams)
	if err != nil {

		return fmt.Errorf("ERROR: Erro na inclusão do registro na tabela autosModel")

	}

	//Faz a deleção do registro na tabela temp_autos
	err = service.tempautosModel.DeleteRow(dataTempautos.IdDoc)
	if err != nil {

		return fmt.Errorf("ERROR: Erro ao deletar registro na tabela temp_autos")
	}

	msg = "Concluído com sucesso!"
	logger.Log.Info(msg)
	return nil

}

func (service *AutosHandlerType) AutuarDocumentos2(c *gin.Context) {

	requestID := uuid.New().String()
	var autuaFiles []regKeys
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&autuaFiles); err != nil {
		logger.Log.Error("JSON com Formato inválido", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		return
	}
	for _, reg := range autuaFiles {

		//REcupero o registro da tabela temp_autos
		dataTempautos, err := service.tempautosModel.SelectByIdDoc(reg.IdDoc)
		if err != nil {
			msg := fmt.Sprintf("Arquivo não encontrato - idDoc=%d - IdContexto=%d", reg.IdDoc, reg.IdContexto)
			logger.Log.Error(msg, err.Error())
			continue
		}
		/* Recupero o prompt da tabela promptsModel*/
		dataPrompt, err := service.promptModel.SelectByNatureza(models.PROMPT_NATUREZA_IDENTIFICA)
		if err != nil {
			msg := fmt.Sprintf("Arquivo não encontrato - id_file=%d - contexto=%d", reg.IdDoc, reg.IdContexto)
			logger.Log.Error(msg, err.Error())
			continue
		}
		var messages openAI.MsgGpt
		messages.CreateMessage("user", dataTempautos.TxtDoc)
		messages.CreateMessage("user", dataPrompt.TxtPrompt)

		retSubmit, err := openAI.Service.SubmitPrompt(messages)
		if err != nil {
			logger.Log.Error("Erro ao submeter o prompt", err.Error())
			response.HandleError(c, http.StatusBadRequest, "Erro ao submeter o prompt", "", requestID)
			return
		}

		/* Verifico se a resposta é um json válido*/
		rspJson := retSubmit.Choices[0].Message.Content

		var objJson DocumentoBase
		err = json.Unmarshal([]byte(rspJson), &objJson)
		if err != nil {
			logger.Log.Error("Erro ao fazer o parse do arquivo JSON", err.Error())
			response.HandleError(c, http.StatusBadRequest, "Erro ao fazer o parse do arquivo JSON", "", requestID)
			return
		}

		//Verificar se documento já existe
		isAutuado, err := service.autosModel.IsDocAutuado(context.Background(), reg.IdContexto, objJson.IdPje)
		if err != nil {
			logger.Log.Error("Erro ao verificar se documento já existe!", err.Error())
			response.HandleError(c, http.StatusBadRequest, "Erro ao verificar se documento já existe!", "", requestID)
			return

		}
		if isAutuado {
			logger.Log.Error("Documento já existe na tabela autosModel!", err.Error())
			response.HandleError(c, http.StatusBadRequest, "Documento já existe na tabela autosModel!", "", requestID)
			return

		}

		//Faz a inclusão do documentos na tabela autos
		autosParams := models.AutosRow{}
		autosParams.IdCtxt = reg.IdContexto
		autosParams.IdNat = objJson.Tipo.Key
		autosParams.IdPje = objJson.IdPje
		autosParams.AutosJson = json.RawMessage(rspJson) // Suporte para JSON nativo no Go
		autosParams.DtInc = time.Now()
		autosParams.Status = "S"

		_, err = service.autosModel.InsertRow(autosParams)
		if err != nil {
			logger.Log.Error("Erro na inclusão do registro na tabela autosModel!", err.Error())
			response.HandleError(c, http.StatusBadRequest, "Erro na inclusão do registro na tabela autosModel!", "", requestID)
			return
		}

		//Faz a deleção do registro na tabela temp_autos
		err = service.tempautosModel.DeleteRow(dataTempautos.IdDoc)
		if err != nil {
			logger.Log.Error("Erro ao deletar registro na tabela temp_autos!", err.Error())
			response.HandleError(c, http.StatusBadRequest, "Erro ao deletar registro na tabela temp_autos!", "", requestID)
			return
		}
		c.JSON(http.StatusOK, gin.H{"response": retSubmit})

	}

	rsp := gin.H{
		"rows":    nil,
		"message": "Todos os registros retornados com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}
