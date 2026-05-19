package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"ocrserver/internal/services"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type ContextoHandlerType struct {
	service *services.ContextoServiceType
}

func NewContextoHandlers(service *services.ContextoServiceType) *ContextoHandlerType {
	return &ContextoHandlerType{service: service}
}

type BodyParamsContextoInsert struct {
	NrProc  string `json:"nr_proc"`
	Juizo   string `json:"juizo"`
	Classe  string `json:"classe"`
	Assunto string `json:"assunto"`
}

/*
*
  - Insere um novo registro de contexto
  - Rota: "/contexto"
  - Método: POST
  - Body: {
    NrProc: string
    Juizo: string
    Classe: string
    Assunto: string
    }
*/
func (obj *ContextoHandlerType) InsertHandler(c *gin.Context) {
	userName := c.GetString("userName")

	bodyParams := BodyParamsContextoInsert{}

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		mslogger.LoggerGlobal.Errorf("Parâmetros inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Parâmetros inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	bodyParams.NrProc = strings.TrimSpace(bodyParams.NrProc)
	bodyParams.Juizo = strings.TrimSpace(bodyParams.Juizo)
	bodyParams.Classe = strings.TrimSpace(bodyParams.Classe)
	bodyParams.Assunto = strings.TrimSpace(bodyParams.Assunto)

	if bodyParams.NrProc == "" {
		mslogger.LoggerGlobal.Error("O campo nr_proc é obrigatório")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"O campo nr_proc é obrigatório",
			msresponse.ErrorValidacao,
			"O campo nr_proc deve ser informado.",
		)
		return
	}

	isExiste, err := obj.service.ContextoExiste(bodyParams.NrProc)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na verificação existência!: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno no servidor ao verificar existência",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	if isExiste {
		mslogger.LoggerGlobal.Error("Processo já existe!")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Processo já existe",
			msresponse.ErrorValidacao,
			"Já existe contexto cadastrado para o número de processo informado.",
		)
		return
	}

	row, err := obj.service.InsertContexto(
		bodyParams.NrProc,
		bodyParams.Juizo,
		bodyParams.Classe,
		bodyParams.Assunto,
		userName,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao inserir contexto: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno no servidor ao inserir contexto",
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

/*
* Altera um registro de contexto
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
	Id      string `json:"id"`
	Juizo   string `json:"juizo"`
	Classe  string `json:"classe"`
	Assunto string `json:"assunto"`
}

func (obj *ContextoHandlerType) UpdateHandler(c *gin.Context) {
	bodyParams := BodyParamsContextoUpdate{}

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		mslogger.LoggerGlobal.Errorf("Parâmetros inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Parâmetros do body inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	bodyParams.Id = strings.TrimSpace(bodyParams.Id)
	bodyParams.Juizo = strings.TrimSpace(bodyParams.Juizo)
	bodyParams.Classe = strings.TrimSpace(bodyParams.Classe)
	bodyParams.Assunto = strings.TrimSpace(bodyParams.Assunto)

	if bodyParams.Id == "" {
		mslogger.LoggerGlobal.Error("O campo id é obrigatório")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"O campo id é obrigatório",
			msresponse.ErrorValidacao,
			"O campo id do contexto deve ser informado.",
		)
		return
	}

	row, err := obj.service.UpdateContexto(
		bodyParams.Id,
		bodyParams.Juizo,
		bodyParams.Classe,
		bodyParams.Assunto,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na alteração do registro!: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro interno no servidor ao alterar o registro",
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

/**
 * Devolve os dados dos usuários cadastrados na tabela 'users'
 * Rota: "/contexto/:id"
 * Método: DELETE
 */
func (obj *ContextoHandlerType) DeleteHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID do contexto não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do contexto não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	// Verifica se o contexto possui registros cadastrados nos autos.
	autos, err := services.AutosJsonServiceGlobal.SelectByContexto(paramID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar os autos do contexto!: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao selecionar os autos do contexto",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	if len(autos) > 0 {
		mslogger.LoggerGlobal.Errorf("Os autos não estão vazios. Contexto não pode ser excluído. id=%s", paramID)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Contexto não pode ser excluído",
			msresponse.ErrorValidacao,
			"Os autos vinculados ao contexto não estão vazios.",
		)
		return
	}

	if err := obj.service.DeletaContexto(paramID); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na deleção do registro!: %v", err)

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

