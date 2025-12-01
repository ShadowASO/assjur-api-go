/*
---------------------------------------------------------------------------------------
File: openaiLib.go
Autor: Aldenor
Data: 15-08-2025
Finalidade: Funções que fazer o uso direto dos serviços da OpenAI e devem ser chamadas
indiretamente, por meio do pacote services(openaiServices)
---------------------------------------------------------------------------------------
*/

package ialib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"ocrserver/internal/config"
	"ocrserver/internal/services/tools"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared/constant"
	"github.com/tiktoken-go/tokenizer"
)

// **************** MENSAGENS - OpenAI ************************************
//
// Observação: a quantidade de tokens reportada pela OpenAI pode incluir
// um pequeno overhead. Esta constante ajusta a estimativa local.
const OPENAI_TOKENS_AJUSTE = 6
const OPENAI_TOKENS_OVERHEAD_MSG = 3 // overhead aproximado por mensagem

// Roles
const (
	ROLE_USER      = "user"
	ROLE_ASSISTANT = "assistant"
	ROLE_SYSTEM    = "system"
	ROLE_DEVELOPER = "developer"
)

// Reasoning
const (
	REASONING_MIN    = responses.ReasoningEffortMinimal
	REASONING_LOW    = responses.ReasoningEffortLow
	REASONING_MEDIUM = responses.ReasoningEffortMedium
	REASONING_HIGH   = responses.ReasoningEffortHigh
)

// Verbosity
const (
	VERBOSITY_LOW    = responses.ResponseTextConfigVerbosityLow
	VERBOSITY_MEDIUM = responses.ResponseTextConfigVerbosityMedium
	VERBOSITY_HIGH   = responses.ResponseTextConfigVerbosityHigh
)

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
func (m *MsgGpt) GetMessages() []MessageResponseItem { return m.Messages }

// ***********************************************************************

type OpenaiType struct {
	client openai.Client
	cfg    *config.Config
}

var (
	OpenaiGlobal   *OpenaiType
	onceInitOpenAI sync.Once
)

// InitOpenai inicializa o cliente global uma única vez.
func InitOpenai(apiKey string, cfg *config.Config) {
	onceInitOpenAI.Do(func() {
		c := openai.NewClient(option.WithAPIKey(apiKey)) // valor
		OpenaiGlobal = &OpenaiType{
			client: c, // pega endereço
			cfg:    cfg,
		}
		logger.Log.Info("Global OpenaiService configurado com sucesso.")
	})
}
func NewOpenaiClient(apiKey string, cfg *config.Config) *OpenaiType {
	c := openai.NewClient(option.WithAPIKey(apiKey)) // valor
	return &OpenaiType{
		client: c, // pega endereço
		cfg:    cfg,
	}
}

/*
GetEmbeddingFromText_openapi
Obtém a representação vetorial do texto. Caso precise de float32, use
Float64ToFloat32Slice() após o retorno.
*/
func (obj *OpenaiType) GetEmbeddingFromText_openai(
	ctx context.Context,
	inputTxt string,
) ([]float32, *openai.CreateEmbeddingResponse, error) {
	if obj == nil {
		return nil, nil, fmt.Errorf("serviço OpenAI não iniciado")
	}

	//Timeout defensivo se caller não definiu
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
	}

	var (
		resp *openai.CreateEmbeddingResponse
		err  error
	)

	// retry 3x em 429/5xx com backoff
	for attempt := 1; attempt <= 3; attempt++ {
		// se o contexto já foi cancelado, saia cedo
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
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
	vec32 := obj.Float64ToFloat32Slice(embedding)

	// (Opcional) apenas loga se vier dimensão inesperada
	if l := len(embedding); l != 3072 {
		logger.Log.Warningf("Dimensão do embedding inesperada: %d (esperado 3072 para text-embedding-3-large)", l)
	}

	usage := resp.Usage
	logger.Log.Infof("Modelo: %s - TOKENS Embeddings - Prompt: %d - Total: %d",
		resp.Model, usage.PromptTokens, usage.TotalTokens)

	return vec32, resp, nil
}

