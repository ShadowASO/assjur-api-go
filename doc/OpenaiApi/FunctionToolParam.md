type FunctionToolParam struct {
    // Whether to enforce strict parameter validation. Default `true`.
    Strict param.Opt[bool] `json:"strict,omitzero,required"`
    // A JSON schema object describing the parameters of the function.
    Parameters map[string]any `json:"parameters,omitzero,required"`
    // The name of the function to call.
    Name string `json:"name,required"`
    // A description of the function. Used by the model to determine whether or not to
    // call the function.
    Description param.Opt[string] `json:"description,omitzero"`
    // The type of the function tool. Always `function`.
    //
    // This field can be elided, and will marshal its zero value as "function".
    Type constant.Function `json:"type,required"`
    paramObj
}
Defines a function in your own code the model can choose to call. Learn more about [function calling](https://platform.openai.com/docs/guides/function-calling).

The properties Name, Parameters, Strict, Type are required.

func (m param.metadata) ExtraFields() map[string]any
func (r responses.FunctionToolParam) MarshalJSON() (data []byte, err error)
func (m param.metadata) Overrides() (any, bool)
func (m *param.metadata) SetExtraFields(extraFields map[string]any)
func (r *responses.FunctionToolParam) UnmarshalJSON(data []byte) error
responses.FunctionToolParam on pkg.go.dev


type ToolUnionParam struct {
    OfFunction           *FunctionToolParam        `json:",omitzero,inline"`
    OfFileSearch         *FileSearchToolParam      `json:",omitzero,inline"`
    OfWebSearchPreview   *WebSearchToolParam       `json:",omitzero,inline"`
    OfComputerUsePreview *ComputerToolParam        `json:",omitzero,inline"`
    OfMcp                *ToolMcpParam             `json:",omitzero,inline"`
    OfCodeInterpreter    *ToolCodeInterpreterParam `json:",omitzero,inline"`
    OfImageGeneration    *ToolImageGenerationParam `json:",omitzero,inline"`
    OfLocalShell         *ToolLocalShellParam      `json:",omitzero,inline"`
    paramUnion
}
Only one field can be non-zero.

Use [param.IsOmitted] to confirm if a field is set.

func (m param.metadata) ExtraFields() map[string]any
func (u responses.ToolUnionParam) GetAllowedTools() *responses.ToolMcpAllowedToolsUnionParam
func (u responses.ToolUnionParam) GetBackground() *string
func (u responses.ToolUnionParam) GetContainer() *responses.ToolCodeInterpreterContainerUnionParam
func (u responses.ToolUnionParam) GetDescription() *string
func (u responses.ToolUnionParam) GetDisplayHeight() *int64
func (u responses.ToolUnionParam) GetDisplayWidth() *int64
func (u responses.ToolUnionParam) GetEnvironment() *string
func (u responses.ToolUnionParam) GetFilters() *responses.FileSearchToolFiltersUnionParam
func (u responses.ToolUnionParam) GetHeaders() map[string]string
func (u responses.ToolUnionParam) GetInputImageMask() *responses.ToolImageGenerationInputImageMaskParam
func (u responses.ToolUnionParam) GetMaxNumResults() *int64
func (u responses.ToolUnionParam) GetModel() *string
func (u responses.ToolUnionParam) GetModeration() *string
func (u responses.ToolUnionParam) GetName() *string
func (u responses.ToolUnionParam) GetOutputCompression() *int64
func (u responses.ToolUnionParam) GetOutputFormat() *string
func (u responses.ToolUnionParam) GetParameters() map[string]any
func (u responses.ToolUnionParam) GetPartialImages() *int64
func (u responses.ToolUnionParam) GetQuality() *string
func (u responses.ToolUnionParam) GetRankingOptions() *responses.FileSearchToolRankingOptionsParam
func (u responses.ToolUnionParam) GetRequireApproval() *responses.ToolMcpRequireApprovalUnionParam
func (u responses.ToolUnionParam) GetSearchContextSize() *string
func (u responses.ToolUnionParam) GetServerLabel() *string
func (u responses.ToolUnionParam) GetServerURL() *string
func (u responses.ToolUnionParam) GetSize() *string
func (u responses.ToolUnionParam) GetStrict() *bool
func (u responses.ToolUnionParam) GetType() *string
func (u responses.ToolUnionParam) GetUserLocation() *responses.WebSearchToolUserLocationParam
func (u responses.ToolUnionParam) GetVectorStoreIDs() []string
func (u responses.ToolUnionParam) MarshalJSON() ([]byte, error)
func (m param.metadata) Overrides() (any, bool)
func (m *param.metadata) SetExtraFields(extraFields map[string]any)
func (u *responses.ToolUnionParam) UnmarshalJSON(data []byte) error
responses.ToolUnionParam on pkg.go.dev



