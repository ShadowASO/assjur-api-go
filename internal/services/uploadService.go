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

	for _, reg := range bodyParams {
		autuar := true
		//row, err := uploadModel.SelectRowById(reg.IdFile)
		row, err := obj.Model.SelectRowById(reg.IdFile)
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

		//******   TEXTO **************************
		if ext == ".txt" {
			bytesContent, err := os.ReadFile(filePath)
			if err != nil {
				logger.Log.Errorf("Erro ao ler arquivo txt - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}
			resultText = string(bytesContent)
			natuDoc, err := Autos_tempServiceGlobal.VerificarNaturezaDocumento(ctx, resultText)
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
				logger.Log.Errorf("Erro na extração do texto - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}

			//Fazendo a extração dos documentos contidos no arquivo texto
			_, err = obj.extrairDocumentosProcessuais(reg.IdContexto, row.NmFileOri, txtPath)
			if err != nil {
				logger.Log.Errorf("Erro na extração do texto - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}

			if err := obj.deletarArquivo(txtPath); err != nil {
				logger.Log.Errorf("Erro ao deletar o arquivo físico - %s", txtPath)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}

		}

		if autuar {
			err = obj.SalvaTextoExtraido(reg.IdContexto, 0, row.NmFileNew, resultText)
			if err != nil {
				logger.Log.Errorf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", row.NmFileNew, reg.IdContexto)
				extractedErros = append(extractedErros, reg.IdFile)
				continue
			}
		}

		if err := obj.DeleteRegistro(reg.IdFile); err != nil {
			logger.Log.Errorf("Erro ao deletar o registro no banco - id_file=%d", reg.IdFile)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		if err := obj.deletarArquivo(filePath); err != nil {
			logger.Log.Errorf("Erro ao deletar o arquivo físico - %s", filePath)
			extractedErros = append(extractedErros, reg.IdFile)
			continue
		}

		extractedFiles = append(extractedFiles, row.NmFileNew)
	}

	return extractedFiles, extractedErros
}

// func (obj *UploadServiceType) VerificarNaturezaDocumento(ctx context.Context, texto string) (*NaturezaDoc, error) {

// 	var msgs MsgGpt
// 	assistente := `O seguinte texto pertence aos autos de um processo judicial.

// Primeiramente, verifique se o texto é uma movimentação, registro ou anotação processual, contendo expressões como:
// "Mov.", "Movimentação", "Observações dos Movimentos", "Registro", "Publicação", "Entrada", "Intimação", "Anotação".
// Se essas expressões estiverem presentes, e o texto não contiver o corpo formal completo da decisão (com fundamentação e conclusão explícita do juiz),
// classifique o documento como:
// - { "key": 1003, "description": "movimentação/processo" }.

// Em seguida, verifique se o texto contém alguma das expressões indicativas de certidões ou outros documentos, tais como:
// "certidão", "certifico que", "Por ordem do MM. Juiz", "teor do ato", "o referido é verdade, dou fé",
// "encaminhado edital/relação para publicação", "ato ordinatório".

// Se qualquer dessas expressões estiver presente em qualquer parte do texto, incluindo cabeçalhos, movimentações ou descrições, classifique o documento imediatamente como:
// - { "key": 1002, "description": "certidões" } se for claramente certidão,
// - caso contrário, classifique como { "key": 1001, "description": "outros documentos" }.

// Somente se nenhuma dessas expressões estiver presente, analise o conteúdo para identificar a natureza do documento conforme as opções a seguir:

// { "key": 1, "description": "Petição inicial" }
// { "key": 2, "description": "Contestação" }
// { "key": 3, "description": "Réplica" }
// { "key": 4, "description": "Despacho inicial" }
// { "key": 5, "description": "Despacho" }
// { "key": 6, "description": "Petição" }
// { "key": 7, "description": "Decisão" }
// { "key": 8, "description": "Sentença" }
// { "key": 9, "description": "Embargos de declaração" }
// { "key": 10, "description": "Contra-razões" }
// { "key": 11, "description": "Recurso" }
// { "key": 12, "description": "Procuração" }
// { "key": 13, "description": "Rol de Testemunhas" }
// { "key": 14, "description": "Contrato" }
// { "key": 15, "description": "Laudo Pericial" }
// { "key": 16, "description": "Ata de audiência" }
// { "key": 17, "description": "Parecer do Ministério Público" }

// Se não puder identificar claramente a natureza do texto, classifique como { "key": 1001, "description": "outros documentos" }.

// Responda apenas com um JSON no formato: {"key": int, "description": string }.`

// 	msgs.CreateMessage("", ROLE_USER, assistente)
// 	msgs.CreateMessage("", ROLE_USER, texto)

// 	//retSubmit, err := services.OpenaiServiceGlobal.SubmitPromptResponse(ctx, msgs, nil, "gpt-4.1-nano")
// 	retSubmit, err := OpenaiServiceGlobal.SubmitPromptResponse(ctx, msgs, nil, "gpt-4.1-mini")
// 	if err != nil {
// 		logger.Log.Errorf("Erro no SubmitPrompt: %s", err)
// 		return nil, erros.CreateError("Erro ao verificar a  natureza do  documento!")
// 	}

// 	resp := strings.TrimSpace(retSubmit.OutputText())
// 	//logger.Log.Infof("Resposta do modelo: %s", resp)

// 	var natureza NaturezaDoc
// 	err = json.Unmarshal([]byte(resp), &natureza)
// 	if err != nil {
// 		logger.Log.Warningf("Erro ao parsear JSON da resposta: %v", err)
// 		logger.Log.Warningf("Resposta recebida: %s", resp)
// 		return nil, erros.CreateError("Resposta inesperada ou formato inválido do modelo")
// 	}

// 	logger.Log.Infof("Natureza documento identificada: key=%d, description=%s", natureza.Key, natureza.Description)

// 	return &natureza, nil
// }

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

func (obj *UploadServiceType) extrairDocumentosProcessuais(IdContexto int, NmFileOri string, txtPath string) (string, error) {

	// Extrai índice do arquivo baixado do PJE. Ele será usado para identificar o tipo/natureza
	// de cada documento.
	indice, err := obj.extrairIndice(txtPath)
	if err != nil {
		return "", fmt.Errorf("erro ao extrair índice: %w", err)
	}

	//Abre o arquivo TXT obtido da conversão do arquivo PDF do PJE

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
		line := obj.normalizaURLRodape(linhaOriginal)

		// Sempre acumula a linha atual
		pageLinesBuffer = append(pageLinesBuffer, line)

		// Verifica se a linha tem o padrão de quebra de página
		numeroDocumento := obj.getDocumentoID(line)
		if numeroDocumento != "" {
			// Se for a primeira vez, só marca o lastDocNumber
			if !firstMarkerFound {
				firstMarkerFound = true
				lastDocNumber = numeroDocumento
			} else if numeroDocumento != lastDocNumber {

				docText, err := obj.removeRodape(docsPages[lastDocNumber])
				if err != nil {
					fmt.Println("Erro ao realizar a limpeza do documento:", err)
					return "", err
				}
				// Salvamos o documento anterior na tabela docsocr antes de avançar
				nmFile := obj.ultimosNDigitos(lastDocNumber, 9)
				docInfo, existe := indice[nmFile]
				if !existe {
					logger.Log.Infof("Documento %s - Não foi salvo, pois não está no índice", nmFile)
				} else if !obj.isDocumentoTipoValido(docInfo.Tipo) {
					logger.Log.Infof("Documento %s - tipo %s - Não foi salvo", nmFile, docInfo.Tipo)
				} else if !obj.isDocumentoSizeValido(docText, maxTextSize) {
					logger.Log.Infof("Documento %s - tipo %s - Não foi salvo", nmFile, docInfo.Tipo)
				} else {
					logger.Log.Infof("Documento %s - tipo %s - SALVO", nmFile, docInfo.Tipo)
					idNatu := consts.GetTipoDocumento(docInfo.Tipo)
					err = obj.SalvaTextoExtraido(IdContexto, idNatu, nmFile, docText)
					if err != nil {
						logger.Log.Errorf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", nmFile, IdContexto)
						continue
					}

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

		docText, err := obj.removeRodape(docsPages[lastDocNumber])
		if err != nil {
			fmt.Println("Erro ao realizar sanitize no documento:", err)
			return "", err
		}
		// Salvamos o documento anterior na tabela docsocr antes de avançar
		nmFile := obj.ultimosNDigitos(lastDocNumber, 9)
		docInfo, existe := indice[nmFile]
		if !existe {
			logger.Log.Infof("Documento %s - Não foi salvo, pois não está no índice", nmFile)
		} else if !obj.isDocumentoTipoValido(docInfo.Tipo) {
			logger.Log.Infof("Documento %s - tipo %s - Não foi salvo", nmFile, docInfo.Tipo)
		} else if !obj.isDocumentoSizeValido(docText, maxTextSize) {
			logger.Log.Infof("Documento %s - tipo %s - Não foi salvo", nmFile, docInfo.Tipo)
		} else {
			idNatu := consts.GetTipoDocumento(docInfo.Tipo)
			err = obj.SalvaTextoExtraido(IdContexto, idNatu, nmFile, docText)
			if err != nil {
				logger.Log.Errorf("Erro ao salvar o texto extraído - fileName=%s - contexto=%d", nmFile, IdContexto)

			}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Log.Errorf("Erro na leitura do arquivo: %v", err)
	}
	return "", nil
}

func (obj *UploadServiceType) deletarArquivo(filePath string) error {
	if files.FileExist(filePath) {
		//err := uploadController.DeletarFile(filePath)
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

	// Usa o modelo para inserir o registro

	//_, err := service.Service.InsertRow(reg)

	//logger.Log.Info("Registro inserido com sucesso: " + fileName)

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
	reLinhaIndice := regexp.MustCompile(`^(\d+)\s+(\d{2}/\d{2}/\d{4})\s+(.*)$`)
	reHora := regexp.MustCompile(`\b(\d{2}:\d{2})\b`)

	indice := make(map[string]*DocumentoIndice)
	var linhaAnteriorIndice *DocumentoIndice

	for scanner.Scan() {
		linha := scanner.Text()
		linha = strings.TrimRight(linha, "\r\n")

		if reLinhaIndice.MatchString(linha) {
			matches := reLinhaIndice.FindStringSubmatch(linha)
			id := matches[1]
			data := matches[2]
			resto := matches[3]

			partes := regexp.MustCompile(`\s{2,}`).Split(resto, -1)
			documento := ""
			tipo := ""

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
			horaMatch := reHora.FindStringSubmatch(linha)
			if len(horaMatch) == 2 {
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

/*
Função utilitária que complementa a extração do ID do documento.
*/
func (obj *UploadServiceType) ultimosNDigitos(s string, n int) string {
	if len(s) > n {
		return s[len(s)-n:]
	}
	return s
}
func (obj *UploadServiceType) isDocumentoSizeValido(
	texto string,
	limiteBytes int,
) bool {
	tamanho := len([]byte(texto))
	if tamanho > limiteBytes {
		logger.Log.Infof("Documento com tamanho %d excede %d bytes", limiteBytes, tamanho)
		return false
	}
	return true
}

// Função que verifica se o tipo de documento deve importado e salvo
func (obj *UploadServiceType) isDocumentoTipoValido(tipo string) bool {
	// tipo = strings.ToLower(tipo)

	// Salvar := consts.GetNaturezaDocumentosImportarPJE()
	// for _, ok := range Salvar {
	// 	if strings.Contains(tipo, ok) {
	// 		return true
	// 	}
	// }
	// return false
	return consts.GetTipoDocumento(tipo) != 0

}

/*
Função utilitária que extrai o ID do documento para ser utilizado como o nome para fins
de registro na tabela "docsocr"
*/
func (obj *UploadServiceType) getDocumentoID(texto string) string {
	// Extrai somente os dígitos do número no formato "Num. 110935393 - Pág. 7"
	re := regexp.MustCompile(`Num\.\s*(\d+)\s*-\s*Pág\.`)
	match := re.FindStringSubmatch(texto)
	if len(match) < 2 {
		return ""
	}
	return match[1]
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
