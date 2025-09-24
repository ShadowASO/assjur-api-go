package handlers

import (
	"encoding/json"
	"fmt"

	"net/http"

	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"os"
	"path/filepath"
	"strconv"
	"time"

	"ocrserver/internal/database/pgdb"

	"github.com/gin-gonic/gin"
)

type UploadHandlerType struct {
	Service *services.UploadServiceType
}

const CONTEXTO_TEMP = 18

func NewUploadHandlers(service *services.UploadServiceType) *UploadHandlerType {

	return &UploadHandlerType{Service: service}
}

// Função para gerar um nome único para o arquivo (essa é apenas uma sugestão, personalize conforme necessário)
func generateUniqueFileName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

/*
*
  - Faz o upload de um arquivo e cria um registro na tabela 'uploads'
  - Rota: "/contexto/documentos/upload"
  - Params:
  - Content-Type: multipart/form-data.
  - Body: {
  - file: File,
  - idContexto: number,
    filename_ori: string,
    }
  - Método: POST
  - Teste: curl -X POST http://localhost:4001/upload -F "file=@replica.pdf"
*/
func (service *UploadHandlerType) UploadFileHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	// Limita tamanho da requisição para 10MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<22)

	handler, err := c.FormFile("file")
	if err != nil {
		logger.Log.Errorf("Erro ao obter arquivo. Arquivo com mais de 40MB: %v", err)
		response.HandleError(c, http.StatusBadRequest, "Erro ao obter arquivo: Arquivo com mais de 40MB", err.Error(), requestID)
		return
	}

	filenameOri := c.PostForm("filename_ori")
	idContextoStr := c.PostForm("idContexto")
	idContexto, err := strconv.Atoi(idContextoStr)
	if err != nil || idContexto == 0 || filenameOri == "" {
		logger.Log.Error("Campos idContexto e filename_ori obrigatórios e válidos")
		response.HandleError(c, http.StatusBadRequest, "Campos idContexto e filename_ori obrigatórios e válidos", "", requestID)
		return
	}

	//uniqueFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(handler.Filename))

	uniqueFileName := generateUniqueFileName() + filepath.Ext(handler.Filename)

	savePath := filepath.Join("uploads", uniqueFileName)

	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		logger.Log.Errorf("Erro ao criar diretório uploads: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao criar diretório uploads", err.Error(), requestID)
		return
	}

	if err := c.SaveUploadedFile(handler, savePath); err != nil {
		logger.Log.Errorf("Erro ao salvar arquivo: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao salvar arquivo", err.Error(), requestID)
		return
	}

	if err := service.InsertUploadedFile(idContexto, uniqueFileName, filenameOri); err != nil {
		logger.Log.Errorf("Erro ao registrar arquivo no banco: %v", err)
		response.HandleError(c, http.StatusInternalServerError, "Erro ao registrar arquivo no banco", err.Error(), requestID)
		return
	}

	rsp := gin.H{
		"message": "Arquivo transferido com sucesso",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/*
 * Devolve os registros da tabela 'uploads' para um determinado contexto.
 *
 * - **Rota**: "/contexto/documentos/upload/:id"
 * - **Params**: ID do Contexto
 * - **Método**: GET
 * - **Resposta**:
 *   {
 *     IdFile    int       // ID do arquivo
 *     IdCtxt    int       // ID do contexto
 *     NmFileNew string    // Nome do arquivo novo
 *     NmFileOri string    // Nome do arquivo original
 *     SnAutos   string    // Indicação se é relacionado a autos
 *     DtInc     time.Time // Data de inclusão
 *     Status    string    // Status do arquivo
 *   }
 */

func (service *UploadHandlerType) SelectHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------
	// Extrai o parâmetro id da rota
	ctxtID := c.Param("id")

	// Converte id para inteiro
	id, err := strconv.Atoi(ctxtID)
	if err != nil {

		logger.Log.Error("ID do contexto inválido:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "ID do contexto inválido:", err.Error(), requestID)
		return
	}

	//rows, err := service.Model.SelectRowsByContextoId(id)
	rows, err := service.Service.SelectByContexto(id)
	if err != nil {

		logger.Log.Error("Erro na inclusão do contexto:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro na inclusão do contexto:", err.Error(), requestID)
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Registros selecionados com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

/*
 * Devolve todos os registros da tabela 'uploads'.
 *
 * - **Rota**: "/contexto/documentos/upload/"
 * - **Params**:
 * - **Método**: GET
 * - **Resposta**:
 *   {
 *     IdFile    int       // ID do arquivo
 *     IdCtxt    int       // ID do contexto
 *     NmFileNew string    // Nome do arquivo novo
 *     NmFileOri string    // Nome do arquivo original
 *     SnAutos   string    // Indicação se é relacionado a autos
 *     DtInc     time.Time // Data de inclusão
 *     Status    string    // Status do arquivo
 *   }
 */
func (service *UploadHandlerType) SelectAllHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------
	//var res string
	var dataRows []models.UploadRow

	uploadModel := models.NewUploadModel(pgdb.DBPoolGlobal.Pool)

	dataRows, err := uploadModel.SelectRows()
	if err != nil {

		logger.Log.Error("Erro ao selecionar arquivos transferidos:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao selecionar arquivos transferidos: ", err.Error(), requestID)
		return
	}

	rsp := gin.H{
		"rows":    dataRows,
		"message": "Executado com sucesso!",
	}

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/*
 * Deleta os registros da tabela 'uploads' e respectivos arquivos da pasta 'upload'.
 *
 * - **Rota**: "/contexto/documentos/upload"
 * - **Params**:
 * - **Método**: DELETE
 * - **Body: regKeys:
 *		[
 *			{
 * 				idContexto: number,
 *	  			idFile: number,
 *	  		},
 *		]
 * - **Resposta**:
 *   {
 *     IdFile    int       // ID do arquivo
 *     IdCtxt    int       // ID do contexto
 *     NmFileNew string    // Nome do arquivo novo
 *     NmFileOri string    // Nome do arquivo original
 *     SnAutos   string    // Indicação se é relacionado a autos
 *     DtInc     time.Time // Data de inclusão
 *     Status    string    // Status do arquivo
 *   }
 */
type paramsBodyUploadDelete struct {
	IdContexto int
	IdFile     int
}

func (service *UploadHandlerType) DeleteHandler(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	var deleteFiles []paramsBodyUploadDelete

	// Decodifica o corpo da requisição
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&deleteFiles); err != nil {

		logger.Log.Error("Dados inválidos!:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Dados inválidos!: ", err.Error(), requestID)
		return
	}

	// Validação inicial
	if len(deleteFiles) == 0 {

		logger.Log.Error("Arquivos não informados")
		response.HandleError(c, http.StatusBadRequest, "Arquivos não informados: ", "", requestID)
		return
	}

	// Rastreamento de resultados
	var deletedFiles []int
	var failedFiles []int

	// Processa os arquivos para deleção
	for _, reg := range deleteFiles {
		// Busca o registro no banco
		row, err := service.Service.SelectById(reg.IdFile)
		if err != nil {

			logger.Log.Error("Arquivo não encontrado:", err.Error())
			failedFiles = append(failedFiles, reg.IdFile)
			continue
		}

		// Deleta o registro do banco
		err = service.Service.DeleteRegistro(reg.IdFile)
		if err != nil {

			logger.Log.Error("Erro ao deletar registro:", err.Error())
			failedFiles = append(failedFiles, reg.IdFile)
			continue
		}

		// Deleta o arquivo do sistema de arquivos
		fullFileName := filepath.Join("uploads", row.NmFileNew)
		if service.FileExist(fullFileName) {
			err = service.DeletarFile(fullFileName)
			if err != nil {

				logger.Log.Error("Erro ao deletar arquivo físico:", err.Error())
				failedFiles = append(failedFiles, reg.IdFile)
				continue
			}
		}

		// Adiciona ao rastreamento de sucessos
		deletedFiles = append(deletedFiles, reg.IdFile)
	}

	// Monta a resposta
	rsp := gin.H{

		"message": "Processamento concluído",
		"deleted": deletedFiles,
		"errors":  failedFiles,
	}

	// Retorna a resposta padronizada

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

func (service *UploadHandlerType) DeleteHandlerById(c *gin.Context) {

	//Generate request ID for tracing
	requestID := middleware.GetRequestID(c)
	//--------------------------------------

	idFile, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Log.Error("IdDoc inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Formado do IdDoc inválidos", "", requestID)
		return
	}

	// Processa os arquivos para deleção

	// Busca o registro no banco
	row, err := service.Service.SelectById(idFile)
	if err != nil {
		logger.Log.Error("Registro não encontrado:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Dados inválidos!: ", err.Error(), requestID)
		return
	}

	// Deleta o registro do banco
	err = service.Service.DeleteRegistro(idFile)
	if err != nil {
		logger.Log.Error("Erro ao deletar registro:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao deletar o registro!: ", err.Error(), requestID)
		return
	}

	// Deleta o arquivo do sistema de arquivos
	fullFileName := filepath.Join("uploads", row.NmFileNew)
	if service.FileExist(fullFileName) {
		err = service.DeletarFile(fullFileName)
		if err != nil {

			logger.Log.Error("Erro ao deletar arquivo físico:"+fullFileName, err.Error())
			response.HandleError(c, http.StatusBadRequest, "Erro ao deletar o arquivo: ", err.Error(), requestID)
			return
		}
	}

	rsp := gin.H{
		"rows":    nil,
		"message": "Documento(s) deletado(s) com sucesso!",
	}

	response.HandleSuccess(c, http.StatusNoContent, rsp, requestID)
}

/* Verifica apenas se o arquivo existe. */
func (service *UploadHandlerType) FileExist(fullFileName string) bool {
	_, err := os.Stat(fullFileName)
	return !os.IsNotExist(err)

}

// Deleta um arquivo
func (service *UploadHandlerType) DeletarFile(fullFileName string) error {
	err := os.Remove(fullFileName)
	if err != nil {

		logger.Log.Error("Erro ao deletar arquivo:", err.Error())
		return err
	}

	logger.Log.Info("Arquivo deletado com sucesso: " + fullFileName)
	return nil
}

/*
Insere um registro na tabela uploads para cada arquivo transferido para o servidor
por upload.
*/
func (service *UploadHandlerType) InsertUploadedFile(idCtxt int, fileName string, fileNameOri string) error {
	// Validações de entrada
	if idCtxt <= 0 {
		return fmt.Errorf("ID de contexto inválido: %d", idCtxt)
	}
	if fileName == "" {
		return fmt.Errorf("Nome do arquivo não pode ser vazio")
	}
	if fileNameOri == "" {
		return fmt.Errorf("Nome original do arquivo não pode ser vazio")
	}

	// Popula o registro
	// reg := models.UploadRow{
	// 	NmFileNew: fileName,
	// 	NmFileOri: fileNameOri,
	// 	IdCtxt:    idCtxt,
	// 	SnAutos:   "N",
	// 	Status:    "S",
	// 	DtInc:     time.Now(),
	// }

	// Usa o modelo para inserir o registro

	_, err := service.Service.InserirRegistro(idCtxt, fileName, fileNameOri)
	if err != nil {

		logger.Log.Error("Erro ao inserir Registro: " + fileName)
		return fmt.Errorf("falha ao inserir registro no banco de dados: %w", err)
	}

	// Log de sucesso

	logger.Log.Info("Registro inserido com sucesso: " + fileName)

	return nil
}

// /*
// Analisa todos os documentos inseridos na tabela "autos_temp", excluindo os registros que não
// correspondam a documentos válidos para a juntada.
// */
// func (service *UploadHandlerType) JuntadaByContextHandler(c *gin.Context) {
// 	requestID := middleware.GetRequestID(c)
// 	idStr := c.Param("id")
// 	if idStr == "" {
// 		c.JSON(http.StatusBadRequest, msgs.CreateResponseMessage("Parâmetro id é obrigatório"))
// 		return
// 	}

// 	idContexto, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, msgs.CreateResponseMessage("Parâmetro id inválido"))
// 		return
// 	}

// 	//Faz um loop nos registros do indice "Autos_temp" para analisar cada uma dos registros,
// 	//e identificar a natureza, excluindo o que for lixo. Esta é a primeira verificação dos
// 	//documentos extraídos do PDF

// 	rows, err := services.Autos_tempServiceGlobal.SelectByContexto(idContexto)
// 	if err != nil {
// 		logger.Log.Errorf("Erro ao buscar arquivos pelo contexto %d: %v", idContexto, err)
// 		c.JSON(http.StatusInternalServerError, msgs.CreateResponseMessage("Erro ao buscar arquivos"))
// 		return
// 	}

// 	if len(rows) == 0 {
// 		c.JSON(http.StatusNotFound, msgs.CreateResponseMessage("Nenhum arquivo encontrado para o contexto informado"))
// 		return
// 	}

// 	var wg sync.WaitGroup
// 	var mu sync.Mutex // Protege chamadas concorrentes de DeleteRow caso não seja thread-safe

// 	// Usar canal para capturar erros na verificação (opcional)
// 	errCh := make(chan error, len(rows))

// 	for _, row := range rows {
// 		wg.Add(1)
// 		deletar := false

// 		// Copiar a variável para evitar problemas com closure
// 		rowCopy := row

// 		go func() {
// 			defer wg.Done()

// 			//Rotina que faz o trabalho pesado de verificação de cada registro
// 			natuDoc, err := service.Service.VerificarNaturezaDocumento(c.Request.Context(), rowCopy.Doc)
// 			if err != nil {
// 				logger.Log.Errorf("Erro ao verificar a natureza do documento: %s", rowCopy.IdPje)
// 				return
// 			}

// 			logger.Log.Infof("Natureza documento %s identificada: key=%d, description=%s", rowCopy.IdPje, natuDoc.Key, natuDoc.Description)

// 			if natuDoc.Key == consts.NATU_DOC_OUTROS || natuDoc.Key == consts.NATU_DOC_CERTIDOES || natuDoc.Key == consts.NATU_DOC_MOVIMENTACAO {
// 				deletar = true
// 			}

// 			if deletar {
// 				mu.Lock()
// 				defer mu.Unlock()
// 				if err := services.Autos_tempServiceGlobal.DeletaAutos(rowCopy.Id); err != nil {
// 					logger.Log.Errorf("Erro ao deletar documento ID %d: %v", rowCopy.Id, err)
// 					errCh <- err
// 				}
// 			}
// 		}()
// 	}

// 	// Aguarda todas as goroutines finalizarem
// 	wg.Wait()
// 	close(errCh)

// 	// Opcional: verificar se houve erros e registrar
// 	var hadErrors bool
// 	for _ = range errCh {
// 		hadErrors = true
// 		// Aqui já logou, pode acumular ou manipular erros se desejar
// 	}

// 	if hadErrors {
// 		c.JSON(http.StatusInternalServerError, msgs.CreateResponseMessage("Alguns erros ocorreram no processamento dos documentos"))
// 		return
// 	}

// 	rsp := gin.H{
// 		"message": "Processamento concluído com sucesso!",
// 	}

// 	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
// }
