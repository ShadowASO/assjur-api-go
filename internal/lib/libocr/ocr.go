package libocr

import (
	"bufio"
	"context"
	"os/exec"
	"regexp"

	"fmt"
	"image/png"

	"net/http"
	"strconv"
	"strings"

	"ocrserver/internal/database/pgdb"
	"ocrserver/internal/handlers"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/models"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"
	"ocrserver/internal/utils/msgs"

	"os"
	"path/filepath"
	"sync"

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

/*
Função genérica destinada a processar a extração dos documentos contidos nos autos de cada
processo, e pode extrarir diretamente do arquivo PDF gerado pelo PJe, ou incorporá arquivos
txt já extraídos externamente. Não estamos utilizadon mais OCR, apesar das rotinas ainda
estarem disponíveis. Utilizaremos o utilitário linux "pdftotext" para converte o PDF p/TXT.
A rotina trabalha tanto com o PDF completo dos autos quanto de pelas individuais.
*/
func processOCRFiles(ctx context.Context, bodyParams []BodyParamsOCR) (extractedFiles []string, extractedErros []int) {
	uploadModel := models.NewUploadModel(pgdb.DBPoolGlobal.Pool)
	uploadController := handlers.NewUploadHandlers(uploadModel)

	for _, reg := range bodyParams {
		autuar := true
		row, err := uploadModel.SelectRowById(reg.IdFile)
		if err != nil {
			logger.Log.Errorf("Arquivo não encontrado em temp_uploads - id_file=%d - contexto=%d", reg.IdFile, reg.IdContexto)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		filePath := filepath.Join("uploads", row.NmFileNew)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			logger.Log.Errorf("Arquivo não encontrado - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		var resultText string
		ext := strings.ToLower(filepath.Ext(row.NmFileNew))
		if ext == ".txt" {
			bytesContent, err := os.ReadFile(filePath)
			if err != nil {
				logger.Log.Errorf("Erro ao ler arquivo txt - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}
			resultText = string(bytesContent)
			autuar, _ = VerificarNaturezaDocumento(ctx, resultText)

		} else {
			//****************************************************
			//TRATAMENTO DO ARQUIVO PDF
			//****************************************************
			autuar = false
			//Usando OCR - desativado
			//resultText, err = processPDFWithPipeline(filePath)
			//Usando o aplicativo pdftotext
			txtPath, err := processPDFToText(filePath)
			if err != nil {
				logger.Log.Errorf("Erro na extração do texto - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}
			//Extrair os documentos contidos no arquivo texto
			_, err = extrairDocumentosAutos(reg.IdContexto, row.NmFileOri, txtPath)
			if err != nil {
				logger.Log.Errorf("Erro na extração do texto - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}

			if err := deletarArquivo(uploadController, txtPath); err != nil {
				logger.Log.Errorf("Erro ao deletar o arquivo físico - %s", txtPath)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}

		}

		if autuar {
			err = SalvaTextoExtraido(reg.IdContexto, 0, row.NmFileNew, resultText)
			if err != nil {
				logger.Log.Errorf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}
		}

		if err := deletarRegistro(uploadModel, reg.IdFile); err != nil {
			logger.Log.Errorf("Erro ao deletar o registro no banco - id_file=%d", reg.IdFile)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		if err := deletarArquivo(uploadController, filePath); err != nil {
			logger.Log.Errorf("Erro ao deletar o arquivo físico - %s", filePath)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		extractedFiles = append(extractedFiles, row.NmFileNew)
	}

	return extractedFiles, extractedErros
}

// Método: POST
// URL: "/contexto/documentos/ocr/"
// Processa e extrai por OCR todos os documentos indicados no body e contidos na tabela "uploads"
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
// Processa e extrai por OCR todos os arquivos do contexto contidos na tabela "uploads"
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
		logger.Log.Errorf("Erro ao buscar arquivos pelo contexto %d: %v", idContexto, err)
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

/*
Analisa todos os documentos inseridos na tabela "autos_temp", excluindo os registros que não
correspondam a documentos válidos para a juntada.
*/
func JuntadaByContextHandler(c *gin.Context) {
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

	//***docsocrModel := models.NewDocsocrModel(pgdb.DBPoolGlobal.Pool)

	// Busca os arquivos com IdContexto
	//rows, err := docsocrModel.SelectByContexto(idContexto)
	rows, err := services.Autos_tempServiceGlobal.SelectByContexto(idContexto)
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

		// Copiar a variável para evitar problemas com closure
		rowCopy := row

		go func() {
			defer wg.Done()

			autuar, err := VerificarNaturezaDocumento(c.Request.Context(), rowCopy.Doc)
			if err != nil {
				logger.Log.Errorf("Erro ao verificar natureza do documento ID %d: %v", rowCopy.Id, err)
				errCh <- err
				return
			}
			if !autuar {
				mu.Lock()
				defer mu.Unlock()
				if err := services.Autos_tempServiceGlobal.DeletaAutos(rowCopy.Id); err != nil {
					logger.Log.Errorf("Erro ao deletar documento ID %d: %v", rowCopy.Id, err)
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

func deletarRegistro(uploadModel *models.UploadModelType, idFile int) error {
	err := uploadModel.DeleteRow(idFile)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar o registro no banco - id_file=%d: %v", idFile, err)
	}
	return err
}

func deletarArquivo(uploadController *handlers.UploadHandlerType, filePath string) error {
	if uploadController.FileExist(filePath) {
		err := uploadController.DeletarFile(filePath)
		if err != nil {
			logger.Log.Errorf("Erro ao deletar o arquivo físico - %s: %v", filePath, err)
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
func SalvaTextoExtraido(idCtxt int, idNatu int, idPje string, texto string) error {

	// var reg models.DocsocrRow
	// reg.IdCtxt = idCtxt
	// reg.NmFileNew = fileNameNew
	// reg.NmFileOri = fileNameOri
	// reg.TxtDoc = texto
	// reg.DtInc = time.Now()
	// reg.Status = "S"

	//serviceTempautos := models.NewDocsocrModel(pgdb.DBPoolGlobal.Pool)
	autos_temp := opensearch.NewAutos_tempIndex()

	//_, err := serviceTempautos.InsertRow(reg)
	_, err := autos_temp.Indexa(idCtxt, idNatu, idPje, texto, "")
	if err != nil {
		logger.Log.Errorf("Erro ao inserir linha: %v", err)
		return err
	}
	//log.Printf("Id do registro: %d.", id)
	return nil

}
func VerificarNaturezaDocumento(ctx context.Context, texto string) (bool, error) {

	//var messages services.MsgGpt
	var msgs services.MsgGpt
	assistente := `O seguinte texto pertence aos autos de um processo judicial. Identifique se é uma: 
	petição, contestação, réplica, despacho, decisão, sentença, embargos de declaração, recursos, contra-razões, apelação,
	procuração, ata de audiência ou laudo pericial. Responda apenas "sim" ou "não". Não confunda com certidões. As certidões possuem
	expressões tais como "certidão, certifico, teor do ato, por ordem do MM. Juiz, o referido é verdade, dou fé etc".  
	Quando for certidão, responda "não"`

	msgs.CreateMessage("", services.ROLE_USER, assistente)
	msgs.CreateMessage("", services.ROLE_USER, texto)

	retSubmit, err := services.OpenaiServiceGlobal.SubmitPromptResponse(ctx, msgs, nil, "gpt-4.1-nano")
	if err != nil {
		logger.Log.Errorf("Erro no SubmitPrompt: %s", err)
		return false, erros.CreateError("Erro ao verificar a natureza do documento!")
	}

	resp := strings.TrimSpace(strings.ToLower(retSubmit.OutputText()))
	logger.Log.Infof("Resposta do modelo: %s", resp)

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

/*
Converte o arquivo PDF baixado do PJe, com todos os documentos dos autos,
para o formato txt, criando um novo arquivo com o mesmo nome, e extensão
.txt
*/
func processPDFToText(pdfPath string) (string, error) {
	txtFile := strings.TrimSuffix(pdfPath, ".pdf") + ".txt"

	cmd := exec.Command("pdftotext", "-layout", pdfPath, txtFile)
	err := cmd.Run()
	if err != nil {
		logger.Log.Errorf("Erro executando pdftotext: %v\n", err)
		return "", err
	}

	logger.Log.Infof("Texto extraído salvo como: %s\n", txtFile)
	return txtFile, nil
}

/*
Rotina utilizada para extrair os documentos contidos no arquivo texto gerado da conversão do arquivo PDF baixado
do PJe e salvá-los individualmente nos registros da tabela "docsocr". O nome do documento(nm_file_new) correspon-
de ao número do ID do documento.
*/
func extrairDocumentosAutos(IdContexto int, NmFileOri string, txtPath string) (string, error) {

	file, err := os.Open(txtPath)
	if err != nil {
		fmt.Println("Erro ao abrir arquivo:", err)
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var lastDocNumber string
	var pageLinesBuffer []string
	docsPages := make(map[string][]string)
	var firstMarkerFound bool = false

	for scanner.Scan() {
		linhaOriginal := scanner.Text()
		line := normalizaURLRodape(linhaOriginal)

		// Sempre acumula a linha atual
		pageLinesBuffer = append(pageLinesBuffer, line)

		// Verifica se a linha tem o padrão de quebra de página
		numeroDocumento := getDocumentoID(line)
		if numeroDocumento != "" {
			// Se for a primeira vez, só marca o lastDocNumber
			if !firstMarkerFound {
				firstMarkerFound = true
				lastDocNumber = numeroDocumento
			} else if numeroDocumento != lastDocNumber {

				docText, err := removeRodape(docsPages[lastDocNumber])
				if err != nil {
					fmt.Println("Erro ao realizar a limpeza do documento:", err)
					return "", err
				}
				// Salvamos o documento anterior na tabela docsocr antes de avançar
				nmFile := ultimosNDigitos(lastDocNumber, 9)
				//Salvamos o texto do documento na tabela "docsocr"
				err = SalvaTextoExtraido(IdContexto, 0, nmFile, docText)
				if err != nil {
					logger.Log.Errorf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", nmFile, IdContexto)
					continue
				}

				// Limpa o buffer do documento anterior
				docsPages[lastDocNumber] = nil
				// O conteúdo da página atual vai para o novo documento
				lastDocNumber = numeroDocumento
			}
			// Acumula as linhas da página no documento correto
			docsPages[lastDocNumber] = append(docsPages[lastDocNumber], pageLinesBuffer...)
			pageLinesBuffer = nil
		}
	}

	// Salva o último documento
	if firstMarkerFound {
		docsPages[lastDocNumber] = append(docsPages[lastDocNumber], pageLinesBuffer...)

		docText, err := removeRodape(docsPages[lastDocNumber])
		if err != nil {
			fmt.Println("Erro ao realizar sanitize no documento:", err)
			return "", err
		}
		// Salvamos o documento anterior na tabela docsocr antes de avançar
		nmFile := ultimosNDigitos(lastDocNumber, 9)

		err = SalvaTextoExtraido(IdContexto, 0, nmFile, docText)
		if err != nil {
			logger.Log.Errorf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", nmFile, IdContexto)

		}
	}

	if err := scanner.Err(); err != nil {
		logger.Log.Errorf("Erro na leitura do arquivo: %v", err)
	}
	return "", nil
}

/*
Função utilitária que extrai o ID do documento para ser utilizado como o nome para fins
de registro na tabela "docsocr"
*/
func getDocumentoID(texto string) string {
	// Extrai somente os dígitos do número no formato "Num. 110935393 - Pág. 7"
	re := regexp.MustCompile(`Num\.\s*(\d+)\s*-\s*Pág\.`)
	match := re.FindStringSubmatch(texto)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

/*
Função utilitária que complementa a extração do ID do documento.
*/
func ultimosNDigitos(s string, n int) string {
	if len(s) > n {
		return s[len(s)-n:]
	}
	return s
}

/*
Rotina que extrai o rodapé das páginas dos documentos, criado pelo PJe
*/
func removeRodape(lines []string) (string, error) {

	// Junta todas as linhas em um texto único
	textoCompleto := strings.Join(lines, "\n")

	// Regex do rodapé (mesma da função extrairMetadadosRodape) - não pode dar enter e quebrar essa linha u o regex falha

	padrao := `(?s)Este documento foi gerado pelo usuário\s+[\d*.\-]+ em \d{2}/\d{2}/\d{4} \d{2}:\d{2}:\d{2}\nNúmero do documento:\s*\d+\nhttps?://[^\n]+\nAssinado eletronicamente por:[^\n]+ - \d{2}/\d{2}/\d{4} \d{2}:\d{2}:\d{2}`
	reRodape := regexp.MustCompile(padrao)

	// Remove o rodapé do texto completo (se existir)
	textoSemRodape := reRodape.ReplaceAllString(textoCompleto, "")

	// Remove espaços em branco no início/fim após remoção
	textoSemRodape = strings.TrimSpace(textoSemRodape)

	//fmt.Printf("Salvo documento %s com %d linhas (rodapé removido)\n", filename, len(strings.Split(textoSemRodape, "\n")))
	return textoSemRodape, nil
}

/*
Rotina que faz o tratamento da URL que vem no rodapé das páginas de cada documento,
inserido automaticamente pelo PJe.
*/
func normalizaURLRodape(linha string) string {
	// Normalizações diversas para o OCR (pode manter se quiser)
	rePontos := regexp.MustCompile(`(\w)\s+(\.)\s*(\w)`)
	linha = rePontos.ReplaceAllString(linha, `$1.$3`)

	rePje1 := regexp.MustCompile(`pje\s+1`)
	linha = rePje1.ReplaceAllString(linha, `pje1`)

	rePje1Grau := regexp.MustCompile(`pje1\s+grau`)
	linha = rePje1Grau.ReplaceAllString(linha, `pje1grau`)

	reEspacosEspeciais := regexp.MustCompile(`\s*([:/?=])\s*`)
	linha = reEspacosEspeciais.ReplaceAllString(linha, `$1`)

	reMultEspaco := regexp.MustCompile(`\s+`)
	linha = reMultEspaco.ReplaceAllString(linha, ` `)

	reParametro := regexp.MustCompile(`\s*\?x=`)
	linha = reParametro.ReplaceAllString(linha, `?x=`)

	linha = strings.TrimSpace(linha)

	return linha
}
