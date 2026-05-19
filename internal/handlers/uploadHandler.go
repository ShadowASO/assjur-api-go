/*
---------------------------------------------------------------------------------------
File: uploadHandler.go
Autor: Aldenor
Data: 17-05-2025
---------------------------------------------------------------------------------------
*/
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ocrserver/internal/database/pgdb"
	"ocrserver/internal/models"
	"ocrserver/internal/services"

	"ocrserver/internal/utils/mslogger"
	"ocrserver/internal/utils/msresponse"

	"github.com/gin-gonic/gin"
)

type UploadHandlerType struct {
	Service *services.UploadServiceType
}

// Tamanho máximo do arquivo aceito no upload.
// 10 << 23 equivale a 83.886.080 bytes, aproximadamente 80MB.
const MAX_SIZE_UPLOAD = 10 << 23

func NewUploadHandlers(service *services.UploadServiceType) *UploadHandlerType {
	return &UploadHandlerType{Service: service}
}

// Função para gerar um nome único para o arquivo.
func generateUniqueFileName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

/*
*
  - Faz o upload de um arquivo e cria um registro na tabela 'uploads'
  - Rota: "/contexto/documentos/upload"
  - Content-Type: multipart/form-data.
  - Body:
  - file: File
  - idContexto: string
  - filename_ori: string
  - Método: POST
  - Teste: curl -X POST http://localhost:4001/upload -F "file=@replica.pdf"
*/
func (service *UploadHandlerType) UploadFileHandler(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MAX_SIZE_UPLOAD)

	handler, err := c.FormFile("file")
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao obter arquivo: %v", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Erro ao obter arquivo",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	filenameOri := c.PostForm("filename_ori")
	idContexto := c.PostForm("idContexto")

	if idContexto == "" || filenameOri == "" {
		mslogger.LoggerGlobal.Error("Campos idContexto e filename_ori obrigatórios e válidos")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Campos obrigatórios ausentes",
			msresponse.ErrorValidacao,
			"Os campos idContexto e filename_ori são obrigatórios.",
		)
		return
	}

	uniqueFileName := generateUniqueFileName() + filepath.Ext(handler.Filename)
	savePath := filepath.Join("uploads", uniqueFileName)

	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao criar diretório uploads: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao criar diretório uploads",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	if err := c.SaveUploadedFile(handler, savePath); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao salvar arquivo: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao salvar arquivo",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	if err := service.InsertUploadedFile(idContexto, uniqueFileName, filenameOri); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao registrar arquivo no banco: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao registrar arquivo no banco",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"filename":     uniqueFileName,
		"filename_ori": filenameOri,
		"id_contexto":  idContexto,
	}

	msresponse.OK(c, http.StatusCreated, "Arquivo transferido com sucesso", rsp)
}

/*
 * Devolve os registros da tabela 'uploads' para um determinado contexto.
 *
 * - **Rota**: "/contexto/documentos/upload/:id"
 * - **Params**: ID do Contexto
 * - **Método**: GET
 */
func (service *UploadHandlerType) SelectHandler(c *gin.Context) {
	ctxtID := c.Param("id")
	if ctxtID == "" {
		mslogger.LoggerGlobal.Error("ID do contexto não informado")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"ID do contexto não informado",
			msresponse.ErrorValidacao,
			"O parâmetro id é obrigatório.",
		)
		return
	}

	rows, err := service.Service.SelectByContexto(ctxtID)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao selecionar arquivos transferidos por contexto:", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao selecionar arquivos transferidos por contexto",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"rows": rows,
	}

	msresponse.OK(c, http.StatusOK, "Registros selecionados com sucesso", rsp)
}

/*
 * Devolve todos os registros da tabela 'uploads'.
 *
 * - **Rota**: "/contexto/documentos/upload/"
 * - **Método**: GET
 */
func (service *UploadHandlerType) SelectAllHandler(c *gin.Context) {
	var dataRows []models.UploadRow

	uploadModel := models.NewUploadModel(pgdb.DBPoolGlobal.Pool)

	dataRows, err := uploadModel.SelectRows()
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao selecionar arquivos transferidos:", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao selecionar arquivos transferidos",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	rsp := gin.H{
		"rows": dataRows,
	}

	msresponse.OK(c, http.StatusOK, "Arquivos transferidos selecionados com sucesso", rsp)
}

/*
 * Deleta os registros da tabela 'uploads' e respectivos arquivos da pasta 'uploads'.
 *
 * - **Rota**: "/contexto/documentos/upload"
 * - **Método**: DELETE
 * - **Body:
 *		[
 *			{
 * 				"idContexto": number,
 *	  			"idFile": number
 *	  		}
 *		]
 */
