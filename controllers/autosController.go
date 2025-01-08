package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"ocrserver/models"
	"ocrserver/services/openAI"
	"strconv"
)

type AutosControllerType struct {
	autosModel     *models.AutosModelType
	promptModel    *models.PromptModelType
	tempautosModel *models.TempautosModelType
}

// Estrutura base para o JSON
type DocumentoBase struct {
	Tipo struct {
		Key         string `json:"key"`
		Description string `json:"description"`
	} `json:"tipo"`
	Processo string `json:"processo"`
	IdPje    string `json:"id_pje"`
}

func NewAutosController() *AutosControllerType {
	return &AutosControllerType{
		promptModel: models.NewPromptModel(),
		autosModel:  models.NewAutosModel(),
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
func (service *AutosControllerType) InsertHandler(c *gin.Context) {
	var requestData models.AutosRow

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if requestData.IdCtxt == 0 || requestData.IdNat == 0 || requestData.IdPje == "" || requestData.AutosJson == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	ret, err := service.autosModel.InsertRow(requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na seleção de sessões!"})
		return
	}
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusCreated,
		"message":    "Record successfully inserted!",
		"rows":       ret,
	}

	c.JSON(http.StatusCreated, response)
}

func (service *AutosControllerType) UpdateHandler(c *gin.Context) {
	var requestData models.AutosRow
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}

	if requestData.IdAutos == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IdPrompt is required"})
		return
	}

	ret, err := service.autosModel.UpdateRow(requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na alteração do registro!"})
		return
	}
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusCreated,
		"message":    "Record successfully updated!",
		"rows":       ret,
	}

	c.JSON(http.StatusOK, response)
}

func (service *AutosControllerType) DeleteHandler(c *gin.Context) {
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

	err = service.autosModel.DeleteRow(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na deleção do registro!"})
		return
	}

	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "registro deletado com sucesso!",
	}

	c.JSON(http.StatusOK, response)
}

func (service *AutosControllerType) SelectByIDHandler(c *gin.Context) {
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

	ret, err := service.autosModel.SelectById(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Registro nçao encontrado!"})
		return
	}
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "registro selecionado com sucesso!",
		"rows":       ret,
	}

	c.JSON(http.StatusOK, response)
}

/**
 * Devolve os registros da tabela 'autos' para um determinado contexto'
 * Rota: "/contexto/documentos/:id"
 * Params: ID do Contexto
 * Método: GET
 */
func (service *AutosControllerType) SelectAllHandler(c *gin.Context) {
	ctxtID := c.Param("id")

	if ctxtID == "" {
		//c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID da sessão não informado!"})
		response := gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "ID do contexto não informado!",
			"rows":       nil,
		}

		c.JSON(http.StatusBadRequest, response)
		return
	}
	idKey, err := strconv.Atoi(ctxtID)
	if err != nil {
		//c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID inválido!"})
		response := gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "ID do contexto inválido!",
			"rows":       nil,
		}

		c.JSON(http.StatusBadRequest, response)
		return
	}

	rows, err := service.autosModel.SelectByContexto(idKey)
	if err != nil {
		//c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro ao buscar registros no banco de dados!"})
		response := gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "Erro na seleção dos registros em autos!",
			"rows":       nil,
		}

		c.JSON(http.StatusBadRequest, response)
		return
	}
	// Verifica se nenhum registro foi encontrado
	if len(rows) == 0 {
		c.JSON(http.StatusNoContent, gin.H{
			"ok":         true,
			"statusCode": http.StatusNoContent,
			"mensagem":   "Nenhum registro encontrado para o ID informado.",
			"rows":       nil,
		})
		return
	}

	// Retorna os dados do usuário
	retOK := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Executado com sucesso!",
		"rows":       rows,
	}

	c.JSON(http.StatusOK, retOK)
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
	idContexto int
	idDoc      int
}

func (service *AutosControllerType) AutuarDocumentos(c *gin.Context) {
	var autuaFiles []regKeys
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&autuaFiles); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Dados inválidos"})
		return
	}
	for _, reg := range autuaFiles {

		//REcupero o registro da tabela temp_autos
		dataTempautos, err := service.tempautosModel.SelectByIdDoc(reg.idDoc)
		if err != nil {
			log.Panicf("Arquivo não encontrato - id_file=%d - contexto=%d", reg.idDoc, reg.idContexto)
			continue
		}
		/* Recupero o prompt da tabela promptsModel*/
		dataPrompt, err := service.promptModel.SelectByNatureza(models.PROMPT_NATUREZA_IDENTIFICA)
		if err != nil {
			log.Panicf("Arquivo não encontrato - id_file=%d - contexto=%d", reg.idDoc, reg.idContexto)
			continue
		}
		var messages openAI.MsgGpt
		messages.CreateMessage("user", dataTempautos.TxtDoc)
		messages.CreateMessage("user", dataPrompt.TxtPrompt)

		retSubmit, err := openAI.Service.SubmitPrompt(messages)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Erro no SubmitPrompt"})
			return
		}

		/* Atualiza o uso de tokens na tabela 'sessions' */
		sessionService := NewSessionsController()
		err = sessionService.UpdateTokensUso(retSubmit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na atualização do uso de tokens!"})
			return
		}

		/* Verifico se a resposta é um json válido*/
		rspJson := retSubmit.Choices[0].Message.Content
		if !isValidJSON(rspJson) {
			//return fmt.Errorf("erro ao fazer o parse do JSON: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "formato do objeto JSON inválido!"})
			return
		}
		var objJson DocumentoBase
		err = json.Unmarshal([]byte(rspJson), &objJson)
		if err != nil {
			//return fmt.Errorf("erro ao fazer o parse do JSON: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "erro ao fazer o parse do arquivo json"})
			return
		}

		//fmt.Printf("ID_PJE: %s\n", objJson.IDPje)
		//Verificar se documento já existe
		isAutuado, err := service.autosModel.IsDocAutuado(reg.idContexto, objJson.IdPje)
		if err != nil {
			log.Printf("Erro ao verificar se documento já existe!")
			c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro ao verificar se documento já existe!"})
			return

		}
		if isAutuado {
			log.Printf("Documento já existe na tabela autosModel!")
			c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Documento já existe na tabela autosModel!"})
			return
		}
		//Faz a inclusão do documentos na tabela autos
		autosParams := models.AutosRow{}
		autosParams.IdCtxt = reg.idContexto
		autosParams.IdPje = objJson.IdPje
		autosParams.AutosJson = rspJson

		_, err = service.autosModel.InsertRow(autosParams)
		if err != nil {
			log.Printf("Erro na inclusão do registro na tabela autosModel!")
			c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro na inclusão do registro na tabela autosModel!"})
			return
		}

		//Faz a deleção do registro na tabela temp_autos
		err = service.tempautosModel.DeleteRow(dataTempautos.IdDoc)
		if err != nil {
			log.Printf("Erro ao deletar registro na tabela temp_autos!")
			c.JSON(http.StatusBadRequest, gin.H{"mensagem": "Erro ao deletar registro na tabela temp_autos!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"response": retSubmit})

	}
	/* Deleção dos arquivos*/

	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Documento(s) autuados com sucesso!",
	}

	c.JSON(http.StatusOK, response)
}

// Verifica se a string é um JSON válido
func isValidJSON(text string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(text), &js) == nil
}
