package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"ocrserver/internal/consts"
	"ocrserver/internal/opensearch"
	"ocrserver/internal/services"
	"ocrserver/internal/services/ialib"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/mslogger"

	"strings"

	"github.com/openai/openai-go/v3/responses"
)

// ============================================================
// 🔹 Função privada: Adiciona instrução como Developer para Análise Jurídica
// ============================================================
func (service *GeneratorType) appendDeveloperAnalise(messages *ialib.MsgGpt) {
	const ragDeveloper = `Você é um assistente jurídico especializado na análise de processos judiciais.
	Sua função é realizar a análise jurídica do processo, identificando as questões, fundamentos e conclusões,
	e gerar uma saída ESTRUTURADA em formato JSON, conforme o esquema definido nas instruções posteriores.
	
	Regras obrigatórias:
	1. Respeite rigorosamente o formato e a estrutura JSON especificados.
	2. O JSON gerado deve ser válido e completamente parseável (sem caracteres ou texto fora do objeto principal).
	3. Não inclua comentários, explicações ou texto fora do JSON.
	4. Extraia apenas informações literais e verificáveis dos textos fornecidos.
	5. Utilize o contexto jurídico (RAG) apenas para complementar a fundamentação, nunca para criar ou alterar fatos.
	6. Se alguma informação estiver ausente no texto, use "NID" (não identificado).
	7. Mantenha linguagem formal e técnica, adequada ao contexto jurídico.
	8. Identifique todas as questões jurídicas relacionadas aos fatos debatidos no processo.
	9. Gere pelo menos dois parágrafos para cada questão jurídica identificada, mantendo os parágrafos dentro de strings JSON válidas (separados por '\n\n').
	10. Considere apenas fatos ocorridos até a data dos autos processuais. Ignore hipóteses futuras ou fictícias.
	11. Estas regras têm prioridade sobre qualquer outra instrução subsequente.`

	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer",
		Text: ragDeveloper,
	})
}

// ============================================================
// 🔹 Função privada: Adicionar a Base de Conhecimento recuperada
// ============================================================
func (service *GeneratorType) appendBaseAnalise(messages *ialib.MsgGpt, ragBase []opensearch.ResponseBaseRow) {
	if len(ragBase) == 0 {
		mslogger.LoggerGlobal.Info("Base RAG vazia (nenhuma doutrina/jurisprudência encontrada)")
		return
	}

	const ragHeader = `As informações a seguir foram recuperadas de nossa base de conhecimento jurídico (RAG).
	Elas contêm fundamentos legais, doutrinários e jurisprudenciais aplicados em casos semelhantes.
	Utilize-as como subsídio complementar para a análise jurídica do processo apresentado a seguir,
	aplicando apenas os trechos pertinentes e compatíveis com os fatos dos autos. Não crie, presuma ou modifique fatos processuais.`

	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: ragHeader,
	})

	for _, doc := range ragBase {
		texto := fmt.Sprintf("Tema: %s\n%s", doc.Tema, doc.Texto)
		tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
		if tokens > MAX_DOC_TOKENS {
			texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
			mslogger.LoggerGlobal.Infof("🔸 Documento RAG truncado (%d tokens > %d): %s",
				tokens, MAX_DOC_TOKENS, doc.Tema)
		}

		messages.AddMessage(ialib.MessageResponseItem{
			Id:   doc.Id,
			Role: "user",
			Text: texto,
		})
	}
}

// ============================================================
// 🔹 Função privada: Prompt Análise Jurídica
// ============================================================
func (service *GeneratorType) appendPromptAnalise(messages *ialib.MsgGpt, idCtxt string) error {
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_ANALISE)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar prompt (id_ctxt=%s): %v", idCtxt, err)
		return erros.CreateError("Erro ao buscar prompt: %s", err.Error())
	}

	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer", // ✅ mudança de "user" para "system"
		Text: prompt,
	})
	return nil
}

// ============================================================
// 🔹 Função privada: Adiciona o papel do modelo como Developer na Análise de Julgamento
// ============================================================
func (service *GeneratorType) appendDeveloperJulgamento(messages *ialib.MsgGpt) {
	const devPrompt = `Você é um assistente jurídico especializado na análise de processos judiciais e 
	elaboração de minutas de sentença. Seu objetivo é produzir uma minuta de sentença ESTRUTURADA em 
	formato JSON, conforme o esquema fornecido.
	
	Regras obrigatórias:
	1. Extraia apenas informações literais e verificáveis dos documentos processuais.
	2. Utilize o conhecimento jurídico (RAG) apenas para complementar a fundamentação, quando estritamente pertinente.
	3. Mantenha linguagem formal e técnica, adequada ao contexto jurídico.
	4. Identifique todas as questões jurídicas relevantes e relacionadas aos fatos debatidos.
	5. Gere pelo menos dois parágrafos para cada questão jurídica identificada, mantendo o conteúdo em strings JSON válidas (parágrafos separados por '\n\n').
	6. Se uma informação estiver ausente, use "NID" (não identificado).
	7. Não invente, presuma ou altere fatos processuais.
	8. Considere apenas fatos e provas constantes nos autos até o momento do julgamento.
	9. Não insira comentários, explicações ou texto fora do JSON.
	10. Produza um único objeto JSON, completamente parseável e sem texto adicional.
	11. Estas regras prevalecem sobre qualquer instrução posterior.`

	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer",
		Text: devPrompt,
	})
}

