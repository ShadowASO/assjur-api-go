

type ResponseInputItemUnionParam struct {
    OfMessage              *EasyInputMessageParam                      `json:",omitzero,inline"`
    OfInputMessage         *ResponseInputItemMessageParam              `json:",omitzero,inline"`
    OfOutputMessage        *ResponseOutputMessageParam                 `json:",omitzero,inline"`
    OfFileSearchCall       *ResponseFileSearchToolCallParam            `json:",omitzero,inline"`
    OfComputerCall         *ResponseComputerToolCallParam              `json:",omitzero,inline"`
    OfComputerCallOutput   *ResponseInputItemComputerCallOutputParam   `json:",omitzero,inline"`
    OfWebSearchCall        *ResponseFunctionWebSearchParam             `json:",omitzero,inline"`
    OfFunctionCall         *ResponseFunctionToolCallParam              `json:",omitzero,inline"`
    OfFunctionCallOutput   *ResponseInputItemFunctionCallOutputParam   `json:",omitzero,inline"`
    OfReasoning            *ResponseReasoningItemParam                 `json:",omitzero,inline"`
    OfImageGenerationCall  *ResponseInputItemImageGenerationCallParam  `json:",omitzero,inline"`
    OfCodeInterpreterCall  *ResponseCodeInterpreterToolCallParam       `json:",omitzero,inline"`
    OfLocalShellCall       *ResponseInputItemLocalShellCallParam       `json:",omitzero,inline"`
    OfLocalShellCallOutput *ResponseInputItemLocalShellCallOutputParam `json:",omitzero,inline"`
    OfMcpListTools         *ResponseInputItemMcpListToolsParam         `json:",omitzero,inline"`
    OfMcpApprovalRequest   *ResponseInputItemMcpApprovalRequestParam   `json:",omitzero,inline"`
    OfMcpApprovalResponse  *ResponseInputItemMcpApprovalResponseParam  `json:",omitzero,inline"`
    OfMcpCall              *ResponseInputItemMcpCallParam              `json:",omitzero,inline"`
    OfItemReference        *ResponseInputItemItemReferenceParam        `json:",omitzero,inline"`
    paramUnion
}
Only one field can be non-zero.

Use [param.IsOmitted] to confirm if a field is set.

func (m param.metadata) ExtraFields() map[string]any
func (u responses.ResponseInputItemUnionParam) GetAcknowledgedSafetyChecks() []responses.ResponseInputItemComputerCallOutputAcknowledgedSafetyCheckParam
func (u responses.ResponseInputItemUnionParam) GetAction() (res responses.responseInputItemUnionParamAction)
func (u responses.ResponseInputItemUnionParam) GetApprovalRequestID() *string
func (u responses.ResponseInputItemUnionParam) GetApprove() *bool
func (u responses.ResponseInputItemUnionParam) GetArguments() *string
func (u responses.ResponseInputItemUnionParam) GetCallID() *string
func (u responses.ResponseInputItemUnionParam) GetCode() *string
func (u responses.ResponseInputItemUnionParam) GetContainerID() *string
func (u responses.ResponseInputItemUnionParam) GetContent() (res responses.responseInputItemUnionParamContent)
func (u responses.ResponseInputItemUnionParam) GetEncryptedContent() *string
func (u responses.ResponseInputItemUnionParam) GetError() *string
func (u responses.ResponseInputItemUnionParam) GetID() *string
func (u responses.ResponseInputItemUnionParam) GetName() *string
func (u responses.ResponseInputItemUnionParam) GetOutput() (res responses.responseInputItemUnionParamOutput)
func (u responses.ResponseInputItemUnionParam) GetPendingSafetyChecks() []responses.ResponseComputerToolCallPendingSafetyCheckParam
func (u responses.ResponseInputItemUnionParam) GetQueries() []string
func (u responses.ResponseInputItemUnionParam) GetReason() *string
func (u responses.ResponseInputItemUnionParam) GetResult() *string
func (u responses.ResponseInputItemUnionParam) GetResults() (res responses.responseInputItemUnionParamResults)
func (u responses.ResponseInputItemUnionParam) GetRole() *string
func (u responses.ResponseInputItemUnionParam) GetServerLabel() *string
func (u responses.ResponseInputItemUnionParam) GetStatus() *string
func (u responses.ResponseInputItemUnionParam) GetSummary() []responses.ResponseReasoningItemSummaryParam
func (u responses.ResponseInputItemUnionParam) GetTools() []responses.ResponseInputItemMcpListToolsToolParam
func (u responses.ResponseInputItemUnionParam) GetType() *string
func (u responses.ResponseInputItemUnionParam) MarshalJSON() ([]byte, error)
func (m param.metadata) Overrides() (any, bool)
func (m *param.metadata) SetExtraFields(extraFields map[string]any)
func (u *responses.ResponseInputItemUnionParam) UnmarshalJSON(data []byte) error
responses.ResponseInputItemUnionParam on pkg.go.dev


