package libocr

import (
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"ocrserver/controllers"
	"ocrserver/models"
	"ocrserver/utils/msgs"

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
 * Extrai com OCR o texto dos arquivos PDF, salva na tabela 'temp_autos'
 *
 * - **Rota**: "/contexto/documentos"
 * - **Params**:
 * - **Método**: POST
 * - **Status**: 201/400/500,
 * - **Body:
 *		{
 *			IdContexto: idContexto,
 *			IdFile: row.IdFile,
 *		};
 * - **Resposta**:
 *  	{
 * 			ExtractedErros: int,
 *		 	ExtractedFiles: string,
 *		}
 */

type BodyParamsOCR struct {
	IdContexto int
	IdFile     int
}

func OcrFileHandler(c *gin.Context) {
	bodyParams := []BodyParamsOCR{}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyParams); err != nil {
		response := msgs.CreateResponseMessage("Body params inválidos!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validação inicial
	if len(bodyParams) == 0 {
		response := msgs.CreateResponseMessage("Body não possui arquivos para extrair!")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Rastreamento de resultados
	var extractedFiles []string
	var extractedErros []int

	uploadModel := models.NewUploadModel()
	uploadController := controllers.NewUploadController()
	// Processa os arquivos para deleção
	for _, reg := range bodyParams {

		// Busca o arquivo pelo IdFile no banco
		row, err := uploadModel.SelectRowById(reg.IdFile)
		if err != nil {
			log.Printf("Arquivo não encontrado em temp_uploads - id_file=%d - contexto=%d", reg.IdFile, reg.IdContexto)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		filePath := filepath.Join("uploads", row.NmFileNew)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("Arquivo não encontrado - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		resultText, err := processPDFWithPipeline(filePath)
		if err != nil {
			log.Printf("Erro na extração do texto - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}
		//Salva o texto extraído na tabela temp_autos
		err = SalvaTextoExtraido(reg.IdContexto, row.NmFileNew, row.NmFileOri, resultText)
		if err != nil {
			log.Printf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		// Deleta o registro do banco
		err = uploadModel.DeleteRow(reg.IdFile)
		if err != nil {
			log.Printf("Erro ao deletar o registro no banco - id_file=%d", reg.IdFile)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		// Deleta o arquivo do sistema de arquivos
		fullFileName := filepath.Join("uploads", row.NmFileNew)
		if uploadController.FileExist(fullFileName) {
			err = uploadController.DeletarFile(fullFileName)
			if err != nil {
				log.Printf("Erro ao deletar o arquivo físico - %s", fullFileName)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}
		}

		// Adiciona ao rastreamento de sucessos
		extractedFiles = append(extractedFiles, row.NmFileNew)
	}

	response := gin.H{
		"extractedErros": extractedErros,
		"extractedFiles": extractedFiles,
	}

	// Retorna a resposta padronizada
	c.JSON(http.StatusOK, response)

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
func SalvaTextoExtraido(idCtxt int, fileNameNew string, fileNameOri string, texto string) error {

	var reg models.TempAutosRow
	reg.IdCtxt = idCtxt
	reg.NmFileNew = fileNameNew
	reg.NmFileOri = fileNameOri
	reg.TxtDoc = texto
	reg.DtInc = time.Now()
	reg.Status = "S"

	serviceTempautos := models.NewTempautosModel()
	_, err := serviceTempautos.InsertRow(reg)
	if err != nil {
		log.Printf("Erro ao inserir linha: %v", err)
		return err
	}
	//log.Printf("Id do registro: %d.", id)
	return nil

}
