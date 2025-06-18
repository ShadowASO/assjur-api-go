type ResponseNewParams struct {
    // Text, image, or file inputs to the model, used to generate a response.
    //
    // Learn more:
    //
    // - [Text inputs and outputs](https://platform.openai.com/docs/guides/text)
    // - [Image inputs](https://platform.openai.com/docs/guides/images)
    // - [File inputs](https://platform.openai.com/docs/guides/pdf-files)
    // - [Conversation state](https://platform.openai.com/docs/guides/conversation-state)
    // - [Function calling](https://platform.openai.com/docs/guides/function-calling)
    Input ResponseNewParamsInputUnion `json:"input,omitzero,required"`
    // Model ID used to generate the response, like `gpt-4o` or `o3`. OpenAI offers a
    // wide range of models with different capabilities, performance characteristics,
    // and price points. Refer to the
    // [model guide](https://platform.openai.com/docs/models) to browse and compare
    // available models.
    Model shared.ResponsesModel `json:"model,omitzero,required"`
    // Whether to run the model response in the background.
    // [Learn more](https://platform.openai.com/docs/guides/background).
    Background param.Opt[bool] `json:"background,omitzero"`
    // Inserts a system (or developer) message as the first item in the model's
    // context.
    //
    // When using along with `previous_response_id`, the instructions from a previous
    // response will not be carried over to the next response. This makes it simple to
    // swap out system (or developer) messages in new responses.
    Instructions param.Opt[string] `json:"instructions,omitzero"`
    // An upper bound for the number of tokens that can be generated for a response,
    // including visible output tokens and
    // [reasoning tokens](https://platform.openai.com/docs/guides/reasoning).
    MaxOutputTokens param.Opt[int64] `json:"max_output_tokens,omitzero"`
    // Whether to allow the model to run tool calls in parallel.
    ParallelToolCalls param.Opt[bool] `json:"parallel_tool_calls,omitzero"`
    // The unique ID of the previous response to the model. Use this to create
    // multi-turn conversations. Learn more about
    // [conversation state](https://platform.openai.com/docs/guides/conversation-state).
    PreviousResponseID param.Opt[string] `json:"previous_response_id,omitzero"`
    // Whether to store the generated model response for later retrieval via API.
    Store param.Opt[bool] `json:"store,omitzero"`
    // What sampling temperature to use, between 0 and 2. Higher values like 0.8 will
    // make the output more random, while lower values like 0.2 will make it more
    // focused and deterministic. We generally recommend altering this or `top_p` but
    // not both.
    Temperature param.Opt[float64] `json:"temperature,omitzero"`
    // An alternative to sampling with temperature, called nucleus sampling, where the
    // model considers the results of the tokens with top_p probability mass. So 0.1
    // means only the tokens comprising the top 10% probability mass are considered.
    //
    // We generally recommend altering this or `temperature` but not both.
    TopP param.Opt[float64] `json:"top_p,omitzero"`
    // A stable identifier for your end-users. Used to boost cache hit rates by better
    // bucketing similar requests and to help OpenAI detect and prevent abuse.
    // [Learn more](https://platform.openai.com/docs/guides/safety-best-practices#end-user-ids).
    User param.Opt[string] `json:"user,omitzero"`
    // Specify additional output data to include in the model response. Currently
    // supported values are:
    //
    //   - `file_search_call.results`: Include the search results of the file search tool
    //     call.
    //   - `message.input_image.image_url`: Include image urls from the input message.
    //   - `computer_call_output.output.image_url`: Include image urls from the computer
    //     call output.
    //   - `reasoning.encrypted_content`: Includes an encrypted version of reasoning
    //     tokens in reasoning item outputs. This enables reasoning items to be used in
    //     multi-turn conversations when using the Responses API statelessly (like when
    //     the `store` parameter is set to `false`, or when an organization is enrolled
    //     in the zero data retention program).
    //   - `code_interpreter_call.outputs`: Includes the outputs of python code execution
    //     in code interpreter tool call items.
    Include []ResponseIncludable `json:"include,omitzero"`
    // Set of 16 key-value pairs that can be attached to an object. This can be useful
    // for storing additional information about the object in a structured format, and
    // querying for objects via API or the dashboard.
    //
    // Keys are strings with a maximum length of 64 characters. Values are strings with
    // a maximum length of 512 characters.
    Metadata shared.Metadata `json:"metadata,omitzero"`
    // Specifies the latency tier to use for processing the request. This parameter is
    // relevant for customers subscribed to the scale tier service:
    //
    //   - If set to 'auto', and the Project is Scale tier enabled, the system will
    //     utilize scale tier credits until they are exhausted.
    //   - If set to 'auto', and the Project is not Scale tier enabled, the request will
    //     be processed using the default service tier with a lower uptime SLA and no
    //     latency guarantee.
    //   - If set to 'default', the request will be processed using the default service
    //     tier with a lower uptime SLA and no latency guarantee.
    //   - If set to 'flex', the request will be processed with the Flex Processing
    //     service tier.
    //     [Learn more](https://platform.openai.com/docs/guides/flex-processing).
    //   - When not set, the default behavior is 'auto'.
    //
    // When this parameter is set, the response body will include the `service_tier`
    // utilized.
    //
    // Any of "auto", "default", "flex".
    ServiceTier ResponseNewParamsServiceTier `json:"service_tier,omitzero"`
    // The truncation strategy to use for the model response.
    //
    //   - `auto`: If the context of this response and previous ones exceeds the model's
    //     context window size, the model will truncate the response to fit the context
    //     window by dropping input items in the middle of the conversation.
    //   - `disabled` (default): If a model response will exceed the context window size
    //     for a model, the request will fail with a 400 error.
    //
    // Any of "auto", "disabled".
    Truncation ResponseNewParamsTruncation `json:"truncation,omitzero"`
    // **o-series models only**
    //
    // Configuration options for
    // [reasoning models](https://platform.openai.com/docs/guides/reasoning).
    Reasoning shared.ReasoningParam `json:"reasoning,omitzero"`
    // Configuration options for a text response from the model. Can be plain text or
    // structured JSON data. Learn more:
    //
    // - [Text inputs and outputs](https://platform.openai.com/docs/guides/text)
    // - [Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs)
    Text ResponseTextConfigParam `json:"text,omitzero"`
    // How the model should select which tool (or tools) to use when generating a
    // response. See the `tools` parameter to see how to specify which tools the model
    // can call.
    ToolChoice ResponseNewParamsToolChoiceUnion `json:"tool_choice,omitzero"`
    // An array of tools the model may call while generating a response. You can
    // specify which tool to use by setting the `tool_choice` parameter.
    //
    // The two categories of tools you can provide the model are:
    //
    //   - **Built-in tools**: Tools that are provided by OpenAI that extend the model's
    //     capabilities, like
    //     [web search](https://platform.openai.com/docs/guides/tools-web-search) or
    //     [file search](https://platform.openai.com/docs/guides/tools-file-search).
    //     Learn more about
    //     [built-in tools](https://platform.openai.com/docs/guides/tools).
    //   - **Function calls (custom tools)**: Functions that are defined by you, enabling
    //     the model to call your own code. Learn more about
    //     [function calling](https://platform.openai.com/docs/guides/function-calling).
    Tools []ToolUnionParam `json:"tools,omitzero"`
    paramObj
}
func (m param.metadata) ExtraFields() map[string]any
func (r responses.ResponseNewParams) MarshalJSON() (data []byte, err error)
func (m param.metadata) Overrides() (any, bool)
func (m *param.metadata) SetExtraFields(extraFields map[string]any)
func (r *responses.ResponseNewParams) UnmarshalJSON(data []byte) error




