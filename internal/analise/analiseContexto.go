package analise

import (
	"log"

	"ocrserver/internal/opensearch"
	"ocrserver/internal/services/openAI"
	"ocrserver/internal/utils/msgs"
	"ocrserver/models"
)

const NM_INDEX_MODELOS = "ml-modelos-msmarco"

func BuildAnaliseContexto(body models.BodyRequestContextoQuery) (*openAI.MsgGpt, error) {

	// log.Println(body.IdCtxt)
	// log.Println(body.Prompt)
	// log.Println(body.ModeloId)
	// log.Println(body.Tipo)
	var Msgs = &openAI.MsgGpt{}

	//PROMPT - Adiciono as mensagens de prompt da interface do cliente
	var Mensagens = body.Prompt.GetMessages()
	for _, msg := range Mensagens {
		//log.Printf("Documento: %s -> %s", msg.Role, msg.Content)
		Msgs.AddMessage(msg)
	}

	//MODELO - Adiciono o modelo a ser utilizado

	var modelos = opensearch.NewIndexModelos()
	doc, err := modelos.ConsultaDocumentoById(body.ModeloId)
	if err != nil {
		msgs.CreateLogTimeMessage("Erro ao selecionar documentos dos autos!")
		return Msgs, err
	}

	Msgs.CreateMessage("user", "use o modelo a seguir:")
	Msgs.CreateMessage("user", doc.Inteiro_teor)

	//AUTOS - Recupera os registros dos autos
	var autos = models.NewAutosModel()
	autosRegs, err := autos.SelectByContexto(body.IdCtxt)
	if err != nil {
		msgs.CreateLogTimeMessage("Erro ao selecionar documentos dos autos!")
		return Msgs, err
	}
	Msgs.CreateMessage("user", "a seguir estão os documentos do processo:")
	for _, reg := range autosRegs {

		Msgs.CreateMessage("user", string(reg.AutosJson))

	}
	lista := Msgs.GetMessages()
	for _, reg := range lista {
		log.Printf("Mensagem: %s - %s", reg.Role, reg.Content)
		//Msgs.CreateMessage("user", string(reg.AutosJson))

	}

	return Msgs, nil
}
