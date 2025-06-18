package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ocrserver/internal/config"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
)

func testeInputFile(storedFileID string, client *openai.Client) {
	ctx := context.Background()

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
}

// GetStockPriceArgs represents the arguments for the get_stock_price function
type GetStockPriceArgs struct {
	Premiado string `json:"premiado"`
}

// GetStockPrice is the implementation of the get_stock_price function
func GetStockPrice(ctx context.Context, args []byte) (string, error) {
	var getArgs GetStockPriceArgs
	if err := json.Unmarshal(args, &getArgs); err != nil {
		return "", fmt.Errorf("failed to parse get_stock_price arguments: %w", err)
	}

	// Validate the stock symbol
	if strings.TrimSpace(getArgs.Premiado) == "" {
		return "", fmt.Errorf("stock symbol is required")
	}

	// Return a static placeholder
	return "$198.53 USD", nil
}

// GetStockPriceArgs represents the arguments for the get_stock_price function

// getStockTool defines the OpenAI tool for getting a single Stock by ID
var getStockTool = responses.ToolUnionParam{
	OfFunction: &responses.FunctionToolParam{
		Name:        "get_stock_price",
		Description: openai.String("The get_stock_price tool retrieves the current price of a single stock by it's ticker symbol"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol": map[string]string{
					"type":        "string",
					"description": "The ticker symbol of the stock to retrieve",
				},
			},
			"required": []string{"symbol"},
		},
	},
}

var getNomeClienteTool = responses.ToolUnionParam{
	OfFunction: &responses.FunctionToolParam{
		Name:        "getNomeCliente",
		Description: openai.String("a função getNomeCliente retorna a qualificação do cliente pelo biblete premiado"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"premiado": map[string]string{
					"type":        "string",
					"description": "O número do bilhete premiado",
				},
			},
			"required": []string{"premiado"},
		},
	},
}

func getNomeCliente(ctx context.Context, args []byte) (string, error) {
	var getArgs GetStockPriceArgs
	log.Printf("Entrei")
	if err := json.Unmarshal(args, &getArgs); err != nil {
		return "", fmt.Errorf("failed to parse get_stock_price arguments: %w", err)
	}

	// Validate the stock symbol
	if strings.TrimSpace(getArgs.Premiado) == "" {
		return "", fmt.Errorf("stock symbol is required")
	}
	log.Printf("Passei")

	return "Aldenor Sombra de Oliveira", nil
}

// agentTools is the list of all tools available to the agent
var agentTools = []responses.ToolUnionParam{
	getNomeClienteTool,
}

// processToolCall handles a tool call from the OpenAI API
func processToolCall(ctx context.Context, toolCall responses.ResponseFunctionToolCall) (string, error) {
	switch toolCall.Name {
	case "getNomeCliente":
		return getNomeCliente(ctx, []byte(toolCall.Arguments))
	default:
		return "", fmt.Errorf("Função desconhecida: %s", toolCall.Name)
	}
}

// func TesteFunctionCall(storedFileID string, client *openai.Client, agentTools []responses.ToolUnionParam) {
func TesteFunctionCall() {
	ctx := context.Background()

	params := responses.ResponseNewParams{
		Model:           openai.ChatModelGPT4o,
		Temperature:     openai.Float(0.7),
		MaxOutputTokens: openai.Int(512),
		Tools:           agentTools,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String("Qual o nome do ganhador com o bilhete premiado número 5?"),
		},
	}
	client := openai.NewClient(
		option.WithAPIKey(config.GlobalConfig.OpenApiKey),
	)

	resp, err := client.Responses.New(ctx, params)
	if err != nil {
		log.Fatalln(err.Error())
	}
	params.PreviousResponseID = openai.String(resp.ID)

	params.Input = responses.ResponseNewParamsInputUnion{}

	for _, output := range resp.Output {
		if output.Type == "function_call" {
			toolCall := output.AsFunctionCall()

			result, err := processToolCall(ctx, toolCall)
			if err != nil {
				params.Input.OfInputItemList = append(params.Input.OfInputItemList, responses.ResponseInputItemParamOfFunctionCallOutput(toolCall.CallID, err.Error()))
			} else {

				params.Input.OfInputItemList = append(params.Input.OfInputItemList, responses.ResponseInputItemParamOfFunctionCallOutput(toolCall.CallID, result))
			}
		}
	}
	// No tools calls made, we already have our final response
	if len(params.Input.OfInputItemList) == 0 {
		log.Println(resp.OutputText())
		return
	}

	// Make a final call with our tools results and no tools to get the final output
	params.Tools = nil
	resp, err = client.Responses.New(ctx, params)
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println(resp.OutputText())

}

