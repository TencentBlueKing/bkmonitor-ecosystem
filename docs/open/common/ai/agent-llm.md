# Agent 与 LLM 通用接入及核心字段

## 0x00 Agent 与 LLM 通用接入

Agent、模型调用和工具执行分别使用独立的 OpenTelemetry Span，上报到蓝鲸 Agent 与 LLM 观测。`gen_ai.operation.name` 标识步骤类型，Span 父子关系串联执行过程。

### 接入准备

应用创建、OTLP 地址、Token 和 Resource 配置可参考 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/python/otlp/README.md" target="_blank">Python 接入文档</a> 或 <a href="https://github.com/TencentBlueKing/bkmonitor-ecosystem/blob/master/docs/open/java/otlp/README.md" target="_blank">Java 接入文档</a>。完成基础上报后，在对应 Span 上设置下文的 GenAI 属性。

同一会话使用框架或应用已有的 `gen_ai.conversation.id`，没有会话 ID 时省略。Agent ID 标识稳定的 Agent 配置，与会话 ID 分开填写。

### Span 示例

以“计算 1 + 2”为例，Agent 调用模型选择 `add` 工具，工具返回结果后，模型生成最终回答。以下保留三类 Span 的核心字段，省略时间戳、Resource 和首轮模型 Span。

```text
invoke_agent Calculator
├── chat gpt-4o          选择 add 工具（示例省略）
├── execute_tool add     执行加法
└── chat gpt-4o          接收工具结果，生成回答
```

**Agent：记录本轮请求与最终回答**

Agent 根 Span 的输入和输出来自应用入口与出口，不以任意子模型的消息代替。

```json
{
  "span_id": "1111111111111111",
  "parent_span_id": "",
  "span_name": "invoke_agent Calculator",
  "attributes": {
    "gen_ai.operation.name": "invoke_agent",
    "gen_ai.conversation.id": "conv-1",
    "gen_ai.agent.id": "calculator",
    "gen_ai.agent.name": "Calculator",
    "gen_ai.input.messages": [
      {"role": "user", "parts": [{"type": "text", "content": "计算 1 + 2"}]}
    ],
    "gen_ai.output.messages": [
      {"role": "assistant", "parts": [{"type": "text", "content": "1 + 2 = 3"}]}
    ]
  }
}
```

参考：<a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/docs/gen-ai/gen-ai-agent-spans.md" target="_blank">GenAI agent spans</a>。

**工具：记录一次执行的参数与结果**

`gen_ai.tool.call.id` 使用模型返回的调用 ID，将工具执行与消息中的工具请求、返回关联起来。

```json
{
  "span_id": "2222222222222222",
  "parent_span_id": "1111111111111111",
  "span_name": "execute_tool add",
  "attributes": {
    "gen_ai.operation.name": "execute_tool",
    "gen_ai.conversation.id": "conv-1",
    "gen_ai.tool.name": "add",
    "gen_ai.tool.type": "function",
    "gen_ai.tool.call.id": "call-1",
    "gen_ai.tool.call.arguments": "{\"a\":1,\"b\":2}",
    "gen_ai.tool.call.result": "{\"sum\":3}"
  }
}
```

参考：<a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/docs/gen-ai/gen-ai-spans.md" target="_blank">GenAI inference and tool spans</a>。

**模型：保留消息顺序与本轮 Token 用量**

此 Span 对应工具执行后的模型调用，输入包含用户问题、首轮工具请求和工具结果。示例假定提供商返回 `stop` 和 Token 用量，实际采集时按原响应填写。

```json
{
  "span_id": "3333333333333333",
  "parent_span_id": "1111111111111111",
  "span_name": "chat gpt-4o",
  "attributes": {
    "gen_ai.operation.name": "chat",
    "gen_ai.conversation.id": "conv-1",
    "gen_ai.provider.name": "openai",
    "gen_ai.request.model": "gpt-4o",
    "gen_ai.input.messages": [
      {"role": "user", "parts": [{"type": "text", "content": "计算 1 + 2"}]},
      {
        "role": "assistant",
        "parts": [
          {"type": "tool_call", "id": "call-1", "name": "add", "arguments": {"a": 1, "b": 2}}
        ]
      },
      {
        "role": "tool",
        "parts": [
          {"type": "tool_call_response", "id": "call-1", "response": {"sum": 3}}
        ]
      }
    ],
    "gen_ai.output.messages": [
      {
        "role": "assistant",
        "parts": [{"type": "text", "content": "1 + 2 = 3"}],
        "finish_reason": "stop"
      }
    ],
    "gen_ai.usage.input_tokens": 100,
    "gen_ai.usage.output_tokens": 20
  }
}
```

