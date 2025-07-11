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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ocrserver/internal/consts"
	"ocrserver/internal/models"
	"ocrserver/internal/opensearch"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"sync"
)

type AutosTempServiceType struct {
	idx *opensearch.AutosTempIndexType
}

var Autos_tempServiceGlobal *AutosTempServiceType
var onceInitAutos_tempService sync.Once

type DocumentoBase struct {
	Tipo *struct {
		Key         int    `json:"key"`
		Description string `json:"description"`
	} `json:"tipo"`
	Processo string `json:"processo"`
	IdPje    string `json:"id_pje"`
}

// type RegKeys struct {
// 	IdContexto int
// 	//IdDoc      int
// 	IdDoc string
// }

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitAutos_tempService(idx *opensearch.AutosTempIndexType) {
	onceInitAutos_tempService.Do(func() {
		Autos_tempServiceGlobal = &AutosTempServiceType{
			idx: idx,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewAutos_tempService(
	idx *opensearch.AutosTempIndexType,
) *AutosTempServiceType {
	return &AutosTempServiceType{
		idx: idx,
	}
}

func (obj *AutosTempServiceType) InserirAutos(IdCtxt int, IdNatu int, IdPje string, doc string) (*consts.ResponseAutosTempRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}
	row, err := obj.idx.Indexa(IdCtxt, IdNatu, IdPje, doc, "")
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return nil, err
	}
	return row, nil
}
func (obj *AutosTempServiceType) UpdateAutos(Id string, IdCtxt int, IdNatu int, IdPje string, doc string) (*consts.ResponseAutosTempRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	row, err := obj.idx.Update(Id, IdCtxt, IdNatu, IdPje, doc)
	if err != nil {
		logger.Log.Error("Erro na inclusão do registro", err.Error())
		return nil, err
	}
	return row, nil
}
func (obj *AutosTempServiceType) DeletaAutos(id string) error {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	err := obj.idx.Delete(id)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return fmt.Errorf("CnjApi global não configurada")
	}
	return nil
}
func (obj *AutosTempServiceType) SelectById(id string) (*consts.ResponseAutosTempRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	row, err := obj.idx.ConsultaById(id)
	if err != nil {
		logger.Log.Error("Erro ao consultar documento %v.", err.Error())
		return nil, fmt.Errorf("Erro ao consultar documento %v.", err.Error())
	}
	return row, nil
}
func (obj *AutosTempServiceType) SelectByContexto(idCtxt int) ([]consts.ResponseAutosTempRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("Tentativa de uso de serviço não iniciado.")
	}

	rows, err := obj.idx.ConsultaByIdCtxt(idCtxt)
	if err != nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return nil, fmt.Errorf("CnjApi global não configurada")
	}
	return rows, nil
}

func (obj *AutosTempServiceType) GetAutosByContexto(id int) ([]consts.ResponseAutosTempRow, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	rows, err := obj.SelectByContexto(id)
	if err != nil {
		logger.Log.Error("erro ao buscar sessão pelo ID")
		return nil, err
	}
	return rows, nil
}

func (obj *AutosTempServiceType) ProcessarDocumento(IdContexto int, IdDoc string) error {
	ctx := context.Background()
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	msg := fmt.Sprintf("Processando documento: IdDoc=%s, IdContexto=%d", IdDoc, IdContexto)
	logger.Log.Info(msg)

	//REcupero o registro da tabela temp_autos
	//row, err := obj.docsocrModel.SelectByIdDoc(reg.IdDoc)
	row, err := obj.idx.ConsultaById(IdDoc)
	if err != nil {

		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%d - IdContexto=%d", IdDoc, IdContexto)
	}

	/* Recupero o prompt da tabela promptsModel*/
	//dataPrompt, err := obj.promptModel.SelectByNatureza(models.PROMPT_NATUREZA_IDENTIFICA)
	dataPrompt, err := PromptServiceGlobal.SelectByNatureza(models.PROMPT_NATUREZA_IDENTIFICA)
	if err != nil {

		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%d - IdContexto=%d", IdDoc, IdContexto)
	}
	//var messages openAI.MsgGpt
	var messages MsgGpt
	messages.CreateMessage("", "user", row.Doc)
	messages.CreateMessage("", "user", dataPrompt.TxtPrompt)

	retSubmit, err := OpenaiServiceGlobal.SubmitPromptResponse(ctx, messages, nil, "")
	if err != nil {
		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%d - IdContexto=%d", IdDoc, IdContexto)
	}

	/* Verifico se a resposta é um json válido*/

	rspJson := retSubmit.Output[0].Content[0].Text
	// Limpar espaços em branco
	rspJson = strings.TrimSpace(rspJson)
	// Opcional: remover possíveis backticks ou aspas extras no início/fim
	rspJson = strings.Trim(rspJson, "`\"")
	// Log para ajudar na depuração
	logger.Log.Infof("JSON retornado OpenAI: %s", rspJson)

	var objJson = DocumentoBase{}
	decoder := json.NewDecoder(strings.NewReader(rspJson))
	//decoder.DisallowUnknownFields() // opcional, para ajudar a detectar campos inesperados

	err = decoder.Decode(&objJson)
	if err != nil {
		return fmt.Errorf("ERROR: Erro ao fazer o parse do JSON: %w", err)
	}

	// err = json.Unmarshal([]byte(rspJson), &objJson)
	// if err != nil {

	// 	return fmt.Errorf("ERROR: Erro ao fazer o parse do JSON")
	// }
	isAutuado, err := AutosServiceGlobal.IsDocAutuado(IdContexto, objJson.IdPje)
	if err != nil {
		return fmt.Errorf("ERROR: Erro ao verificar se documento já existe")

	}
	if isAutuado {
		return fmt.Errorf("ERROR: Documento já existe na tabela autosModel")
	}

	//Faz a inclusão do documentos na tabela autos
	idCtxt := IdContexto
	idNatu := objJson.Tipo.Key
	idPje := objJson.IdPje
	docJson := json.RawMessage(rspJson)

	//autosParams.DocJson =rspJson // Suporte para JSON nativo no Go

	//autosParams.DtInc = time.Now()
	//autosParams.Status = "S"

	_, err = AutosServiceGlobal.InserirAutos(idCtxt, idNatu, idPje, row.Doc, docJson)
	if err != nil {

		return fmt.Errorf("ERROR: Erro na inclusão do registro na tabela autosModel")

	}

	//Faz a deleção do registro na tabela temp_autos

	// err = obj.idx.Delete(reg.IdDoc)
	// if err != nil {

	// 	return fmt.Errorf("ERROR: Erro ao deletar registro na tabela temp_autos")
	// }

	msg = "Concluído com sucesso!"
	logger.Log.Info(msg)
	return nil

}

