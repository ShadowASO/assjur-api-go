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

	//Recupero o registro do índice "autos_temp"
	row, err := AutosTempServiceGlobal.SelectById(IdDoc)
	if err != nil {
		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%s - IdContexto=%d", IdDoc, IdContexto)
	}

	/* Recupero o prompt da tabela "prompts"*/

	dataPrompt, err := PromptServiceGlobal.SelectByNatureza(models.PROMPT_NATUREZA_IDENTIFICA)
	if err != nil {
		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%s - IdContexto=%d", IdDoc, IdContexto)
	}

	var messages MsgGpt

	messages.CreateMessage("", "user", dataPrompt.TxtPrompt)
	messages.CreateMessage("", "user", row.Doc)

	/* Extrai o JSON utilizando o prompt */

	retSubmit, usage, err := OpenaiServiceGlobal.SubmitPromptResponse(ctx, messages, nil, "")
	if err != nil {
		return fmt.Errorf("ERROR: Arquivo não encontrato - idDoc=%s - IdContexto=%d", IdDoc, IdContexto)
	}

	//*** Atualizo o uso de tokens para o contexto
	//idCtxt := IdContexto
	ContextoServiceGlobal.UpdateTokenUso(IdContexto, int(usage.InputTokens), int(usage.OutputTokens))
	//*************************************

	/* Verifico se a resposta é um json válido*/

	rspJson := retSubmit.Output[0].Content[0].Text
	// Limpar espaços em branco
	rspJson = strings.TrimSpace(rspJson)
	// Opcional: remover possíveis backticks ou aspas extras no início/fim
	rspJson = strings.Trim(rspJson, "`\"")
	// Log para ajudar na depuração
	//logger.Log.Infof("JSON retornado OpenAI: %s", rspJson)

	var objJson = DocumentoBase{}
	decoder := json.NewDecoder(strings.NewReader(rspJson))
	//decoder.DisallowUnknownFields() // opcional, para ajudar a detectar campos inesperados

	err = decoder.Decode(&objJson)
	if err != nil {
		return erros.CreateErrorf("ERROR: Erro ao fazer o parse do JSON: %w", err)
	}

	/* Verifica, pelo id_pje se o documentos está sendo inserido em duplicidade*/

	isAutuado, err := AutosServiceGlobal.IsDocAutuado(IdContexto, objJson.IdPje)
	if err != nil {
		return erros.CreateError("Documento já existe no índice 'autos' ")
	}
	if isAutuado {
		//return erros.CreateError("Documento já existe no índice 'autos' ")
		logger.Log.Info("Documento já existe no índice 'autos' ")
	}

	/* Faz a inclusão do documentos na índice "autos" */

	idCtxt := IdContexto
	idNatu := objJson.Tipo.Key
	idPje := objJson.IdPje
	docJson := json.RawMessage(rspJson)

	rowAutos, err := AutosServiceGlobal.InserirAutos(idCtxt, idNatu, idPje, row.Doc, docJson)
	if err != nil {
		return erros.CreateError("Documento inserido no índice 'autos'")
	}
	//************************************************************************************************
	// CRIAR O EMBEDDING a partir do texto do documento inserido em "autos"
	// Vamos criar embedding apenas para identificar a causa, a partir da
	// petição inicial, contestação, réplica e demais petições
	if idNatu == consts.NATU_DOC_INICIAL || idNatu == consts.NATU_DOC_CONTESTACAO ||
		idNatu == consts.NATU_DOC_REPLICA || idNatu == consts.NATU_DOC_PETICAO ||
		idNatu == consts.NATU_DOC_PARECER_MP {

		jsonRaw, _ := parsers.ParserDocumentosJson(idNatu, docJson)

		embVector, err := GetDocumentoEmbeddings(jsonRaw)
		if err != nil {
			logger.Log.Errorf("Erro ao extrair os embeddings do documento: %v", err)
			return erros.CreateErrorf("Erro ao extrair o embedding: Contexto: %d - IdDoc: %s", idCtxt, rowAutos.Id)
		}

		/* Insere o registro no índice "autos_json_embedding" */

		_, err = AutosJsonServiceGlobal.InserirEmbedding(IdDoc, idCtxt, idNatu, embVector)
		if err != nil {
			return fmt.Errorf("ERROR: Erro na inclusão do documento no índice 'autos_json_embedding'")
		}
	}

	/* Faz a deleção do registro na tabela temp_autos  */

	// err = obj.idx.Delete(reg.IdDoc)
	// if err != nil {
	// 	return fmt.Errorf("ERROR: Erro ao deletar registro na tabela temp_autos")
	// }

	msg = "Concluído com sucesso!"
	logger.Log.Info(msg)
	return nil

}