/*
SubmitPromptResponse_openapi
Envia mensagens (sem tools) e retorna a resposta do modelo.
`modelo` pode ser vazio para usar o definido em cfg.
*/
func (obj *OpenaiType) SubmitPromptResponse_openai(
	ctx context.Context,
	inputMsgs MsgGpt,
	prevID string,
	modelo string,
	effort responses.ReasoningEffort,
	verbosity responses.ResponseTextConfigVerbosity,

) (*responses.Response, error) {
	if obj == nil {
		return nil, fmt.Errorf("serviço OpenAI não iniciado")
	}

	if obj.cfg == nil {
		return nil, fmt.Errorf("configuração OpenAI ausente")
	}

	msgs := inputMsgs.GetMessages()
	if len(msgs) == 0 {
		return nil, fmt.Errorf("lista de mensagens vazia")
	}

	// Timeout defensivo: se o contexto não tiver deadline, aplica o da config
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		timeout := 180 // fallback padrão
		if config.GlobalConfig != nil && config.GlobalConfig.OpenOptionTimeoutSeconds > 0 {
			timeout = config.GlobalConfig.OpenOptionTimeoutSeconds
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	items := make([]responses.ResponseInputItemUnionParam, 0, len(msgs))
	for _, it := range msgs {
		items = append(items, responses.ResponseInputItemUnionParam{OfMessage: toEasyInputMessage(it)})
	}

	model := strings.TrimSpace(modelo)
	if model == "" {
		model = obj.cfg.OpenOptionModel
	}
	logger.Log.Infof("Modelo de IA: %s", modelo)

	params := responses.ResponseNewParams{
		Model:           model,
		Reasoning:       responses.ReasoningParam{Effort: effort},
		MaxOutputTokens: openai.Int(int64(config.GlobalConfig.OpenOptionMaxCompletionTokens)),
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: items},
		Text:            responses.ResponseTextConfigParam{Verbosity: verbosity},
	}
	if prevID != "" {
		params.PreviousResponseID = openai.String(prevID)
	}

	var resp *responses.Response
	var err error

	for attempt := 1; attempt <= 3; attempt++ {

		resp, err = obj.client.Responses.New(ctx, params)

		// ✅ Se não houve erro → sair do loop
		if err == nil {
			// Verificar truncamento por política
			if resp != nil && resp.IncompleteDetails.Reason == "content_filter" {
				logger.Log.Errorf("Resposta bloqueada por política de conteúdo")
				return nil, erros.CreateError("Resposta truncada pela política da OpenAI!")
			}
			break
		}

		// ⏳ Timeout → retry se ainda há tentativas
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Log.Errorf("Timeout (%d seg). Tentativa %d/3",
				config.GlobalConfig.OpenOptionTimeoutSeconds, attempt)
			if attempt < 3 {
				time.Sleep(erros.RetryBackoff(attempt))
				continue
			}
			return nil, fmt.Errorf("tempo limite excedido ao aguardar resposta da OpenAI")
		}

		// 🚦 Rate limit 429 ou erros 5xx → retry
		var apiErr *openai.Error
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 429 || apiErr.StatusCode >= 500) {
			if attempt < 3 {
				backoff := erros.RetryBackoff(attempt)
				logger.Log.Warningf("Erro API %d (%s). Retentando em %v...",
					apiErr.StatusCode, apiErr.Message, backoff)
				time.Sleep(backoff)
				continue
			}
		}
		break
	}

	if err != nil {
		logger.Log.Errorf("Falha final na chamada OpenAI: %v", err)
		return nil, err
	}

	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		resp.Model, resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.TotalTokens)
	return resp, nil
}