参考：<a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/docs/gen-ai/gen-ai-spans.md" target="_blank">GenAI inference spans</a>。

- *[1] 三类 Span 属于同一条 Trace，共用同一个 TraceID，示例中的 ID 和业务内容上报时需替换为实际采集值。*
- *[2] 消息、系统指令和工具定义为便于阅读展开为 JSON 数组，工具 Span 的参数与结果展示为 JSON 字符串。*
- *[3] SDK 不支持结构化属性时，将整个数组或对象序列化为 JSON 字符串后设置到对应属性。*

### 核心字段速查

按步骤填写实际采集到的字段。下表以 JSON 值展示字段结构，类型、枚举和消息 Part 的详细规则见 `0x01 标准字段参考`。

**步骤与会话**

| 字段 | 说明 | 示例 |
| --- | --- | --- |
| `gen_ai.operation.name` | 步骤类型，如 Agent 调用、模型调用、工具执行。 | `"chat"` |
| `gen_ai.conversation.id` | 框架或应用已有的真实会话标识。 | `"conv-1"` |
| `gen_ai.provider.name` | 模型服务商。 | `"openai"` |

**Agent 身份**

| 字段 | 说明 | 示例 |
| --- | --- | --- |
| `gen_ai.agent.id` | Agent 稳定标识。 | `"agent-1"` |
| `gen_ai.agent.name` | Agent 展示名称。 | `"Assistant"` |
| `gen_ai.agent.description` | Agent 用途说明。 | `"问答助手"` |
| `gen_ai.agent.version` | Agent 版本。 | `"1.0.0"` |

**消息与系统指令**

| 字段 | 说明 | 示例 |
| --- | --- | --- |
| `gen_ai.input.messages` | 按顺序记录输入消息，保留 role 和 parts 结构。 | `[{"role":"user","parts":[{"type":"text","content":"你好"}]}]` |
| `gen_ai.output.messages` | 输出消息，停止原因有原值时填写。 | `[{"role":"assistant","parts":[{"type":"text","content":"好"}]}]` |
| `gen_ai.system_instructions` | 与聊天历史分开传递的系统指令。 | `[{"type":"text","content":"你是助手"}]` |

**模型请求与响应**

| 字段 | 说明 | 示例 |
| --- | --- | --- |
| `gen_ai.request.model` | 请求指定的模型名称。 | `"gpt-4o"` |
| `gen_ai.request.temperature` | 请求的温度参数。 | `0.7` |
| `gen_ai.request.stream` | 是否为流式请求。 | `true` |
| `gen_ai.request.reasoning.level` | 请求指定的推理级别。 | `"high"` |
| `gen_ai.response.id` | 模型响应标识。 | `"resp-1"` |
| `gen_ai.response.model` | 实际响应的模型名称。 | `"gpt-4o"` |
| `gen_ai.response.status` | 提供商明确返回的响应生命周期状态。 | `"completed"` |
| `gen_ai.response.finish_reasons` | 提供商明确返回的停止原因列表。 | `["stop"]` |
| `gen_ai.response.time_to_first_chunk` | 流式请求收到首个 chunk 的延迟，单位为秒。 | `0.35` |

**工具调用**

| 字段 | 说明 | 示例 |
| --- | --- | --- |
| `gen_ai.tool.name` | 工具步骤执行的工具名称。 | `"add"` |
| `gen_ai.tool.description` | 工具用途说明。 | `"两数相加"` |
| `gen_ai.tool.type` | 工具类型，如 function、extension、datastore。 | `"function"` |
| `gen_ai.tool.call.id` | 关联模型工具调用消息与工具执行步骤的 ID。 | `"call-1"` |
| `gen_ai.tool.call.arguments` | 工具调用参数。 | `{"a":1,"b":2}` |
| `gen_ai.tool.call.result` | 工具调用结果。 | `{"sum":3}` |
| `gen_ai.tool.definitions` | 可用工具定义，按需记录。 | `[{"type":"function","name":"add","description":"加法","parameters":{"type":"object"}}]` |

**Token 用量**

