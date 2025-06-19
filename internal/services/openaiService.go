// package openAI
package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	"ocrserver/internal/config"

	"ocrserver/internal/services/tools"
	"ocrserver/internal/utils/logger"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

// **************** MENSAGENS - OpenAI   **********************************
// Roles

const ROLE_DEVELOPER = "developer"
const ROLE_USER = "user"
const ROLE_ASSISTANT = "assistant"

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
	m.Messages = append(m.Messages, MessageResponseItem{Id: id, Role: "user", Text: message})
}

func (m *MsgGpt) GetMessages() []MessageResponseItem {
	return m.Messages
}

// ***********************************************************************
// type OpenaiServiceType struct{}
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

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewOpenaiClient(apiKey string, cfg *config.Config) *OpenaiServiceType {
	return &OpenaiServiceType{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		cfg:    cfg,
	}
}

/*
*
Obtém a representação vetorial do texto enviado
*/

func (obj *OpenaiServiceType) GetEmbeddingFromText(text string) ([]float32, error) {
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
		Model:          (openai.EmbeddingModelTextEmbedding3Large), // ou use config.OpenEmbeddingModel
		Input:          openai.EmbeddingNewParamsInputUnion{OfString: param.NewOpt(text)},
		EncodingFormat: (openai.EmbeddingNewParamsEncodingFormat("float")),
	})
	if err != nil {
		return nil, err
	}

	//**** Converte para vetor de 32bits
	vetorEmbedding := float64ToFloat32Slice(resp.Data[0].Embedding)
	if len(vetorEmbedding) != 3072 {
		log.Fatalf("Erro: vetor de entrada tem dimensão %d, mas o índice espera 3072.", len(vetorEmbedding))
	}

	usage := resp.Usage
	// Registro de uso (tokens)
	log.Printf("Modelo - %s :Uso da API OpenAI (embeddings) - TOKENS - Prompt: %d - Total: %d",
		resp.Model,
		int64(usage.PromptTokens),
		int64(usage.TotalTokens))

	/* Atualiza o uso de tokens na tabela 'sessions' */
	SessionServiceGlobal.UpdateTokensUso(usage.PromptTokens, usage.TotalTokens-usage.PromptTokens, usage.TotalTokens)
	//return resp, nil
	return vetorEmbedding, nil
}

// Converte o slice do embedding de []float64 para []float32, formato reconhecido pelo OpenSearch
func float64ToFloat32Slice(input []float64) []float32 {
	output := make([]float32, len(input))
	for i, val := range input {
		output[i] = float32(val)
	}
	return output
}

func (obj *OpenaiServiceType) SubmitPromptResponse(inputMsgs MsgGpt, prevID *string) (*responses.Response, error) {
	ctx := context.Background()
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	inpuItemList := []responses.ResponseInputItemUnionParam{}

	msgs := inputMsgs.GetMessages()
	for _, item := range msgs {
		msg := &responses.EasyInputMessageParam{
			Type: "message",
			Role: responses.EasyInputMessageRole(item.Role),
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

		inpuItemList = append(inpuItemList, responses.ResponseInputItemUnionParam{
			OfMessage: msg,
		})

	}

	params := responses.ResponseNewParams{
		Model:           obj.cfg.OpenOptionModel,
		Temperature:     openai.Float(0.2),
		MaxOutputTokens: openai.Int(int64(config.GlobalConfig.OpenOptionMaxCompletionTokens)),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: inpuItemList,
		},
	}

	if prevID != nil && *prevID != "" {
		params.PreviousResponseID = openai.String(*prevID)
	}
	//Faz a chamda à API
	rsp, err := obj.client.Responses.New(ctx, params)
	if err != nil {
		logger.Log.Error("Erro ao realizar uma chamada à API da OpenAI")
		return nil, err
	}

	/* Atualiza o uso de tokens na tabela 'sessions' */

	SessionServiceGlobal.UpdateTokensUso(rsp.Usage.InputTokens, rsp.Usage.OutputTokens, rsp.Usage.InputTokens+rsp.Usage.OutputTokens)

	log.Printf("Modelo - %s: Uso da API OpenAI - TOKENS - Prompt: %d - Completion: %d - Total: %d",
		rsp.Model,
		rsp.Usage.InputTokens,
		rsp.Usage.OutputTokens,
		rsp.Usage.InputTokens+rsp.Usage.TotalTokens)

	return rsp, err
}

