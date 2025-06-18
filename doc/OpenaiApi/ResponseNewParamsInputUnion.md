INPUT:

type ResponseNewParamsInputUnion struct {
    OfString        param.Opt[string]  `json:",omitzero,inline"`
    OfInputItemList ResponseInputParam `json:",omitzero,inline"`
    paramUnion
}
Only one field can be non-zero.

Use [param.IsOmitted] to confirm if a field is set.

func (m param.metadata) ExtraFields() map[string]any
func (u responses.ResponseNewParamsInputUnion) MarshalJSON() ([]byte, error)
func (m param.metadata) Overrides() (any, bool)
func (m *param.metadata) SetExtraFields(extraFields map[string]any)
func (u *responses.ResponseNewParamsInputUnion) UnmarshalJSON(data []byte) error
responses.ResponseNewParamsInputUnion on pkg.go.dev
