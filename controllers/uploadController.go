package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ocrserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
	"time"
)

type UploadControllerType struct {
	uploadModel *models.UploadModelType
}

const CONTEXTO_TEMP = 18

func NewUploadController() *UploadControllerType {
	model := models.NewUploadModel()
	return &UploadControllerType{uploadModel: model}
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
func (service *UploadControllerType) UploadFileHandler(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Método não permitido"})
		return
	}

	// Limita o tamanho do corpo da requisição (10 MB neste exemplo)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)

	// Parse da requisição multipart/form-data
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar o formulário"})
		return
	}
	// Obtém os valores do formulário
	filenameOri := c.PostForm("filename_ori")

	idCtxt := c.PostForm("idContexto")
	idContexto, err := strconv.Atoi(idCtxt)
	if err != nil {
		//c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID inválido!"})
		response := gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "ID do contexto inválido!",
			"rows":       nil,
		}

		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Valida os valores extras enviados
	if idContexto == 0 || filenameOri == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campos idContexto ou filename_ori estão ausentes"})
		return
	}

	// Obtém o arquivo enviado
	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao obter o arquivo"})
		return
	}
	defer file.Close()

	// Gera um nome único para o arquivo
	uniqueFileName := generateUniqueFileName() + filepath.Ext(handler.Filename)
	//FileNameOri := handler.Filename

	// Define o caminho para salvar o arquivo
	savePath := filepath.Join("uploads", uniqueFileName)

	// Cria o diretório "uploads" se não existir
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar o diretório"})
		return
	}

	// Cria o arquivo no disco
	dst, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar o arquivo"})
		return
	}
	defer dst.Close()

	// Copia o conteúdo do arquivo enviado para o arquivo no disco
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar o arquivo"})
		return
	}
	if err := service.InsertUploadedFile(idContexto, uniqueFileName, filenameOri); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Erro ao registrar o arquivo em temp_updatefile", "file": uniqueFileName})
		return
	}
	// Retorna sucesso com o nome do arquivo salvo
	c.JSON(http.StatusOK, gin.H{"message": "Arquivo enviado com sucesso", "file": uniqueFileName})

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

func (service *UploadControllerType) SelectHandler(c *gin.Context) {
	// Extrai o parâmetro id da rota
	ctxtID := c.Param("id")

	// Converte id para inteiro
	id, convErr := strconv.Atoi(ctxtID)
	if convErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensagem": "ID do contexto inválido"})
		return
	}
	log.Printf("contexto ID: Params=%v", ctxtID)

	rows, err := service.uploadModel.SelectRowsByContextoId(id)
	if err != nil {
		response := gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "Erro na inclusão do contexto!",
			"rows":       nil,
		}

		c.JSON(http.StatusCreated, response)
		return
	}

	// Retorna os dados do usuário
	retOK := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Executado com sucesso!",
		"rows":       rows,
	}
	c.JSON(http.StatusOK, retOK)
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
func (service *UploadControllerType) SelectAllUploadFilesHandler(c *gin.Context) {
	//var res string
	var dataRows []models.UploadRow

	uploadModel := models.NewUploadModel()

	dataRows, err := uploadModel.SelectRows()
	if err != nil {
		response := gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "Erro na inclusão do contexto!",
			"rows":       nil,
		}

		c.JSON(http.StatusCreated, response)
		return
	}
	// Retorna os dados do usuário
	retRows := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Executado com sucesso!",
		"rows":       dataRows,
	}
	c.JSON(http.StatusOK, retRows)

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

func (service *UploadControllerType) DeleteHandler(c *gin.Context) {
	var deleteFiles []paramsBodyUploadDelete

	// Decodifica o corpo da requisição
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&deleteFiles); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "Dados inválidos",
		})
		return
	}

	// Validação inicial
	if len(deleteFiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":         false,
			"statusCode": http.StatusBadRequest,
			"message":    "Nenhum arquivo para deletar",
		})
		return
	}

	// Rastreamento de resultados
	var deletedFiles []int
	var failedFiles []int

	// Processa os arquivos para deleção
	for _, reg := range deleteFiles {
		// Busca o registro no banco
		row, err := service.uploadModel.SelectRowById(reg.IdFile)
		if err != nil {
			log.Printf("Arquivo não encontrado - id_file=%d - contexto=%d", reg.IdFile, reg.IdContexto)
			failedFiles = append(failedFiles, reg.IdFile)
			continue
		}

		// Deleta o registro do banco
		err = service.uploadModel.DeleteRow(reg.IdFile)
		if err != nil {
			log.Printf("Erro ao deletar o registro no banco - id_file=%d", reg.IdFile)
			failedFiles = append(failedFiles, reg.IdFile)
			continue
		}

		// Deleta o arquivo do sistema de arquivos
		fullFileName := filepath.Join("uploads", row.NmFileNew)
		if fileExist(fullFileName) {
			err = service.deletarFile(fullFileName)
			if err != nil {
				log.Printf("Erro ao deletar o arquivo físico - %s", fullFileName)
				failedFiles = append(failedFiles, reg.IdFile)
				continue
			}
		}

		// Adiciona ao rastreamento de sucessos
		deletedFiles = append(deletedFiles, reg.IdFile)
	}

	// Monta a resposta
	response := gin.H{
		"ok":         true,
		"statusCode": http.StatusOK,
		"message":    "Processamento concluído",
		"deleted":    deletedFiles,
		"errors":     failedFiles,
	}

	// Retorna a resposta padronizada
	c.JSON(http.StatusOK, response)
}

/* Verifica apenas se o arquivo existe. */
func fileExist(fullFileName string) bool {
	_, err := os.Stat(fullFileName)
	return !os.IsNotExist(err)

}

// Deleta um arquivo
func (service *UploadControllerType) deletarFile(fullFileName string) error {
	err := os.Remove(fullFileName)
	if err != nil {
		fmt.Printf("Erro ao deletar o arquivo: %s\n", err)
		return err
	}
	fmt.Printf("Arquivo \"%s\" deletado com sucesso.\n", fullFileName)
	return nil
}

/*
Insere um registro na tabela temp_uploadfiles para cada arquivo transferido para o servidor
por upload.
*/
func (service *UploadControllerType) InsertUploadedFile(idCtxt int, fileName string, fileNameOri string) error {
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
	id, err := service.uploadModel.InsertRow(reg)
	if err != nil {
		log.Printf("Erro ao inserir registro (ID Contexto: %d, Arquivo: %s): %v", idCtxt, fileName, err)
		return fmt.Errorf("falha ao inserir registro no banco de dados: %w", err)
	}

	// Log de sucesso
	log.Printf("Registro inserido com sucesso. ID: %d, ID Contexto: %d, Arquivo: %s.", id, idCtxt, fileName)

	return nil
}
