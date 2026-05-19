package handlers

import (
	"net/http"
	"strings"

	"ocrserver/internal/services"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

// Estrutura do Handler
type EmbeddingHandlerType struct {
	service *services.AutosJsonServiceType
}

// Construtor do Handler
func NewEmbeddingHandlers(service *services.AutosJsonServiceType) *EmbeddingHandlerType {
	return &EmbeddingHandlerType{service: service}
}

type BodyAutosInsert struct {
	IdCtxt  string `json:"id_ctxt"`
	IdNatu  int    `json:"id_natu"`
	IdPje   string `json:"id_pje"`
	DocText string `json:"doc_text"`
}

/*
  - Insere um novo documento no banco vetorial mantido no OpenSearch, nos índices
    autos_embedding e decisoes.

* Rota: "/tabelas/modelos/autos/:id"
- Método: POST
*/
func (handler *EmbeddingHandlerType) InsertHandler(c *gin.Context) {
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

	rspSuc, err := handler.service.IncluirDocumento(
		"idDoc",
		paramID,
		0,
		"idPje",
		"doc",
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao inserir documento: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao inserir documento",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"sucesso": rspSuc,
	}

	msresponse.OK(c, http.StatusCreated, "Documento inserido com sucesso", rsp)
}

/*
  - Insere um novo documento no banco vetorial mantido no OpenSearch, nos índices
    autos_embedding e decisoes.

- Rota: "/tabelas/modelos/autos/doc"

- Método: POST

  - Body: {
    IdCtxt  string `json:"id_ctxt"`
    IdNatu  int    `json:"id_natu"`
    IdPje   string `json:"id_pje"`
    DocText string `json:"doc_text"`
    }
*/
func (handler *EmbeddingHandlerType) InsertDocumentoHandler(c *gin.Context) {
	var bodyParams BodyAutosInsert

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		mslogger.LoggerGlobal.Errorf("Dados do body de requisição inválidos: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Parâmetros do body da requisição inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	bodyParams.IdCtxt = strings.TrimSpace(bodyParams.IdCtxt)
	bodyParams.IdPje = strings.TrimSpace(bodyParams.IdPje)
	bodyParams.DocText = strings.TrimSpace(bodyParams.DocText)

	if bodyParams.IdCtxt == "" || bodyParams.IdNatu == 0 || bodyParams.DocText == "" {
		mslogger.LoggerGlobal.Error("Um dos campos obrigatórios está ausente: id_ctxt, id_natu e doc_text")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Todos os campos são obrigatórios: id_ctxt, id_natu e doc_text.",
		)
		return
	}

	id, err := handler.service.IncluirDocumento(
		"idDoc",
		bodyParams.IdCtxt,
		bodyParams.IdNatu,
		bodyParams.IdPje,
		bodyParams.DocText,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao inserir documento: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao inserir documento",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"id": id,
	}

	msresponse.OK(c, http.StatusCreated, "Registro inserido com sucesso", rsp)
}

/*
- Modifica um documento existente no Elasticsearch/OpenSearch
  - Rota: "/tabelas/modelos/{id}"

- Método: PUT
*/
func (handler *EmbeddingHandlerType) UpdateHandler(c *gin.Context) {
	idDoc := strings.TrimSpace(c.Param("id"))

	if idDoc == "" {
		mslogger.LoggerGlobal.Error("Id do documento é obrigatório")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Id do documento é obrigatório",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	var bodyParams BodyAutosInsert

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

	/*
		doc, err := handler.service.IndexAutos.UpdateDocumento(idDoc, bodyParams)
		if err != nil {
			mslogger.LoggerGlobal.Errorf("Erro ao atualizar documento: %v", err)

			msresponse.Fail(
				c,
				http.StatusInternalServerError,
				"Erro ao atualizar documento",
				msresponse.ErrorInterno,
				err.Error(),
			)
			return
		}

		rsp := gin.H{
			"doc": doc,
		}

		msresponse.OK(c, http.StatusOK, "Registro alterado com sucesso", rsp)
	*/

	_ = bodyParams

	msresponse.OK(c, http.StatusOK, "Registro alterado com sucesso")
}

/*
- Deleta um documento existente no Elasticsearch/OpenSearch
  - Rota: "/tabelas/modelos/:id"

- Método: DELETE
*/
func (handler *EmbeddingHandlerType) DeleteHandler(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	if id == "" {
		mslogger.LoggerGlobal.Error("ID do documento não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do documento não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	if err := handler.service.DeletaEmbedding(id); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao deletar documento: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao deletar documento",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	msresponse.OK(c, http.StatusOK, "Registro excluído com sucesso")
}

/*
- Busca um documento pelo ID no OpenSearch
  - Rota: "/tabelas/elastic/{id}"

- Método: GET
*/
func (handler *EmbeddingHandlerType) SelectByIdHandler(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	if id == "" {
		mslogger.LoggerGlobal.Error("ID do documento não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do documento não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	documento, err := handler.service.SelectById(id)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar documento: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao buscar documento",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	if documento == nil {
		mslogger.LoggerGlobal.Error("Documento não encontrado")

		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Documento não encontrado",
			msresponse.ErrorNaoEncontrado,
			"Não foi localizado documento para o id informado.",
		)
		return
	}

	rsp := gin.H{
		"doc": documento,
	}

	msresponse.OK(c, http.StatusOK, "Registro selecionado com sucesso", rsp)
}

/*
- Seleciona documentos que sejam da natureza apontada e contenham o conteúdo search_texto
  - Rota: "/tabelas/modelos/search"

- Método: POST
*/
type BodySearchEmbedding struct {
	IdCtxt      string `json:"id_ctxt"`
	IdNatu      int    `json:"id_natu"`
	SearchTexto string `json:"search_texto"`
}

// Busca documentos pelo conteúdo no OpenSearch
func (handler *EmbeddingHandlerType) SearchEmbeddingHandler(c *gin.Context) {
	bodyParams := BodySearchEmbedding{}

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

	bodyParams.IdCtxt = strings.TrimSpace(bodyParams.IdCtxt)
	bodyParams.SearchTexto = strings.TrimSpace(bodyParams.SearchTexto)

	if bodyParams.IdCtxt == "" || bodyParams.IdNatu == 0 || bodyParams.SearchTexto == "" {
		mslogger.LoggerGlobal.Error("contexto, natureza e search_texto são obrigatórios")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Os campos id_ctxt, id_natu e search_texto são obrigatórios.",
		)
		return
	}

	/*
		Converte a string de busca num embedding:

		rspEmbeddings, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(
			c.Request.Context(),
			bodyParams.SearchTexto,
		)
		if err != nil {
			mslogger.LoggerGlobal.Errorf("Erro ao converter a string de busca em embeddings: %v", err)

			msresponse.Fail(
				c,
				http.StatusInternalServerError,
				"Erro ao converter a string de busca em embeddings",
				msresponse.ErrorInterno,
				err.Error(),
			)
			return
		}

		vector32 := services.OpenaiServiceGlobal.Float64ToFloat32Slice(rspEmbeddings)

		documentos, err := handler.idx.ConsultaSemantica(vector32, bodyParams.Natureza)
		if err != nil {
			mslogger.LoggerGlobal.Errorf("Erro ao buscar documentos: %v", err)

			msresponse.Fail(
				c,
				http.StatusInternalServerError,
				"Erro ao buscar documentos",
				msresponse.ErrorInterno,
				err.Error(),
			)
			return
		}

		message := "Consulta realizada com sucesso"
		if len(documentos) == 0 {
			message = "Consulta realizada com sucesso: nenhum documento retornado"
		}

		rsp := gin.H{
			"docs": documentos,
		}

		msresponse.OK(c, http.StatusOK, message, rsp)
		return
	*/

	rsp := gin.H{}

	msresponse.OK(c, http.StatusOK, "Consulta realizada com sucesso", rsp)
}
