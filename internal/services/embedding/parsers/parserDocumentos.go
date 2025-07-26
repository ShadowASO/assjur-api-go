/*
File: parserDocumentos.go
Extrai os campos relevantes do json que representa cada documento e os concatena para a geração do embedding.
Data: 16-07-2025
*/
package parsers

import (
	"encoding/json"
	"ocrserver/internal/consts"
)

func ParserDocumentosJson(idNatu int, docJson json.RawMessage) (string, error) {
	jsonRaw := ""
	switch idNatu {

	case consts.NATU_DOC_INICIAL:
		jsonRaw, _ = ParserInicialJson(idNatu, docJson)
	case consts.NATU_DOC_CONTESTACAO:
		jsonRaw, _ = ParserContestacaoJson(idNatu, docJson)
	case consts.NATU_DOC_REPLICA:
		jsonRaw, _ = ParserReplicaJson(idNatu, docJson)
	case consts.NATU_DOC_DESPACHO:
		jsonRaw, _ = ParserDespachoJson(idNatu, docJson)
	case consts.NATU_DOC_PETICAO:
		jsonRaw, _ = ParserPeticaoJson(idNatu, docJson)
	case consts.NATU_DOC_DECISAO:
		jsonRaw, _ = ParserDecisaoJson(idNatu, docJson)
	case consts.NATU_DOC_SENTENCA:
		jsonRaw, _ = ParserSentencaJson(idNatu, docJson)
	case consts.NATU_DOC_EMBARGOS:
		jsonRaw, _ = ParserEmbargosDeclaracaoJson(idNatu, docJson)
	case consts.NATU_DOC_APELACAO:
		jsonRaw, _ = ParserApelacaoJson(idNatu, docJson)
	case consts.NATU_DOC_PROCURACAO:
		jsonRaw, _ = ParserProcuracaoJson(idNatu, docJson)
	case consts.NATU_DOC_ROL_TESTEMUNHAS:
		jsonRaw, _ = ParserRolTestemunhasJson(idNatu, docJson)
	case consts.NATU_DOC_LAUDO_PERICIAL:
		jsonRaw, _ = ParserLaudoPericialJson(idNatu, docJson)
	default:
		jsonRaw = string(docJson)
	}
	return jsonRaw, nil
}
