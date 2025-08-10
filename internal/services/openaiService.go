package services

import (
	"context"
	"errors"
	"math"

	"time"

	"fmt"

	"strings"
	"sync"

	"ocrserver/internal/config"

	"ocrserver/internal/services/tools"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/tiktoken-go/tokenizer"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/responses"
)

// **************** MENSAGENS - OpenAI   **********************************
// a quandiade de tokens calculada pela OpenAI parecer ser acrescida sempre
// de 7 tokens. Essa constante será utilizada para ajustar o cálculo feito
// pela função "TokensCounter"
const OPENAI_TOKENS_AJUSTE = 7

// Roles
const ROLE_USER = "user"
const ROLE_ASSISTANT = "assistant"
const ROLE_SYSTEM = "system"
const ROLE_DEVELOPER = "developer"

// Reasoning

const REASONING_MIN = responses.ReasoningEffortMinimal
const REASONING_LOW = responses.ReasoningEffortLow
const REASONING_MEDIUM = responses.ReasoningEffortMedium
const REASONING_HIGH = responses.ReasoningEffortHigh

// Verbosity
const VERBOSITY_LOW = responses.ResponseTextConfigVerbosityLow
const VERBOSITY_MEDIUM = responses.ResponseTextConfigVerbosityMedium
const VERBOSITY_HIGH = responses.ResponseTextConfigVerbosityHigh

type MessageResponseItem struct {
	Id   string `json:"id"`
	Role string `json:"role"`
	Text string `json:"text"`
}

type MsgGpt struct {
	Messages []MessageResponseItem `json:"messages"`
}

func (m *MsgGpt) AddMessage(message MessageResponseItem) {
	m.Messages = append(m.Messages, message)
}
func (m *MsgGpt) CreateMessage(id string, role string, message string) {

	m.Messages = append(m.Messages, MessageResponseItem{Id: id, Role: role, Text: message})
}

func (m *MsgGpt) GetMessages() []MessageResponseItem {
	return m.Messages
}

// ***********************************************************************

type OpenaiServiceType struct {
	client openai.Client
	cfg    *config.Config
}

var OpenaiServiceGlobal *OpenaiServiceType
var onceInitOpenAIService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitOpenaiService(apiKey string, cfg *config.Config) {
	onceInitOpenAIService.Do(func() {

		OpenaiServiceGlobal = &OpenaiServiceType{
			client: openai.NewClient(
				option.WithAPIKey(apiKey),
			),
			cfg: cfg,
		}

		logger.Log.Info("Global OpenaiService configurado com sucesso.")
	})
}

func NewOpenaiClient(apiKey string, cfg *config.Config) *OpenaiServiceType {
	return &OpenaiServiceType{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		cfg:    cfg,
	}
}

// // backoff simples
// func retryBackoff(attempt int) time.Duration {
// 	// 200ms, 400ms, 800ms, máx 2s + jitter
// 	base := 200 * time.Millisecond
// 	d := base << (attempt - 1)
// 	if d > 2*time.Second {
// 		d = 2 * time.Second
// 	}
// 	jitter := time.Duration(rand.Int63n(int64(100 * time.Millisecond)))
// 	return d + jitter
// }

