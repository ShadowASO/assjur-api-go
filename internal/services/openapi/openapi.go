package openapi

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

type OpenaiType struct {
	client openai.Client
	cfg    *config.Config
}

var OpenaiGlobal *OpenaiType
var onceInitOpenAI sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitOpenai(apiKey string, cfg *config.Config) {
	onceInitOpenAI.Do(func() {

		OpenaiGlobal = &OpenaiType{
			client: openai.NewClient(
				option.WithAPIKey(apiKey),
			),
			cfg: cfg,
		}

		logger.Log.Info("Global OpenaiService configurado com sucesso.")
	})
}

func NewOpenaiClient(apiKey string, cfg *config.Config) *OpenaiType {
	return &OpenaiType{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		cfg:    cfg,
	}
}

/*
*
Obtém a representação vetorial do texto enviado. Quem for utilizar o valor retornadotem
que saber que se precisar converter para float32, deverá fazê-lo onde necessário.
*/
func (obj *OpenaiType) GetEmbeddingFromText_openapi(
	ctx context.Context,
	inputTxt string,
) ([]float64, *openai.CreateEmbeddingResponse, error) {
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

	// if nrTokens, _ := OpenaiGlobal.StringTokensCounter(inputTxt); nrTokens > 0 {
	// 	logger.Log.Infof("Estimativa de tokens no texto: %d", nrTokens)
	// }

	return embedding, resp, nil
}

/*
modelo: nome do modelo a usar, ou uma string vazia("")
*/
func (obj *OpenaiType) SubmitPromptResponse_openapi(
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
	msgs := inputMsgs.GetMessages()
	if len(msgs) == 0 {
		return nil, fmt.Errorf("lista de mensagens vazia")
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

	resp, err := obj.client.Responses.New(ctx, params)
	if err != nil {
		logger.Log.Errorf("OpenAI Responses.New falhou: %v", err)
		return nil, err
	}
	usage := resp.Usage
	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		resp.Model, usage.InputTokens, usage.OutputTokens, usage.TotalTokens)

	return resp, nil
}

func (obj *OpenaiType) SubmitResponseFunctionRAG_openapi(
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

	msgs := inputMsgs.GetMessages()
	if len(msgs) == 0 {
		return nil, fmt.Errorf("lista de mensagens vazia")
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
	resp, err := obj.client.Responses.New(ctx, params)
	if err != nil {
		logger.Log.Errorf("OpenAI Responses.New (passo ferramentas) falhou: %v", err)
		return nil, err
	}

	// Preparar 2ª chamada com outputs das funções
	params.PreviousResponseID = openai.String(resp.ID)
	params.Tools = nil
	params.Input = responses.ResponseNewParamsInputUnion{} // limpa
	hasToolOutputs := false

	// 2ª chamada: consolidar resposta final
	if hasToolOutputs {
		resp, err = obj.client.Responses.New(ctx, params)
		if err != nil {
			logger.Log.Errorf("OpenAI Responses.New (passo consolidação) falhou: %v", err)
			return nil, err
		}
	}

	usage := resp.Usage
	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		resp.Model, usage.InputTokens, usage.OutputTokens, usage.TotalTokens)

	return resp, nil
}

/*
*
Função ainda em desenvolvimento para a busca de arquivos com auxílio da IA. A COMPLETAR
*/
func (obj *OpenaiType) SubmitResponseFileSearch_openapi(storedFileID string) (*responses.Response, error) {
	ctx := context.Background()
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	params := responses.ResponseNewParams{
		Model: openai.ChatModel("gpt-5-mini"),
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
	usage := resp.Usage
	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		resp.Model, usage.InputTokens, usage.OutputTokens, usage.TotalTokens)

	return resp, nil
}

/*
Função destinada a calcular a quantidade de tokens constantes de um vetor de mensagtens
*/
//const OPENAI_TOKENS_AJUSTE = 7
const OPENAI_TOKENS_OVERHEAD_MSG = 3 // chutinho por mensagem

