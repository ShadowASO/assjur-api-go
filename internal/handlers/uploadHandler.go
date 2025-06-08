package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/utils/logger"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ocrserver/internal/database/pgdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandlerType struct {
	Model *models.UploadModelType
}

const CONTEXTO_TEMP = 18

func NewUploadHandlers(model *models.UploadModelType) *UploadHandlerType {

	return &UploadHandlerType{Model: model}
}

// Função para gerar um nome único para o arquivo (essa é apenas uma sugestão, personalize conforme necessário)
func generateUniqueFileName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

/*
*
  - Faz o upload de um arquivo e cria um registro na tabela 'temp_uploadfiles'
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
	requestID := uuid.New().String()
	log.Println("Iniciando o processamento do upload de arquivo")

	if c.Request.Method != http.MethodPost {

		logger.Log.Error("Método não permitido")
		response.HandleError(c, http.StatusMethodNotAllowed, "Método não permitido", "", requestID)
		return
	}

	// Limita o tamanho do corpo da requisição (10 MB neste exemplo)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)

	// Parse da requisição multipart/form-data
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {

		logger.Log.Error("Erro ao processar o formulário:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao processar o formulário", err.Error(), requestID)
		return
	}
	// Obtém os valores do formulário
	filenameOri := c.PostForm("filename_ori")

	idCtxt := c.PostForm("idContexto")
	idContexto, err := strconv.Atoi(idCtxt)
	if err != nil {

		logger.Log.Error("ID do contexto inválido:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "ID do contexto inválido:", err.Error(), requestID)
		return
	}

	// Valida os valores extras enviados
	if idContexto == 0 || filenameOri == "" {

		logger.Log.Error("Campos idContexto/filename_ori ausentes:")
		response.HandleError(c, http.StatusBadRequest, "Campos idContexto/filename_ori ausentes", "", requestID)
		return
	}

	// Obtém o arquivo enviado
	file, handler, err := c.Request.FormFile("file")
	if err != nil {

		logger.Log.Error("Erro ao obter o arquivo:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro ao obter o arquivo:", err.Error(), requestID)
		return

	}
	defer file.Close()

	// Gera um nome único para o arquivo
	uniqueFileName := generateUniqueFileName() + filepath.Ext(handler.Filename)

	// Define o caminho para salvar o arquivo
	savePath := filepath.Join("uploads", uniqueFileName)

	// Cria o diretório "uploads" se não existir
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {

		logger.Log.Error("Erro ao criar o diretório uploads:", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Erro ao criar o diretório uploads:", err.Error(), requestID)
		return
	}

	// Cria o arquivo no disco
	dst, err := os.Create(savePath)
	if err != nil {

		logger.Log.Error("Erro ao salvar o arquivo:", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Erro ao salvar o arquivo:", err.Error(), requestID)
		return
	}
	defer dst.Close()

	// Copia o conteúdo do arquivo enviado para o arquivo no disco
	if _, err := io.Copy(dst, file); err != nil {

		logger.Log.Error("Erro ao salvar o conteúdo do arquivo:", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Erro ao salvar o conteúdo do arquivo:", err.Error(), requestID)
		return
	}
	if err := service.InsertUploadedFile(idContexto, uniqueFileName, filenameOri); err != nil {

		logger.Log.Error("Erro ao registrar o arquivo no banco de dados:", err.Error())
		response.HandleError(c, http.StatusInternalServerError, "Erro ao registrar o arquivo no banco de dados:", err.Error(), requestID)
		return
	}
	// Retorna sucesso com o nome do arquivo salvo

	rsp := gin.H{
		"message": "Arquivo transferido com sucesso",
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)

}

/*
 * Devolve os registros da tabela 'temp_uploadfiles' para um determinado contexto.
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
	requestID := uuid.New().String()
	// Extrai o parâmetro id da rota
	ctxtID := c.Param("id")

	// Converte id para inteiro
	id, err := strconv.Atoi(ctxtID)
	if err != nil {

		logger.Log.Error("ID do contexto inválido:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "ID do contexto inválido:", err.Error(), requestID)
		return
	}

	rows, err := service.Model.SelectRowsByContextoId(id)
	if err != nil {

		logger.Log.Error("Erro na inclusão do contexto:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Erro na inclusão do contexto:", err.Error(), requestID)
		return
	}

	rsp := gin.H{
		"rows":    rows,
		"message": "Registros selecionados com sucesso!",
	}
	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

/*
 * Devolve todos os registros da tabela 'temp_uploadfiles'.
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
func (service *UploadHandlerType) SelectAllUploadFilesHandler(c *gin.Context) {
	//Generate request ID for tracing
	requestID := uuid.New().String()
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
 * Deleta os registros da tabela 'temp_uploadfiles' e respectivos arquivos da pasta 'upload'.
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
	requestID := uuid.New().String()
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
		row, err := service.Model.SelectRowById(reg.IdFile)
		if err != nil {

			logger.Log.Error("Arquivo não encontrado:", err.Error())
			failedFiles = append(failedFiles, reg.IdFile)
			continue
		}

		// Deleta o registro do banco
		err = service.Model.DeleteRow(reg.IdFile)
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

	response.HandleSuccess(c, http.StatusCreated, rsp, requestID)
}

func (service *UploadHandlerType) DeleteHandlerById(c *gin.Context) {

	requestID := uuid.New().String()
	idFile, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Log.Error("IdDoc inválidos", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Formado do IdDoc inválidos", "", requestID)
		return
	}

	// Processa os arquivos para deleção

	// Busca o registro no banco
	row, err := service.Model.SelectRowById(idFile)
	if err != nil {
		logger.Log.Error("Registro não encontrado:", err.Error())
		response.HandleError(c, http.StatusBadRequest, "Dados inválidos!: ", err.Error(), requestID)
		return
	}

	// Deleta o registro do banco
	err = service.Model.DeleteRow(idFile)
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
Insere um registro na tabela temp_uploadfiles para cada arquivo transferido para o servidor
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
	reg := models.UploadRow{
		NmFileNew: fileName,
		NmFileOri: fileNameOri,
		IdCtxt:    idCtxt,
		SnAutos:   "N",
		Status:    "S",
		DtInc:     time.Now(),
	}

	// Usa o modelo para inserir o registro

	_, err := service.Model.InsertRow(reg)
	if err != nil {

		logger.Log.Error("Erro ao inserir Registro: " + fileName)
		return fmt.Errorf("falha ao inserir registro no banco de dados: %w", err)
	}

	// Log de sucesso

	logger.Log.Info("Registro inserido com sucesso: " + fileName)

	return nil
}