/*
TokensCounter
Calcula estimativa de tokens em um conjunto de mensagens.
*/
func (obj *OpenaiType) TokensCounter(inputMsgs MsgGpt) (int, error) {
	msgs := inputMsgs.GetMessages()
	if len(msgs) == 0 {
		return 0, fmt.Errorf("lista de mensagens vazia")
	}
	enc, err := tokenizer.Get(tokenizer.Encoding(tokenizer.O200kBase))
	if err != nil {
		return 0, fmt.Errorf("falha ao obter tokenizer: %w", err)
	}

	const roleOverhead = OPENAI_TOKENS_AJUSTE // heurística leve por mensagem
	total := 0
	for _, it := range msgs {
		ids, _, err := enc.Encode(strings.TrimSpace(it.Text))
		if err != nil {
			return 0, fmt.Errorf("falha ao codificar texto: %w", err)
		}
		total += len(ids) + roleOverhead
	}
	return total, nil
}
func (obj *OpenaiType) StringTokensCounter(inputStr string) (int, error) {
	msg := MsgGpt{}
	msg.CreateMessage("", ROLE_USER, inputStr)
	return obj.TokensCounter(msg)
}

func (obj *OpenaiType) Float64ToFloat32Slice(input []float64) []float32 {
	out := make([]float32, len(input))
	for i, v := range input {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			logger.Log.Warningf("Valor inválido no embedding (índice %d): %v. Substituindo por 0.", i, v)
			v = 0
		}
		out[i] = float32(v)
	}
	return out
}

// GetDocumentoEmbeddings gera embedding em float32 para um texto.
func GetDocumentoEmbeddings(docText string) ([]float32, error) {
	if OpenaiGlobal == nil {
		return nil, fmt.Errorf("OpenaiGlobal não inicializado")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vec32, _, err := OpenaiGlobal.GetEmbeddingFromText_openai(ctx, docText)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar embedding: %w", err)
	}
	return vec32, nil
}

func EnsureJSONPayload(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return `""`
	}
	// heurística simples: se começa com { ou [ assume JSON
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return s
	}
	// caso contrário, serializa como JSON string
	b, _ := json.Marshal(s) // nunca deve falhar
	return string(b)
}

// ------------------------- Helpers de Mensagens --------------------------

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

