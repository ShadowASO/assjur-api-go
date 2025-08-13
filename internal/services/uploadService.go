/*
---------------------------------------------------------------------------------------
File: userService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
---------------------------------------------------------------------------------------
*/
package services

import (
	"bufio"
	"context"

	"fmt"

	"ocrserver/internal/consts"
	"ocrserver/internal/models"
	"ocrserver/internal/opensearch"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"ocrserver/internal/utils/files"
	"ocrserver/internal/utils/logger"
	"sync"
)

type UploadServiceType struct {
	Model *models.UploadModelType
}

var UploadServiceGlobal *UploadServiceType
var onceInitUploadService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitUploadService(model *models.UploadModelType) {
	onceInitUploadService.Do(func() {

		UploadServiceGlobal = &UploadServiceType{
			Model: model,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewUploadService(model *models.UploadModelType,
) *UploadServiceType {
	return &UploadServiceType{

		Model: model,
	}
}

type DocumentoIndice struct {
	Id        string
	Data      string
	Hora      string
	Documento string
	Tipo      string
}

type NaturezaDoc struct {
	Key         int    `json:"key"`
	Description string `json:"description"`
}

var naturezasValidasImportarPJE = []int{
	consts.NATU_DOC_INICIAL,
	consts.NATU_DOC_CONTESTACAO,
	consts.NATU_DOC_REPLICA,
	consts.NATU_DOC_DESPACHO,
	consts.NATU_DOC_PETICAO,
	consts.NATU_DOC_DECISAO,
	consts.NATU_DOC_SENTENCA,
	consts.NATU_DOC_APELACAO,
	consts.NATU_DOC_CONTRA_RAZOES,
	consts.NATU_DOC_TERMO_AUDIENCIA,
	// Acrescente outras constantes que desejar incluir aqui
}

type BodyParamsPDF struct {
	IdContexto int
	IdFile     int
}

const maxTextSize = 60 * 1024 // 60 KB em bytes

/*
Função genérica destinada a processar a extração dos documentos contidos nos autos de cada
processo, e pode extrarir diretamente do arquivo PDF gerado pelo PJe, ou incorporá arquivos
txt já extraídos externamente. Não estamos utilizadon mais OCR, apesar das rotinas ainda
estarem disponíveis. Utilizaremos o utilitário linux "pdftotext" para converte o PDF p/TXT.
A rotina trabalha tanto com o PDF completo dos autos quanto de pelas individuais.
*/
func (obj *UploadServiceType) ProcessaPDF(ctx context.Context, bodyParams []BodyParamsPDF) (extractedFiles []string, extractedErros []int) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return
	}

	for _, doc := range bodyParams {
		autuar := true
		idCtxt := doc.IdContexto
		idFile := doc.IdFile

		row, err := obj.Model.SelectRowById(idFile)
		if err != nil {
			logger.Log.Errorf("Arquivo não encontrado em temp_uploads - id_file=%d - contexto=%d", idFile, idCtxt)
			extractedErros = append(extractedErros, idFile)
			continue
		}

		filePath := filepath.Join("uploads", row.NmFileNew)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			logger.Log.Errorf("Arquivo não encontrado - fileName=%s - contexto=%d", row.NmFileNew, idCtxt)
			extractedErros = append(extractedErros, idFile)
			continue
		}

		var resultText string
		ext := strings.ToLower(filepath.Ext(row.NmFileNew))

		//******   TEXTO **************************
		if ext == ".txt" {
			bytesContent, err := os.ReadFile(filePath)
			if err != nil {
				logger.Log.Errorf("Erro ao ler arquivo txt - fileName=%s - contexto=%d", row.NmFileNew, idCtxt)
				extractedErros = append(extractedErros, idFile)
				continue
			}
			resultText = string(bytesContent)
			natuDoc, err := AutosTempServiceGlobal.VerificarNaturezaDocumento(ctx, idCtxt, resultText)
			if err != nil {
				autuar = false
			} else {
				logger.Log.Infof("natuDoc=%d - %s", natuDoc.Key, natuDoc.Description)
			}
			// if autuar {
			// 	err = SalvaTextoExtraido(reg.IdContexto, 0, row.NmFileNew, resultText)
			// 	if err != nil {
			// 		logger.Log.Errorf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
			// 		extractedErros = append(extractedErros, reg.IdFile)
			// 		continue
			// 	}
			// }

		} else {
			//****************************************************
			//TRATAMENTO DO ARQUIVO PDF
			//****************************************************
			autuar = false
			//Usando OCR - desativado
			//resultText, err = processPDFWithPipeline(filePath)

			//Convertendo PDF para TXT com o aplicativo "pdftotext"
			txtPath, err := obj.convertePDFParaTexto(filePath)
			if err != nil {
				logger.Log.Errorf("Erro na extração do texto - fileName=%s - contexto=%d", row.NmFileNew, idCtxt)
				extractedErros = append(extractedErros, idFile)
				continue
			}

			//Fazendo a extração dos documentos contidos no arquivo texto
			_, err = obj.extrairDocumentosProcessuais(idCtxt, row.NmFileOri, txtPath)
			if err != nil {
				logger.Log.Errorf("Erro na extração do texto - fileName=%s - contexto=%d", row.NmFileNew, doc.IdContexto)
				extractedErros = append(extractedErros, idFile)
				continue
			}
			//DELETA o arquivo .TXT
			if err := obj.deletarArquivo(txtPath); err != nil {
				logger.Log.Errorf("Erro ao deletar o arquivo físico - %s", txtPath)
				extractedErros = append(extractedErros, idFile)
				continue
			}

		}

		if autuar {
			err = obj.SalvaTextoExtraido(idCtxt, 0, row.NmFileNew, resultText)
			if err != nil {
				logger.Log.Errorf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", row.NmFileNew, idCtxt)
				extractedErros = append(extractedErros, idFile)
				continue
			}
		}
		//DELETA o registro em "uploads"
		if err := obj.DeleteRegistro(doc.IdFile); err != nil {
			logger.Log.Errorf("Erro ao deletar o registro no banco - id_file=%d", idFile)
			extractedErros = append(extractedErros, idFile)
			continue
		}
		//DELETA o arquivo .PDF
		if err := obj.deletarArquivo(filePath); err != nil {
			logger.Log.Errorf("Erro ao deletar o arquivo físico - %s", filePath)
			extractedErros = append(extractedErros, idFile)
			continue
		}

		extractedFiles = append(extractedFiles, row.NmFileNew)
	}

	return extractedFiles, extractedErros
}

/*
Converte o arquivo PDF baixado do PJe, com todos os documentos dos autos,
para o formato txt, criando um novo arquivo com o mesmo nome, e extensão
.txt
*/
func (obj *UploadServiceType) convertePDFParaTexto(pdfPath string) (string, error) {
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

func (obj *UploadServiceType) extrairDocumentosProcessuais(
	IdContexto int,
	NmFileOri string,
	txtPath string,
) (string, error) {

	// 1) Extrai o índice para mapear ID → {Documento, Tipo, Data, Hora}
	indice, err := obj.extrairIndice(txtPath)
	if err != nil {
		return "", fmt.Errorf("erro ao extrair índice: %w", err)
	}
	logger.Log.Infof("[CTX=%d] Iniciando extração do TXT='%s' (itens no índice=%d)", IdContexto, txtPath, len(indice))

	// 2) Abre o TXT para varrer páginas/linhas
	file, err := os.Open(txtPath)
	if err != nil {
		logger.Log.Errorf("[CTX=%d] Erro ao abrir TXT: %v", IdContexto, err)
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Aumenta o buffer para evitar truncamentos em linhas longas
	const maxTokenSize = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxTokenSize)

	var (
		lastDocNumber    string
		pageLinesBuffer  []string
		docsPages        = make(map[string][]string) // acumula “pedaços de página” por documento
		firstMarkerFound bool
		lineNo           int
		totalSalvos      int
		totalIgnorados   int
		totalFechados    int
	)

	// Helper: tenta salvar/descartar o documento anterior com logs detalhados
	saveOrSkip := func(docNumber string) {
		if docNumber == "" {
			return
		}
		totalFechados++
		docLines := docsPages[docNumber]
		logger.Log.Debugf("[CTX=%d] Fechando doc Num=%s (linhas acumuladas=%d)", IdContexto, docNumber, len(docLines))

		docText, err := obj.removeRodape(docLines)
		if err != nil {
			logger.Log.Errorf("[CTX=%d] Erro limpando rodapé do Num=%s: %v", IdContexto, docNumber, err)
			return
		}

		nmFile := obj.ultimosNDigitos(docNumber, 9)
		docInfo, existe := indice[nmFile]

		switch {
		case !existe:
			totalIgnorados++
			logger.Log.Infof("[CTX=%d] IGNORADO Num=%s (nmFile=%s) — não consta no índice", IdContexto, docNumber, nmFile)

		case !obj.isDocumentoTipoValido(docInfo.Tipo):
			totalIgnorados++
			logger.Log.Infof("[CTX=%d] IGNORADO Num=%s (tipo=%s) — tipo não importável", IdContexto, docNumber, docInfo.Tipo)

		case !obj.isDocumentoSizeValido(docText, maxTextSize):
			totalIgnorados++
			logger.Log.Infof("[CTX=%d] IGNORADO Num=%s (tipo=%s) — tamanho excede limite (%d bytes)",
				IdContexto, docNumber, docInfo.Tipo, len([]byte(docText)))

		default:
			idNatu := consts.GetCodigoNatureza(docInfo.Tipo)
			if err := obj.SalvaTextoExtraido(IdContexto, idNatu, nmFile, docText); err != nil {
				logger.Log.Errorf("[CTX=%d] ERRO ao salvar Num=%s (nmFile=%s, tipo=%s): %v",
					IdContexto, docNumber, nmFile, docInfo.Tipo, err)
			} else {
				totalSalvos++
				logger.Log.Infof("[CTX=%d] SALVO Num=%s (nmFile=%s, tipo=%s, bytes=%d)",
					IdContexto, docNumber, nmFile, docInfo.Tipo, len([]byte(docText)))
			}
		}

		// limpa o acumulador do doc anterior para liberar memória
		docsPages[docNumber] = nil
	}

	// 3) Varre o arquivo linha a linha
	for scanner.Scan() {
		lineNo++
		linhaOriginal := scanner.Text()
		linha := obj.normalizaURLRodape(linhaOriginal) // já remove \f e normaliza espaços

		// Sempre acumula a linha atual como parte do “bloco” corrente
		pageLinesBuffer = append(pageLinesBuffer, linha)

		// Tenta detectar o marcador de página/ID: "Num. <digits> - Pág."
		numeroDocumento := obj.getDocumentoID(linha)
		if numeroDocumento != "" {
			logger.Log.Debugf("[CTX=%d][L%d] Encontrado marcador: Num=%s", IdContexto, lineNo, numeroDocumento)

			if !firstMarkerFound {
				firstMarkerFound = true
				lastDocNumber = numeroDocumento
				logger.Log.Debugf("[CTX=%d] Primeiro marcador definido: lastDoc=%s", IdContexto, lastDocNumber)
			} else if numeroDocumento != lastDocNumber {
				// Fechamos o documento anterior e iniciamos um novo
				logger.Log.Debugf("[CTX=%d] Troca de doc: %s → %s", IdContexto, lastDocNumber, numeroDocumento)
				saveOrSkip(lastDocNumber)
				lastDocNumber = numeroDocumento
			}

			// Move o bloco acumulado para o doc atual e zera o buffer
			docsPages[lastDocNumber] = append(docsPages[lastDocNumber], pageLinesBuffer...)
			logger.Log.Debugf("[CTX=%d] Acumulado em Num=%s (chunk linhas=%d, total=%d)",
				IdContexto, lastDocNumber, len(pageLinesBuffer), len(docsPages[lastDocNumber]))
			pageLinesBuffer = nil
		}
	}

	// 4) Fecha o último documento (se houver)
	if firstMarkerFound {
		// Acrescenta o que sobrou do buffer ao último doc
		if len(pageLinesBuffer) > 0 {
			docsPages[lastDocNumber] = append(docsPages[lastDocNumber], pageLinesBuffer...)
			logger.Log.Debugf("[CTX=%d] EOF: anexado restante ao Num=%s (restante linhas=%d, total=%d)",
				IdContexto, lastDocNumber, len(pageLinesBuffer), len(docsPages[lastDocNumber]))
		}
		saveOrSkip(lastDocNumber)
	} else {
		logger.Log.Warningf("[CTX=%d] Nenhum marcador 'Num. <id> - Pág.' encontrado no arquivo — nada a salvar.", IdContexto)
	}

	if err := scanner.Err(); err != nil {
		logger.Log.Errorf("[CTX=%d] Erro na leitura do arquivo: %v", IdContexto, err)
	}

	logger.Log.Infof("[CTX=%d] Finalizado TXT='%s' — fechados=%d, salvos=%d, ignorados=%d",
		IdContexto, txtPath, totalFechados, totalSalvos, totalIgnorados)

	return "", nil
}

func (obj *UploadServiceType) deletarArquivo(filePath string) error {
	if files.FileExist(filePath) {
		err := files.DeletarFile(filePath)
		if err != nil {
			logger.Log.Errorf("Erro ao deletar o arquivo físico - %s: %v", filePath, err)
			return err
		}
	}
	return nil
}

func (obj *UploadServiceType) SalvaTextoExtraido(idCtxt int, idNatu int, idPje string, texto string) error {

	autos_temp := opensearch.NewAutos_tempIndex()

	_, err := autos_temp.Indexa(idCtxt, idNatu, idPje, texto, "")
	if err != nil {
		logger.Log.Errorf("Erro ao inserir linha: %v", err)
		return err
	}
	logger.Log.Infof("Doc %s - idNatu=%d", idPje, idNatu)
	return nil

}

func (obj *UploadServiceType) InserirRegistro(IdCtxt int, newFile string, oriFile string) (int64, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return 0, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	SnAutos := "N"
	Status := "S"
	DtInc := time.Now()

	row, err := obj.Model.InsertRow(IdCtxt, newFile, oriFile, SnAutos, DtInc, Status)
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return 0, err
	}
	return row, nil
}

func (obj *UploadServiceType) DeleteRegistro(idFile int) error {
	err := obj.Model.DeleteRow(idFile)
	if err != nil {
		logger.Log.Errorf("Erro ao deletar o registro no banco - id_file=%d: %v", idFile, err)
	}
	return err
}

// extrairIndice extrai o índice do arquivo texto, devolvendo um mapa id → DocumentoIndice
func (obj *UploadServiceType) extrairIndice(txtPath string) (map[string]*DocumentoIndice, error) {
	file, err := os.Open(txtPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Aumenta limite de token para linhas atípicas (opcional, mas seguro)
	const maxTokenSize = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxTokenSize)

	// Aceita controles/brancos no início da linha antes do ID
	reLinhaIndice := regexp.MustCompile(`^[\f\t\r ]*(\d+)\s+(\d{2}/\d{2}/\d{4})\s+(.*)$`)
	reHora := regexp.MustCompile(`\b(\d{2}:\d{2})\b`)

	indice := make(map[string]*DocumentoIndice)
	var linhaAnteriorIndice *DocumentoIndice

	for scanner.Scan() {
		linha := scanner.Text()
		// Sanitiza: remove form-feed e outros controles não impressos, preservando \n (já removido pelo Scanner)
		linha = strings.Map(func(r rune) rune {
			// Remove form-feed e demais controles (exceto TAB, que pode existir entre colunas)
			if r == '\f' || (r < 32 && r != '\t') {
				return -1
			}
			return r
		}, linha)
		linha = strings.TrimRight(linha, " \r")

		if reLinhaIndice.MatchString(linha) {
			matches := reLinhaIndice.FindStringSubmatch(linha)
			id := matches[1]
			data := matches[2]
			resto := matches[3]

			// Divide por 2+ espaços (colunas); o último item tende a ser o "Tipo"
			partes := regexp.MustCompile(`\s{2,}`).Split(resto, -1)
			documento := ""
			tipo := ""

			logger.Log.Debugf("linha índice: %s", linha)

			if len(partes) == 1 {
				documento = strings.TrimSpace(partes[0])
			} else if len(partes) >= 2 {
				tipo = strings.TrimSpace(partes[len(partes)-1])
				documento = strings.TrimSpace(strings.Join(partes[:len(partes)-1], " "))
			}

			doc := &DocumentoIndice{
				Id:        id,
				Data:      data,
				Documento: documento,
				Tipo:      tipo,
			}

			indice[id] = doc
			linhaAnteriorIndice = doc

		} else if linhaAnteriorIndice != nil {
			// A linha da hora costuma vir sozinha na linha seguinte
			if horaMatch := reHora.FindStringSubmatch(linha); len(horaMatch) == 2 {
				linhaAnteriorIndice.Hora = horaMatch[1]
				linhaAnteriorIndice = nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return indice, nil
}

/*
Rotina que faz o tratamento da URL que vem no rodapé das páginas de cada documento,
inserido automaticamente pelo PJe.
*/
func (obj *UploadServiceType) normalizaURLRodape(linha string) string {
	// Remove form-feed e controles não necessários
	linha = strings.Map(func(r rune) rune {
		if r == '\f' || (r < 32 && r != '\t') {
			return -1
		}
		return r
	}, linha)

	// Normalizações diversas...
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

	return strings.TrimSpace(linha)
}

/*
Função utilitária que complementa a extração do ID do documento.
*/
func (obj *UploadServiceType) ultimosNDigitos(s string, n int) string {
	if len(s) > n {
		return s[len(s)-n:]
	}
	return s
}

// Função que verifica se o tipo de documento deve importado e salvo
func (obj *UploadServiceType) isDocumentoTipoValido(tipo string) bool {

	logger.Log.Infof("Tipo: %s", tipo)
	natu := consts.GetCodigoNatureza(tipo)

	for _, v := range naturezasValidasImportarPJE {
		if v == natu {
			return true
		}
	}

	return false

}

func (obj *UploadServiceType) isDocumentoSizeValido(texto string, limiteBytes int) bool {
	tamanho := len([]byte(texto))
	if tamanho > limiteBytes {
		logger.Log.Infof("Documento com tamanho %d excede %d bytes", tamanho, limiteBytes)
		return false
	}
	return true
}

/*
Função utilitária que extrai o ID do documento para ser utilizado como o nome para fins
de registro na tabela "docsocr"
*/
func (obj *UploadServiceType) getDocumentoID(texto string) string {
	// "Num. 110935393 - Pág." ou "Num.110935393- Pág." etc.
	re := regexp.MustCompile(`Num\.\s*(\d+)\s*-\s*Pág\.`)
	if m := re.FindStringSubmatch(texto); len(m) == 2 {
		return m[1]
	}
	return ""
}

/*
Rotina que extrai o rodapé das páginas dos documentos, criado pelo PJe
*/
func (obj *UploadServiceType) removeRodape(lines []string) (string, error) {

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
func (obj *UploadServiceType) SelectById(id int) (*models.UploadRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	row, err := obj.Model.SelectRowById(id)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return nil, fmt.Errorf("CnjApi global não configurada")
	}
	return row, nil
}
func (obj *UploadServiceType) SelectByContexto(idCtxt int) ([]models.UploadRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	rows, err := obj.Model.SelectRowsByContextoId(idCtxt)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return nil, fmt.Errorf("CnjApi global não configurada")
	}
	return rows, nil
}
