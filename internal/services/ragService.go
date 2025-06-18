package services

import (
	"context"
	"encoding/json"
	"fmt"

	"ocrserver/internal/utils/logger"

	"strings"

	"github.com/openai/openai-go"

	"github.com/openai/openai-go/responses"
)

// Funções a serem chamadas pelo Modelo RAG
type GetNomeClienteArgs struct {
	Premiado string `json:"premiado"`
}

func getNomeCliente(ctx context.Context, args []byte) (string, error) {
	var getArgs GetNomeClienteArgs

	if err := json.Unmarshal(args, &getArgs); err != nil {
		return "", fmt.Errorf("failed to parse get_stock_price arguments: %w", err)
	}

	// Validate the stock symbol
	if strings.TrimSpace(getArgs.Premiado) == "" {
		return "", fmt.Errorf("stock symbol is required")
	}

	return "Aldenor Sombra de Oliveira", nil
}

// DESCRIÇÃO
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

// Relação de ferramentas a serem chamadas pelo modelo
var AgentTools = []responses.ToolUnionParam{
	getNomeClienteTool,
}

// Função utilitária para a chamada das funções from the OpenAI API
func processToolCall(ctx context.Context, toolCall responses.ResponseFunctionToolCall) (string, error) {
	switch toolCall.Name {
	case "getNomeCliente":
		return getNomeCliente(ctx, []byte(toolCall.Arguments))
	default:
		logger.Log.Error("função desconhecida: ", toolCall.Name)
		return "", fmt.Errorf("função desconhecida")
	}
}