func (obj *AutosTempServiceType) VerificarNaturezaDocumento(ctx context.Context, texto string) (*NaturezaDoc, error) {

	var msgs MsgGpt
	assistente := `O seguinte texto pertence aos autos de um processo judicial. 

Primeiramente, verifique se o texto é uma movimentação, registro ou anotação processual, contendo expressões como:
"Mov.", "Movimentação", "Observações dos Movimentos", "Registro", "Publicação", "Entrada", "Intimação", "Anotação".
Se essas expressões estiverem presentes, e o texto não contiver o corpo formal completo da decisão (com fundamentação e conclusão explícita do juiz),
classifique o documento como:
- { "key": 1003, "description": "movimentação/processo" }.

Em seguida, verifique se o texto contém alguma das expressões indicativas de certidões ou outros documentos, tais como:
"certidão", "certifico que", "Por ordem do MM. Juiz", "teor do ato", "o referido é verdade, dou fé",
"encaminhado edital/relação para publicação", "ato ordinatório".

Se qualquer dessas expressões estiver presente em qualquer parte do texto, incluindo cabeçalhos, movimentações ou descrições, classifique o documento imediatamente como:
- { "key": 1002, "description": "certidões" } se for claramente certidão,
- caso contrário, classifique como { "key": 1001, "description": "outros documentos" }.

Somente se nenhuma dessas expressões estiver presente, analise o conteúdo para identificar a natureza do documento conforme as opções a seguir:

{ "key": 1, "description": "Petição inicial" }
{ "key": 2, "description": "Contestação" }
{ "key": 3, "description": "Réplica" }
{ "key": 4, "description": "Despacho inicial" }
{ "key": 5, "description": "Despacho" }
{ "key": 6, "description": "Petição" }
{ "key": 7, "description": "Decisão" }
{ "key": 8, "description": "Sentença" }
{ "key": 9, "description": "Embargos de declaração" }
{ "key": 10, "description": "Contra-razões" }
{ "key": 11, "description": "Recurso" }
{ "key": 12, "description": "Procuração" }
{ "key": 13, "description": "Rol de Testemunhas" }
{ "key": 14, "description": "Contrato" }
{ "key": 15, "description": "Laudo Pericial" }
{ "key": 16, "description": "Ata de audiência" }
{ "key": 17, "description": "Parecer do Ministério Público" }

Se não puder identificar claramente a natureza do texto, classifique como { "key": 1001, "description": "outros documentos" }.

Responda apenas com um JSON no formato: {"key": int, "description": string }.`

	msgs.CreateMessage("", ROLE_USER, assistente)
	msgs.CreateMessage("", ROLE_USER, texto)

	//retSubmit, err := services.OpenaiServiceGlobal.SubmitPromptResponse(ctx, msgs, nil, "gpt-4.1-nano")
	retSubmit, err := OpenaiServiceGlobal.SubmitPromptResponse(ctx, msgs, nil, "gpt-4.1-mini")
	if err != nil {
		logger.Log.Errorf("Erro no SubmitPrompt: %s", err)
		return nil, erros.CreateError("Erro ao verificar a  natureza do  documento!")
	}

	resp := strings.TrimSpace(retSubmit.OutputText())
	//logger.Log.Infof("Resposta do modelo: %s", resp)

	var natureza NaturezaDoc
	err = json.Unmarshal([]byte(resp), &natureza)
	if err != nil {
		logger.Log.Warningf("Erro ao parsear JSON da resposta: %v", err)
		logger.Log.Warningf("Resposta recebida: %s", resp)
		return nil, erros.CreateError("Resposta inesperada ou formato inválido do modelo")
	}

	logger.Log.Infof("Natureza documento identificada: key=%d, description=%s", natureza.Key, natureza.Description)

	return &natureza, nil
}
