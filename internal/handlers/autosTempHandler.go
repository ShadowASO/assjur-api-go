package handlers

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type AutosTempHandlerType struct {
	Service *services.AutosTempServiceType
	Idx     *opensearch.AutosTempIndexType
}

func NewAutosTempHandlers(service *services.AutosTempServiceType) *AutosTempHandlerType {
	return &AutosTempHandlerType{
		Service: service,
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

type BodyAutosTempInserir struct {
	IdCtxt string `json:"id_ctxt"`
	IdNatu int    `json:"id_natu"`
	IdPje  string `json:"id_pje"`
	Doc    string `json:"doc"`
}

// Método: POST
// URL: "/contexto/documentos/ocr/"
// Processa e extrai todos os documentos indicados no body e contidos na tabela "uploads"
func (obj *AutosTempHandlerType) PDFHandler(c *gin.Context) {
	bodyParams := []services.BodyParamsPDF{}

	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Body params inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if len(bodyParams) == 0 {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Body não possui arquivos para extrair",
			msresponse.ErrorValidacao,
			"A lista de arquivos está vazia.",
		)
		return
	}

	extractedFiles, extractedErros := services.UploadServiceGlobal.ProcessaPDF(
		c.Request.Context(),
		bodyParams,
	)

	rsp := gin.H{
		"extractedErros": extractedErros,
		"extractedFiles": extractedFiles,
	}

	msresponse.OK(c, http.StatusOK, "Registros selecionados com sucesso", rsp)
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
type BodyAutos struct {
	IdContexto string
	IdDoc      string
}

type resultadoProcessamento struct {
	IdDoc string
	Erro  error
}

func (obj *AutosTempHandlerType) AutuarDocumentosHandler(c *gin.Context) {
	start := time.Now()
	defer func() {
		mslogger.LoggerGlobal.Infof("Autuação concluída: %v", time.Since(start))
	}()

	var autuaFiles []BodyAutos
	if err := c.ShouldBindJSON(&autuaFiles); err != nil {
		mslogger.LoggerGlobal.Errorf("Formato inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato do request.body inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if len(autuaFiles) == 0 {
		mslogger.LoggerGlobal.Error("Nenhum documento informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Nenhum documento informado",
			msresponse.ErrorValidacao,
			"A lista de documentos para autuação está vazia.",
		)
		return
	}

	msresponse.LogTime("Iniciando processamento")

	resultChan := make(chan resultadoProcessamento, len(autuaFiles))

	var wg sync.WaitGroup

	var (
		mu             sync.Mutex
		extractedFiles []string
		extractedErros []string
	)

	doneAgg := make(chan struct{})
	go func() {
		defer close(doneAgg)

		for res := range resultChan {
			mu.Lock()

			if res.Erro != nil {
				msg := fmt.Sprintf("Erro ao processar documento IdDoc=%s: %v", res.IdDoc, res.Erro)
				mslogger.LoggerGlobal.Error(msg)
				extractedErros = append(extractedErros, res.Erro.Error())
			} else {
				extractedFiles = append(extractedFiles, res.IdDoc)
			}

			mu.Unlock()
		}
	}()

	for _, reg := range autuaFiles {
		idCtxt := reg.IdContexto
		idDoc := reg.IdDoc

		wg.Add(1)

		go func(idCtxt, idDoc string) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					err := fmt.Errorf(
						"panic em ProcessarDocumento idCtxt=%s idDoc=%s: %v",
						idCtxt,
						idDoc,
						r,
					)

					mslogger.LoggerGlobal.Errorf("%v\n%s", err, stack)

					resultChan <- resultadoProcessamento{
						IdDoc: idDoc,
						Erro:  err,
					}
				}
			}()

			if idCtxt == "" || idDoc == "" {
				resultChan <- resultadoProcessamento{
					IdDoc: idDoc,
					Erro: fmt.Errorf(
						"idCtxt ou idDoc vazio (idCtxt=%q idDoc=%q)",
						idCtxt,
						idDoc,
					),
				}
				return
			}

			err := services.ProcessarDocumento(idCtxt, idDoc)

			resultChan <- resultadoProcessamento{
				IdDoc: idDoc,
				Erro:  err,
			}
		}(idCtxt, idDoc)
	}

	wg.Wait()
	close(resultChan)
	<-doneAgg

	msresponse.LogTime("Processamento concluído")

	sucesso := true
	message := "Processamento concluído com sucesso"

	if len(extractedErros) > 0 {
		sucesso = false
		message = "Processamento concluído com erros"
	}

	rsp := gin.H{
		"sucesso":        sucesso,
		"extractedErros": extractedErros,
		"extractedFiles": extractedFiles,
	}

	if !sucesso {
		rsp["erros"] = strings.Join(extractedErros, "; ")
	}

	msresponse.OK(c, http.StatusCreated, message, rsp)
}

