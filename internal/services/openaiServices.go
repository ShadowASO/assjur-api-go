/*
---------------------------------------------------------------------------------------
File: openaiServices.go
Autor: Aldenor
Data: 15-08-2025
Finalidade: Funções que servem como intermediárias nas chamadas aos serviços da OpenAI,
usando as funções do pacote "ialib" e devem ser chamadas
indiretamente, por meio do pacote services(openaiServices)
---------------------------------------------------------------------------------------
*/
package services

import (
	"context"

	"math"

	"time"

	"fmt"

	"strings"
	"sync"

	"ocrserver/internal/config"

	"ocrserver/internal/services/ialib"
	"ocrserver/internal/services/tools"

	"ocrserver/internal/utils/logger"

	"github.com/openai/openai-go/v2"

	"github.com/openai/openai-go/v2/responses"
)

type OpenaiServiceType struct {
	cfg *config.Config
}

var OpenaiServiceGlobal *OpenaiServiceType
var onceInitOpenAIService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitOpenaiService(apiKey string, cfg *config.Config) {
	onceInitOpenAIService.Do(func() {

		OpenaiServiceGlobal = &OpenaiServiceType{
			cfg: cfg,
		}

		logger.Log.Info("Global OpenaiService configurado com sucesso.")
	})
}

/*
*
Obtém a representação vetorial do texto enviado. Quem for utilizar o valor retornadotem
que saber que se precisar converter para float32, deverá fazê-lo onde necessário.
*/
func (obj *OpenaiServiceType) GetEmbeddingFromText(
	ctx context.Context,
	inputTxt string,
) ([]float32, *openai.CreateEmbeddingResponseUsage, error) {
	if obj == nil {
		return nil, nil, fmt.Errorf("serviço OpenAI não iniciado")
	}
	//Timeout defensivo se caller não definiu
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	vec32, resp, err := ialib.OpenaiGlobal.GetEmbeddingFromText_openai(ctx, inputTxt)
	if err != nil {
		return nil, nil, fmt.Errorf("falha ao obter embedding: %w", err)
	}

	usage := resp.Usage
	if SessionServiceGlobal != nil {
		SessionServiceGlobal.UpdateTokensUso(usage.PromptTokens, usage.TotalTokens-usage.PromptTokens, usage.TotalTokens)
	}

	return vec32, &usage, nil
}

/*
modelo: nome do modelo a usar, ou uma string vazia("")
*/
func (obj *OpenaiServiceType) SubmitPromptResponse(
	ctx context.Context,
	inputMsgs ialib.MsgGpt,
	prevID string,
	modelo string,
	effort responses.ReasoningEffort,
	verbosity responses.ResponseTextConfigVerbosity,
) (*responses.Response, error) {
	if obj == nil {
		return nil, fmt.Errorf("serviço OpenAI não iniciado")
	}

	rsp, err := ialib.OpenaiGlobal.SubmitPromptResponse_openai(ctx,
		inputMsgs,
		prevID,
		modelo,
		effort,
		verbosity)
	if err != nil {
		return nil, fmt.Errorf("falha ao submeter o prompt: %w", err)
	}

	usage := rsp.Usage
	if SessionServiceGlobal != nil {
		SessionServiceGlobal.UpdateTokensUso(usage.InputTokens, usage.OutputTokens, usage.TotalTokens)
	}

	return rsp, err
}

// func (obj *OpenaiServiceType) SubmitResponseFunctionRAG(
// 	ctx context.Context,
// 	idCtxt string,
// 	inputMsgs ialib.MsgGpt,
// 	toolManager *tools.ToolManager,
// 	prevID string,
// 	effort responses.ReasoningEffort,
// 	verbosity responses.ResponseTextConfigVerbosity,
// ) (*responses.Response, error) {
// 	if obj == nil {
// 		return nil, fmt.Errorf("serviço OpenAI não iniciado")
// 	}

// 	rsp, err := ialib.OpenaiGlobal.SubmitResponseFunctionRAG_openai(ctx,
// 		idCtxt,
// 		inputMsgs,
// 		toolManager,
// 		prevID,
// 		effort,
// 		verbosity)
// 	if err != nil {
// 		logger.Log.Errorf("OpenAI Responses.New (passo consolidação) falhou: %v", err)
// 		return nil, err
// 	}

// 	usage := rsp.Usage
// 	if SessionServiceGlobal != nil {
// 		SessionServiceGlobal.UpdateTokensUso(usage.InputTokens, usage.OutputTokens, usage.TotalTokens)
// 	}

// 	return rsp, nil
// }