func (obj *OpenaiServiceType) SubmitResponseFunctionRAG(inputMsg string, toolManager *tools.ToolManager, prevID string) (*responses.Response, error) {
	ctx := context.Background()
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}

	params := responses.ResponseNewParams{
		Model:           obj.cfg.OpenOptionModel,
		Temperature:     openai.Float(0.2),
		MaxOutputTokens: openai.Int(int64(config.GlobalConfig.OpenOptionMaxCompletionTokens)),

		Tools: toolManager.GetAgentTools(),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(inputMsg),
		},
	}

	if prevID != "" {
		params.PreviousResponseID = openai.String(prevID)
	}
	//Faz a chamda à API
	rsp, err := obj.client.Responses.New(ctx, params)
	if err != nil {
		logger.Log.Error("Erro ao realizar uma chamada à API da OpenAI")
		return nil, err
	}
	//Pego o ID da resposta anterior
	params.PreviousResponseID = openai.String(rsp.ID)

	//Crio um novo params.Input
	params.Input = responses.ResponseNewParamsInputUnion{}

	//Faço a chamada de todas as funções escolhidas pelo modelo
	for _, output := range rsp.Output {
		if output.Type == "function_call" {
			//Extraio as funções escolhidas pelo modelo
			toolCall := output.AsFunctionCall()
			//Função utilitária que efetivamente chama as funções
			//result, err := processToolCall(ctx, toolCall)
			result, err := toolManager.ProcessToolCall(ctx, toolCall)
			if err != nil {
				params.Input.OfInputItemList = append(params.Input.OfInputItemList, responses.ResponseInputItemParamOfFunctionCallOutput(toolCall.CallID, err.Error()))
			} else {

				params.Input.OfInputItemList = append(params.Input.OfInputItemList, responses.ResponseInputItemParamOfFunctionCallOutput(toolCall.CallID, result))
			}
		}
	}
	//Se não houve nenhuma chama de função, já temos nossa resposta final
	if len(params.Input.OfInputItemList) == 0 {
		log.Println(rsp.OutputText())
		/* Atualiza o uso de tokens na tabela 'sessions' */
		SessionServiceGlobal.UpdateTokensUso(rsp.Usage.InputTokens, rsp.Usage.OutputTokens, rsp.Usage.InputTokens+rsp.Usage.OutputTokens)
		return rsp, nil
	}

	// Limpa as ferramentas e faz uma nova chamada com o resultado das funções para obter o resultado final
	params.Tools = nil
	//Chama a API
	rsp, err = obj.client.Responses.New(ctx, params)
	if err != nil {
		logger.Log.Error("Erro ao realizar uma chamada à API da OpenAI")
	}

	/* Atualiza o uso de tokens na tabela 'sessions' */
	SessionServiceGlobal.UpdateTokensUso(rsp.Usage.InputTokens, rsp.Usage.OutputTokens, rsp.Usage.InputTokens+rsp.Usage.OutputTokens)

	log.Printf("Modelo - %s: Uso da API OpenAI - TOKENS - Prompt: %d - Completion: %d - Total: %d",
		rsp.Model,
		rsp.Usage.InputTokens,
		rsp.Usage.OutputTokens,
		rsp.Usage.InputTokens+rsp.Usage.TotalTokens)

	return rsp, nil
}

/*
*
Função ainda em desenvolvimento para a busca de arquivos com auxílio da IA. A COMPLETAR
*/
func (obj *OpenaiServiceType) SubmitResponseFileSearch(storedFileID string, client *openai.Client) (*responses.Response, error) {
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

	resp, err := client.Responses.New(ctx, params)
	if err != nil {
		log.Fatalf("Erro ao chamar a API OpenAI: %v", err)
	}

	fmt.Printf("Resposta: %+v\n", resp.Output)

	return resp, nil
}
