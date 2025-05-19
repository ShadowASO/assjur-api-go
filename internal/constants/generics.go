package consts

type TokenData struct {
	PromptTokens     int64 `json:"PromptTokens"`
	CompletionTokens int64 `json:"CompletionTokens"`
	TotalTokens      int64 `json:"TotalTokens"`
}