| 字段 | 说明 | 示例 |
| --- | --- | --- |
| `gen_ai.usage.input_tokens` | 输入 Token 总数，包含缓存读取和写入。 | `100` |
| `gen_ai.usage.output_tokens` | 输出 Token 总数，包含推理 Token。 | `20` |
| `gen_ai.usage.cache_read.input_tokens` | 输入 Token 中的缓存读取部分。 | `30` |
| `gen_ai.usage.cache_write.input_tokens` | 输入 Token 中的缓存写入部分。 | `10` |
| `gen_ai.usage.reasoning.output_tokens` | 输出 Token 中的推理部分。 | `5` |

**用户与错误**

| 字段 | 说明 | 示例 |
| --- | --- | --- |
| `user.id` | 用户标识。 | `"user-1"` |
| `user.name` | 用户名称。 | `"demo"` |
| `user.hash` | 用户哈希标识。 | `"a1b2c3"` |
| `error.type` | 失败时记录的低基数错误类型。 | `"timeout"` |

- *[1] 缓存读取和写入包含在输入 Token 总数中，推理 Token 包含在输出 Token 总数中，统计总量时不重复相加。*
- *[2] 响应状态、停止原因和 Token 用量仅在原数据提供时填写，不从 Span 状态推导，也不用示例值补齐缺失数据。*

## 0x01 标准字段参考

> 核心字段以 <a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/docs/gen-ai/gen-ai-agent-spans.md" target="_blank">Semantic Conventions for GenAI agent and framework spans</a> 为准；该规范继承的推理与工具字段以其引用的 <a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/docs/gen-ai/gen-ai-spans.md" target="_blank">GenAI spans</a> 为准。

### a. 通用字段

**稳定性**

