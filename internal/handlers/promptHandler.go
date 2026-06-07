/*
---------------------------------------------------------------------------------------
File: promptHandler.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 17-05-2025
---------------------------------------------------------------------------------------
*/
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ocrserver/internal/models/postgres"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type PromptHandlerType struct {
	service *services.PromptServiceType
}

func NewPromptHandlers(service *services.PromptServiceType) *PromptHandlerType {
	return &PromptHandlerType{service: service}
}

/*
- Insere um novo prompt na tabela 'prompts'
  - Rota: "/tabelas/prompt"
  - Método: POST
  - Body: {
    "IdNat": int
    "IdDoc": int
    "IdClasse": int
    "IdAssunto": int
    "NmDesc": string
    "TxtPrompt": string
    }
*/
func (obj *PromptHandlerType) InsertHandler(c *gin.Context) {
	bodyParams := postgres.BodyParamsPromptInsert{}

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		mslogger.LoggerGlobal.Errorf("JSON com formato inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if bodyParams.IdNat == 0 ||
		bodyParams.IdDoc == 0 ||
		bodyParams.IdClasse == 0 ||
		bodyParams.IdAssunto == 0 {
		mslogger.LoggerGlobal.Error("Faltam campos obrigatórios")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Faltam campos obrigatórios",
			msresponse.ErrorValidacao,
			"Os campos id_nat, id_doc, id_classe e id_assunto são obrigatórios.",
		)
		return
	}

	row, err := obj.service.InsertPrompt(bodyParams)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na inserção do registro: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro na inserção do registro",
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
- Modifica o registro na tabela 'prompts'
  - Rota: "/tabelas/prompt"
  - Método: PUT
  - Body: {
    "IdPrompt": int
    "NmDesc": string
    "TxtPrompt": string
    }
*/
func (obj *PromptHandlerType) UpdateHandler(c *gin.Context) {
	bodyParams := postgres.BodyParamsPromptUpdate{}

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		mslogger.LoggerGlobal.Errorf("Dados inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Parâmetros do body inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if bodyParams.IdPrompt == 0 {
		mslogger.LoggerGlobal.Error("O campo IdPrompt é obrigatório")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"O campo IdPrompt é obrigatório",
			msresponse.ErrorValidacao,
			"O campo id_prompt deve ser informado.",
		)
		return
	}

	rows, err := obj.service.UpdatePrompt(bodyParams)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na alteração do registro!: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro na alteração do registro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"rows": rows,
	}

	msresponse.OK(c, http.StatusOK, "Registro alterado com sucesso", rsp)
}

func (obj *PromptHandlerType) DeleteHandler(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))

	if idStr == "" {
		mslogger.LoggerGlobal.Error("ID não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("ID inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	rows, err := obj.service.DeletaPrompt(id)
	if err != nil {
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

	rsp := gin.H{
		"rows": rows,
	}

	msresponse.OK(c, http.StatusOK, "Registro deletado com sucesso", rsp)
}

func (obj *PromptHandlerType) SelectByIDHandler(c *gin.Context) {
	paramID := strings.TrimSpace(c.Param("id"))

	if paramID == "" {
		mslogger.LoggerGlobal.Error("ID não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	id, err := strconv.Atoi(paramID)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("ID inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	row, err := obj.service.SelectById(id)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar o registro: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao selecionar o registro",
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

func (obj *PromptHandlerType) SelectAllHandler(c *gin.Context) {
	rows, err := obj.service.SelectAll()
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao selecionar os registros: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao selecionar os registros",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"rows": rows,
	}

	msresponse.OK(c, http.StatusOK, "Todos os registros retornados com sucesso", rsp)
}
