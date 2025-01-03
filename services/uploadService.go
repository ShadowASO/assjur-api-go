package uploadServices

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"ocrserver/models"

	"github.com/gin-gonic/gin"

	//"ocrserver/models/uploadModel"
	"os"
	"path/filepath"
	"time"
)

const CONTEXTO_TEMP = 18

// Função para gerar um nome único para o arquivo (essa é apenas uma sugestão, personalize conforme necessário)
func generateUniqueFileName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func UploadFileHandler(c *gin.Context) {
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

	// Obtém o arquivo enviado
	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao obter o arquivo"})
		return
	}
	defer file.Close()

	// Gera um nome único para o arquivo
	uniqueFileName := generateUniqueFileName() + filepath.Ext(handler.Filename)
	FileNameOri := handler.Filename

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
	if err := RegistraUploadedFile(CONTEXTO_TEMP, uniqueFileName, FileNameOri); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Erro ao registrar o arquivo em temp_updatefile", "file": uniqueFileName})
		return
	}
	// Retorna sucesso com o nome do arquivo salvo
	c.JSON(http.StatusOK, gin.H{"message": "Arquivo enviado com sucesso", "file": uniqueFileName})

}
func ListaUploadFileHandler(c *gin.Context) {
	//var res string
	var todos []models.UploadRow
	//todos, err := upload.SelectRows()
	uploadModel := models.NewUploadModel()
	//todos, err := models.UploadModel.SelectRows()
	todos, err := uploadModel.SelectRows()
	if err != nil {
		log.Fatalln(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar tabela temp_uploadfiles\n"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"todos": todos})

}
func RegistraUploadedFile(idCtxt int, fileName string, fileNameOri string) error {
	var reg models.UploadRow
	reg.NmFileNew = fileName
	reg.NmFileOri = fileNameOri
	reg.IdCtxt = idCtxt
	reg.SnAutos = "N"
	reg.Status = "S"
	reg.DtInc = time.Now()

	uploadModel := models.NewUploadModel()
	id, err := uploadModel.InsertRow(reg)

	log.Printf("Id do registro: %d.", id)

	return err
}