/*
*
Obtém a representação vetorial do texto enviado. Quem for utilizar o valor retornadotem
que saber que se precisar converter para float32, deverá fazê-lo onde necessário.
*/
func (obj *OpenaiServiceType) GetEmbeddingFromText(
	ctx context.Context,
	inputTxt string,
) ([]float64, *openai.CreateEmbeddingResponseUsage, error) {
	if obj == nil {
		return nil, nil, fmt.Errorf("serviço OpenAI não iniciado")
	}

	// timeout defensivo se o caller não definiu
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	var (
		resp *openai.CreateEmbeddingResponse
		err  error
	)

	// retry 3x em 429/5xx com backoff exponencial simples
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = obj.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
			Model:          openai.EmbeddingModelTextEmbedding3Large,
			Input:          openai.EmbeddingNewParamsInputUnion{OfString: openai.String(inputTxt)},
			EncodingFormat: openai.EmbeddingNewParamsEncodingFormat("float"),
		})
		if err == nil {
			break
		}

		var apiErr *openai.Error
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 429 || apiErr.StatusCode >= 500) && attempt < 3 {
			// 200ms, 400ms, 800ms
			//time.Sleep(time.Duration(1<<uint(attempt-1)) * 200 * time.Millisecond)
			time.Sleep(erros.RetryBackoff(attempt))
			continue
		}
		break
	}

	if err != nil {
		return nil, nil, fmt.Errorf("falha ao obter embedding: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, nil, fmt.Errorf("nenhum embedding retornado")
	}

	embedding := resp.Data[0].Embedding // []float64

	// (Opcional) apenas loga se vier dimensão inesperada
	if l := len(embedding); l != 3072 {
		logger.Log.Warningf("Dimensão do embedding inesperada: %d (esperado 3072 para text-embedding-3-large)", l)
	}

	usage := resp.Usage
	logger.Log.Infof("Modelo: %s - TOKENS Embeddings - Prompt: %d - Total: %d",
		resp.Model, usage.PromptTokens, usage.TotalTokens)

	if nrTokens, _ := OpenaiServiceGlobal.StringTokensCounter(inputTxt); nrTokens > 0 {
		logger.Log.Infof("Estimativa de tokens no texto: %d", nrTokens)
	}

	if SessionServiceGlobal != nil {
		SessionServiceGlobal.UpdateTokensUso(usage.PromptTokens, usage.TotalTokens-usage.PromptTokens, usage.TotalTokens)
	}

	return embedding, &usage, nil
}

/*
modelo: nome do modelo a usar, ou uma string vazia("")
*/
func (obj *OpenaiServiceType) SubmitPromptResponse(
	ctx context.Context,
	inputMsgs MsgGpt,
	prevID string,
	modelo string,
	effort responses.ReasoningEffort,
	verbosity responses.ResponseTextConfigVerbosity,
) (*responses.Response, *responses.ResponseUsage, error) {
	if obj == nil {
		return nil, nil, fmt.Errorf("serviço OpenAI não iniciado")
	}
	msgs := inputMsgs.GetMessages()
	if len(msgs) == 0 {
		return nil, nil, fmt.Errorf("lista de mensagens vazia")
	}

	items := make([]responses.ResponseInputItemUnionParam, 0, len(msgs))
	for _, it := range msgs {
		items = append(items, responses.ResponseInputItemUnionParam{OfMessage: toEasyInputMessage(it)})
	}

	model := obj.cfg.OpenOptionModel
	if strings.TrimSpace(modelo) != "" {
		model = modelo
	}

	params := responses.ResponseNewParams{
		Model:           model,
		Reasoning:       openai.ReasoningParam{Effort: effort},
		MaxOutputTokens: openai.Int(int64(config.GlobalConfig.OpenOptionMaxCompletionTokens)),
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: items},
		Text:            responses.ResponseTextConfigParam{Verbosity: verbosity},
	}
	if prevID != "" {
		params.PreviousResponseID = openai.String(prevID)
	}

	rsp, err := obj.client.Responses.New(ctx, params)
	if err != nil {
		logger.Log.Errorf("OpenAI Responses.New falhou: %v", err)
		return nil, nil, err
	}

	usage := rsp.Usage
	if SessionServiceGlobal != nil {
		SessionServiceGlobal.UpdateTokensUso(usage.InputTokens, usage.OutputTokens, usage.TotalTokens)
	}

	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		rsp.Model, usage.InputTokens, usage.OutputTokens, usage.TotalTokens)

	return rsp, &usage, nil
}

