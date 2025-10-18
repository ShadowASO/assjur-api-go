package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// Tipo base do Handler interno (que recebe []byte vindo da API)
type ToolHandlerFunc func(ctx context.Context, args []byte) (string, error)

// ToolManager: Gerenciador de Tools
type ToolManager struct {
	Registry map[string]struct {
		Schema  responses.ToolUnionParam
		Handler ToolHandlerFunc
	}
}

// Criador de nova instância
func NewToolManager() *ToolManager {
	return &ToolManager{
		Registry: make(map[string]struct {
			Schema  responses.ToolUnionParam
			Handler ToolHandlerFunc
		}),
	}
}

// Função utilitária: Geração automática de JSON Schema via reflexão
func generateJSONSchemaFromStruct[T any]() map[string]any {
	t := reflect.TypeOf((*T)(nil)).Elem()
	properties := make(map[string]any)
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		fieldName := strings.Split(jsonTag, ",")[0]

		fieldSchema := map[string]any{}
		switch field.Type.Kind() {
		case reflect.String:
			fieldSchema["type"] = "string"
		case reflect.Int, reflect.Int64:
			fieldSchema["type"] = "integer"
		case reflect.Float64, reflect.Float32:
			fieldSchema["type"] = "number"
		case reflect.Bool:
			fieldSchema["type"] = "boolean"
		}

		descTag := field.Tag.Get("description")
		if descTag != "" {
			fieldSchema["description"] = descTag
		}

		properties[fieldName] = fieldSchema
		required = append(required, fieldName)
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// Função de registro genérico de Tool (externa à ToolManager por limitação do Go)
func RegisterGenericTool[T any](tm *ToolManager, name, description string, handler func(ctx context.Context, args T) (string, error)) {
	schema := generateJSONSchemaFromStruct[T]()

	tm.Registry[name] = struct {
		Schema  responses.ToolUnionParam
		Handler ToolHandlerFunc
	}{
		Schema: responses.ToolUnionParam{
			OfFunction: &responses.FunctionToolParam{
				Name:        name,
				Description: openai.String(description),
				Parameters:  schema,
			},
		},
		Handler: func(ctx context.Context, rawArgs []byte) (string, error) {
			var parsedArgs T
			if err := json.Unmarshal(rawArgs, &parsedArgs); err != nil {
				return "", fmt.Errorf("erro ao fazer unmarshal dos argumentos: %w", err)
			}
			return handler(ctx, parsedArgs)
		},
	}
}

// Processador central de chamadas OpenAI -> Tools
func (tm *ToolManager) ProcessToolCall(ctx context.Context, toolCall responses.ResponseFunctionToolCall) (string, error) {
	tool, ok := tm.Registry[toolCall.Name]
	if !ok {
		return "", fmt.Errorf("função desconhecida: %s", toolCall.Name)
	}
	return tool.Handler(ctx, []byte(toolCall.Arguments))

}

// Exportador para o OpenAI
func (tm *ToolManager) GetAgentTools() []responses.ToolUnionParam {
	tools := make([]responses.ToolUnionParam, 0, len(tm.Registry))
	for _, v := range tm.Registry {
		tools = append(tools, v.Schema)
	}
	return tools
}
