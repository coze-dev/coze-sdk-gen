# 执行对话流
执行已发布的对话流。
## 接口说明
对话流是基于对话场景的特殊工作流，专门<span style="color: #060709">用于处理对话类请求。对话流通过对话的方式和用户交互，并完成复杂的业务逻辑。</span>在应用中添加对话流，将对话中的用户指令拆分为一个个步骤节点，并为其设计用户界面，你可以搭建出适用于移动端或网页端的对话式 AI 应用，<span style="color: #060607">实现自动化、智能化的对话流程。</span>关于对话流的详细说明可参考[工作流与对话流](https://docs.coze.cn/api/open/docs/guides/workflow_and_chatflow)。
此接口为流式响应模式，允许客户端在接收到完整的数据流之前就开始处理数据，例如在对话界面实时展示回复内容，减少客户端等待模型完整回复的时间。 
此接口支持包括问答节点、输入节点等可能导致对话中断的节点，对话中断时只需再次调用对话流，在 additional_messages 中指定输入内容，即可继续对话。
* 如果对话流的输入中包含文件、图片等多模态内容，需要先上传多模态内容以获取文件 ID 或 URL 地址，再将其作为对话流的输入。上传方式包括：
   * 调用[上传文件](https://www.coze.cn/open/docs/developer_guides/upload_files) API，获取文件 ID，将此 ID 作为对话流的输入。 
   * 上传到第三方存储工具中，并获取一个公开可访问的 URL 地址，将此 URL 作为对话流的输入。
* 调用接口后，你可以从响应的 Done 事件中获得 debug_url，访问链接即可通过可视化界面查看对话流的试运行过程，其中包含每个执行节点的输入输出等详细信息，帮助你在线调试或排障。debug_url 的访问有效期为 7 天，过期后将无法访问。

此接口可用于调用空间资源库中的对话流，或低代码应用中的对话流。调用这两种对话流时，入参不同：
| **入参** | **资源库对话流** |  | **低代码应用中的对话流** |
| --- | --- | --- | --- |
|  | **在智能体中执行** | **在低代码应用中执行** |  |
| workflow_id | 必选 | 必选 | 必选 |
| app_id | 不传 | 必选 | 必选 |
| bot_id | 必选 | 不传 | 不传 |
| conversation_id | 可选 | 可选 | 可选 |
## 限制说明

* 通过 API 方式执行对话流前，应确认此对话流已发布，执行从未发布过的对话流时会返回错误码 4200。如果是低代码应用中的对话流，应先发布低代码应用为 API 服务；如果是空间资源库中的对话流，应先在资源库中发布对话流。
* 此接口暂不支持异步运行。
* 对话流中不支持提交端插件结果。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflows/chat <br> ``` <br>  |
| **权限** | `run` <br> 确保调用该接口使用的令牌开通了 `run` 权限，详细信息参考[鉴权方式概述](https://docs.coze.cn/api/open/docs/developer_guides/authentication)。 |
| **接口说明** | 执行已发布的对话流。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer <span style="color: #D83931"><em>$Access_Token</em></span> | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/api/open/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |


### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workflow_id | String  | 必选 | 73505836754923*** | 待执行的对话流 ID，此对话流应已发布。 <br> 进入对话流编排页面，在页面 URL 中，`workflow` 参数后的数字就是 Workflow ID。例如 `https://www.coze.cn/work_flow?space_id=42463***&workflow_id=73505836754923***`，Workflow ID 为 `73505836754923***`。 |
| additional_messages  <br>   <br>   | Array of additional_messages object  <br>   | 必选 <br>   | [ { "role": "user", "content_type": "text", "content": "你好" } ] | 对话中用户问题和历史消息。数组长度限制为 50，即最多传入 50 条消息。  <br>  <br> * 你需要通过此字段传入本次对话中用户的问题，也就是对话流的输入参数 USER_INPUT 的值。 <br> * 可以同时传入多条历史消息，也就是本次对话的上下文。多条消息应按对话顺序排列，最后一条消息应为 role=user 的记录，也就是本次对话中用户的问题；其他消息为历史消息。  <br>  <br> 对话流执行到问答节点、输入节点等节点时可能导致对话中断，此时只需再次调用对话流，在 additional_messages 中指定输入内容即可继续对话。 <br>  |
| parameters | map[String]any | 必选 | {"image": "{\"file_id\":\"1122334455\"}","user_name":"George"} | 设置对话流输入参数中的自定义参数。以 JSON 序列化字符串形式传入。你可以在指定对话流的编排页面查看自定义输入参数列表。 <br>  <br> * **设置文件类型的自定义参数** <br>    如果对话流输入参数为 Image 等类型的文件，你可以传入文件 URL 或调用[上传文件](https://www.coze.cn/open/docs/developer_guides/upload_files) API 获取 file_id 后传入 file_id。示例： <br>    * 上传文件并传入 file_id： <br>       * 单个文件示例：`"parameters": { "image": "{\"file_id\":\"1122334455\"}" }` <br>       * 文件数组示例：`"parameters": { "image": [ "{\"file_id\":\"1122334455\"}" ] }`。 <br>    * 传入文件 URL：`“parameters” :{"input":"请总结图片内容", "image": "https://example.com/tos-cn-i-mdko3gqilj/example.png" } ` <br> * **会话名称绑定** <br>    你可以通过输入参数中的 `CONVERSATION_NAME` 绑定对应的会话名称，对话流产生的消息将保存到该会话中。推荐使用 `conversation_id` 绑定会话。 <br>    * **生效范围**：此设置仅对低代码应用中的对话流生效。对于资源库对话流，你需要同时传入 `app_id` 以关联一个低代码应用，该设置才会生效。 <br>    * **获取方法**：你可以在应用编排页面的会话管理中查看会话名称，或通过工作流中的**查询会话列表**节点查询会话名称。 <br>       ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/edc7a1a76fd24606947236d20c7ca9ce~tplv-goo7wpa0wc-image.image) <br>  <br> * 对话流的输入参数  USER_INPUT 应在 additional_messages 中传入，在 parameters 中的 USER_INPUT 不生效。 <br> * 如果 parameters 中未指定 CONVERSATION_NAME 或自定义输入参数，则使用参数默认值运行对话流；如果指定了这些参数，则使用指定值。 <br>  |
|  app_id | String | 可选 | 744208683** | 需要关联的低代码应用 ID。调用对话流时，必须指定 app_id 或 bot_id，便于模型调用智能体或应用的数据库、变量等数据处理问题。 <br> 进入应用开发界面，开发页面URL中的project-ide参数后的数字就是AppID，例如`https://www.coze.cn/space/74421656*****/project-ide/744208683**` ，低代码应用 ID 为`744208683**`**。** <br> 运行低代码应用中的对话流时，app_id 必选。详细说明可参考[接口说明](https://docs.coze.cn/api/open/docs/developer_guides/workflow_chat#15461feb)。 <br>  |
| bot_id <br>  | String  <br>  | 可选 <br>  | 75049216555930**** | 需要关联的智能体 ID。 调用对话流时，必须指定 app_id 或 bot_id，便于模型调用智能体或应用的数据库、变量等数据处理问题。 <br> 进入智能体开发页面，开发页面 URL 中 `bot` 参数后的数字就是智能体 ID。例如`https://www.coze.cn/space/73428668341****/bot/73428668*****`，bot ID 为`73428668*****`。  <br> 运行资源库中的对话流时，根据对话流执行的位置（智能体或低代码应用）选择设置 bot_id 还是 app_id。详细说明可参考[接口说明](https://docs.coze.cn/api/open/docs/developer_guides/workflow_chat#15461feb)。 <br>  |
| conversation_id | String | 可选 | 748348012449138**** | 对话流对应的会话 ID，对话流产生的消息会保存到此会话中。会话默认为开始节点设置的 CONVERSATION_NAME，也可以通过 conversation_id 参数指定会话。 <br> * 指定 conversation_id 时，parameters 中设置的 CONVERSATION_NAME 不生效。 <br> * 会话的创建者必须和执行对话流的用户一致，即 API 访问令牌的创建者一致，否则无法执行对话流。 <br> * 会话与 app_id、渠道匹配，不同渠道的会话隔离。 <br> * 指定 bot_id 时，如果没有传入 conversation_id ，扣子编程会创建一个新的 conversation_id。不支持同时指定 bot_id 和 app_id。详细说明可参考[接口说明](https://docs.coze.cn/api/open/docs/developer_guides/workflow_chat#15461feb)。 <br>  |
| ext | Map[String][String] <br>  | 可选 | {"latitude":"39.9042","longitude":"116.4074","user_id":"123456789"} | 用于指定一些额外的字段，以 Map[String][String] 格式传入。例如某些插件会隐式用到的经纬度等字段。 <br> 目前仅支持以下字段： <br>  <br> * latitude：String 类型，表示纬度。 <br> * longitude：String 类型，表示经度。 <br> * user_id：String 类型，表示用户 ID。 |
| workflow_version <br>  <br>  | String <br>  | 可选 <br>  | v0.0.5 | 对话流的版本号，仅当运行的对话流属于资源库对话流时有效。未指定版本号时默认执行最新版本的对话流。 <br>  |
| connector_id | String <br>  | 可选 <br>  | 1024 <br>  | 渠道 ID，用于配置该对话流在什么渠道执行。 <br> 当智能体或低代码应用发布到某个渠道后，可以通过该参数指定对话流在对应的渠道执行。 <br> 扣子编程的渠道 ID 包括： <br>  <br> * 1024（默认值）：API 渠道。 <br> * 999：Chat SDK。 <br> * 998：WebSDK。 <br> * 10000122：扣子商店。 <br> * 10000113：微信客服。 <br> * 10000120：微信服务号。 <br> * 10000121：微信订阅号。 <br> * 10000126：抖音小程序。 <br> * 10000127：微信小程序。 <br> * 10000011：飞书。 <br> * 10000128：飞书多维表格。 <br> * 10000117：掘金。 <br> * 自定义渠道 ID。自定义渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 <br>  <br> 不同渠道的用户数据、会话记录等相互隔离，进行数据分析统计时，不支持跨渠道数据调用。 <br>  |
#### additional_messages object
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| role | String | 必选 | user | 发送这条消息的实体。取值： <br>  <br> * **user**：代表该条消息内容是用户发送的。 <br> * **assistant**：代表该条消息内容是模型发送的。 |
| type <br>  | String | 可选 <br>  | question | 消息类型。默认为 **question。** <br>  <br> * **question**：用户输入内容。 <br> * **answer**：模型返回给用户的消息内容，支持增量返回。如果对话流绑定了输出节点，可能会存在多 answer 场景，此时可以用流式返回的结束标志来判断所有 answer 完成。 <br> * **function_call**：智能体对话过程中调用函数（function call）的中间结果。  <br> * **tool_response**：调用工具 （function call）后返回的结果。 |
| content | String | 可选 | 北京今天的天气怎么样 | 消息的内容，仅支持纯文本。 <br> 如果需要通过**输入节点**输入消息，可通过如下两种格式： <br>  <br> * JSON 格式的键值对，例如：`{"input":"北京今天的天气怎么样"}`。 <br> * 用`\n`分隔的键值对，例如：`{key1:val1\nkey2:val2\nkey3:val3}`。 <br>  <br> * 暂不支持在 `content` 中输入多模态（文本、图片、文件混合输入）、卡片等类型的内容。 <br> * 如果需要传入图片、文件等多模态内容，你可以在 `parameters` 中传入对应的自定义参数及其值。 <br>  |
| content_type | String | 可选 | text | 消息内容的类型。 <br> content_type **** 固定为 text，表示普通文本。 <br> 指定 content 时，应同时设置 content_type。  <br>  |
|  meta_data | Map  | 可选 | {"key1":"value1", "key2":"value2"} | 创建消息时的附加消息，获取消息时也会返回此附加消息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 |
## 返回参数
流式响应允许客户端在接收到完整的数据流之前就开始处理数据，例如在对话界面实时展示回复内容，减少客户端等待模型完整回复的时间。 
流式响应的整体流程如下：

<div type="doc-tabs">
<div type="tab-item" title="流式响应流程" key="RzH6LYdwg4">

```Plain Text
######### 整体概览 （chat, MESSAGE 两级） 
# chat - 开始
# chat - 处理中 
#   MESSAGE - 知识库召回 
#   MESSAGE - function_call 
#   MESSAGE - tool_output 
#   MESSAGE - answer is normal text 
#   MESSAGE - 多 answer 的情况下，会继续有 message.delta 
# chat - 完成 
# 流结束 event: done 
######### 
```


</div>
<div type="tab-item" title="流式响应示例" key="OQzIiFWnWz">

```JSON
# chat - 开始
event: conversation.chat.created
data: {"id":"120","conversation_id":"456","created_at":1733407180,"last_error":{"code":0,"msg":""},"status":"created","usage":{"token_count":0,"output_count":0,"input_count":0},"section_id":"789"}
# chat - 处理中 
event: conversation.chat.in_progress
data: {"id":"121","conversation_id":"456","created_at":1733407180,"last_error":{"code":0,"msg":""},"status":"in_progress","usage":{"token_count":0,"output_count":0,"input_count":0},"section_id":"789"}
# MESSAGE - answer is normal text 
event: conversation.message.delta
data: {"id":"122","conversation_id":"456","role":"assistant","type":"answer","content":"中午吃啥了","content_type":"text","chat_id":"567","section_id":"789","created_at":1733407182}

# MESSAGE - 消息结束
event: conversation.message.completed
data: {"id":"124","conversation_id":"456","role":"assistant","type":"answer","content":"中午吃啥了","content_type":"text","chat_id":"567","section_id":"789","created_at":1733407182}

event: conversation.message.completed
data: {"id":"125","conversation_id":"456","role":"assistant","type":"verbose","content":"{\"msg_type\":\"interrupt\",\"data\":\"\",\"from_module\":null,\"from_unit\":null}","content_type":"text","chat_id":"567","section_id":"789","created_at":1733407182,"updated_at":1733407182}

event: conversation.message.completed
data: {"id":"130","conversation_id":"456","role":"assistant","type":"verbose","content":"{\"msg_type\":\"generate_answer_finish\",\"data\":\"{\\\"finish_reason\\\":1,\\\"FinData\\\":\\\"\\\"}\",\"from_module\":null,\"from_unit\":null}","content_type":"text","chat_id":"567","section_id":"789","created_at":1733407182,"updated_at":1733407182}

event: done
data: {"debug_url":"https://www.coze.cn/work_flow?execute_id=74449256856****\u0026space_id=7442165654356*****\u0026workflow_id=744224337778*****"}
```


</div>
</div>
返回的事件消息体结构如下：
| **参数** | **类型** | **说明** |
| --- | --- | --- |
| event | String | 当前流式返回的数据包事件。在流式响应中，服务端不会一次性发送所有数据，而是以数据流的形式逐条发送数据给客户端，数据流中包含对话过程中触发的各种事件（event），直至处理完毕或处理中断。处理结束后，服务端会通过 conversation.message.completed 事件返回拼接后完整的模型回复信息。各个事件的说明可参考下表。  |
| data | Object | 消息内容。其中，chat 事件和 message 事件的格式不同。 <br>  <br> * chat 事件中，data 为 **Chat** **Object**。 <br> * message 事件中，data 为 **Message** **Object**。 |
流式响应事件列表：
| **事件（event）名称** | **说明** |
| --- | --- |
| conversation.chat.created | 创建对话的事件，表示对话开始。 |
| conversation.chat.in_progress | 服务端正在处理对话。 |
| conversation.message.delta | 增量消息，通常是 type=answer 时的增量消息。 |
| conversation.message.completed | message 已回复完成。此时流式包中带有所有 message.delta 的拼接结果，且每个消息均为 completed 状态。 |
| conversation.chat.completed | 对话完成。 |
| conversation.chat.failed | 此事件用于标识对话失败。 |
| conversation.chat.requires_action | 对话中断，需要使用方上报工具的执行结果。通常是触发了问答节点或输入节点，需要再次调用此接口继续对话。 |
| error | 流式响应过程中的错误事件。关于 code 和 msg 的详细说明，可参考[错误码](https://docs.coze.cn/api/open/docs/developer_guides/coze_error_codes)。 |
| done | 本次会话的流式返回正常结束。 |
#### Chat Object
| 参数 | 类型 | 是否可选 | 说明 |
| --- | --- | --- | --- |
| id | String | 必填 | 对话 ID，即对话的唯一标识。 |
| conversation_id | String | 必填 | 会话 ID，即会话的唯一标识。 |
| bot_id | String | 必填 | 要进行会话聊天的智能体 ID。 |
| created_at | Integer | 选填 | 对话创建的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| completed_at | Integer | 选填 | 对话结束的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| failed_at | Integer | 选填 | 对话失败的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| meta_data | Map<String,String> | 选填 | 创建消息时的附加消息，用于传入使用方的自定义数据，获取消息时也会返回此附加消息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 |
| last_error <br>  | Object | 选填 | 对话运行异常时，此字段中返回详细的错误信息，包括： <br>  <br> * Code：错误码。Integer 类型。0 表示成功，其他值表示失败。 <br> * Msg：错误信息。String 类型。 <br>  <br> * 对话正常运行时，此字段返回 null。 <br> * suggestion 失败不会被标记为运行异常，不计入 last_error。 <br>  |
| status <br>  | String | 必填 | 对话的运行状态。取值为： <br>  <br> * created：对话已创建。 <br> * in_progress：智能体正在处理中。 <br> * completed：智能体已完成处理，本次对话结束。 <br> * failed：对话失败。 <br> * requires_action：对话中断，需要进一步处理。 <br> * canceled：对话已取消。 |
| required_action | Object | 选填 | 需要运行的信息详情。 |
| usage | Object | 选填 | Token 消耗的详细信息。实际的 Token 消耗以对话结束后返回的值为准。 |
| »token_count | Integer | 选填 | 本次对话消耗的 Token 总数，包括 input 和 output 部分的消耗。 |
| »output_count | Integer | 选填 | output 部分消耗的 Token 总数。 |
| »input_count | Integer | 选填 | input 部分消耗的 Token 总数。 |
Chat Object 的示例如下：
```JSON
{
// 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。
    "id": "737662389258662****",
    "conversation_id": "737554565555041****",
    "bot_id": "736661612448078****",
    "completed_at": 1717508113,
    "last_error": {
        "code": 0,
        "msg": ""
    },
    "status": "completed",
    "usage": {
        "token_count": 6644,
        "output_count": 766,
        "input_count": 5878
    }
}
```

#### Message Object
| 参数 | 类型 | 说明 |
| --- | --- | --- |
| id | String | Message ID，即消息的唯一标识。 |
| conversation_id | String | 此消息所在的会话 ID。 |
| bot_id | String | 编写此消息的智能体ID。此参数仅在对话产生的消息中返回。 |
| chat_id | String | Chat ID。此参数仅在对话产生的消息中返回。 |
| meta_data | Map | 创建消息时的附加消息，获取消息时也会返回此附加消息。 |
| role | String | 发送这条消息的实体。取值： <br>  <br> * **user**：代表该条消息内容是用户发送的。 <br> * **assistant**：代表该条消息内容是智能体发送的。 |
| content | String | 消息的内容，支持纯文本、多模态（文本、图片、文件混合输入）等多种类型的内容。 |
| content_type | String | 消息内容的类型，取值包括： <br>  <br> * text：文本。 <br> * object_string：多模态内容，即文本和文件的组合、文本和图片的组合。 |
| created_at | Integer | 消息的创建时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| updated_at | Integer | 消息的更新时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| type | String | 消息类型。 <br>  <br> * **question**：用户输入内容。 <br> * **answer**：智能体返回给用户的消息内容，支持增量返回。如果对话流绑定了 messge 节点，可能会存在多 answer 场景，此时可以用流式返回的结束标志来判断所有 answer 完成。 <br> * **function_call**：智能体对话过程中调用函数（function call）的中间结果。  <br> * **tool_response**：调用工具 （function call）后返回的结果。 |
| section_id | String | 预留字段，仅在调用[发起对话](https://docs.coze.cn/api/open/docs/developer_guides/chat_v3) API 时的部分场景下返回，无需关注。 |
## 示例
### 请求示例

<div type="doc-tabs">
<div type="tab-item" title="执行应用对话流" key="lfeBwssMXn">

执行低代码应用中的对话流时，需要通过 app_id 指定应用 ID。
示例如下：
```Shell
curl --location 'https://api.coze.cn/v1/workflows/chat' \
--header 'Authorization: Bearer pat_*****' \
--header 'Content-Type: application/json' \
--data '{
    "workflow_id": "74423***",
    "app_id": "7439828073***",
    "additional_messages": [
        {
            "role": "user",
            "content_type": "text",
            "content": "你好"
        }
    ],
    "parameters": {
        "image":"{\"file_id\":\"1122334455\"}"
    }
}'
```


</div>
<div type="tab-item" title="执行资源库对话流" key="jMTGj01KqN">

执行资源库中的对话流时，需要通过 bot_id 参数绑定一个智能体 ID，或通过 app_id 绑定一个应用。
执行对话流且绑定智能体的示例如下：
```Shell
curl --location 'https://api.coze.cn/v1/workflows/chat' \
--header 'Authorization: Bearer pat_*****' \
--header 'Content-Type: application/json' \
--data '{
    "workflow_id": "74423***",
    "bot_id": "7439828073***",
    "conversation_id": "7237029***",
    "additional_messages": [
        {
            "role": "user",
            "content_type": "text",
            "content": "你好"
        }
    ],
    "parameters": {
        "image":"{\"file_id\":\"1122334455\"}"
    }
}'
```


</div>
</div>
### 返回示例

<div type="doc-tabs">
<div type="tab-item" title="执行成功" key="KxyJMa0xpf">

```JSON
# chat - 开始
event: conversation.chat.created
data: {"last_error":{"msg":"","code":0},"usage":{"token_count":0,"output_count":0,"input_count":0},"id":"75598600924738*****","conversation_id":"75598599835687*****","section_id":"75598599835687*****","status":"created","created_at":1760167093}
# chat - 处理中
event: conversation.chat.in_progress
data: {"usage":{"token_count":0,"output_count":0,"input_count":0},"conversation_id":"75598599835687*****","id":"75598600924738*****","section_id":"75598599835687*****","last_error":{"code":0,"msg":""},"status":"in_progress","created_at":1760167093}

event: conversation.message.delta
data: {"conversation_id":"75598599835687*****","role":"assistant","section_id":"75598599835687*****","chat_id":"75598600924738*****","id":"75598601273532*****","type":"answer","content":"那我给你讲","content_type":"text"}

event: conversation.message.delta
data: {"content":"个会冒冷气的笑话哦！从前有只小企鹅问","content_type":"text","section_id":"75598599835687*****","conversation_id":"75598599835687*****","type":"answer","chat_id":"75598600924738*****","id":"75598601273532*****","role":"assistant"}

event: conversation.message.delta
data: {"id":"75598601273532*****","conversation_id":"75598599835687*****","role":"assistant","content_type":"text","chat_id":"75598600924738*****","type":"answer","content":"妈妈：\"为什么我们住在南极呀","section_id":"75598599835687*****"}

event: conversation.message.delta
data: {"content":"？\"妈妈摸着它的圆脑袋说","role":"assistant","conversation_id":"75598599835687*****","type":"answer","content_type":"text","chat_id":"75598600924738*****","section_id":"75598599835687*****","id":"75598601273532*****"}

event: conversation.message.delta
data: {"role":"assistant","type":"answer","chat_id":"75598600924738*****","section_id":"75598599835687*****","id":"75598601273532*****","conversation_id":"75598599835687*****","content":"：\"因为这里有好多好多鱼呀~\"小","content_type":"text"}

event: conversation.message.delta
data: {"role":"assistant","chat_id":"75598600924738*****","id":"75598601273532*****","conversation_id":"75598599835687*****","type":"answer","content":"企鹅眨巴眨巴眼睛：\"可是北极熊住在","content_type":"text","section_id":"75598599835687*****"}

event: conversation.message.delta
data: {"content":"北极也有鱼呀！\"妈妈突然把翅膀","chat_id":"75598600924738*****","section_id":"75598599835687*****","role":"assistant","type":"answer","content_type":"text","id":"75598601273532*****","conversation_id":"75598599835687*****"}

event: conversation.message.delta
data: {"id":"75598601273532*****","conversation_id":"75598599835687*****","role":"assistant","type":"answer","section_id":"75598599835687*****","content":"搭在它肩上，压低声音说","content_type":"text","chat_id":"75598600924738*****"}

event: conversation.message.delta
data: {"role":"assistant","type":"answer","content":"：\"傻孩子...因为如果我们搬到北极，就","content_type":"text","id":"75598601273532*****","conversation_id":"75598599835687*****","chat_id":"75598600924738*****","section_id":"75598599835687*****"}

event: conversation.message.delta
data: {"conversation_id":"75598599835687*****","role":"assistant","type":"answer","content_type":"text","section_id":"75598599835687*****","id":"75598601273532*****","content":"会变成'北极大企鹅啦！","chat_id":"75598600924738*****"}

# MESSAGE - 消息结束
event: conversation.message.completed
data: {"chat_id":"75598600924738*****","created_at":1760167105,"id":"75598601273532*****","content":"那我给你讲个会冒冷气的笑话哦！从前有只小企鹅问妈妈：\"为什么我们住在南极呀？\"妈妈摸着它的圆脑袋说：\"因为这里有好多好多鱼呀~\"小企鹅眨巴眨巴眼睛：\"可是北极熊住在北极也有鱼呀！\"妈妈突然把翅膀搭在它肩上，压低声音说：\"傻孩子...因为如果我们搬到北极，就会变成'北极大企鹅'啦！\"","content_type":"text","section_id":"75598599835687*****","conversation_id":"75598599835687*****","role":"assistant","type":"answer"}

event: conversation.message.completed
data: {"updated_at":1760167107,"id":"75598601516393*****","role":"assistant","content_type":"text","chat_id":"75598600924738*****","created_at":1760167107,"conversation_id":"75598599835687*****","type":"verbose","content":"{\"msg_type\":\"empty result\",\"data\":\"empty result\",\"from_module\":null,\"from_unit\":null}","section_id":"75598599835687*****"}

event: conversation.message.completed
data: {"updated_at":1760167107,"id":"75598601569435*****","content":"{\"msg_type\":\"generate_answer_finish\",\"data\":\"{\\\"finish_reason\\\":0,\\\"FinData\\\":\\\"\\\"}\",\"from_module\":null,\"from_unit\":null}","content_type":"text","chat_id":"75598600924738*****","section_id":"75598599835687*****","created_at":1760167107,"conversation_id":"75598599835687*****","role":"assistant","type":"verbose"}
# chat - 结束
event: conversation.chat.completed
data: {"status":"completed","usage":{"token_count":1736,"output_count":498,"input_count":1238},"created_at":1760167093,"id":"75598600924738*****","section_id":"75598599835687*****","completed_at":1760167107,"conversation_id":"75598599835687*****","last_error":{"code":0,"msg":""}}

event: done
data: {"debug_url":"https://www.coze.cn/work_flow?execute_id=75598600951038*****&space_id=74982048832804*****&workflow_id=75228046974940*****&execute_mode=2"}
```


</div>
<div type="tab-item" title="执行失败" key="Emu90f8BwZ">

```JSON
event: conversation.chat.failed
data: {"code": "720702204","msg": "会话名不存在"}
```


</div>
</div>
## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/api/open/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。