func (obj *OpenaiType) TokensCounter(inputMsgs MsgGpt) (int, error) {
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

func (obj *OpenaiType) StringTokensCounter(inputStr string) (int, error) {
	msg := MsgGpt{}
	msg.CreateMessage("", ROLE_USER, inputStr)
	return obj.TokensCounter(msg)
}

func (obj *OpenaiType) Float64ToFloat32Slice(input []float64) []float32 {
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

	vec64, _, err := OpenaiGlobal.GetEmbeddingFromText_openapi(ctx, docText)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar embedding: %w", err)
	}
	vec32 := OpenaiGlobal.Float64ToFloat32Slice(vec64)
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
func (obj *OpenaiType) SubmitPromptTools_openapi(
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
	//Extrai as mensagens passadas ao modelo
	msgs := inputMsgs.GetMessages()
	if len(msgs) == 0 {
		return nil, fmt.Errorf("lista de mensagens vazia")
	}

	// Monta input inicial com as mensagens
	inputItems := make([]responses.ResponseInputItemUnionParam, 0, len(msgs))
	for _, msg := range msgs {
		inputItems = append(inputItems, responses.ResponseInputItemUnionParam{
			OfMessage: toEasyInputMessage(msg),
		})
	}

	// Obtém tools (se houver gerenciador)
	toolsCfg := toolManager.GetAgentTools()
	if toolsCfg == nil {
		toolsCfg = []responses.ToolUnionParam{}
		logger.Log.Error("Tools está vazia")
	}

	params := responses.ResponseNewParams{
		Model:           obj.cfg.OpenOptionModel,
		MaxOutputTokens: openai.Int(int64(config.GlobalConfig.OpenOptionMaxCompletionTokens)),
		Tools:           toolsCfg, // ok ser nil/empty: o modelo pode responder sem tools
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: inputItems},
		Reasoning:       responses.ReasoningParam{Effort: effort},
		Text:            responses.ResponseTextConfigParam{Verbosity: verbosity},
	}
	if prevID != "" {
		params.PreviousResponseID = openai.String(prevID)
	}

	// 1ª chamada — o modelo decide se chama ferramentas
	resp, err := obj.client.Responses.New(ctx, params)
	if err != nil {
		logger.Log.Errorf("OpenAI Responses.New (passo ferramentas) falhou: %v", err)
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("resposta nula do provedor na 1ª chamada")
	}
	usage := resp.Usage
	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		resp.Model, usage.InputTokens, usage.OutputTokens, usage.TotalTokens)

	// devolve os parâmetros prontos para a 2ª chamada e a usage da 1ª
	return resp, nil
}

/*
2ª CHAMADA
Extrai as funções chamadas pelo modelo e executa as funções reais, utilizando o handlerFunc.
Cada resposta é adicionada a "params" e será devolvida para ser envido ao modelo pel função
"SubmitResponseTools". Esta função executa apenas as funções apoontadas pelo modelo e monta
params com as respostas.
*/
func (obj *OpenaiType) ExtraiResponseTools_openapi(
	idCtxt string,
	rsp *responses.Response,
	handlerFunc func(idCtxt string, output responses.ResponseOutputItemUnion) (string, error),
) (responses.ResponseNewParams, bool, error) {
	if obj == nil {
		return responses.ResponseNewParams{}, false, fmt.Errorf("serviço OpenAI não iniciado")
	}

	// Prepara params para a 2ª chamada (consolidação)
	params := responses.ResponseNewParams{}

	// Coleta function_call(s) e prepara os outputs para o modelo
	hasToolOutputs := false
	for _, out := range rsp.Output {
		if out.Type != "function_call" {
			continue
		}

		call := out.AsFunctionCall()
		callID := call.CallID
		funcName := call.Name
		if funcName == "" {
			funcName = out.Name // fallback (depende do SDK)
		}
		logger.Log.Debugf("(%s) Chamando função: %s (call_id=%s)", idCtxt, funcName, callID)

		result, err := handlerFunc(idCtxt, out)
		payload := result
		if err != nil {
			// Nunca deixa de montar o payload para o modelo entender a falha da tool.
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

	// devolve os parâmetros prontos para a 2ª chamada e a usage da 1ª
	return params, hasToolOutputs, nil

}

// ExtraiResponseTools executa a 2ª chamada (consolidação da resposta final)
// usando os params retornados por SubmitRequestTools.
func (obj *OpenaiType) SubmitResponseTools_openapi(
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

	// Sanidade: ou continuamos a partir de uma resposta anterior, ou enviamos algum input.
	// Isso evita 2ª chamada completamente vazia em caso de integração incorreta.
	if (params.PreviousResponseID.Value == "") &&
		len(params.Input.OfInputItemList) == 0 {
		return nil, fmt.Errorf("parâmetros inválidos para 2ª chamada: sem PreviousResponseID e sem Input")
	}

	params.Model = obj.cfg.OpenOptionModel
	params.PreviousResponseID = openai.String(reqID)
	params.Tools = nil
	params.MaxOutputTokens = openai.Int(int64(config.GlobalConfig.OpenOptionMaxCompletionTokens))
	params.Reasoning = responses.ReasoningParam{Effort: effort}
	params.Text = responses.ResponseTextConfigParam{Verbosity: verbosity}

	resp, err := obj.client.Responses.New(ctx, params)
	if err != nil {
		logger.Log.Errorf("OpenAI Responses.New (passo consolidação) falhou: %v", err)
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("resposta nula do provedor na 2ª chamada")
	}
	usage := resp.Usage
	logger.Log.Infof("Modelo: %s - TOKENS - Input: %d - Output: %d - Total: %d",
		resp.Model, usage.InputTokens, usage.OutputTokens, usage.TotalTokens)

	return resp, nil
}
