package handlers

import (
	"net/http"
	"strings"

	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

// Estrutura do Handler
type ModelosHandlerType struct {
	idx *opensearch.ModelosIndexType
}

// Construtor do Handler
func NewModelosHandlers(index *opensearch.ModelosIndexType) *ModelosHandlerType {
	return &ModelosHandlerType{idx: index}
}

/*
- Insere um novo documento no OpenSearch
  - Rota: "/tabelas/modelos"
  - Método: POST
  - Body: {
    Natureza     string `json:"natureza"`
    Ementa       string `json:"ementa"`
    Inteiro_teor string `json:"inteiro_teor"`
    }
*/
func (handler *ModelosHandlerType) InsertHandler(c *gin.Context) {
	var bodyParams opensearch.BodyModelosInsert

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

	bodyParams.Natureza = strings.TrimSpace(bodyParams.Natureza)
	bodyParams.Ementa = strings.TrimSpace(bodyParams.Ementa)
	bodyParams.Inteiro_teor = strings.TrimSpace(bodyParams.Inteiro_teor)

	if bodyParams.Natureza == "" || bodyParams.Ementa == "" || bodyParams.Inteiro_teor == "" {
		mslogger.LoggerGlobal.Error("Todos os campos são obrigatórios: natureza, ementa e inteiro_teor")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Todos os campos são obrigatórios: natureza, ementa e inteiro_teor.",
		)
		return
	}

	ementaVector, err := services.GetDocumentoEmbeddings(bodyParams.Ementa)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao extrair os embeddings da ementa: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao extrair os embeddings da ementa",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	teorVector, err := services.GetDocumentoEmbeddings(bodyParams.Inteiro_teor)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao extrair os embeddings do inteiro teor: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao extrair os embeddings do inteiro teor",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	resp, err := handler.idx.Indexa(
		bodyParams.Natureza,
		bodyParams.Ementa,
		bodyParams.Inteiro_teor,
		ementaVector,
		teorVector,
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
		"id": resp.ID,
	}

	msresponse.OK(c, http.StatusCreated, "Registro inserido com sucesso", rsp)
}

/*
- Modifica um documento existente no OpenSearch
  - Rota: "/tabelas/modelos/{id}"
  - Método: PUT
  - Body: {
    Natureza     string `json:"natureza"`
    Ementa       string `json:"ementa"`
    Inteiro_teor string `json:"inteiro_teor"`
    }
*/
func (handler *ModelosHandlerType) UpdateHandler(c *gin.Context) {
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

	var bodyParams opensearch.ModelosText

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

	doc, err := handler.idx.Update(idDoc, bodyParams)
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
}

/*
- Deleta um documento existente no OpenSearch
  - Rota: "/tabelas/modelos/:id"

- Método: DELETE
*/
func (handler *ModelosHandlerType) DeleteHandler(c *gin.Context) {
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

	if err := handler.idx.Delete(id); err != nil {
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
func (handler *ModelosHandlerType) SelectByIdHandler(c *gin.Context) {
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

	documento, err := handler.idx.ConsultaById(id)
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
  - Body: {
    Index_name   string `json:"index_name"`
    Natureza     string `json:"natureza"`
    Search_texto string `json:"search_texto"`
    }
*/
type BodySearchModelos struct {
	IndexName   string `json:"index_name"`
	Natureza    string `json:"natureza"`
	SearchTexto string `json:"search_texto"`
}

// Busca documentos pelo conteúdo no OpenSearch
func (handler *ModelosHandlerType) SearchModelosHandler(c *gin.Context) {
	bodyParams := BodySearchModelos{}

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

	bodyParams.IndexName = strings.TrimSpace(bodyParams.IndexName)
	bodyParams.Natureza = strings.TrimSpace(bodyParams.Natureza)
	bodyParams.SearchTexto = strings.TrimSpace(bodyParams.SearchTexto)

	if bodyParams.IndexName == "" || bodyParams.Natureza == "" || bodyParams.SearchTexto == "" {
		mslogger.LoggerGlobal.Error("index_name, natureza e search_texto são obrigatórios")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Os campos index_name, natureza e search_texto são obrigatórios.",
		)
		return
	}

	vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(
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

	docs, err := handler.idx.ConsultaSemantica(vec32, bodyParams.Natureza)
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
	if len(docs) == 0 {
		message = "Consulta realizada com sucesso: nenhum documento retornado"
	}

	rsp := gin.H{
		"docs": docs,
	}

	msresponse.OK(c, http.StatusOK, message, rsp)
}
