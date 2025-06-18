package analise

import (
	"log"

	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/utils/msgs"
)

const NM_INDEX_MODELOS = "ml-modelos-msmarco"

// Tipos de Análise
const TIPO_ANALISE_PROMPT = 1
const TIPO_ANALISE_CONTEXTO = 2

type BodyRequestContextoQuery struct {
	IdCtxt   int
	Prompt   services.MsgGpt
	ModeloId string
	Tipo     int
}

func BuildAnaliseContexto(body BodyRequestContextoQuery) (*services.MsgGpt, error) {

	log.Println(body.IdCtxt)
	log.Println(body.Prompt)
	log.Println(body.ModeloId)
	log.Println(body.Tipo)
	var Msgs = &services.MsgGpt{}

	//Msgs.CreateMessage(openAI.ROLE_DEVELOPER, "você deve responder e perguntar utilizando um objeto JSON no seguinte formato: { 'cod': int,'msg': string}. O código para uma ")

	//PROMPT - Adiciono as mensagens de prompt da interface do cliente
	var Mensagens = body.Prompt.GetMessages()
	for _, msg := range Mensagens {
		//log.Printf("Documento: %s -> %s", msg.Role, msg.Content)
		Msgs.AddMessage(msg)
	}

	if body.Tipo == TIPO_ANALISE_CONTEXTO {
		//MODELO - Adiciono o modelo a ser utilizado

		//var modelos = opensearch.NewIndexModelos()
		//doc, err := modelos.ConsultaDocumentoById(body.ModeloId)
		doc, err := opensearch.IndexService.GetDocumentoById(body.ModeloId)
		if err != nil {
			msgs.CreateLogTimeMessage("Erro ao selecionar documentos dos autos!")
			return Msgs, err
		}

		Msgs.CreateMessage("", "user", "use o modelo a seguir:")
		Msgs.CreateMessage("", "user", doc.Inteiro_teor)

		//AUTOS - Recupera os registros dos autos
		//var autos = models.NewAutosModel()
		//autosRegs, err := autos.SelectByContexto(body.IdCtxt)
		autosRegs, err := services.AutosService.GetAutosByContexto(body.IdCtxt)
		if err != nil {
			msgs.CreateLogTimeMessage("Erro ao selecionar documentos dos autos!")
			return Msgs, err
		}
		Msgs.CreateMessage("", "user", "a seguir estão os documentos do processo:")
		for _, reg := range autosRegs {

			Msgs.CreateMessage("", "user", string(reg.AutosJson))

		}
		lista := Msgs.GetMessages()
		for _, reg := range lista {
			log.Printf("Mensagem: %s - %s", reg.Role, reg.Text)
			//Msgs.CreateMessage("user", string(reg.AutosJson))

		}
	}

	return Msgs, nil
}