| 属性 | 类型 | 稳定性 | 示例值 | 描述 |
|------|------|--------|--------|------|
| `error.type` | `string` | ![Stable](https://img.shields.io/badge/-stable-lightgreen) | `"timeout"` / `"rate_limit"` | 操作结束时的低基数错误类型 [1] |
| `user.id` | `string` | ![Stable](https://img.shields.io/badge/-stable-lightgreen) | `"user-12345"` | 用户标识 |
| `user.name` | `string` | ![Stable](https://img.shields.io/badge/-stable-lightgreen) | `"zhang.san"` | 用户登录名 |
| `user.hash` | `string` | ![Stable](https://img.shields.io/badge/-stable-lightgreen) | `"a1b2c3d4"` | 用户哈希标识 |

**[1] `error.type`:** 应与 GenAI 提供商或客户端库返回的错误代码、发生的异常的规范名称或另一个低基数错误标识符匹配。插桩库应记录其报告的错误列表。禁止填高基数值（如完整错误消息）。

`error.type` 取值（开放枚举，无固定集合）：

| 值 | 说明 |
|---|---|
| `"timeout"` | 超时类错误 |
| `"rate_limit"` | 限流类错误 |
| 异常规范名称（如 `"java.net.UnknownHostException"`） | 客户端库抛出的异常类名 |
| 提供商错误代码（如 `"server_certificate_invalid"` / `"500"`） | 提供商返回的低基数错误码 |

### b. GenAI 操作

| 属性 | 类型 | 稳定性 | 示例值 | 描述 |
|------|------|--------|--------|------|
| `gen_ai.operation.name`  | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"chat"` / `"execute_tool"` / `"invoke_agent"` | 操作名称 [1]（开放枚举，取值见注解） |
| `gen_ai.provider.name` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"openai"` / `"anthropic"` | 客户端或服务端插桩识别到的 GenAI 提供商 [2] |
| `gen_ai.conversation.id` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"conv-abc123"` | 会话、Session 或 Thread 的唯一标识 [3] |

**[1] `gen_ai.operation.name`:** 如果预定义值之一适用，但特定系统使用不同的名称，建议在特定 GenAI 系统的语义约定中记录它，并在插桩中使用系统特定的名称。如果未记录不同的名称，插桩库应使用适用的预定义值。

`gen_ai.operation.name` 取值（开放枚举，允许产品自定义扩展值）：

| 值 | 说明 |
|---|---|
| `chat` | 对话补全（LLM 对话调用） |
| `generate_content` | 内容生成 |
| `text_completion` | 文本补全 |
| `embeddings` | 向量化 |
| `retrieval` | 检索 |
| `fetch_response` | 拉取响应 |
| `create_agent` | 创建 Agent |
| `invoke_agent` | 调用 Agent |
| `execute_tool` | 执行工具 |
| `invoke_workflow` | 调用工作流 |
| `plan` | 规划 |
| `create_memory` | 创建记忆记录 |
| `create_memory_store` | 创建或初始化记忆存储 |
| `delete_memory` | 删除记忆记录 |
| `delete_memory_store` | 删除记忆存储 |
| `search_memory` | 查询记忆存储 |
| `update_memory` | 更新记忆记录 |
| `upsert_memory` | 创建或更新记忆记录 |
| 产品自定义值 | 需在特定系统的语义约定中记录；适配层无法归一时保留原值，前端步骤分类降级为 `other` |

适配层到步骤类型 `kind` 的映射：`chat` / `generate_content` / `text_completion` / `fetch_response` / `embeddings` → `llm`；`execute_tool` → `tool`；`create_agent` / `invoke_agent` / `plan` → `agent`；`invoke_workflow` → `workflow`；`retrieval` → `retriever`；无法识别的开放枚举值 → `other`。这里的 `other` 只是前端展示分类，不会覆盖 `gen_ai.operation.name` 的原始值。

**[2] `gen_ai.provider.name`:** 应基于插桩的最佳知识设置，可能与实际上游提供商不同。例如，客户端 SDK 可能配置为代理或托管平台，透明地将请求转发到不同的提供商。`gen_ai.provider.name` 属性充当鉴别器，标识 GenAI 语义约定中特定于该提供商的遥测格式风格。它应与提供商特定的属性和信号保持一致设置。

`gen_ai.provider.name` 取值（开放枚举，GenAI SemConv well-known 值）：

| 值 | 说明 |
|---|---|
| `openai` | OpenAI |
| `anthropic` | Anthropic |
| `azure.ai.inference` | Azure AI Inference |
| `azure.ai.openai` | Azure OpenAI |
| `aws.bedrock` | AWS Bedrock |
| `gcp.gen_ai` | 无法确定具体后端时的 Google GenAI |
| `gcp.vertex_ai` | GCP Vertex AI |
| `gcp.gemini` | GCP Gemini |
| `cohere` | Cohere |
| `groq` | Groq |
| `ibm.watsonx.ai` | IBM watsonx.ai |
| `mistral_ai` | Mistral AI |
| `moonshot_ai` | Moonshot AI |
| `perplexity` | Perplexity |
| `x_ai` | xAI |
| `deepseek` | DeepSeek |
| 其他提供商/平台自定义值 | 允许扩展；适配层仅对能识别的值归一小写，未识别值透传 |

**[3] `gen_ai.conversation.id`:** 当插桩库有现成的标识符，或用户应用程序通过 OpenTelemetry 上下文或库特定机制提供时，应填充会话 ID。例如：当客户端框架管理会话历史时（如 LlamaIndex chat store、LangChain `session_id`、Google ADK sessions），或当插桩维护后端会话的 GenAI 客户端库时（如 AWS Bedrock agent sessions、OpenAI Assistant threads）。当没有可用的会话标识符时，不应填充会话 ID。例如，不应使用新 UUID、跟踪标识符或请求内容的哈希作为后备值。前端概览会话列表与抽屉标题展示此 ID，缺失时降级为短 trace ID。

### c. Agent 身份

| 属性 | 类型 | 稳定性 | 示例值 | 描述 |
|------|------|--------|--------|------|
| `gen_ai.agent.id` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"3841"` | Agent 的稳定唯一标识 |
| `gen_ai.agent.name` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"DevOpsAgent"` | Agent 的可读名称 |
| `gen_ai.agent.description` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"运维智能助手"` | Agent 的自由文本描述 |
| `gen_ai.agent.version` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"1.2.0"` | Agent 版本 |

> 前端消费：`agent.id` / `agent.name` 用于步骤搜索与 agent 步骤展示；`agent.description` / `agent.version` 前端当前不消费。

### d. 请求参数

| 属性 | 类型 | 稳定性 | 示例值 | 描述 |
|------|------|--------|--------|------|
| `gen_ai.request.model` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"gpt-4o"` | Agent 配置的 GenAI 模型名称 [1] |
| `gen_ai.request.temperature` | `double` | ![Development](https://img.shields.io/badge/-development-blue) | `0.7` | GenAI 请求的 Temperature 设置 |
| `gen_ai.request.stream` | `boolean` | ![Development](https://img.shields.io/badge/-development-blue) | `true` | 请求是否使用流式模式 [2] |
| `gen_ai.request.reasoning.level` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"high"` | 请求模型使用的推理或思考强度 [3] |

**[1] `gen_ai.request.model`:** 如果模型由供应商提供，则值必须是请求的模型的确切名称。如果模型是微调的自定义模型，则值应具有比基础模型更具体的名称。仅当库每个 Agent 只支持单一模型时才应填充；如果库支持多模型或动态选择模型，则不应填充此属性。前端模型横幅、步骤事实栏与概览会话 models 列表消费；缺失时可显示 `--`，或回退到 `response.model`。

**[2] `gen_ai.request.stream`:** 当且仅当请求是流式时设置。如果未设置，则假定请求是非流式的。前端当前不消费，仅作为 TTFC 的前提条件。

**[3] `gen_ai.request.reasoning.level`:** 值应是发送给提供商的确切字符串值。特定提供商的语义约定应记录哪个输入参数映射到此属性（如 OpenAI `reasoning_effort`）。

`gen_ai.request.reasoning.level` 取值（开放枚举，视提供商而定）：

| 值 | 说明 |
|---|---|
| `low` | 低推理层级 |
| `medium` | 中推理层级 |
| `high` | 高推理层级 |
| 提供商确切字符串 | 以提供商 API 实际接受的值为限 |

### e. 响应信息

| 属性 | 类型 | 稳定性 | 示例值 | 描述 |
|------|------|--------|--------|------|
| `gen_ai.response.id` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"chatcmpl-abc123"` | 模型完成结果的唯一标识 |
| `gen_ai.response.model` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"gpt-4o-2024-05-13"` | 生成响应的模型名称 [1] |
| `gen_ai.response.finish_reasons` | `string[]` | ![Development](https://img.shields.io/badge/-development-blue) | `["stop"]` | 模型停止生成 Token 的原因数组，与收到的每个生成结果一一对应 |
| `gen_ai.response.status` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"completed"` | 获取或轮询响应时，提供商返回的响应生命周期状态 |
| `gen_ai.response.time_to_first_chunk` | `double` | ![Development](https://img.shields.io/badge/-development-blue) | `0.35` | 流式首 chunk 延迟（秒） [2] |

**[1] `gen_ai.response.model`:** 如果模型由供应商提供，则值必须是实际使用的模型的确切名称。如果模型是微调的自定义模型，则值应具有比基础模型更具体的名称。前端优先于 `request.model` 展示，缺失回退 `request.model`。`gen_ai.response.id` 前端当前不消费。

**[2] `gen_ai.response.time_to_first_chunk`:** 从客户端发出生成请求到响应流中接收到第一个 chunk 的时间，仅流式请求时填充。

`gen_ai.response.finish_reasons` 取值（开放枚举；归一值如下，供应商扩展值可原样保留，适配层尽量归一如 `tool_calls` → `tool_call`）：

| 值 | 说明 |
|---|---|
| `stop` | 正常结束，模型自行完成回答 |
| `length` | 达到 max_tokens / 上下文长度上限，输出被截断 |
| `content_filter` | 触发内容安全过滤被拦截 |
| `tool_call` | 模型决定调用工具而停止生成文本（agent loop 核心信号） |
| `compaction` | 上下文压缩触发停止 |
| `error` | 异常结束（模型侧错误、超时等），通常伴随 `status.code=2` |

`gen_ai.response.status` 主要用于 `fetch_response` 等查询或轮询已有响应的场景，表示提供商明确返回的响应生命周期。它是开放枚举；GenAI SemConv 定义的 well-known 值如下，其他产品扩展值可原样保留：

| 值 | 说明 |
|---|---|
| `completed` | 响应正常完成 |
| `incomplete` | 完成但不完整（如被 max_tokens 截断） |
| `failed` | 响应失败 |
| `in_progress` | 处理中（异步/流式中间态） |
| `queued` | 排队中（异步任务） |
| `cancelled` | 被取消 |

与 `finish_reasons` 的区别：`status` 是整个响应的生命周期（响应级、单值），`finish_reasons` 是每个输出候选的停止原因（候选级、数组）。

> **填充规则**：只有原数据明确提供响应生命周期状态时才填充 `gen_ai.response.status`。不得从 OTel `status.code` 的 `0`（UNSET）、`1`（OK）或 `2`（ERROR）反推，也不得从 span 已结束或 `finish_reason` 反推。原数据未提供或语义不确定时省略该字段，前端统一显示“未知”；“未知”只是展示兜底，不作为字符串写入协议字段。

### f. Token 用量

| 属性 | 类型 | 稳定性 | 示例值 | 描述 |
|------|------|--------|--------|------|
| `gen_ai.usage.input_tokens` | `int` | ![Development](https://img.shields.io/badge/-development-blue) | `236` | 输入 Token 总数，包含缓存 Token [1] |
| `gen_ai.usage.output_tokens` | `int` | ![Development](https://img.shields.io/badge/-development-blue) | `18` | 输出 Token 总数 [2] |
| `gen_ai.usage.cache_read.input_tokens` | `int` | ![Development](https://img.shields.io/badge/-development-blue) | `100` | 从提供商托管缓存读取的输入 Token 数 [3] |
| `gen_ai.usage.cache_write.input_tokens` | `int` | ![Development](https://img.shields.io/badge/-development-blue) | `50` | 写入提供商托管缓存的输入 Token 数 [4] |
| `gen_ai.usage.reasoning.output_tokens` | `int` | ![Development](https://img.shields.io/badge/-development-blue) | `10` | 用于推理的输出 Token 数 [5] |

**[1] `gen_ai.usage.input_tokens`:** 此值应包括所有类型的输入 Token，包括缓存 Token。插桩应尽最大努力填充此值，使用提供商提供的总计数（当可用时）或根据提供商 API 将从提供商输出解析的不同 Token 类型相加。当提供商同时报告计费 Token 计数和模型消耗 Token 计数时（例如 Cohere 同时暴露 `usage.billed_units` 和 `usage.tokens`），插桩应报告计费计数，以便值与客户被收费的单位匹配。前端 TokenSummary 与概览 Token 趋势核心指标，缺失导致 Token 指标整体失真。

**[2] `gen_ai.usage.output_tokens`:** 当提供商同时报告计费 Token 计数和模型消耗 Token 计数时，插桩应报告计费计数，以便值与客户被收费的单位匹配。

**[3] `gen_ai.usage.cache_read.input_tokens`:** 值应计入 `gen_ai.usage.input_tokens`。

**[4] `gen_ai.usage.cache_write.input_tokens`:** 值应计入 `gen_ai.usage.input_tokens`。

**[5] `gen_ai.usage.reasoning.output_tokens`:** 值应计入 `gen_ai.usage.output_tokens`。

> **注意**：缓存读取和缓存写入都是 `gen_ai.usage.input_tokens` 的子集，统计总 Token 时不得重复相加。`derived.total_tokens = gen_ai.usage.input_tokens + gen_ai.usage.output_tokens`。固定 GenAI commit 没有"缓存输出 Token"字段，不得自造 `gen_ai.usage.cache_*.output_tokens`。

### g. 消息与系统指令

| 属性 | 类型 | 稳定性 | 示例值 | 描述 |
|------|------|--------|--------|------|
| `gen_ai.system_instructions` | `TextPart[] \| GenericPart[]` | ![Development](https://img.shields.io/badge/-development-blue) | `[{"type": "text", "content": "你是运维助手"}]` | 与聊天历史分离的系统指令 [1]（前端 SpanPage 系统指令区消费） |
| `gen_ai.input.messages` | `ChatMessage[]` | ![Development](https://img.shields.io/badge/-development-blue) | `[{"role": "user", "parts": [{"type": "text", "content": "北京天气？"}]}]` | 发送给模型的输入消息列表 [2] |
| `gen_ai.output.messages` | `OutputMessage[]` | ![Development](https://img.shields.io/badge/-development-blue) | `[{"role": "assistant", "parts": [...], "finish_reason": "stop"}]` | 模型生成的输出消息列表；`finish_reason` 按原数据条件填充 [3] |

**[1] `gen_ai.system_instructions`:** 当提供商或 API 将系统指令与聊天历史分开传递时，记录到该属性。聊天历史中的系统消息保留在 `gen_ai.input.messages`。该属性可能包含敏感信息。

**举例写法**

```json
[
  {"type": "text", "content": "你是运维助手，负责帮助用户排查系统问题。"}
]
```

---

**[2] `gen_ai.input.messages`:** 消息必须按发送到模型的顺序提供。此属性可能包含敏感信息，包括用户/PII 数据。<a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/model/gen-ai/gen-ai-input-messages.json" target="_blank">JSON Schema</a>。前端 LLM 详情输入面板与 trace_io 会话首问的核心数据源。Agent / Workflow 根步骤上的原始输入同样需要保留：若来源是根 span 的 `input.value` 或 `invoke_agent_request` event，也应转换成根步骤的输入消息，不能因为子 LLM 已有消息而丢弃。

`ChatMessage.role` 取值（封闭枚举，扩展值前端按原值展示）：

| 值 | 说明 |
|---|---|
| `user` | 用户消息 |
| `assistant` | 助手（模型）消息 |
| `system` | 系统消息（与 `system_instructions` 分离时此处不再重复） |
| `tool` | 工具返回消息 |
| 扩展字符串 | 框架私有角色允许，前端按原值展示 |

Part `type` 取值（well-known 集合 + GenericPart 自定义，前端对未知类型按 JSON 原样渲染）：

| 值 | 说明 |
|---|---|
| `text` | 纯文本内容 |
| `tool_call` | 模型请求调用工具 |
| `tool_call_response` | 工具执行结果 |
| `reasoning` | 模型推理过程 |
| `compaction` | 压缩后的上下文 |
| `blob` | 内联二进制（base64） |
| `file` | 预上传文件引用 |
| `uri` | 外部 URI 引用 |
| `server_tool_call` | 服务端工具调用 |
| `server_tool_call_response` | 服务端工具返回 |
| 自定义（GenericPart） | 允许任意自定义 `type` |

**举例写法**

```json
[
  {
    "role": "user",
    "parts": [
      {"type": "text", "content": "北京今天天气怎么样？"}
    ]
  },
  {
    "role": "assistant",
    "parts": [
      {"type": "reasoning", "content": "用户问天气，需要先调用天气工具获取实时数据。"},
      {"type": "tool_call", "id": "call-123", "name": "get_weather", "arguments": {"city": "北京"}}
    ]
  },
  {
    "role": "tool",
    "parts": [
      {"type": "tool_call_response", "id": "call-123", "response": {"condition": "晴", "temperature_c": 25}}
    ]
  },
  {
    "role": "assistant",
    "parts": [
      {"type": "text", "content": "北京今天晴，25°C。"}
    ]
  }
]
```

各 Part 类型举例（均可出现在 `parts` 数组中）：

```json
{"type": "text", "content": "纯文本内容"}
{"type": "tool_call", "id": "call-123", "name": "get_weather", "arguments": {"city": "北京"}}
{"type": "tool_call_response", "id": "call-123", "response": {"condition": "晴"}}
{"type": "reasoning", "content": "模型推理过程"}
{"type": "compaction", "id": "cmp-1", "content": "压缩后的上下文摘要"}
{"type": "blob", "mime_type": "image/png", "modality": "image", "content": "iVBORw0KGgo..."}
{"type": "file", "mime_type": "application/pdf", "modality": "document", "file_id": "file-abc123"}
{"type": "uri", "mime_type": "image/jpeg", "modality": "image", "uri": "https://example.com/pic.jpg"}
{"type": "server_tool_call", "name": "web_search", "server_tool_call": {"query": "北京天气"}}
{"type": "server_tool_call_response", "server_tool_call_response": {"results": []}}
{"type": "custom_type", "任意扩展字段": "任意值"}
```

---

**[3] `gen_ai.output.messages`:** 每条消息代表模型或 Agent 生成的单个输出选择/候选。前端输出面板、trace_io 终答、tool_call 决策链渲染的核心数据源。<a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/model/gen-ai/gen-ai-output-messages.json" target="_blank">JSON Schema</a>。Agent / Workflow 根步骤上的 `output.value` 或 `invoke_agent_response` event 也必须转换并保留，不能以子 LLM 输出替代根步骤的最终响应。

`OutputMessage.role`：基本固定为 `assistant`（可扩展字符串）。

`OutputMessage.finish_reason`（消息级，条件填充）：只有原数据明确提供时才进行枚举归一并写入；原数据未提供时省略，不从 Part 类型、span 状态或是否结束推导。span 级 `response.finish_reasons` 仅汇总有明确原值的消息级停止原因；无有效原值时同样省略。

**与 InputMessages 的关键差异**

| 差异点 | InputMessages | OutputMessages |
|--------|---------------|----------------|
| 消息类型 | `ChatMessage` | `OutputMessage` |
| `finish_reason` | ❌ 无 | ⚠️ 原数据提供时填充 |
| `role` 常见值 | `user` / `assistant` / `tool` / `system` | 主要是 `assistant` |

**举例写法**

```json
[
  {
    "role": "assistant",
    "parts": [
      {"type": "text", "content": "北京今天晴，25°C，适合出行。"}
    ],
    "finish_reason": "stop"
  }
]
```

工具调用决策轮（`finish_reason: "tool_call"`，前端据此展示"结束原因：调用工具"标签）：

```json
[
  {
    "role": "assistant",
    "parts": [
      {"type": "tool_call", "id": "call-123", "name": "get_weather", "arguments": {"city": "北京"}}
    ],
    "finish_reason": "tool_call"
  }
]
```

输出被长度截断（`finish_reason: "length"`）：

```json
[
  {
    "role": "assistant",
    "parts": [
      {"type": "text", "content": "北京今天的天气情况如下：首先……（输出被截断）"}
    ],
    "finish_reason": "length"
  }
]
```

---

### h. 工具

| 属性 | 类型 | 稳定性 | 示例值 | 描述 |
|------|------|--------|--------|------|
| `gen_ai.tool.name` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"add"` | Agent 使用的工具名称 |
| `gen_ai.tool.description` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"加法计算器"` | 工具用途说明 [1] |
| `gen_ai.tool.type` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"function"` | Agent 使用的工具类型 [2] |
| `gen_ai.tool.call.id` | `string` | ![Development](https://img.shields.io/badge/-development-blue) | `"call-1"` | 工具调用关联 ID，关联消息 tool_call 与 tool 步骤 |
| `gen_ai.tool.definitions` | `ToolDefinition[]` | ![Development](https://img.shields.io/badge/-development-blue) | `[{"type": "function", "name": "get_weather", ...}]` | Agent 或模型可用的工具定义列表 [3] |
| `gen_ai.tool.call.arguments` | `object` | ![Development](https://img.shields.io/badge/-development-blue) | `{"city": "北京"}` | 传递给工具调用的参数 [4] |
| `gen_ai.tool.call.result` | `object` | ![Development](https://img.shields.io/badge/-development-blue) | `{"temperature": 25}` | 工具调用成功时返回的结果 [5] |

**[1] `gen_ai.tool.description`:** 此属性可能包含敏感信息。前端工具面板副标题。`gen_ai.tool.name` 由前端工具事实栏与步骤搜索消费。`gen_ai.tool.call.id` 由前端工具面板调用 ID 与消息-工具关联消费。

**[2] `gen_ai.tool.type`:** 封闭枚举，仅三个固定值：

| 值 | 说明 |
|---|---|
| `extension` | 在 Agent 端执行的工具，直接调用外部 API，弥合 Agent 与现实世界系统之间的差距 |
| `function` | 在客户端执行的工具，Agent 为预定义函数生成参数，客户端执行逻辑 |
| `datastore` | Agent 用于访问和查询结构化或非结构化外部数据，以进行检索增强任务或知识更新的工具 |

**[3] `gen_ai.tool.definitions`:** 工具定义列表，遵循 <a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/model/gen-ai/gen-ai-tool-definitions.json" target="_blank">JSON Schema</a>。`type` 为 `"function"` 时含 `name` / `description` / `parameters`（JSON Schema draft-07）；非 `"function"` 的自定义类型仅 `name` 必填。此属性可能很大且含敏感信息，不建议默认填充。前端当前不直接消费。

`ToolDefinition.type` 取值：

| 值 | 说明 |
|---|---|
| `function` | 函数工具（well-known） |
| 自定义类型 | 非 `function` 时仅 `name` 必填（GenericToolDefinition） |

**举例写法**

```json
[
  {
    "type": "function",
    "name": "get_weather",
    "description": "获取指定城市的天气信息",
    "parameters": {
      "type": "object",
      "properties": {
        "city": {"type": "string", "description": "城市名称"},
        "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
      },
      "required": ["city"]
    }
  }
]
```

**[4] `gen_ai.tool.call.arguments`:** 参数应为符合 <a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/model/gen-ai/gen-ai-tool-call-arguments.json" target="_blank">JSON Schema</a> 的 JSON 对象，具体字段由对应工具的 `definitions.parameters` 约束。插桩只有序列化字符串时，应尽最大努力反序列化为对象。此属性可能包含敏感信息。前端工具面板"调用参数"核心数据源。

**举例写法**

```json
{"city": "北京", "unit": "celsius"}
```

**[5] `gen_ai.tool.call.result`:** 返回值应为符合 <a href="https://github.com/open-telemetry/semantic-conventions-genai/blob/main/model/gen-ai/gen-ai-tool-call-result.json" target="_blank">JSON Schema</a> 的 JSON 对象。插桩只有序列化字符串时，应尽最大努力反序列化为对象。此属性可能包含敏感信息。前端工具面板"返回结果"核心数据源。

**举例写法**

```json
{"temperature": 25, "condition": "晴", "humidity": 40}
```