func (obj *OpenaiServiceType) SubmitResponseFunctionRAG(
	ctx context.Context,
	idCtxt string,
	inputMsgs MsgGpt,
	toolManager *tools.ToolManager,
	prevID string,
	effort responses.ReasoningEffort,
	verbosity responses.ResponseTextConfigVerbosity,
) (*responses.Response, *responses.ResponseUsage, error) {
	if obj == nil {
		return nil, nil, fmt.Errorf("serviço OpenAI não iniciado")
	}

	msgs := inputMsgs.GetMessages()
	if len(msgs) == 0 {
		return nil, nil, fmt.Errorf("lista de mensagens vazia")
	}

	inputItems := make([]responses.ResponseInputItemUnionParam, 0, len(msgs))
	for _, it := range msgs {
		inputItems = append(inputItems, responses.ResponseInputItemUnionParam{OfMessage: toEasyInputMessage(it)})
	}

	params := responses.ResponseNewParams{
		Model:           obj.cfg.OpenOptionModel,
		MaxOutputTokens: openai.Int(int64(config.GlobalConfig.OpenOptionMaxCompletionTokens)),
		Tools:           toolManager.GetAgentTools(),
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems},
		Reasoning:       responses.ReasoningParam{Effort: effort},
		Text:            responses.ResponseTextConfigParam{Verbosity: verbosity},
	}
	if prevID != "" {
		params.PreviousResponseID = openai.String(prevID)
	}

	// 1ª chamada: o modelo decide ferramentas
	rsp, err := obj.client.Responses.New(ctx, params)
	if err != nil {
		logger.Log.Errorf("OpenAI Responses.New (passo ferramentas) falhou: %v", err)
		return nil, nil, err
	}

	// Preparar 2ª chamada com outputs das funções
	params.PreviousResponseID = openai.String(rsp.ID)
	params.Tools = nil
	params.Input = responses.ResponseNewParamsInputUnion{} // limpa
	hasToolOutputs := false

	for _, out := range rsp.Output {
		if out.Type != "function_call" {
			continue
		}
		call := out.AsFunctionCall()
		logger.Log.Infof("Chamando função: %s (call_id=%s)", out.Name, call.CallID)

		result, err := HandlerToolsFunc(idCtxt, out)
		payload := result
		if err != nil {
			payload = fmt.Sprintf(`{"error": %q}`, err.Error())
		}

		params.Input.OfInputItemList = append(params.Input.OfInputItemList,
			responses.ResponseInputItemParamOfFunctionCallOutput(call.CallID, payload),
		)
		hasToolOutputs = true
	}

	// 2ª chamada: consolidar resposta final
	if hasToolOutputs {
		rsp, err = obj.client.Responses.New(ctx, params)
		if err != nil {
			logger.Log.Errorf("OpenAI Responses.New (passo consolidação) falhou: %v", err)
			return nil, nil, err
		}
	}

	usage := rsp.Usage
	if SessionServiceGlobal != nil {
		SessionServiceGlobal.UpdateTokensUso(usage.InputTokens, usage.OutputTokens, usage.TotalTokens)
	}

	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		rsp.Model, usage.InputTokens, usage.OutputTokens, usage.TotalTokens)

	return rsp, &usage, nil
}

