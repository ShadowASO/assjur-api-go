package handlers

import (
	"encoding/json"
	"fmt"

	"net/http"

	"ocrserver/internal/consts"
	"ocrserver/internal/handlers/response"

	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"
	"ocrserver/internal/utils/msgs"

	"strconv"

	"github.com/gin-gonic/gin"
)

type AutosHandlerType struct {
	service *services.AutosServiceType
	idx     *opensearch.AutosIndexType
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
 * Deleta os registros da tabela 'uploads' e respectivos arquivos da pasta 'upload'.
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

type BodyAutosInserir struct {
	IdCtxt  int             `json:"id_ctxt"`
	IdNatu  int             `json:"id_natu"`
	IdPje   string          `json:"id_pje"`
	Doc     string          `json:"doc"`
	DocJson json.RawMessage `json:"doc_json"`
}

func (obj *AutosHandlerType) InsertHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var data BodyAutosInserir

	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Dados inválidos", "", requestID)
		return
	}

	if data.IdCtxt == 0 || data.IdNatu == 0 || data.IdPje == "" {
		logger.Log.Error("Campos obrigatórios ausentes!")
		response.HandleError(c, http.StatusBadRequest, "Campos obrigatórios ausentes!", "", requestID)
		return
	}

	row, err := obj.service.InserirAutos(data.IdCtxt, data.IdNatu, data.IdPje, data.Doc, data.DocJson)

	if err != nil {
		logger.Log.Errorf("Erro na inclusão do registro %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno no servidor, durante inclusão do registro", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro inserido com sucesso!",
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

func (obj *AutosHandlerType) UpdateHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var requestData consts.AutosRow
	if err := c.ShouldBindJSON(&requestData); err != nil {
		logger.Log.Errorf("Dados do request.body inválidos %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato inválidos", "", requestID)
		return
	}

	if requestData.Id == "" {
		logger.Log.Error("Campos IdAutos inválidos")
		response.HandleError(c, http.StatusBadRequest, "Campos IdAutos com valor zero", "", requestID)
		return
	}

	row, err := obj.service.UpdateAutos(requestData)
	if err != nil {
		logger.Log.Errorf("Erro no update do registro! %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro interno do servidor durante o update", "", requestID)
		return
	}

	rsp := gin.H{
		"row":     row,
		"message": "Registro alterado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

func (obj *AutosHandlerType) DeleteHandler(c *gin.Context) {

	requestID := middleware.GetRequestID(c)

	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID ausente")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}

	err := obj.service.DeletaAutos(paramID)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar o registro: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro na deleção do registro", "", requestID)
		return
	}

	rsp := gin.H{
		"ok":      true,
		"message": "Registro deletado com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

func (obj *AutosHandlerType) SelectByIdHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID ausente na requisição")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}

	row, err := obj.service.SelectById(paramID)

	if err != nil {
		logger.Log.Errorf("Registro não localizado pelo ID: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Registro não localizado pelo ID", "", requestID)
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
	requestID := middleware.GetRequestID(c)

	ctxtID := c.Param("id")
	if ctxtID == "" {
		logger.Log.Error("ID não informado")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}
	idKey, err := strconv.Atoi(ctxtID)
	if err != nil {
		logger.Log.Errorf("ID inválidos: %v", err)
		response.HandleError(c, http.StatusBadRequest, "ID inválidos", "", requestID)
		return
	}

	rows, err := obj.service.SelectByContexto(idKey)
	if err != nil {
		logger.Log.Error("Erro ao realizar busca pelo contexto", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Erro ao realizar busca pelo contexto", "", requestID)
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

func (obj *AutosHandlerType) AutuarDocumentos(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var autuaFiles []services.RegKeys
	if err := c.ShouldBindJSON(&autuaFiles); err != nil {
		logger.Log.Errorf("Formato inválidos: %v", err)
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