/*
Função destinada a calcular a quantidade de tokens constantes de um vetor de mensagtens
*/

func (obj *OpenaiServiceType) TokensCounter(inputMsgs ialib.MsgGpt) (int, error) {
	// msgs := inputMsgs.GetMessages()
	// if len(msgs) == 0 {
	// 	return 0, fmt.Errorf("lista de mensagens vazia")
	// }

	// enc, err := tokenizer.Get(tokenizer.Encoding(tokenizer.O200kBase))
	// if err != nil {
	// 	return 0, fmt.Errorf("falha ao obter tokenizer: %w", err)
	// }

	// total := 0
	// for _, it := range msgs {
	// 	ids, _, err := enc.Encode(it.Text)
	// 	if err != nil {
	// 		return 0, fmt.Errorf("falha ao codificar texto: %w", err)
	// 	}
	// 	total += len(ids) + ialib.OPENAI_TOKENS_OVERHEAD_MSG
	// }

	//return total + ialib.OPENAI_TOKENS_AJUSTE, nil
	return ialib.OpenaiGlobal.TokensCounter(inputMsgs)
}

func (obj *OpenaiServiceType) StringTokensCounter(inputStr string) (int, error) {
	msg := ialib.MsgGpt{}
	msg.CreateMessage("", ialib.ROLE_USER, inputStr)
	return obj.TokensCounter(msg)
}

func (obj *OpenaiServiceType) Float64ToFloat32Slice(input []float64) []float32 {
	out := make([]float32, len(input))
	for i, v := range input {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			// substitui por zero e loga
			logger.Log.Warningf("Valor inválido no embedding (índice %d): %v. Substituindo por 0.", i, v)
			v = 0
		}
		out[i] = float32(v)
	}
	return out
}

//Obtem o embedding de cada campo texto do index Modelos e devolve uma strutura.

func GetDocumentoEmbeddings(docText string) ([]float32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vec32, _, err := OpenaiServiceGlobal.GetEmbeddingFromText(ctx, docText)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar embedding: %w", err)
	}
	//vec32 := OpenaiServiceGlobal.Float64ToFloat32Slice(vec32)
	return vec32, nil
}

// helper: normaliza role conhecido
func normalizeRole(role string) responses.EasyInputMessageRole {
	switch role {
	case ialib.ROLE_USER:
		return responses.EasyInputMessageRoleUser
	case ialib.ROLE_ASSISTANT:
		return responses.EasyInputMessageRoleAssistant
	case ialib.ROLE_SYSTEM:
		return responses.EasyInputMessageRoleSystem
	case ialib.ROLE_DEVELOPER:
		return responses.EasyInputMessageRoleDeveloper
	default:
		return responses.EasyInputMessageRoleUser
	}
}

// util: pega o primeiro texto disponível da resposta do modelo
// func FirstTextFromSubmit(retSubmit *responses.Response) (string, error) {
// FirstMessageFromSubmit retorna o primeiro item de output do tipo "message"
// que contenha pelo menos um content part textual ("output_text") não vazio.
func FirstMessageFromSubmit(retSubmit *responses.Response) (responses.ResponseOutputItemUnion, error) {
	// 0) nulo
	if retSubmit == nil {
		return responses.ResponseOutputItemUnion{}, fmt.Errorf("resposta nula do provedor")
	}

	// 2) precisa ter outputs
	if len(retSubmit.Output) == 0 {
		return responses.ResponseOutputItemUnion{}, fmt.Errorf("resposta sem Output")
	}

	// 3) varre por "message" completo
	for _, out := range retSubmit.Output {
		// Alguns SDKs expõem out.Type / out.Status diretamente.
		// Se o seu tiver "Status", checar "completed" ajuda:
		// if out.Status != "completed" { continue }

		if out.Type != "message" {
			continue
		}
		if len(out.Content) == 0 {
			continue
		}

		// Garante que há ao menos um part textual não vazio
		hasText := false
		for _, c := range out.Content {
			if c.Type == "output_text" && strings.TrimSpace(c.Text) != "" {
				hasText = true
				break
			}
		}
		if !hasText {
			// pode ser message com tool result, imagem, etc. Procura o próximo
			continue
		}

		return out, nil
	}

	return responses.ResponseOutputItemUnion{}, fmt.Errorf("nenhum conteúdo textual encontrado na resposta (message/output_text)")
}