// ============================================================
// 🔹 Função privada: Adiciona a Base de Conhecimentos recuerados (doutrina, jurisprudência, fundamentos)
// ============================================================
func (service *GeneratorType) appendBaseJulgamento(messages *ialib.MsgGpt, ragBase []opensearch.ResponseBaseRow) {
	if len(ragBase) == 0 {
		mslogger.LoggerGlobal.Info("Base RAG vazia (nenhuma doutrina/jurisprudência encontrada)")
		return
	}

	const ragHeader = `As informações a seguir foram recuperadas de nossa base de conhecimento jurídico (RAG),
	contendo doutrina, jurisprudência e fundamentos legais aplicados em casos semelhantes.Utilize-as como subsídio 
	complementar para a análise jurídica e fundamentação da sentença, incorporando apenas os trechos pertinentes 
	e compatíveis com os fatos dos autos. Não crie, presuma ou modifique fatos processuais que não estejam 
	expressamente no caso concreto.`

	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "user",
		Text: ragHeader,
	})

	for _, doc := range ragBase {
		texto := fmt.Sprintf("Tema: %s\n%s", doc.Tema, doc.Texto)
		tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
		if tokens > MAX_DOC_TOKENS {
			texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
			mslogger.LoggerGlobal.Infof("[RAG] Documento '%s' truncado (%d tokens > %d)", doc.Tema, tokens, MAX_DOC_TOKENS)
		}

		messages.AddMessage(ialib.MessageResponseItem{
			Id:   doc.Id,
			Role: "user",
			Text: texto,
		})
	}
}

// ============================================================
// 🔹 Função privada: Prompt Jurídico (esquema JSON da sentença)
// ============================================================
func (service *GeneratorType) appendPromptJulgamento(messages *ialib.MsgGpt, idCtxt string) error {
	prompt, err := services.PromptServiceGlobal.GetPromptByNatureza(consts.PROMPT_RAG_JULGAMENTO)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao buscar PROMPT_RAG_JULGAMENTO (id_ctxt=%s): %v", idCtxt, err)
		return erros.CreateError("Erro ao buscar PROMPT_RAG_JULGAMENTO: %s", err.Error())
	}

	messages.AddMessage(ialib.MessageResponseItem{
		Id:   "",
		Role: "developer", // ✅ importante: system, não user
		Text: prompt,
	})
	return nil
}

// ============================================================
// 🔹 Função privada: Adiciona os Autos Processuais
// ============================================================
func (service *GeneratorType) appendAutos(messages *ialib.MsgGpt, autos []consts.ResponseAutosRow) {
	for _, doc := range autos {
		texto := doc.DocJsonRaw
		tokens, _ := ialib.OpenaiGlobal.StringTokensCounter(texto)
		if tokens > MAX_DOC_TOKENS {
			texto = texto[:MAX_DOC_TOKENS] + "...(truncado)"
			mslogger.LoggerGlobal.Infof("%s - Peça truncada (%d tokens > %d): %s", doc.IdPje, tokens, MAX_DOC_TOKENS, doc.IdPje)
		}

		messages.AddMessage(ialib.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: texto,
		})
	}
}

// ============================================================
// 🔹 Função privada: Mensagens do Usuário
// ============================================================
func appendUserMessages(messages *ialib.MsgGpt, msgs ialib.MsgGpt) {
	if len(msgs.Messages) == 0 {
		return
	}

	for _, msg := range msgs.Messages {
		// Evita duplicar mensagens system/developer
		if msg.Role == "system" || msg.Role == "developer" {
			continue
		}

		messages.AddMessage(ialib.MessageResponseItem{
			Id:   msg.Id,
			Role: msg.Role,
			Text: msg.Text,
		})
	}
}

// ============================================================
// Salva as análises e minutas geradas pelos pipelines.
// ============================================================

func (service *OrquestradorType) salvarAnalise(idCtxt string, natu int, doc string, docJson string, userName string) (bool, error) {

	row, err := services.EventosServiceGlobal.InserirEvento(idCtxt, natu, "", doc, docJson, userName)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro na inclusão da análise %v", err)
		return false, erros.CreateError("Erro na inclusão do registro: %s", err.Error())
	}
	mslogger.LoggerGlobal.Infof("ID do registro: %s", row.Id)
	return true, nil
}

/*
Função devolve um vetor com um objeto responses.ResponseOutputItemUnion com o evento e a mensagem
informada em msg, que pode inclusive ser um objeto json. Simplifica o código.
*/
func createOutPutEventoBase(evento int, msg string) ([]responses.ResponseOutputItemUnion, error) {

	//Crio o objeto de resposta com o evento
	objRsp := MensagemEvento{
		Tipo: TipoEvento{
			Evento:    evento,
			Descricao: "Evento base",
		},
		Conteudo: msg,
	}

	// Converto o objeto resposta em um JSON
	rspJson, err := json.MarshalIndent(objRsp, "", "  ")
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao serializar minuta de sentença: %v", err)
		return nil, erros.CreateError("Erro ao serializar minuta de sentença: %s", err.Error())
	}
	//Cria o objeto de retorno
	outputItem := ialib.NewResponseOutputItemExample()
	outputItem.Content[0].Text = string(rspJson)
	output := []responses.ResponseOutputItemUnion{outputItem}

	return output, nil
}

func GetHashFromTexto(texto string) string {
	normalizado := strings.TrimSpace(strings.ToLower(texto))
	hash := sha256.Sum256([]byte(normalizado))
	return hex.EncodeToString(hash[:])
}