// toEasyInputMessage
// - user/system/developer => input_text
// - assistant             => output_text
//
// Observação: Alguns SDKs usam o mesmo struct para "input_text"/"output_text"
// (com um campo Type). Este código segue esse padrão.
func toEasyInputMessage(item MessageResponseItem) *responses.EasyInputMessageParam {
	role := normalizeRole(item.Role)

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

// FirstMessageFromSubmit retorna o primeiro Output "message" com texto.
func FirstMessageFromSubmit(retSubmit *responses.Response) (responses.ResponseOutputItemUnion, error) {
	if retSubmit == nil {
		return responses.ResponseOutputItemUnion{}, fmt.Errorf("resposta nula do provedor")
	}
	if len(retSubmit.Output) == 0 {
		return responses.ResponseOutputItemUnion{}, fmt.Errorf("resposta sem Output")
	}
	for _, out := range retSubmit.Output {
		if out.Type != "message" {
			continue
		}
		msg := out.AsMessage()
		for _, c := range msg.Content {
			if c.Type == "output_text" && strings.TrimSpace(c.Text) != "" {
				return out, nil
			}
		}
	}
	return responses.ResponseOutputItemUnion{}, fmt.Errorf("nenhum conteúdo textual encontrado na resposta (message/output_text)")
}

// ExtractOutputText retorna o primeiro "output_text" não vazio de um message.
func ExtractOutputText(msg responses.ResponseOutputItemUnion) (string, error) {
	if msg.Type != "message" {
		return "", fmt.Errorf("tipo não é message: %s", msg.Type)
	}
	m := msg.AsMessage()
	for _, c := range m.Content {
		if c.Type == "output_text" {
			if t := strings.TrimSpace(c.Text); t != "" {
				return t, nil
			}
		}
	}
	return "", fmt.Errorf("message sem output_text utilizável")
}

// *****************************  RAG  *************************************

/*
SubmitPromptTools_openapi (1ª chamada)
Prepara e envia a primeira rodada para o modelo decidir por function_call(s).
*/
func (obj *OpenaiType) SubmitPromptTools_openai(
	ctx context.Context,
	idCtxt string,
	inputMsgs MsgGpt,
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

	msgs := inputMsgs.GetMessages()
	if len(msgs) == 0 {
		return nil, fmt.Errorf("lista de mensagens vazia")
	}

	// Timeout defensivo
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
	}

	inputItems := make([]responses.ResponseInputItemUnionParam, 0, len(msgs))
	for _, msg := range msgs {
		inputItems = append(inputItems, responses.ResponseInputItemUnionParam{
			OfMessage: toEasyInputMessage(msg),
		})
	}

	var toolsCfg []responses.ToolUnionParam
	if toolManager != nil {
		toolsCfg = toolManager.GetAgentTools()
		if len(toolsCfg) == 0 {
			logger.Log.Warning("Tools vazia — o modelo poderá responder sem tools")
		}
	} else {
		logger.Log.Warning("toolManager nil — seguindo sem tools")
	}

	params := responses.ResponseNewParams{
		Model:           obj.cfg.OpenOptionModel,
		MaxOutputTokens: openai.Int(int64(config.GlobalConfig.OpenOptionMaxCompletionTokens)),
		Tools:           toolsCfg,
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems},
		Reasoning:       responses.ReasoningParam{Effort: effort},
		Text:            responses.ResponseTextConfigParam{Verbosity: verbosity},
	}
	if prevID != "" {
		params.PreviousResponseID = openai.String(prevID)
	}

	var (
		resp *responses.Response
		err  error
	)
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = obj.client.Responses.New(ctx, params)
		if err == nil {
			break
		}
		var apiErr *openai.Error
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 429 || apiErr.StatusCode >= 500) && attempt < 3 {
			time.Sleep(erros.RetryBackoff(attempt))
			continue
		}
		break
	}
	if err != nil {
		logger.Log.Errorf("OpenAI Responses.New (passo ferramentas) falhou: %v", err)
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("resposta nula do provedor na 1ª chamada")
	}

	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		resp.Model, resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.TotalTokens)

	return resp, nil
}

/*
ExtraiResponseTools_openapi (parsing/local-exec)
Lê as function_call(s), executa o handler fornecido e monta os inputs para a 2ª chamada.
*/
func (obj *OpenaiType) ExtraiResponseTools_openai(
	idCtxt string,
	rsp *responses.Response,
	handlerFunc func(idCtxt string, output responses.ResponseOutputItemUnion) (string, error),
) (responses.ResponseNewParams, bool, error) {
	if obj == nil {
		return responses.ResponseNewParams{}, false, fmt.Errorf("serviço OpenAI não iniciado")
	}
	if rsp == nil {
		return responses.ResponseNewParams{}, false, fmt.Errorf("resposta nula do provedor")
	}

	params := responses.ResponseNewParams{
		PreviousResponseID: openai.String(rsp.ID), // já devolva preenchido
	}
	hasToolOutputs := false

	for _, out := range rsp.Output {
		if out.Type != "function_call" {
			continue
		}
		call := out.AsFunctionCall()
		callID := call.CallID
		funcName := call.Name
		if funcName == "" {
			funcName = out.Name // fallback conforme SDK
		}
		logger.Log.Debugf("(%s) Chamando função: %s (call_id=%s)", idCtxt, funcName, callID)

		result, err := handlerFunc(idCtxt, out)
		payload := result
		if err != nil {
			payload = fmt.Sprintf(`{"error": %q}`, err.Error())
			logger.Log.Errorf("(%s) Erro na tool %s (call_id=%s): %v", idCtxt, funcName, callID, err)
		}

		params.Input.OfInputItemList = append(params.Input.OfInputItemList,
			responses.ResponseInputItemParamOfFunctionCallOutput(callID, payload),
		)
		hasToolOutputs = true
	}

	if !hasToolOutputs {
		logger.Log.Infof("(%s) Nenhuma function_call retornada; 2ª chamada seguirá sem tool outputs.", idCtxt)
	}
	return params, hasToolOutputs, nil
}