// ExtractOutputText varre o content do Union (já sabido "message")
// e retorna o primeiro output_text não vazio.
func ExtractOutputText(msg responses.ResponseOutputItemUnion) (string, error) {
	if msg.Type != "message" {
		return "", fmt.Errorf("tipo não é message: %s", msg.Type)
	}
	for _, c := range msg.Content {
		if c.Type == "output_text" {
			t := strings.TrimSpace(c.Text)
			if t != "" {
				return t, nil
			}
		}
	}
	return "", fmt.Errorf("message sem output_text utilizável")
}

//********************************************************
//               FUNÇÕES RAG
//********************************************************

/*
*
1ª PRIMEIRA A SER CHAMADA:
Deve ser chamada para obter as function_call(s) retornadas e que deverão ser extraídas e re-
submetidas ao modelo, com a indicação do ID. Ela retorna a primeria resposta que deverá con-
ter as funções a serem chamadas pela função "ExtrairResponseTools" e inseridos num
responses.ResponseNewParams{} para ser passada à função "SubmitResponseTools". Com a quebra
das funções, ganhamos mais flexibilidade no manuseio do RAG.
*/
func (obj *OpenaiServiceType) SubmitPromptTools(
	ctx context.Context,
	idCtxt string,
	inputMsgs ialib.MsgGpt,
	toolManager *tools.ToolManager,
	prevID string,
	effort responses.ReasoningEffort,
	verbosity responses.ResponseTextConfigVerbosity,
) (*responses.Response, error) {
	if obj == nil {
		return nil, fmt.Errorf("serviço OpenAI não iniciado")
	}

	if obj.cfg == nil {
		return nil, fmt.Errorf("configuração OpenAI ausente")
	}
	rsp, err := ialib.OpenaiGlobal.SubmitPromptTools_openai(ctx,
		idCtxt,
		inputMsgs,
		toolManager,
		prevID,
		effort,
		verbosity)

	if err != nil {
		return nil, fmt.Errorf("erro ao executar SubmitPromptTools_openai()")
	}
	usage := rsp.Usage
	if SessionServiceGlobal != nil {
		SessionServiceGlobal.UpdateTokensUso(usage.InputTokens, usage.OutputTokens, usage.TotalTokens)
	}

	// devolve os parâmetros prontos para a 2ª chamada e a usage da 1ª
	return rsp, nil
}

/*
2ª CHAMADA
Extrai as funções chamadas pelo modelo e executa as funções reais, utilizando o handlerFunc.
Cada resposta é adicionada a "params" e será devolvida para ser envido ao modelo pel função
"SubmitResponseTools". Esta função executa apenas as funções apoontadas pelo modelo e monta
params com as respostas.
*/
func (obj *OpenaiServiceType) ExtraiResponseTools(
	idCtxt string,
	rsp *responses.Response,
	handlerFunc func(idCtxt string, output responses.ResponseOutputItemUnion) (string, error),
) (responses.ResponseNewParams, bool, error) {
	if obj == nil {
		return responses.ResponseNewParams{}, false, fmt.Errorf("serviço OpenAI não iniciado")
	}

	resp, has, err := ialib.OpenaiGlobal.ExtraiResponseTools_openai(idCtxt, rsp, handlerFunc)
	if err != nil {
		return responses.ResponseNewParams{}, has, fmt.Errorf("serviço OpenAI não iniciado")
	}

	// devolve os parâmetros prontos para a 2ª chamada e a usage da 1ª
	return resp, has, nil

}

// ExtraiResponseTools executa a 2ª chamada (consolidação da resposta final)
// usando os params retornados por SubmitRequestTools.
func (obj *OpenaiServiceType) SubmitResponseTools(
	ctx context.Context,
	reqID string,
	params responses.ResponseNewParams,
	effort responses.ReasoningEffort,
	verbosity responses.ResponseTextConfigVerbosity,
) (*responses.Response, error) {
	if obj == nil {
		return nil, fmt.Errorf("serviço OpenAI não iniciado")
	}

	if len(params.Input.OfInputItemList) == 0 {

		logger.Log.Debug("nenhuma function_call retornada ")

		return nil, fmt.Errorf("nenhuma function_call retornada; 2ª chamada seguirá sem tool outputs")
	}

	resp, err := ialib.OpenaiGlobal.SubmitResponseTools_openai(ctx, reqID, params, effort, verbosity)
	if err != nil {
		return nil, fmt.Errorf("serviço OpenAI não iniciado: %v", err)
	}
	usage := resp.Usage
	if SessionServiceGlobal != nil {
		SessionServiceGlobal.UpdateTokensUso(usage.InputTokens, usage.OutputTokens, usage.TotalTokens)
	}

	return resp, nil
}
