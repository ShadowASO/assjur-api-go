package libocr

import (
	"fmt"
	"image/png"
	"log"
	"net/http"
	"ocrserver/models"

	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/gin-gonic/gin"
	"github.com/otiai10/gosseract/v2"
)

const CONTEXTO_TEMP = 18

// Exclui arquivo após uso
func deleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("erro ao excluir arquivo %s: %w", filePath, err)
	}
	return nil
}

// Realiza OCR em uma imagem
func extractTextFromImage(imagePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage(imagePath)
	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("erro ao extrair texto da imagem: %w", err)
	}
	return text, nil
}

/*
*
  - Extrai com OCR o texto dos arquivos PDF, salva na tabela 'temp_autos'
  - e deleta o arquivo.
  - Rota: "/contexto/documentos" *
  - Body: regKeys: [ {
    idContexto: number,
    idFile: number,
    }]
  - Método: POST
*/
func OcrFileHandler(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nome do arquivo não fornecido"})
		return
	}

	filePath := filepath.Join("uploads", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}
	resultText, err := processPDFWithPipeline(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Arquivo processado com erro"})
		return
	}
	SalvaTextoExtraido(CONTEXTO_TEMP, filename, filename, resultText)

	c.Data(http.StatusOK, "text/plain", []byte(fmt.Sprintf("Texto extraído do arquivo:\n\n%s", resultText)))
}

func processPDFWithPipeline(pdfPath string) (string, error) {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir PDF: %w", err)
	}
	defer doc.Close()

	var wg sync.WaitGroup
	texts := make([]string, doc.NumPage()) // Slice para armazenar o texto de cada página
	errChan := make(chan error, doc.NumPage())

	for i := 0; i < doc.NumPage(); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			img, err := doc.Image(i)
			if err != nil {
				errChan <- fmt.Errorf("erro ao converter página %d: %w", i+1, err)
				return
			}

			imagePath := fmt.Sprintf("uploads/images/page_%d.png", i+1)
			file, err := os.Create(imagePath)
			if err != nil {
				errChan <- fmt.Errorf("erro ao criar arquivo de imagem: %w", err)
				return
			}
			defer file.Close()

			if err := png.Encode(file, img); err != nil {
				errChan <- fmt.Errorf("erro ao salvar imagem PNG: %w", err)
				return
			}

			text, err := extractTextFromImage(imagePath)
			if err != nil {
				errChan <- fmt.Errorf("erro ao extrair texto da imagem: %w", err)
				return
			}

			texts[i] = text // Armazena o texto na posição correspondente à página
			deleteFile(imagePath)
		}(i)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return "", <-errChan
	}

	// Concatena os textos na ordem correta
	var resultText string
	for _, text := range texts {
		resultText += text + "\n"
	}

	return resultText, nil
}
func SalvaTextoExtraido(idCtxt int, fileNameNew string, fileNameOri string, texto string) {
	//var reg autos.AutosRow
	var reg models.TempAutosRow
	reg.IdCtxt = idCtxt
	reg.NmFileNew = fileNameNew
	reg.NmFileOri = fileNameOri
	reg.TxtDoc = texto
	reg.DtInc = time.Now()
	reg.Status = "S"

	serviceTempautos := models.NewTempautosModel()
	id, err := serviceTempautos.InsertRow(reg)
	if err != nil {
		log.Printf("Erro ao inserir linha: %v", err)
		return
	}
	log.Printf("Id do registro: %d.", id)

}
