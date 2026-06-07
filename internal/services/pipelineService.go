/*
---------------------------------------------------------------------------------------
File: pipelineService.go
Autor: Aldenor
Data: 27-07-2025
Finalidade: Faz os processamentos necessários para a ingestão dos documentos na tabela
autos e autos_json_embedding.
---------------------------------------------------------------------------------------
*/
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"ocrserver/internal/config"
	"ocrserver/internal/consts"

	openaiservice "ocrserver/internal/services/openai"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/mslogger"

	"strings"
)

/*
**  Pipeline de ingestão dos documentos do processo, sendo salvos nas tabelas "autos", "autos_json_embedding"
 */
func ProcessarDocumento(IdContexto string, IdDoc string) error {
	ctx := context.Background()
	/*   ****  Modelo da OpenAI a ser utilizado */
	modeloOpenAI := config.GlobalConfig.OpenOptionModel
	/*------------------------------------------------*/
	if IdContexto == "" || IdDoc == "" {
		//return fmt.Errorf("idContexto ou idDoc vazio")
		mslogger.LoggerGlobal.Error("IdContexto ou IdDoc vazio.")
		return erros.CreateError("IdContexto ou IdDoc vazio.")
	}
	if AutosTempServiceGlobal == nil {
		mslogger.LoggerGlobal.Error("Objeto global 'AutosTempServiceGlobal' não foi inicializado.")
		return erros.CreateError("Objeto global 'AutosTempServiceGlobal' não foi inicializado.")
	}

	msg := fmt.Sprintf("Processando documento: IdContexto=%s - IdDoc=%s", IdContexto, IdDoc)
	mslogger.LoggerGlobal.Info(msg)

	/*01 - AUTOS_TEMP: Recupero o registro do índice "autos_temp" */

	row, err := AutosTempServiceGlobal.SelectById(IdDoc)
	if err != nil {
		return fmt.Errorf("Documento  não encontrato no índice 'autos_temp' - idDoc=%s - IdContexto=%s", IdDoc, IdContexto)
	}
	mslogger.LoggerGlobal.Infof("\nID PJe: %s - INÍCIO", row.IdPje)
	/*02 - DUPLICIDADE: Verifica, pelo id_pje se o documentos está sendo inserido em duplicidade*/

	isAutuado, err := AutosServiceGlobal.IsDocAutuado(IdContexto, row.IdPje)
	if err != nil {
		mslogger.LoggerGlobal.Infof("Erro ao verificar a existência do documento em 'autos': %v", err)
		return erros.CreateErrorf("Erro ao verificar a existência do documento em 'autos': %v", err.Error())
	}
	if isAutuado {
		mslogger.LoggerGlobal.Errorf("Documento %s já existe no índice 'autos'", IdDoc)
		return erros.CreateErrorf("Documento %s já existe no índice 'autos'", IdDoc)
	}

	/*03 - PROMPT: Recupero o natuPrompt da tabela "prompts"*/
	natuPrompt := consts.PROMPT_ANALISE_AUTUACAO
	//alterado em 22/05/2026
	// if row.IdNatu == consts.NATU_DOC_SENTENCA {
	// 	natuPrompt = consts.PROMPT_AUTUACAO_SENTENCA
	// 	/* Se for uma sentença, utilizamos o modelo mais avançado */
	// 	modeloOpenAI = config.GlobalConfig.OpenOptionModelTop
	// } else if row.IdNatu == consts.NATU_DOC_CERTIDAO {
	// 	natuPrompt = consts.PROMPT_AUTUACAO_CERTIDAO
	// }
	switch row.IdNatu {
	case consts.NATU_DOC_SENTENCA:
		natuPrompt = consts.PROMPT_AUTUACAO_SENTENCA
		/* Se for uma sentença, utilizamos o modelo mais avançado */
		modeloOpenAI = config.GlobalConfig.OpenOptionModelTop
	case consts.NATU_DOC_CERTIDAO:
		natuPrompt = consts.PROMPT_AUTUACAO_CERTIDAO
	}

	prompt, err := PromptServiceGlobal.GetPromptByNatureza(natuPrompt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar prompt natureza=%d: %v", natuPrompt, err)
		return erros.CreateError("Erro ao buscar prompt: %s", err.Error())
	}

	var messages openaiservice.MsgGpt

	messages.CreateMessage("", "user", prompt)
	messages.CreateMessage("", "user", row.Doc)

	/*04 - CHATGPT:  Extrai o JSON utilizando o prompt */

	retSubmit, err := OpenaiServiceGlobal.SubmitPromptResponse(
		ctx,
		messages,
		"",
		//config.GlobalConfig.OpenOptionModel,
		modeloOpenAI,
		openaiservice.REASONING_LOW,
		openaiservice.VERBOSITY_LOW)
	if err != nil {
		return erros.CreateErrorf("idDoc=%s : %s", IdDoc, err.Error())
	}
	usage := retSubmit.Usage

	mslogger.LoggerGlobal.Infof("json=%s", retSubmit.Truncation)
	mslogger.LoggerGlobal.Infof("json=%s", retSubmit.Error.Message)

	/*05 - TOKENS:= Atualiza o uso de tokens no contexto */

	ContextoServiceGlobal.UpdateTokenUso(IdContexto, int(usage.InputTokens), int(usage.OutputTokens))
	//*************************************

	// 06 - Limpa e prepara a resposta JSON

	item, err := FirstMessageFromSubmit(retSubmit)
	if err != nil {
		return erros.CreateErrorf("resposta do modelo sem texto: %v", err)
	}
	rspJson, err := ExtractOutputText(item)

	if err != nil {
		return erros.CreateErrorf("falha ao extrair texto da resposta: %v", err)
	}

	rspJson = strings.Trim(rspJson, "`\"")
	//mslogger.LoggerGlobal.Infof("json=%s", rspJson)

	// 07 - Verifica se o JSON é válido
	var objJson DocumentoBase
	if err := json.Unmarshal([]byte(rspJson), &objJson); err != nil {
		return erros.CreateErrorf("ERROR: Erro ao fazer o parse do JSON: %w", err)
	}

	/*06 - AUTOS: Faz a inclusão do documentos na índice "autos" */

	idCtxt := IdContexto
	idNatu := objJson.Tipo.Key
	idPje := objJson.IdPje

	// rowAutos, err := AutosServiceGlobal.InserirAutos(idCtxt, idNatu, idPje, row.Doc, rspJson)
	_, err = AutosServiceGlobal.InserirAutos(idCtxt, idNatu, idPje, row.Doc, rspJson)
	if err != nil {
		mslogger.LoggerGlobal.Error("Erro ao inserir documento no índice 'autos'")
		return erros.CreateError("Erro ao inserir documento no índice 'autos'")
	}
	//************************************************************************************************
	// CRIAR O EMBEDDING a partir do texto do documento inserido em "autos"
	// Vamos criar embedding apenas para identificar a causa, a partir da
	// petição inicial, contestação, réplica e demais petições
	// if idNatu == consts.NATU_DOC_INICIAL ||
	// 	idNatu == consts.NATU_DOC_CONTESTACAO ||
	// 	idNatu == consts.NATU_DOC_REPLICA ||
	// 	idNatu == consts.NATU_DOC_PETICAO ||
	// 	idNatu == consts.NATU_DOC_PARECER_MP {
	// if false {

	// 	jsonRaw, _ := parsers.ParserDocumentosJson(idNatu, json.RawMessage(rspJson)) // se parser espera RawMessage

	// 	embVector, err := openaiservice.GetDocumentoEmbeddings(jsonRaw)
	// 	if err != nil {
	// 		mslogger.LoggerGlobal.Errorf("Erro ao extrair os embeddings do documento: %v", err)
	// 		return erros.CreateErrorf("Erro ao extrair o embedding: Contexto: %d - IdDoc: %s", idCtxt, rowAutos.Id)
	// 	}

	// 	/* Insere o registro no índice "autos_json_embedding" */

	// 	_, err = AutosJsonServiceGlobal.InserirEmbedding(rowAutos.Id, idCtxt, idNatu, embVector)
	// 	if err != nil {
	// 		mslogger.LoggerGlobal.Errorf("ERROR: Erro na inclusão do documento no índice 'autos_json_embedding'")
	// 		return fmt.Errorf("ERROR: Erro na inclusão do documento no índice 'autos_json_embedding'")
	// 	}
	// }

	/*07 - DELETA TEMP_AUTOS:  Faz a deleção do registro na tabela temp_autos  */

	err = AutosTempServiceGlobal.DeletaAutos(IdDoc)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("ERROR: Erro ao deletar registro no índice 'temp_autos'")
		return fmt.Errorf("ERROR: Erro ao deletar registro no índice 'temp_autos'")
	}

	//msg = "Concluído com sucesso!"
	//mslogger.LoggerGlobal.Info(msg)
	mslogger.LoggerGlobal.Infof("\nID PJe: %s - CONCLUÍDO", row.IdPje)
	return nil

}
