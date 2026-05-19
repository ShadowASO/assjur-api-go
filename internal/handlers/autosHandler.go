package handlers

import (
	"encoding/json"
	"net/http"

	"ocrserver/internal/consts"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type AutosHandlerType struct {
	service *services.AutosServiceType
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
	IdCtxt     string          `json:"id_ctxt"`
	IdNatu     int             `json:"id_natu"`
	IdPje      string          `json:"id_pje"`
	Doc        string          `json:"doc"`
	DocJsonRaw json.RawMessage `json:"doc_json_raw"`
}

func (obj *AutosHandlerType) InsertHandler(c *gin.Context) {
	var data BodyAutosInserir

	if err := c.ShouldBindJSON(&data); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao decodificar JSON: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Dados inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if data.IdCtxt == "" || data.IdNatu == 0 {
		mslogger.LoggerGlobal.Error("Campos obrigatórios ausentes!")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Os campos id_ctxt e id_natu são obrigatórios.",
		)
		return
	}

	docJsonRaw := string(data.DocJsonRaw)

	row, err := obj.service.InserirAutos(
		data.IdCtxt,
		data.IdNatu,
		data.IdPje,
		data.Doc,
		docJsonRaw,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na inclusão do registro %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno no servidor durante inclusão do registro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"row": row,
	}

	msresponse.OK(c, http.StatusCreated, "Registro inserido com sucesso", rsp)
}

func (obj *AutosHandlerType) UpdateHandler(c *gin.Context) {
	var requestData consts.ResponseAutosRow

	if err := c.ShouldBindJSON(&requestData); err != nil {
		mslogger.LoggerGlobal.Errorf("Dados do request.body inválidos %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if requestData.Id == "" {
		mslogger.LoggerGlobal.Error("Campo IdAutos inválido")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do registro não informado",
			msresponse.ErrorValidacao,
			"O campo id é obrigatório.",
		)
		return
	}

	row, err := obj.service.UpdateAutos(requestData)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro no update do registro! %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno do servidor durante o update",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"row": row,
	}

	msresponse.OK(c, http.StatusOK, "Registro alterado com sucesso", rsp)
}

/*
Deleção de documentos do índice "autos" e o embedding em "autos_json_embedding"
*/
func (obj *AutosHandlerType) DeleteHandler(c *gin.Context) {
	paramID := c.Param("id")
	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID ausente")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID ausente",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	if err := obj.service.DeletaAutos(paramID); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao deletar o registro: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro na deleção do registro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Registro deletado com sucesso")
}

func (obj *AutosHandlerType) SelectByIdHandler(c *gin.Context) {
	paramID := c.Param("id")
	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID ausente na requisição")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID ausente",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	row, err := obj.service.SelectById(paramID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Registro não localizado pelo ID: %v", err)

		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Registro não localizado pelo ID",
			msresponse.ErrorNaoEncontrado,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"row": row,
	}

	msresponse.OK(c, http.StatusOK, "Registro selecionado com sucesso", rsp)
}

/**
 * Devolve os registros da tabela 'autos' para um determinado contexto'
 * Rota: "/contexto/documentos/:id"
 * Params: ID do Contexto
 * Método: GET
 */
func (obj *AutosHandlerType) SelectAllHandler(c *gin.Context) {
	ctxtID := c.Param("id")
	if ctxtID == "" {
		mslogger.LoggerGlobal.Error("ID não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID ausente",
			msresponse.ErrorValidacao,
			"O parâmetro id do contexto é obrigatório.",
		)
		return
	}

	rows, err := obj.service.SelectByContexto(ctxtID)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao realizar busca pelo contexto", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao realizar busca pelo contexto",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"rows": rows,
	}

	msresponse.OK(c, http.StatusOK, "Registros selecionados com sucesso", rsp)
}
