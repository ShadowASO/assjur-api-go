package libocr

import (
	"context"

	"fmt"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"

	//services "ocrserver/doc/OpenaiApi"
	"ocrserver/internal/database/pgdb"
	"ocrserver/internal/handlers"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"
	"ocrserver/internal/utils/msgs"

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

// Função genérica para processar OCR dado um slice de BodyParamsOCR
func processOCRFiles(ctx context.Context, bodyParams []BodyParamsOCR) (extractedFiles []string, extractedErros []int) {
	uploadModel := models.NewUploadModel(pgdb.DBPoolGlobal.Pool)
	uploadController := handlers.NewUploadHandlers(uploadModel)

	for _, reg := range bodyParams {
		autuar := true
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

		var resultText string
		ext := strings.ToLower(filepath.Ext(row.NmFileNew))
		if ext == ".txt" {
			bytesContent, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Erro ao ler arquivo txt - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}
			resultText = string(bytesContent)
			autuar, _ = VerificarNaturezaDocumento(ctx, resultText)

		} else {
			resultText, err = processPDFWithPipeline(filePath)
			if err != nil {
				log.Printf("Erro na extração do texto - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}
		}

		if autuar {
			err = SalvaTextoExtraido(reg.IdContexto, row.NmFileNew, row.NmFileOri, resultText)
			if err != nil {
				log.Printf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}
		}

		if err := deletarRegistro(uploadModel, reg.IdFile); err != nil {
			log.Printf("Erro ao deletar o registro no banco - id_file=%d", reg.IdFile)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		if err := deletarArquivo(uploadController, filePath); err != nil {
			log.Printf("Erro ao deletar o arquivo físico - %s", filePath)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		extractedFiles = append(extractedFiles, row.NmFileNew)
	}

	return extractedFiles, extractedErros
}

// Método: POST
// URL: "/contexto/documentos/ocr/"
// Processa e extrai por OCR todos os documentos indicados no body e contidos na tabela "uploadfiles"
func OcrFileHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	bodyParams := []BodyParamsOCR{}
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

	extractedFiles, extractedErros := processOCRFiles(c.Request.Context(), bodyParams)

	rsp := gin.H{
		"extractedErros": extractedErros,
		"extractedFiles": extractedFiles,
		"message":        "Registros selecionados com sucesso!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

// Método: POST
// URL: "/contexto/documentos/ocr/:id"
// Processa e extrai por OCR todos os arquivos do contexto contidos na tabela "uploadfiles"
func OcrByContextHandler(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, msgs.CreateResponseMessage("Parâmetro id é obrigatório"))
		return
	}

	idContexto, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, msgs.CreateResponseMessage("Parâmetro id inválido"))
		return
	}

	uploadModel := models.NewUploadModel(pgdb.DBPoolGlobal.Pool)
	// Busca os arquivos com IdContexto
	rows, err := uploadModel.SelectRowsByContextoId(idContexto)
	if err != nil {
		log.Printf("Erro ao buscar arquivos pelo contexto %d: %v", idContexto, err)
		c.JSON(http.StatusInternalServerError, msgs.CreateResponseMessage("Erro ao buscar arquivos"))
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusNotFound, msgs.CreateResponseMessage("Nenhum arquivo encontrado para o contexto informado"))
		return
	}

	// Monta slice BodyParamsOCR para processar
	var bodyParams []BodyParamsOCR
	for _, row := range rows {
		bodyParams = append(bodyParams, BodyParamsOCR{
			IdContexto: idContexto,
			IdFile:     row.IdFile,
		})
	}

	_, _ = processOCRFiles(c.Request.Context(), bodyParams)

	rsp := gin.H{
		"message": "Aguarde a conclusão do processamento!",
	}

	response.HandleSuccess(c, http.StatusOK, rsp, requestID)

}

func deletarRegistro(uploadModel *models.UploadModelType, idFile int) error {
	err := uploadModel.DeleteRow(idFile)
	if err != nil {
		log.Printf("Erro ao deletar o registro no banco - id_file=%d: %v", idFile, err)
	}
	return err
}

func deletarArquivo(uploadController *handlers.UploadHandlerType, filePath string) error {
	if uploadController.FileExist(filePath) {
		err := uploadController.DeletarFile(filePath)
		if err != nil {
			log.Printf("Erro ao deletar o arquivo físico - %s: %v", filePath, err)
			return err
		}
	}
	return nil
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

	serviceTempautos := models.NewTempautosModel(pgdb.DBPoolGlobal.Pool)
	_, err := serviceTempautos.InsertRow(reg)
	if err != nil {
		log.Printf("Erro ao inserir linha: %v", err)
		return err
	}
	//log.Printf("Id do registro: %d.", id)
	return nil

}
func VerificarNaturezaDocumento(ctx context.Context, texto string) (bool, error) {

	//var messages services.MsgGpt
	var msgs services.MsgGpt
	assistente := `O seguinte texto pertence aos autos de um processo judicial. Analise o texto e verifique se o documento pode ser 
	classificado como uma petição inicial, contestação, réplica, despacho inicial, despacho ordinatório, petição diversa, decisão
	interlocutória, sentença, embargos de declaração, contra-razões, apelação ou laudo pericial. Não confunda certidões com as peças
	processuais. Muitas certidões reproduzem tais peças, mão não são a mesma coisa. Fique atendo para as indicações de certidão. 
	Responda apenas "sim" ou "não"`

	msgs.CreateMessage("", services.ROLE_USER, assistente)
	msgs.CreateMessage("", services.ROLE_USER, texto)

	retSubmit, err := services.OpenaiServiceGlobal.SubmitPromptResponse(ctx, msgs, nil, "gpt-4.1-nano")
	if err != nil {
		logger.Log.Errorf("Erro no SubmitPrompt: %s", err)
		return false, erros.CreateError("Erro ao verificar a natureza do documento!")
	}

	resp := strings.TrimSpace(strings.ToLower(retSubmit.OutputText()))
	logger.Log.Infof("Resposta do nano: %s", resp)

	switch resp {
	case "sim":
		return true, nil
	case "não", "nao", "nâo": // cobre variações de "não"
		return false, nil
	default:
		// Caso a resposta não seja clara, considera erro ou false
		logger.Log.Warningf("Resposta inesperada do modelo: %q", resp)
		return false, erros.CreateError("Resposta inesperada do modelo")
	}
}
