package handlers

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"

	"ocrserver/internal/consts"
	"ocrserver/internal/handlers/response"

	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"
	"ocrserver/internal/utils/msgs"

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
	requestID := middleware.GetRequestID(c)

	bodyParams := []services.BodyParamsPDF{}
	if err := c.ShouldBindJSON(&bodyParams); err != nil {
		response := msgs.CreateResponseMessage("Body params inválidos!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if len(bodyParams) == 0 {
		response := msgs.CreateResponseMessage("Body não possui arquivos para extrair!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	extractedFiles, extractedErros := services.UploadServiceGlobal.ProcessaPDF(c.Request.Context(), bodyParams)

	rsp := gin.H{
		"extractedErros": extractedErros,
		"extractedFiles": extractedFiles,
		"message":        "Registros selecionados com sucesso!",
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
type BodyAutos struct {
	IdContexto string
	IdDoc      string
}

type resultadoProcessamento struct {
	IdDoc string
	Erro  error
}

func (obj *AutosTempHandlerType) AutuarDocumentosHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var autuaFiles []BodyAutos
	if err := c.ShouldBindJSON(&autuaFiles); err != nil {
		logger.Log.Errorf("Formato inválido: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato do request.body inválido", "", requestID)
		return
	}
	if len(autuaFiles) == 0 {
		logger.Log.Error("Nenhum documento informado")
		response.HandleError(c, http.StatusBadRequest, "Nenhum documento informado", "", requestID)
		return
	}

	msgs.CreateLogTimeMessage("Iniciando processamento")

	type resultadoProcessamento struct {
		IdDoc string
		Erro  error
	}

	// Se quiser limitar concorrência (RECOMENDADO p/ OCR/OpenSearch):
	// const maxWorkers = 8
	// sem := make(chan struct{}, maxWorkers)

	resultChan := make(chan resultadoProcessamento, len(autuaFiles))

	var wg sync.WaitGroup

	// Consumidor: agrega resultados enquanto as goroutines rodam
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
				logger.Log.Error(msg)
				//extractedErros = append(extractedErros, res.IdDoc)
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

			// sem <- struct{}{}        // se habilitar o limitador
			// defer func() { <-sem }() // libera o slot

			// RECOVER para evitar derrubar o processo inteiro
			defer func() {
				if r := recover(); r != nil {
					// stacktrace completo para diagnosticar o nil
					stack := debug.Stack()
					err := fmt.Errorf("panic em ProcessarDocumento idCtxt=%s idDoc=%s: %v", idCtxt, idDoc, r)
					logger.Log.Errorf("%v\n%s", err, stack)

					resultChan <- resultadoProcessamento{
						IdDoc: idDoc,
						Erro:  err,
					}
				}
			}()

			// validação mínima
			if idCtxt == "" || idDoc == "" {
				resultChan <- resultadoProcessamento{
					IdDoc: idDoc,
					Erro:  fmt.Errorf("idCtxt ou idDoc vazio (idCtxt=%q idDoc=%q)", idCtxt, idDoc),
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

	msgs.CreateLogTimeMessage("Processamento concluído")

	sucesso := true
	message := "Processamento concluído com sucesso!"
	if len(extractedErros) > 0 {
		sucesso = false
		message = strings.Join(extractedErros, "; ")
	}

	rsp := gin.H{
		"sucesso":        sucesso,
		"extractedErros": extractedErros,
		"extractedFiles": extractedFiles,
		"message":        message,
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

func (obj *AutosTempHandlerType) InsertHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var data BodyAutosTempInserir

	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Dados inválidos", "", requestID)
		return
	}

	if data.IdCtxt == "" || data.IdNatu == 0 || data.IdPje == "" {
		logger.Log.Error("Campos obrigatórios ausentes!")
		response.HandleError(c, http.StatusBadRequest, "Campos obrigatórios ausentes!", "", requestID)
		return
	}

	row, err := obj.Service.InserirAutos(data.IdCtxt, data.IdNatu, data.IdPje, data.Doc)

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

func (obj *AutosTempHandlerType) UpdateHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var body consts.ResponseAutosTempRow
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Errorf("Dados do request.body inválidos %v", err)
		response.HandleError(c, http.StatusBadRequest, "Formato inválidos", "", requestID)
		return
	}

	if body.Id == "" {
		logger.Log.Error("Campos IdAutos inválidos")
		response.HandleError(c, http.StatusBadRequest, "Campos IdAutos com valor zero", "", requestID)
		return
	}

	row, err := obj.Service.UpdateAutos(body.Id, body.IdCtxt, body.IdNatu, body.IdPje, body.Doc)
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

func (obj *AutosTempHandlerType) DeleteHandler(c *gin.Context) {

	requestID := middleware.GetRequestID(c)

	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID ausente")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}

	err := obj.Service.DeletaAutos(paramID)
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

func (obj *AutosTempHandlerType) SelectByIdHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	paramID := c.Param("id")
	if paramID == "" {
		logger.Log.Error("ID ausente na requisição")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}

	row, err := obj.Service.SelectById(paramID)

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
func (obj *AutosTempHandlerType) SelectAllHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	ctxtID := c.Param("id")
	if ctxtID == "" {
		logger.Log.Error("ID não informado")
		response.HandleError(c, http.StatusBadRequest, "ID ausente", "", requestID)
		return
	}
	// idKey, err := strconv.Atoi(ctxtID)
	// if err != nil {
	// 	logger.Log.Errorf("ID inválidos: %v", err)
	// 	response.HandleError(c, http.StatusBadRequest, "ID inválidos", "", requestID)
	// 	return
	// }
	idKey := (ctxtID)

	rows, err := obj.Service.SelectByContexto(idKey)
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
Analisa todos os documentos inseridos na tabela "autos_temp", excluindo os registros que não
correspondam a documentos válidos para a juntada.
*/
func (service *AutosTempHandlerType) SanearByContextHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, msgs.CreateResponseMessage("Parâmetro id é obrigatório"))
		return
	}

	// idContexto, err := strconv.Atoi(idStr)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, msgs.CreateResponseMessage("Parâmetro id inválido"))
	// 	return
	// }

	idContexto := (idStr)

	//Faz um loop nos registros do indice "Autos_temp" para analisar cada uma dos registros,
	//e identificar a natureza, excluindo o que for lixo. Esta é a primeira verificação dos
	//documentos extraídos do PDF

	rows, err := services.AutosTempServiceGlobal.SelectByContexto(idContexto)
	if err != nil {
		logger.Log.Errorf("Erro ao buscar arquivos pelo contexto %d: %v", idContexto, err)
		c.JSON(http.StatusInternalServerError, msgs.CreateResponseMessage("Erro ao buscar arquivos"))
		return
	}

	if len(rows) == 0 {
		c.JSON(http.StatusNotFound, msgs.CreateResponseMessage("Nenhum arquivo encontrado para o contexto informado"))
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex // Protege chamadas concorrentes de DeleteRow caso não seja thread-safe

	// Usar canal para capturar erros na verificação (opcional)
	errCh := make(chan error, len(rows))

	for _, row := range rows {
		wg.Add(1)
		deletar := false

		// Copiar a variável para evitar problemas com closure
		rowCopy := row

		go func() {
			defer wg.Done()

			//Rotina que faz o trabalho pesado de verificação de cada registro
			natuDoc, err := service.Service.VerificarNaturezaDocumento(c.Request.Context(), idContexto, rowCopy.Doc)
			if err != nil {
				logger.Log.Errorf("Erro ao verificar a natureza do documento: %s", rowCopy.IdPje)
				return
			}

			logger.Log.Infof("Natureza documento %s identificada: key=%d, description=%s", rowCopy.IdPje, natuDoc.Key, natuDoc.Description)

			if natuDoc.Key == consts.NATU_DOC_OUTROS || natuDoc.Key == consts.NATU_DOC_CERTIDOES || natuDoc.Key == consts.NATU_DOC_MOVIMENTACAO {
				deletar = true
			}

			if deletar {
				mu.Lock()
				defer mu.Unlock()
				if err := services.AutosTempServiceGlobal.DeletaAutos(rowCopy.Id); err != nil {
					logger.Log.Errorf("Erro ao deletar documento ID %s: %v", rowCopy.Id, err)
					errCh <- err
				}
			}
		}()
	}

	// Aguarda todas as goroutines finalizarem
	wg.Wait()
	close(errCh)

	// Opcional: verificar se houve erros e registrar
	var hadErrors bool
	for _ = range errCh {
		hadErrors = true
		// Aqui já logou, pode acumular ou manipular erros se desejar
	}

	if hadErrors {
		c.JSON(http.StatusInternalServerError, msgs.CreateResponseMessage("Alguns erros ocorreram no processamento dos documentos"))
		return
	}

	rsp := gin.H{
		"message": "Processamento concluído com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}