/*
*
Função ainda em desenvolvimento para a busca de arquivos com auxílio da IA. A COMPLETAR
*/
func (obj *OpenaiServiceType) SubmitResponseFileSearch(storedFileID string) (*responses.Response, error) {
	ctx := context.Background()
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	params := responses.ResponseNewParams{
		Model: openai.ChatModel("gpt-4.1"),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: []responses.ResponseInputItemUnionParam{
				{
					OfMessage: &responses.EasyInputMessageParam{
						Type: "message",
						Role: "user",
						Content: responses.EasyInputMessageContentUnionParam{
							OfInputItemContentList: responses.ResponseInputMessageContentListParam{
								responses.ResponseInputContentUnionParam{
									OfInputFile: &responses.ResponseInputFileParam{
										Type:   "input_file",
										FileID: openai.String(storedFileID),
									},
								},
								responses.ResponseInputContentUnionParam{
									OfInputText: &responses.ResponseInputTextParam{
										Type: "input_text",
										Text: "Provide a one paragraph summary of the provided document.",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resp, err := obj.client.Responses.New(ctx, params)
	if err != nil {

		msg := fmt.Sprintf("Erro ao chamar a API OpenAI: %v", err)
		logger.Log.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	fmt.Printf("Resposta: %+v\n", resp.Output)

	return resp, nil
}

/*
Função destinada a calcular a quantidade de tokens constantes de um vetor de mensagtens
*/
//const OPENAI_TOKENS_AJUSTE = 7
const OPENAI_TOKENS_OVERHEAD_MSG = 3 // chutinho por mensagem

func (obj *OpenaiServiceType) TokensCounter(inputMsgs MsgGpt) (int, error) {
	msgs := inputMsgs.GetMessages()
	if len(msgs) == 0 {
		return 0, fmt.Errorf("lista de mensagens vazia")
	}

	enc, err := tokenizer.Get(tokenizer.Encoding(tokenizer.O200kBase))
	if err != nil {
		return 0, fmt.Errorf("falha ao obter tokenizer: %w", err)
	}

	total := 0
	for _, it := range msgs {
		ids, _, err := enc.Encode(it.Text)
		if err != nil {
			return 0, fmt.Errorf("falha ao codificar texto: %w", err)
		}
		total += len(ids) + OPENAI_TOKENS_OVERHEAD_MSG
	}

	return total + OPENAI_TOKENS_AJUSTE, nil
}

func (obj *OpenaiServiceType) StringTokensCounter(inputStr string) (int, error) {
	msg := MsgGpt{}
	msg.CreateMessage("", ROLE_USER, inputStr)
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

	vec64, _, err := OpenaiServiceGlobal.GetEmbeddingFromText(ctx, docText)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar embedding: %w", err)
	}
	vec32 := OpenaiServiceGlobal.Float64ToFloat32Slice(vec64)
	return vec32, nil
}

// helper: normaliza role conhecido
func normalizeRole(role string) responses.EasyInputMessageRole {
	switch role {
	case ROLE_USER:
		return responses.EasyInputMessageRoleUser
	case ROLE_ASSISTANT:
		return responses.EasyInputMessageRoleAssistant
	case ROLE_SYSTEM:
		return responses.EasyInputMessageRoleSystem
	case ROLE_DEVELOPER:
		return responses.EasyInputMessageRoleDeveloper
	default:
		return responses.EasyInputMessageRoleUser
	}
}

// Constrói EasyInputMessageParam com o item correto:
// - user/system/developer => InputText
// - assistant             => OutputText
func toEasyInputMessage(item MessageResponseItem) *responses.EasyInputMessageParam {
	role := normalizeRole(item.Role)

	// Input := responses.ResponseNewParamsInputUnion{OfString: openai.String(question)}
	// OfInputItemList:= ResponseInputParam{[]}

	if role == responses.EasyInputMessageRoleAssistant {
		// Mensagem prévia do assistente volta como OUTPUT_TEXT
		return &responses.EasyInputMessageParam{
			Type: "message",
			Role: role,
			Content: responses.EasyInputMessageContentUnionParam{
				OfInputItemContentList: responses.ResponseInputMessageContentListParam{
					responses.ResponseInputContentUnionParam{
						OfInputText: &responses.ResponseInputTextParam{
							Type: "output_text",
							Text: item.Text,
						},
					},
				},
			},
		}
	}

	// Demais roles entram como INPUT_TEXT
	return &responses.EasyInputMessageParam{
		Type: "message",
		Role: role,
		Content: responses.EasyInputMessageContentUnionParam{
			OfInputItemContentList: responses.ResponseInputMessageContentListParam{
				responses.ResponseInputContentUnionParam{
					OfInputText: &responses.ResponseInputTextParam{
						Type: "input_text",
						Text: item.Text,
					},
				},
			},
		},
	}
}
