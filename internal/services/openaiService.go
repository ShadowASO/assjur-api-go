// package openAI
package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	"ocrserver/internal/config"

	"ocrserver/internal/utils/logger"

	"github.com/openai/openai-go" // imported as openai
	"github.com/openai/openai-go/option"
)

// **************** MENSAGENS - OpenAI   **********************************
// Roles
// type RoleType = 'developer' | 'user' | 'assistant';
const ROLE_DEVELOPER = "developer"
const ROLE_USER = "user"
const ROLE_ASSISTANT = "assistant"

type MessageOpenai struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MsgGpt struct {
	Messages []MessageOpenai `json:"messages"`
}

func (m *MsgGpt) AddMessage(message MessageOpenai) {
	m.Messages = append(m.Messages, message)
}
func (m *MsgGpt) CreateMessage(role string, message string) {
	m.Messages = append(m.Messages, MessageOpenai{Role: "user", Content: message})
}

func (m *MsgGpt) GetMessages() []MessageOpenai {
	return m.Messages
}

// ***********************************************************************
// type OpenaiServiceType struct{}
type OpenaiServiceType struct {
	client *openai.Client
	cfg    *config.Config
}

var OpenaiServiceGlobal *OpenaiServiceType
var onceInitOpenAIService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitOpenaiService(apiKey string, cfg *config.Config) {
	onceInitOpenAIService.Do(func() {

		OpenaiServiceGlobal = &OpenaiServiceType{
			client: openai.NewClient(option.WithAPIKey(apiKey)),
			cfg:    cfg,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewOpenaiClient(apiKey string, cfg *config.Config) *OpenaiServiceType {
	return &OpenaiServiceType{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		cfg:    cfg,
	}
}

//var Service = OpenAIClient{}

/*
*
Consulta com o envio do contexto na forma de mensagens
*/
func (obj *OpenaiServiceType) SubmitPrompt(messages MsgGpt) (*openai.ChatCompletion, error) {

	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	var msg []openai.ChatCompletionMessageParamUnion
	client := openai.NewClient(
		option.WithAPIKey(obj.cfg.OpenApiKey), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	ctx := context.Background()

	for _, m := range messages.Messages {
		if m.Role == string(openai.ChatCompletionUserMessageParamRoleUser) {
			msg = append(msg, openai.UserMessage(m.Content))
		}
	}

	completion, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: openai.F(msg),
			Seed:     openai.Int(1),

			Model:               openai.F(obj.cfg.OpenOptionModel),
			Temperature:         openai.Float(0),
			MaxCompletionTokens: openai.Int(int64(obj.cfg.OpenOptionMaxCompletionTokens)),
			FrequencyPenalty:    openai.Float(0),
			PresencePenalty:     openai.Float(0),
		})

	if err != nil {
		panic(err)
	}
	/* Atualiza o uso de tokens na tabela 'sessions' */
	//server.UpdateTokensUso(completion)
	SessionServiceGlobal.UpdateTokensUso(completion)

	log.Printf("Uso da API OpenAI - TOKENS - Prompt: %d - Completion: %d - Total: %d",
		completion.Usage.PromptTokens,
		completion.Usage.CompletionTokens,
		completion.Usage.TotalTokens)

	return completion, err
}

/*
*
Obtém a representação vetorial do texto enviado
*/
func (obj *OpenaiServiceType) GetEmbeddingFromText(text string) (*openai.CreateEmbeddingResponse, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	client := openai.NewClient(
		option.WithAPIKey(obj.cfg.OpenApiKey),
	)
	ctx := context.Background()

	// Chamada à API de embeddings
	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model:          openai.F(openai.EmbeddingModelTextEmbedding3Large), // ou use config.OpenEmbeddingModel
		Input:          openai.F(openai.EmbeddingNewParamsInputUnion(openai.EmbeddingNewParamsInputArrayOfStrings{text})),
		EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormat("float")),
	})
	if err != nil {
		return nil, err
	}

	// Verifica retorno
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("nenhum embedding retornado")
	}

	// Registro de uso (tokens)
	log.Printf("Uso da API OpenAI (embeddings) - TOKENS - Prompt: %d - Total: %d",
		resp.Usage.PromptTokens,
		resp.Usage.TotalTokens)

	return resp, nil
}

// Converte o slice do embedding de []float64 para []float32, formato reconhecido pelo OpenSearch
func Float64ToFloat32Slice(input []float64) []float32 {
	output := make([]float32, len(input))
	for i, val := range input {
		output[i] = float32(val)
	}
	return output
}
