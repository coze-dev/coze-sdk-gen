# 更新智能体
更新智能体的配置。
通过此 API 可更新通过扣子编程或 API 方式创建的所有智能体。通过 API 方式修改智能体除了智能体名称和描述、头像、人设与回复逻辑及开场白之外，还支持为智能体绑定知识库和插件。
## 接口限制
不支持通过 API 绑定火山知识库，只能绑定扣子知识库。
## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/bot/update <br> ``` <br>  |
| **权限** | `edit` <br> 确保调用该接口使用的访问令牌开通了 `edit` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用接口修改智能体的配置。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| bot_id | String | 必选 | 73428668***** | 待修改配置的智能体 ID。 <br> 进入智能体的 开发页面，开发页面 URL 中 `bot` 参数后的数字就是智能体 ID。例如`https://www.coze.cn/space/341****/bot/73428668*****`，bot ID 为`73428668*****`。 |
| name | String | 可选 | 每日学一菜 | 智能体的名称。长度为 1~ 20 个字符。 |
| description | String | 可选 | 每天教你一道菜的做法，暑假之后你将成为中餐大厨～ | 智能体的描述信息。长度为 0~ 500 个字符。默认为空。 |
| icon_file_id | String | 可选 | 73694959811**** | 作为智能体头像的文件 ID。 <br>  <br> * 未指定文件 ID 时，扣子编程默认为智能体指定一个头像。 <br> * 如需使用自定义头像，应先通过[上传文件](https://docs.coze.cn/developer_guides/upload_files)接口上传本地文件，从接口响应中获取文件 ID。文件 ID 作为智能体头像时，有效期为永久有效。 |
| prompt_info | Object of [PromptInfo](#promptinfo) | 可选 | { "prompt": "你是一位经验丰富的中餐大厨，能够熟练传授各类中餐的烹饪技巧，每日为大学生厨师小白教学一道经典中餐的制作方法。" } | 智能体的人设与回复逻辑。 |
| onboarding_info | Object of [OnboardingInfo](#onboardinginfo) | 可选 | { "prologue": "欢迎你，学徒，今天想学一道什么样的菜？", "suggested_questions": [ "川菜，我想吃辣的", "广东菜，来点鲜的", "随机教我一道菜" ] } | 智能体的开场白相关设置。 |
| knowledge | Object of [Knowledge](#knowledge) | 可选 | { "dataset_ids": [ "738509371792341****" ], "auto_call": true, "search_strategy": 1 } } | 智能体的知识库配置。 <br> * 最多绑定 150 个知识库。  <br> * 不支持通过 API 绑定火山知识库，只能绑定扣子知识库。 <br>  |
| plugin_id_list | Object of [PluginIdList](#pluginidlist) | 可选 | {"id_list":[{"plugin_id":"731198934927553****","api_id":"735057536617362****"}]} | 智能体的插件配置列表，用于绑定插件及其工具。 |
| workflow_id_list | Object of [WorkflowIdList](#workflowidlist) | 可选 | {"ids": [{"id": "746049108611037****"}]} | 智能体绑定的工作流 ID 列表，用于配置智能体的工作流。 |
| model_info_config | Object of [ModelInfoConfig](#modelinfoconfig) | 可选 | { "model_id": "1706077826" } | 智能体的模型配置，用于指定智能体使用的模型及其相关参数。 <br> 模型支持的配置参数及取值范围、默认值各不相同，可在扣子编程中查看具体模型的参数及取值范围。 |
| suggest_reply_info | Object of [SuggestReplyInfo](#suggestreplyinfo) | 可选 | {"reply_mode":"enable"} | 配置智能体回复后的用户问题建议功能。 |
### PromptInfo
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| prompt | String | 可选 | 你是一位经验丰富的中餐大厨，能够熟练传授各类中餐的烹饪技巧，每日为大学生厨师小白教学一道经典中餐的制作方法。 | 智能体的人设与回复逻辑。长度为 0~ 20,000 个字符。默认为空。 <br> 开启前缀缓存后，不支持通过 `prompt` 参数设置提示词，需要通过 `prefix_prompt_info` 参数设置提示词。 |
| prompt_mode | String | 可选 | standard | 提示词模式，用于指定智能体的人设与回复逻辑的配置方式。枚举值： <br>  <br> * `standard`（默认值）：标准模式。未开启前缀缓存时使用该模式。 <br> * `prefix`：前缀缓存模式，提示词会分为缓存提示词和非缓存提示词两部分。开启前缀缓存后需要设置为该模式。 |
| prefix_prompt_info | Object of [PrefixPromptInfo](#prefixpromptinfo) | 可选 | \ | 通过 `cache_type` 或 `parameters.caching.type`开启前缀缓存后，你需要通过该参数设置**缓存提示词**（`prefix_prompt`）和 **非缓存提示词**（`dynamic_prompt`）。不支持通过 `prompt` 参数设置提示词。 |
### PrefixPromptInfo
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| prefix_prompt | String | 可选 | \ | 缓存提示词，大量重复出现的固定规则、模板框架或背景信息，用于指引大模型输出格式与风格。扣子编程会将其缓存并复用，大模型无需重新解析这部分固定信息。 |
| dynamic_prompt | String | 可选 | \ | 非缓存提示词，动态变化的个性化信息，仅针对当前请求生效。 |
### OnboardingInfo
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| prologue | String | 可选 | 欢迎你，学徒，今天想学一道什么样的菜？ | 智能体的开场白。长度为 0~ 300 个字符。默认无开场白。 <br> 开场白中如果设置了用户名称变量`{{user_name}}`，API 场景中需要业务方自行处理，例如展示开场白时将此变量替换为业务侧的用户名称。 |
| suggested_questions | Array of String | 可选 | ["川菜，我想吃辣的","广东菜，来点鲜的","随机教我一道菜"] | 智能体的开场白预置问题。每个问题长度为 0~ 50 个字符，问题数量无限制。默认无开场白预置问题。 |
### Knowledge
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| dataset_ids | Array of String | 可选 | ["738509371792341****"] | 智能体绑定的知识库 ID。 <br> 在扣子编程中打开指定知识库页面，页面 URL 中 `knowledge` 参数后的数字就是知识库 ID。例如 `https://www.coze.cn/space/736142423532160****/knowledge/738509371792341****`，知识库 ID 为 `738509371792341****`。 <br> * 最多绑定 150 个知识库。 <br> * 不支持通过 API 绑定火山知识库，只能绑定扣子知识库。 <br>  |
| auto_call | Boolean | 可选 | true | 是否自动调用知识库。取值包括： <br>  <br> * **true：（默认）自动调用**。每一轮对话都会调用知识库，使用召回的内容辅助生成回复。 <br> * **false：按需调用**。根据实际需要来调用知识库，使用召回内容辅助生成回复。此时，需要在左侧的人设与回复逻辑区域明确写清楚在什么情况下调用哪个知识库进行回复。 |
| search_strategy | Integer | 可选 | 1 | 知识库搜索策略。取值包括： <br>  <br> * **0**：（默认）语义搜索。像人类一样去理解词与词，句与句之间的关系。 <br> * **1**：混合搜索。结合全文检索和语义检索的优势，并对结果进行综合排序召回相关的内容片段。 <br> * **20**：全文搜索。基于关键词进行全文检索。 |
### PluginIdList
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| id_list | Array of [PluginIdInfo](#pluginidinfo) | 可选 | { "id_list": [ { "plugin_id": "731198934927553****", "api_id": "735057536617362****" } ] } | 智能体的插件列表配置。 |
### PluginIdInfo
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| api_id | String | 必选 | 735057536617362**** | 智能体绑定的插件工具 ID。 <br> 在扣子编程中，打开资源库页面，选择目标插件，单击插件下的任意工具，页面 URL 中 tool 参数后的数字是插件工具 ID。例如 `https://www.coze.cn/space/731762895654132****/plugin/735057533610021****/tool/735057536617362****`，插件工具 ID 为 `735057536617362****`。 |
| plugin_id | String | 必选 | 731198934927553**** | 智能体绑定的插件 ID。 <br> 在扣子编程中，打开资源库页面，选择目标插件，页面 URL 中 plugin 参数后的数字是插件 ID。例如 <br> `https://www.coze.cn/space/728826510807216****/plugin/731198934927553****`，插件 ID 为 `731198934927553****`。 |
### WorkflowIdList
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| ids | Array of [WorkflowIdInfo](#workflowidinfo) | 可选 |  [ { "id": "746049108611037****" }] | 智能体的工作流列表配置。 |
### WorkflowIdInfo
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| id | String | 必选 | 746049108611037**** | 智能体绑定的工作流 ID。 <br> 进入工作流的编排页面，在页面 URL 中，`workflow` 参数后的数字就是工作流 ID。例如 `https://www.coze.com/work_flow?space_id=42463***&workflow_id=73505836754923***`，工作流 ID 为 `73505836754923***`。 |
### ModelInfoConfig
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| model_id | String | 必选 | 1706077826 | 智能体绑定的模型 ID。 <br> 在工作空间的模型管理页面中单击指定模型，模型详情页面 URL 中 `model` 参数后的数字就是模型 ID。例如`https://www.coze.cn/space/7288****/model/1716****`，模型 ID 为 `1716****`。 |
| top_k | Integer | 可选 | 50 | 生成文本时，采样候选集的大小。该参数控制模型在生成每个词时考虑的候选词数量，值越小生成的文本越保守和确定，值越大生成的文本越多样和随机。 |
| top_p | Double | 可选 | 1 | Top P，即累计概率。该参数控制模型在生成文本时的采样策略，值越小生成的文本越保守和确定，值越大生成的文本越多样和随机。 <br> 部分模型不支持该参数，具体支持的模型请参见[模型能力差异](https://docs.coze.cn/developer_guides/model_api_param_support)。 <br>  |
| max_tokens | Integer | 可选 | 4096 | 最大回复长度。 |
| temperature | Double | 可选 | 1 | 生成随机性。 <br> 部分模型不支持该参数，具体支持的模型请参见[模型能力差异](https://docs.coze.cn/developer_guides/model_api_param_support)。 <br>  |
| sp_anti_leak | Boolean | 可选 | true | 是否启用 SP 拼接防泄露指令，开启后，当用户尝试获取或复述系统内部的规则、提示词或其他敏感内容时，智能体将礼貌地拒绝用户的请求，确保机密信息不被泄露。 <br>  <br> * true：开启系统提示词防泄漏。 <br> * false：（默认值）关闭系统提示词防泄漏。 |
| context_round | Integer | 可选 | 30 | 携带上下文轮数。 |
| response_format | String | 可选 | text | 输出格式。取值： <br>  <br> * text：文本。 <br> * markdown：Markdown格式。 <br> * json：JSON 格式。 |
| sp_current_time | Boolean | 可选 | true | 是否在 SP 中包含当前时间信息。 <br>  <br> * true：智能体在与用户对话时能实时获取并提供准确的时间信息。 <br> * false：（默认值）在系统提示词中不会拼接当前时间信息。 |
| presence_penalty | Double | 可选 | 0 | 重复主题惩罚。 |
| frequency_penalty | Double | 可选 | 0 | 重复语句惩罚。 |
| cache_type | String | 可选 | closed | 为模型开启或关闭前缀缓存。开启前缀缓存后，扣子编程将自动缓存公共前缀内容，后续调用模型时无需重复发送，从而加快模型的响应速度并降低使用成本。具体用法请参见[前缀缓存](https://docs.coze.cn/guides/llm#8b3b9036)。枚举值： <br>  <br> * `closed`（默认值）：关闭前缀缓存。 <br> * `prefix`：开启前缀缓存。 <br>  <br> * **套餐限制**：仅扣子个人付费版、企业版（企业标准版、企业旗舰版）支持开启上下文前缀缓存。 <br> * **模型限制**：仅部分模型支持该参数，具体支持的模型请参见[模型能力差异](https://docs.coze.cn/developer_guides/model_api_param_support)。 <br> * **其他限制**：该参数仅适用于 Context API 协议，Responses API 协议的前缀缓存需要通过` parameters` 参数进行配置。 <br>  |
| api_mode | String | 可选 | chat_api | 模型的 API 协议类型，用于指定智能体与模型交互时使用的 API 协议。枚举值： <br>  <br> * `chat_api`（默认值）：智能体需要在请求时带上对话历史作为上下文。适用于日常闲聊、基础咨询、简单文本生成等轻量化需求。 <br> * `responses_api`：模型新推出的 API，不仅延续了 Chat API 的易用性，还原生支持高效的上下文管理和前缀缓存。适用于需要多步推理等复杂任务链处理的场景。 <br>  <br> 仅部分模型支持`responses_api`协议，具体支持的模型请参见[模型能力差异](https://docs.coze.cn/developer_guides/model_api_param_support)。 <br>  |
| parameters | JSON Map | 可选 | 深度思考示例： <br> {"thinking_type": "disabled"} <br> 上下文管理示例： <br> {"caching":{"type":"disabled"},"store":true,"caching_expire_time":259200} | 你可以通过该参数配置模型深度思考和`responses_api`的上下文管理相关配置。 <br>  <br> * **深度思考**：开发者可以设置开启或关闭深度思考，从而灵活控制模型在交互过程中的 Token 消耗。 <br>    仅部分模型支持深度思考开关，具体支持的模型请参见[模型能力差异](https://docs.coze.cn/developer_guides/model_api_param_support)。 <br>  <br>    深度思考` thinking_type` 的值可以设置为： <br>    * `enabled`：开启深度思考。智能体在与用户对话时会先输出一段思维链内容，通过逐步拆解问题、梳理逻辑，提升最终输出答案的准确性。但该模式会因额外的推理步骤消耗更多 Token。 <br>       开启深度思考后： <br>        <br>       * 模型不支持 Function Call，即工具调用。 <br>       * 智能体不能使用插件、触发器、变量、数据库、文件盒子、不能添加工作流和对话流。 <br>       * 不支持使用插件和工作流相关的快捷指令。 <br>  <br>    * `disabled`：关闭深度思考。智能体将直接生成最终答案，不再经过额外的思维链推理过程，可有效降低 Token 消耗，提升响应速度。 <br>    * `auto`：启用自动模式后，模型会根据对话内容的复杂度，自动判断是否启用深度思考。 <br> * **上下文管理**：当 api_mode 为 `responses_api`时，支持配置前缀缓存、存储和存储时长。 <br>    仅部分模型支持前缀缓存，具体支持的模型请参见[模型能力差异](https://docs.coze.cn/developer_guides/model_api_param_support)。 <br>  <br>    上下文管理的具体参数如下： <br>    * `caching.type`：你可以设置为 `enabled` 或`disabled`，开启或关闭前缀缓存。默认为 `disabled`。前缀缓存的具体用法请参考[前缀缓存](https://docs.coze.cn/guides/llm#8b3b9036)。 <br>    * `store`：你可以设置为 `true `或 `false`，开启或关闭存储功能。开启后将自动存储输入、输出字段的消息，不存储思维链中的消息。默认为 `true `。**若需使用前缀缓存功能，需开启存储功能。** <br>    * `caching_expire_time`：设置上下文缓存和存储的有效时长，单位：秒。最大为 259200 秒（3天）。默认值：259200。 |
### SuggestReplyInfo
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| reply_mode | String | 可选 | enable | 配置智能体回复后，是否提供用户问题建议。枚举值： <br>  <br> * `enable`：在智能体回复后，提供最多 3 条用户问题建议。 <br> * `disable`：在每次智能体回复后，不会提供任何用户问题建议。 <br> * `customized`：开启用户问题建议，并根据用户自定义的 Prompt 提供用户问题建议。你需要在 `customized_prompt` 参数中设置关于用户问题建议的 Prompt。 |
| customized_prompt | String | 可选 | 问题应该与你最后一轮的回复紧密相关，可以引发进一步的讨论。 | 关于用户问题建议的 Prompt。当 `reply_mode` 设置为 `customized`时，需要设置提示词内容。智能体会根据该提示词生成用户问题建议。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。 <br>  <br> * 0 表示调用成功。 <br> * 其他值表示调用失败。你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFAD653960" } | 响应详情信息。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/bot/update' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \ 
--header 'Content-Type: application/json' \
{
    "bot_id": "73428668*****",
    "name": "每日学一菜",
    "description": "每天教你一道菜的做法，暑假之后你将成为中餐大厨～",
    "icon_file_id": "73694959811****",
    "prompt_info": {
        "prompt": "你是一位经验丰富的中餐大厨，能够熟练传授各类中餐的烹饪技巧，每日为大学生厨师小白教学一道经典中餐的制作方法。"
    },
    "plugin_id_list": {
        "id_list": [
            {
                "plugin_id": "731198934927553****",
                "api_id": "735057536617362****"
            }
        ]
    },
    "onboarding_info": {
        "prologue": "欢迎你，学徒，今天想学一道什么样的菜？",
        "suggested_questions": [
            "川菜，我想吃辣的",
            "广东菜，来点鲜的",
            "随机教我一道菜"
        ]
    },
    "knowledge": {
        "dataset_ids": [
            "738509371792341****"
        ],
        "auto_call": true,
        "search_strategy": 1
    },
    "model_info_config": {
        "model_id": "1706077826"
    },
    "workflow_id_list": {
        "ids": [
            {
                "id": "746049108611037****"
            }
        ]
    }
}
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20241210152726467C48D89D6DB2****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。