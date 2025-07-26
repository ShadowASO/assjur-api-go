package services

import (
	"context"
	"encoding/json"
	"fmt"

	"ocrserver/internal/consts"
	"ocrserver/internal/models"
	"ocrserver/internal/services/embedding/parsers"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

/*
**  Pipeline de ingestão dos documentos do processo, sendo salvos nas tabelas "autos", "autos_json_embedding"
 */
func ProcessarDocumento(IdContexto int, IdDoc string) error {
	ctx := context.Background()
	if AutosTempServiceGlobal == nil {
		logger.Log.Error("Tentativa de uso de AutosTempServiceGlobal não iniciado.")
		return fmt.Errorf("tentativa de uso de AutosTempServiceGlobal não iniciado")
	}

	msg := fmt.Sprintf("Processando documento: IdContexto=%d - IdDoc=%s", IdContexto, IdDoc)
	logger.Log.Info(msg)

	/*01 - AUTOS_TEMP: Recupero o registro do índice "autos_temp" */

	row, err := AutosTempServiceGlobal.SelectById(IdDoc)
	if err != nil {
		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%s - IdContexto=%d", IdDoc, IdContexto)
	}

	/*02 - DUPLICIDADE: Verifica, pelo id_pje se o documentos está sendo inserido em duplicidade*/

	isAutuado, err := AutosServiceGlobal.IsDocAutuado(IdContexto, row.IdPje)
	if err != nil {
		logger.Log.Infof("Erro ao verificar a existência  no índice 'autos' %v", err)
		return erros.CreateError("Documento já existe no índice 'autos' ")
	}
	if isAutuado {
		//return erros.CreateError("Documento já existe no índice 'autos' ")
		logger.Log.Info("Documento já existe no índice 'autos' ")
		return erros.CreateError("Documento já existe no índice 'autos' ")
	}

	/*03 - PROMPT: Recupero o prompt da tabela "prompts"*/

	dataPrompt, err := PromptServiceGlobal.SelectByNatureza(models.PROMPT_NATUREZA_IDENTIFICA)
	if err != nil {
		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%s - IdContexto=%d", IdDoc, IdContexto)
	}

	var messages MsgGpt

	messages.CreateMessage("", "user", dataPrompt.TxtPrompt)
	messages.CreateMessage("", "user", row.Doc)

	/*04 - CHATGPT:  Extrai o JSON utilizando o prompt */

	retSubmit, usage, err := OpenaiServiceGlobal.SubmitPromptResponse(ctx, messages, nil, "")
	if err != nil {
		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%s - IdContexto=%d", IdDoc, IdContexto)
	}

	/*05 - TOKENS:= Atualiza o uso de tokens no contexto */

	ContextoServiceGlobal.UpdateTokenUso(IdContexto, int(usage.InputTokens), int(usage.OutputTokens))
	//*************************************

	// 06 - Limpa e prepara a resposta JSON
	rspJson := retSubmit.Output[0].Content[0].Text
	rspJson = strings.TrimSpace(rspJson)
	rspJson = strings.Trim(rspJson, "`\"")

	// 07 - Verifica se o JSON é válido
	var objJson DocumentoBase
	if err := json.Unmarshal([]byte(rspJson), &objJson); err != nil {
		return erros.CreateErrorf("ERROR: Erro ao fazer o parse do JSON: %w", err)
	}

	/*06 - AUTOS: Faz a inclusão do documentos na índice "autos" */

	idCtxt := IdContexto
	idNatu := objJson.Tipo.Key
	idPje := objJson.IdPje

	rowAutos, err := AutosServiceGlobal.InserirAutos(idCtxt, idNatu, idPje, row.Doc, rspJson)
	if err != nil {
		return erros.CreateError("Erro ao inserir documento no índice 'autos'")
	}
	//************************************************************************************************
	// CRIAR O EMBEDDING a partir do texto do documento inserido em "autos"
	// Vamos criar embedding apenas para identificar a causa, a partir da
	// petição inicial, contestação, réplica e demais petições
	if idNatu == consts.NATU_DOC_INICIAL ||
		idNatu == consts.NATU_DOC_CONTESTACAO ||
		idNatu == consts.NATU_DOC_REPLICA ||
		idNatu == consts.NATU_DOC_PETICAO ||
		idNatu == consts.NATU_DOC_PARECER_MP {

		jsonRaw, _ := parsers.ParserDocumentosJson(idNatu, json.RawMessage(rspJson)) // se parser espera RawMessage

		embVector, err := GetDocumentoEmbeddings(jsonRaw)
		if err != nil {
			logger.Log.Errorf("Erro ao extrair os embeddings do documento: %v", err)
			return erros.CreateErrorf("Erro ao extrair o embedding: Contexto: %d - IdDoc: %s", idCtxt, rowAutos.Id)
		}

		/* Insere o registro no índice "autos_json_embedding" */

		_, err = AutosJsonServiceGlobal.InserirEmbedding(rowAutos.Id, idCtxt, idNatu, embVector)
		if err != nil {
			return fmt.Errorf("ERROR: Erro na inclusão do documento no índice 'autos_json_embedding'")
		}
	}

	/*07 - DELETA TEMP_AUTOS:  Faz a deleção do registro na tabela temp_autos  */

	// err = obj.idx.Delete(reg.IdDoc)
	// if err != nil {
	// 	return fmt.Errorf("ERROR: Erro ao deletar registro na tabela temp_autos")
	// }

	msg = "Concluído com sucesso!"
	logger.Log.Info(msg)
	return nil

}