type paramsBodyUploadDelete struct {
	IdContexto int `json:"idContexto"`
	IdFile     int `json:"idFile"`
}

func (service *UploadHandlerType) DeleteHandler(c *gin.Context) {
	var deleteFiles []paramsBodyUploadDelete

	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&deleteFiles); err != nil {
		mslogger.LoggerGlobal.ErrorErr("Dados inválidos:", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Dados inválidos",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	if len(deleteFiles) == 0 {
		mslogger.LoggerGlobal.Error("Arquivos não informados")

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Arquivos não informados",
			msresponse.ErrorValidacao,
			"A lista de arquivos para exclusão está vazia.",
		)
		return
	}

	var deletedFiles []int
	var failedFiles []int

	for _, reg := range deleteFiles {
		row, err := service.Service.SelectById(reg.IdFile)
		if err != nil {
			mslogger.LoggerGlobal.ErrorErr("Arquivo não encontrado:", err)
			failedFiles = append(failedFiles, reg.IdFile)
			continue
		}

		if err := service.Service.DeleteRegistro(reg.IdFile); err != nil {
			mslogger.LoggerGlobal.ErrorErr("Erro ao deletar registro:", err)
			failedFiles = append(failedFiles, reg.IdFile)
			continue
		}

		fullFileName := filepath.Join("uploads", row.NmFileNew)
		if service.FileExist(fullFileName) {
			if err := service.DeletarFile(fullFileName); err != nil {
				mslogger.LoggerGlobal.ErrorErr("Erro ao deletar arquivo físico:", err)
				failedFiles = append(failedFiles, reg.IdFile)
				continue
			}
		}

		deletedFiles = append(deletedFiles, reg.IdFile)
	}

	rsp := gin.H{
		"deleted": deletedFiles,
		"errors":  failedFiles,
	}

	message := "Processamento concluído"
	if len(failedFiles) > 0 {
		message = "Processamento concluído com falhas"
	}

	msresponse.OK(c, http.StatusOK, message, rsp)
}

func (service *UploadHandlerType) DeleteHandlerById(c *gin.Context) {
	idParam := c.Param("id")

	idFile, err := strconv.Atoi(idParam)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("IdDoc inválido", err)

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"IdDoc inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	row, err := service.Service.SelectById(idFile)
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("Registro não encontrado:", err)

		msresponse.Fail(
			c,
			http.StatusNotFound,
			"Registro não encontrado",
			msresponse.ErrorNaoEncontrado,
			err.Error(),
		)
		return
	}

	if err := service.Service.DeleteRegistro(idFile); err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao deletar registro:", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao deletar o registro",
			msresponse.ErrorInterno,
			err.Error(),
		)
		return
	}

	fullFileName := filepath.Join("uploads", row.NmFileNew)
	if service.FileExist(fullFileName) {
		if err := service.DeletarFile(fullFileName); err != nil {
			mslogger.LoggerGlobal.ErrorErr("Erro ao deletar arquivo físico: "+fullFileName, err)

			msresponse.Fail(
				c,
				http.StatusInternalServerError,
				"Erro ao deletar o arquivo",
				msresponse.ErrorInterno,
				err.Error(),
			)
			return
		}
	}

	msresponse.OK(c, http.StatusOK, "Documento deletado com sucesso")
}

/* Verifica apenas se o arquivo existe. */
func (service *UploadHandlerType) FileExist(fullFileName string) bool {
	_, err := os.Stat(fullFileName)
	return !os.IsNotExist(err)
}

// Deleta um arquivo.
func (service *UploadHandlerType) DeletarFile(fullFileName string) error {
	if err := os.Remove(fullFileName); err != nil {
		mslogger.LoggerGlobal.ErrorErr("Erro ao deletar arquivo:", err)
		return err
	}

	mslogger.LoggerGlobal.Info("Arquivo deletado com sucesso: " + fullFileName)
	return nil
}

/*
Insere um registro na tabela uploads para cada arquivo transferido para o servidor
por upload.
*/
func (service *UploadHandlerType) InsertUploadedFile(idCtxt string, fileName string, fileNameOri string) error {
	if idCtxt == "" {
		return fmt.Errorf("ID de contexto inválido: %s", idCtxt)
	}

	if fileName == "" {
		return fmt.Errorf("nome do arquivo não pode ser vazio")
	}

	if fileNameOri == "" {
		return fmt.Errorf("nome original do arquivo não pode ser vazio")
	}

	if _, err := service.Service.InserirRegistro(idCtxt, fileName, fileNameOri); err != nil {
		mslogger.LoggerGlobal.Error("Erro ao inserir Registro: " + fileName)
		return fmt.Errorf("falha ao inserir registro no banco de dados: %w", err)
	}

	mslogger.LoggerGlobal.Info("Registro inserido com sucesso: " + fileName)

	return nil
}
