package openAI

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/openai-go" // imported as openai
	"github.com/openai/openai-go/option"

	//"log"
	"ocrserver/internal/config"
)

// type PromptType string
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

type OpenAIClient struct{}

var Service = OpenAIClient{}

func (c *OpenAIClient) SubmitPrompt(messages MsgGpt) (*openai.ChatCompletion, error) {
	var msg []openai.ChatCompletionMessageParamUnion

	client := openai.NewClient(
		option.WithAPIKey(config.OpenApiKey), // defaults to os.LookupEnv("OPENAI_API_KEY")
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

			//Model:       openai.F(openai.ChatModelGPT4oMini),
			//Temperature: openai.Float(0),
			//MaxTokens:        openai.Int(int64(config.OpenOptionMaxTokens)),

			Model:               openai.F(config.OpenOptionModel),
			MaxCompletionTokens: openai.Int(int64(config.OpenOptionMaxCompletionTokens)),
			FrequencyPenalty:    openai.Float(0),
			PresencePenalty:     openai.Float(0),
		})

	if err != nil {
		panic(err)
	}
	//log.Printf("Modelo: %v", completion.Model)
	//return completion.Choices[0].Message.Content, err
	/**
	Insiro um registro do log para cada consulta à API da OpenAI
	*/
	log.Printf("Uso da API OpenAI - TOKENS - Prompt: %d - Completion: %d - Total: %d",
		completion.Usage.PromptTokens,
		completion.Usage.CompletionTokens,
		completion.Usage.TotalTokens)

	return completion, err
}
func (c *OpenAIClient) GetEmbeddingFromText(text string) (*openai.CreateEmbeddingResponse, error) {
	client := openai.NewClient(
		option.WithAPIKey(config.OpenApiKey),
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

	// Retorna o vetor de floats
	//return resp.Data[0].Embedding, nil

	return resp, nil
}

func Float64ToFloat32Slice(input []float64) []float32 {
	output := make([]float32, len(input))
	for i, val := range input {
		output[i] = float32(val)
	}
	return output
}