/*
SubmitResponseTools_openapi (2ª chamada)
Consolida a resposta final no modelo, alimentando os function_call_outputs gerados.
*/
func (obj *OpenaiType) SubmitResponseTools_openai(
	ctx context.Context,
	reqID string,
	params responses.ResponseNewParams,
	effort responses.ReasoningEffort,
	verbosity responses.ResponseTextConfigVerbosity,
) (*responses.Response, error) {
	if obj == nil {
		return nil, fmt.Errorf("serviço OpenAI não iniciado")
	}

	if obj.cfg == nil {
		return nil, fmt.Errorf("configuração OpenAI ausente")
	}
	if params.PreviousResponseID.Value == "" {
		return nil, fmt.Errorf("PreviousResponseID ausente para a 2ª chamada")
	}

	if len(params.Input.OfInputItemList) == 0 {
		logger.Log.Debug("nenhuma function_call retornada")
		return nil, fmt.Errorf("nenhuma function_call retornada; 2ª chamada não é necessária")
	}

	// Timeout defensivo
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
	}

	// Configura parâmetros da 2ª chamada
	params.Model = obj.cfg.OpenOptionModel
	params.PreviousResponseID = openai.String(reqID)
	params.Tools = nil
	params.MaxOutputTokens = openai.Int(int64(config.GlobalConfig.OpenOptionMaxCompletionTokens))
	params.Reasoning = responses.ReasoningParam{Effort: effort}
	params.Text = responses.ResponseTextConfigParam{Verbosity: verbosity}

	var (
		resp *responses.Response
		err  error
	)
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = obj.client.Responses.New(ctx, params)
		if err == nil {
			break
		}
		var apiErr *openai.Error
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 429 || apiErr.StatusCode >= 500) && attempt < 3 {
			time.Sleep(erros.RetryBackoff(attempt))
			continue
		}
		break
	}
	if err != nil {
		logger.Log.Errorf("OpenAI Responses.New (passo consolidação) falhou: %v", err)
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("resposta nula do provedor na 2ª chamada")
	}

	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		resp.Model, resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.TotalTokens)

	return resp, nil
}

/*
SubmitResponseFileSearch_openapi
Exemplo de uso com input_file + input_text.
*/
func (obj *OpenaiType) SubmitResponseFileSearch_openai(storedFileID string) (*responses.Response, error) {
	if obj == nil {
		return nil, fmt.Errorf("serviço OpenAI não iniciado")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	params := responses.ResponseNewParams{
		Model: obj.cfg.OpenOptionModel,
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
										Text: "Provide a one-paragraph summary of the provided document.",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	var (
		resp *responses.Response
		err  error
	)
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = obj.client.Responses.New(ctx, params)
		if err == nil {
			break
		}
		var apiErr *openai.Error
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 429 || apiErr.StatusCode >= 500) && attempt < 3 {
			time.Sleep(erros.RetryBackoff(attempt))
			continue
		}
		break
	}
	if err != nil {
		logger.Log.Errorf("Erro ao chamar a API OpenAI: %v", err)
		return nil, fmt.Errorf("erro ao chamar a API OpenAI: %w", err)
	}

	//if resp.Usage != nil {
	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		resp.Model, resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.TotalTokens)
	//}
	return resp, nil
}

// Função que retorna um ResponseOutputItemUnion válido para modifica-
// ção pelo usuário
func NewResponseOutputItemExample() responses.ResponseOutputItemUnion {
	id := uuid.New().String()

	// 1️⃣ Cria o conteúdo textual da resposta
	content := []responses.ResponseOutputMessageContentUnion{
		{
			Text: "mensagem default",
		},
	}

	// 2️⃣ Monta o item principal (union)
	item := responses.ResponseOutputItemUnion{
		ID:      id,
		Role:    constant.Assistant(openai.MessageRoleAssistant),
		Type:    "message", // ou outro tipo válido (ex: "reasoning", "function_call", etc.)
		Status:  "ok",
		Content: content,
	}

	return item
}
