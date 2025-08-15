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

	"ocrserver/internal/config"
	"ocrserver/internal/consts"
	"ocrserver/internal/services/openapi"

	"ocrserver/internal/opensearch"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"sync"
)

type AutosTempServiceType struct {
	idx *opensearch.AutosTempIndexType
}

var AutosTempServiceGlobal *AutosTempServiceType
var onceInitAutosTempService sync.Once

type DocumentoBase struct {
	Tipo *struct {
		Key         int    `json:"key"`
		Description string `json:"description"`
	} `json:"tipo"`
	Processo string `json:"processo"`
	IdPje    string `json:"id_pje"`
}

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitAutos_tempService(idx *opensearch.AutosTempIndexType) {
	onceInitAutosTempService.Do(func() {
		AutosTempServiceGlobal = &AutosTempServiceType{
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

func (obj *AutosTempServiceType) VerificarNaturezaDocumento(ctx context.Context, idCtxt int, texto string) (*NaturezaDoc, error) {

	var msgs openapi.MsgGpt
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

	msgs.CreateMessage("", openapi.ROLE_USER, assistente)
	msgs.CreateMessage("", openapi.ROLE_USER, texto)

	//retSubmit, err := services.OpenaiServiceGlobal.SubmitPromptResponse(ctx, msgs, nil, "gpt-5-nano")
	retSubmit, err := OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		msgs,
		"",
		config.GlobalConfig.OpenOptionModel,
		openapi.REASONING_LOW,
		openapi.VERBOSITY_LOW)
	if err != nil {
		logger.Log.Errorf("Erro no SubmitPrompt: %s", err)
		return nil, erros.CreateError("Erro ao verificar a  natureza do  documento!")
	}
	usage := retSubmit.Usage
	//*** Atualizo o uso de tokens para o contexto
	ContextoServiceGlobal.UpdateTokenUso(idCtxt, int(usage.InputTokens), int(usage.OutputTokens))
	//******************************************

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
