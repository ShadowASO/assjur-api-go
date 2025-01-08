package consts

// Estrutura de resposta para erro
type ResponseStatus struct {
	Ok         bool   `json:"ok"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}
type TokenData struct {
	PromptTokens     int64 `json:"PromptTokens"`
	CompletionTokens int64 `json:"CompletionTokens"`
	TotalTokens      int64 `json:"TotalTokens"`
}
