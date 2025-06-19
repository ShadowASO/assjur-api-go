
// Funções a serem chamadas pelo Modelo RAG
type GetNomeClienteArgs struct {
	Premiado string `json:"premiado" description:"Número do bilhete premiado"`
}

func GetNomeCliente(ctx context.Context, args GetNomeClienteArgs) (string, error) {
	logger.Log.Info("entrei")
	logger.Log.Info(args.Premiado)
	return "Aldenor Sombra de Oliveira", nil
}

### Cria o ToolManager
tools := rag.NewToolManager()

### Faz o registro da função
rag.RegisterGenericTool(tools, "premiado", "retorna nome do cliente pelo biblete premiado", rag.GetNomeCliente)

### Retorna todas as tools registradas. Server para atribuir a Params.Tools
Tools: toolManager.GetAgentTools()

### Executa a tools
result, err := toolManager.ProcessToolCall(ctx, toolCall)