/**
 * Devolve os dados do contexto indicado
 * Rota: "/contexto/:id"
 * Método: GET
 */
func (obj *ContextoHandlerType) SelectByIDHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID do contexto não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do contexto não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	row, statusCode, err := obj.service.SelectContextoById(paramID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar contexto pelo ID: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao buscar contexto pelo ID",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	if statusCode == http.StatusNotFound {
		mslogger.LoggerGlobal.Errorf("Documento não encontrado ID: %s", paramID)

		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Documento não encontrado",
			msresponse.ErrorNaoEncontrado,
			"Não foi localizado contexto para o id informado.",
		)
		return
	}

	rsp := gin.H{
		"row": row,
	}

	msresponse.OK(c, http.StatusOK, "Registro selecionado com sucesso", rsp)
}

/**
 * Devolve o uso de tokens por contexto
 * Rota: "/tokens/:id"
 * Método: GET
 */
func (obj *ContextoHandlerType) SelectByIdCtxtHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID_CTXT da sessão não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID_CTXT da sessão não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	row, err := obj.service.SelectContextoByIdCtxt(paramID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Registro não encontrado!: %v", err)

		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Registro não encontrado",
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
 * Devolve os dados do contexto indicado pelo número do processo
 * Rota: "/contexto/processo/:id"
 * Método: GET
 */
func (obj *ContextoHandlerType) SelectByProcessoHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID do processo não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do processo não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id deve conter o número do processo.",
		)
		return
	}

	row, err := obj.service.SelectContextoByProcesso(paramID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			mslogger.LoggerGlobal.Errorf("Nenhum registro encontrado para o processo informado: %v", err)

			msresponse.Fail(
				c,
				http.StatusNotFound,
				"Nenhum registro encontrado para o processo informado",
				msresponse.ErrorNaoEncontrado,
				err.Error(),
			)
			return
		}

		mslogger.LoggerGlobal.Errorf("Erro ao buscar o registro no banco de dados: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao buscar o registro no banco de dados",
			msresponse.ErrorInterno,
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
 * Devolve o uso de tokens por contexto
 * Rota: "/tokens/:id"
 * Método: GET
 */
func (obj *ContextoHandlerType) SelectTokenUsoHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID da sessão não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID da sessão não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	row, err := obj.service.SelectContextoByIdCtxt(paramID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Registro não encontrado!: %v", err)

		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Registro não encontrado",
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
 * Devolve os registros que possuam o número do processo semelhante ao valor informado
 * Rota: "/contexto/processo/:id"
 * Método: GET
 */
type BodySearchContexto struct {
	SearchProcesso string `json:"search_processo"`
}

func (obj *ContextoHandlerType) SearchByProcessoHandler(c *gin.Context) {
	bodyParams := BodySearchContexto{}

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		mslogger.LoggerGlobal.Errorf("Formato inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	bodyParams.SearchProcesso = strings.TrimSpace(bodyParams.SearchProcesso)

	if bodyParams.SearchProcesso == "" {
		mslogger.LoggerGlobal.Error("search_processo é obrigatório")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"search_processo é obrigatório",
			msresponse.ErrorValidacao,
			"O campo search_processo deve ser informado.",
		)
		return
	}

	rows, err := obj.service.SelectContextoByProcessoLike(bodyParams.SearchProcesso)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			mslogger.LoggerGlobal.Errorf("Nenhum registro encontrado para o processo informado: %v", err)

			msresponse.Fail(
				c,
				http.StatusNotFound,
				"Nenhum registro encontrado para o processo informado",
				msresponse.ErrorNaoEncontrado,
				err.Error(),
			)
			return
		}

		mslogger.LoggerGlobal.Errorf("Erro ao buscar o registro no banco de dados: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao buscar o registro no banco de dados",
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

/**
 * Devolve os dados de todos os contextos
 * Rota: "/contexto"
 * Método: GET
 */
func (obj *ContextoHandlerType) SelectAllHandler(c *gin.Context) {
	rows, err := obj.service.SelectContextos(5, 0)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar contextos: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao selecionar contextos",
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