func (obj *AutosTempHandlerType) InsertHandler(c *gin.Context) {
	var data BodyAutosTempInserir

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

	if data.IdCtxt == "" || data.IdNatu == 0 || data.IdPje == "" {
		mslogger.LoggerGlobal.Error("Campos obrigatórios ausentes!")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Os campos id_ctxt, id_natu e id_pje são obrigatórios.",
		)
		return
	}

	row, err := obj.Service.InserirAutos(data.IdCtxt, data.IdNatu, data.IdPje, data.Doc)
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

func (obj *AutosTempHandlerType) UpdateHandler(c *gin.Context) {
	var body consts.ResponseAutosTempRow

	if err := c.ShouldBindJSON(&body); err != nil {
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

	if body.Id == "" {
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

	row, err := obj.Service.UpdateAutos(body.Id, body.IdCtxt, body.IdNatu, body.IdPje, body.Doc)
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

func (obj *AutosTempHandlerType) DeleteHandler(c *gin.Context) {
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

	if err := obj.Service.DeletaAutos(paramID); err != nil {
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

func (obj *AutosTempHandlerType) SelectByIdHandler(c *gin.Context) {
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

	row, err := obj.Service.SelectById(paramID)
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
func (obj *AutosTempHandlerType) SelectAllHandler(c *gin.Context) {
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

	rows, err := obj.Service.SelectByContexto(ctxtID)
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

/*
Analisa todos os documentos inseridos na tabela "autos_temp", excluindo os registros que não
correspondam a documentos válidos para a juntada.
*/
func (obj *AutosTempHandlerType) SanearByContextHandler(c *gin.Context) {
	idContexto := c.Param("id")

	if idContexto == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Parâmetro id é obrigatório",
			msresponse.ErrorValidacao,
			"O parâmetro id do contexto é obrigatório.",
		)
		return
	}

	rows, err := services.AutosTempServiceGlobal.SelectByContexto(idContexto)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar arquivos pelo contexto %s: %v", idContexto, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao buscar arquivos",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	if len(rows) == 0 {
		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Nenhum arquivo encontrado para o contexto informado",
			msresponse.ErrorNaoEncontrado,
			"Não há documentos temporários vinculados ao contexto informado.",
		)
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	errCh := make(chan error, len(rows))

	for _, row := range rows {
		wg.Add(1)

		rowCopy := row

		go func() {
			defer wg.Done()

			natuDoc, err := obj.Service.VerificarNaturezaDocumento(
				c.Request.Context(),
				idContexto,
				rowCopy.Doc,
			)
			if err != nil {
				mslogger.LoggerGlobal.Errorf(
					"Erro ao verificar a natureza do documento %s: %v",
					rowCopy.IdPje,
					err,
				)

				errCh <- err
				return
			}

			mslogger.LoggerGlobal.Infof(
				"Natureza documento %s identificada: key=%d, description=%s",
				rowCopy.IdPje,
				natuDoc.Key,
				natuDoc.Description,
			)

			deletar := natuDoc.Key == consts.NATU_DOC_OUTROS ||
				natuDoc.Key == consts.NATU_DOC_MOVIMENTACAO

			if !deletar {
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if err := services.AutosTempServiceGlobal.DeletaAutos(rowCopy.Id); err != nil {
				mslogger.LoggerGlobal.Errorf("Erro ao deletar documento ID %s: %v", rowCopy.Id, err)
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	var erros []string
	for err := range errCh {
		if err != nil {
			erros = append(erros, err.Error())
		}
	}

	if len(erros) > 0 {
		rsp := gin.H{
			"erros": erros,
		}

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Alguns erros ocorreram no processamento dos documentos",
			msresponse.ErrorInterno,
			strings.Join(erros, "; "),
		)

		_ = rsp
		return
	}

	msresponse.OK(c, http.StatusOK, "Processamento concluído com sucesso")
}
