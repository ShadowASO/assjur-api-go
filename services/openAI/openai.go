package openAI

import (
	"context"
	"github.com/openai/openai-go" // imported as openai
	"github.com/openai/openai-go/option"
	"log"
	"ocrserver/config"
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
			Model:            openai.F(config.OpenOptionModel),
			Temperature:      openai.Float(0),
			MaxTokens:        openai.Int(int64(config.OpenOptionMaxTokens)),
			FrequencyPenalty: openai.Float(0),
			PresencePenalty:  openai.Float(0),
		})

	if err != nil {
		panic(err)
	}
	log.Printf("Modelo: %v", completion.Model)
	//return completion.Choices[0].Message.Content, err
	return completion, err
}
