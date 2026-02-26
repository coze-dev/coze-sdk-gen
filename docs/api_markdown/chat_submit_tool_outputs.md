# 提交工具执行结果
调用此接口提交工具执行的结果。
## 接口说明
你可以将需要客户端执行的操作定义为插件，对话中如果触发这个插件，流式 event 响应信息会提示“conversation.chat.requires_action”，此时需要执行客户端的操作后，通过此接口提交插件执行后的结果。
* 调用[发起对话](https://docs.coze.cn/developer_guides/chat_v3) API 时，`auto_save_history` 参数需要设置为 `true`，否则调用本 API 提交工具执行结果时会提示 5000 错误。
* 仅触发了端插件的对话需要调用此接口提交执行结果。端插件是非扣子服务端执行的插件，需要开发者自行执行任务后提交结果，通常用于 IoT 等设备控制场景。详细说明可参考[通过 API 使用端插件](https://docs.coze.cn/guides/use_local_plugin)。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v3/chat/submit_tool_outputs <br> ``` <br>  |
| **权限** | `submitToolChat` <br> 确保调用该接口使用的访问令牌开通了 `submitToolChat` 权限，详细信息参考[准备工作](https://www.coze.com/docs/developer_guides/preparation)。 |
| **接口说明** | 调用接口提交工具执行结果。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| conversation_id | String | 必选 | 748348012449138*** | Conversation ID，即会话的唯一标识。可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口 Response 中查看 conversation_id 字段。 |
| chat_id | String | 必选 | 738137187639794*** | Chat ID，即对话的唯一标识。可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口 Response 中查看 id 字段，如果是流式响应，则在 Response 的 chat 事件中查看 id 字段。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| stream | Boolean | 可选 | true | 是否开启流式响应。 <br>  <br> * true：填充之前对话中的上下文，继续流式响应。 <br> * false：（默认）非流式响应，仅回复对话的基本信息。 |
| tool_outputs | Array of [ToolOutput](#tooloutput) | 必选 | [{"tool_call_id":"BUJJF0dAQ0NAEBVeQkVKEV5HFURFXhFCEhFeFxdHShcSQEtFSxY****","output":"Tokyo"}] | 工具执行结果。 |
### ToolOutput
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| output | String | 必选 | {\"file\": \"12345\"} | 工具的执行结果。 <br> 如果端插件工具某个请求参数包含 Image 等文件类型，可以先调用[上传文件](https://docs.coze.cn/developer_guides/upload_files) API 获取 `file_id`，然后调用此 API 在 output 中以序列化之后的 JSON 格式传入 `file_id`。例如入参 `file` 为文件类型，文件的 `file_id` 为 `12345`，则 `output` 可以指定为 `"output": "{\"file\": \"12345\"}"`。 |
| tool_call_id | String | 必选 | BUJJF0dAQ0NAEBVeQkVKEV5HFURFXhFCEhFeFxdHShcSQEtFSxY**** | 上报运行结果的 ID。你可以在[发起对话（V3）](https://docs.coze.cn/api/open/docs/docs/developer_guides/chat_v3)接口响应的 tool_calls 字段下查看此 ID。 |

## 返回参数
### 非流式响应
在非流式响应中，无论服务端是否处理完毕，立即发送响应消息。其中包括本次对话的 chat_id、状态等元数据信息，但不包括模型处理的最终结果。
非流式响应不需要维持长连接，在场景实现上更简单，但通常需要客户端主动查询对话状态和消息详情才能得到完整的数据。你可以通过接口[查看对话详情](https://docs.coze.cn/api/open/docs/developer_guides/retrieve_chat)确认本次对话处理结束后，再调用[查看对话消息详情](https://docs.coze.cn/api/open/docs/developer_guides/list_chat_messages)接口查看模型回复等完整响应内容。流程如下：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/79f9ce1cefc94f4daf20cad241382cdf~tplv-goo7wpa0wc-image.image)
非流式响应的结构如下：
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| data | Object of [ChatV3ChatDetail](#chatv3chatdetail) |  | 本次对话的基本信息。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 响应的详细信息。 |
### ChatV3ChatDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 738137187639794**** | 对话 ID，即对话的唯一标识。 |
| conversation_id | String | 738136585609548**** | 会话 ID，即会话的唯一标识。 |
| bot_id | String | 737946218936519**** | 该会话所属的智能体的 ID。 |
| status | String | completed | 对话的运行状态。取值为： <br>  <br> * created：对话已创建。 <br> * in_progress：智能体正在处理中。 <br> * completed：智能体已完成处理，本次对话结束。 <br> * failed：对话失败。 <br> * requires_action：对话中断，需要进一步处理。 <br> * canceled：对话已取消。 |
| created_at | Integer | 1718609571 | 对话创建的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| completed_at | Integer | 1718609575 | 对话结束的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| failed_at | Integer | 1718609571 | 对话失败的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| meta_data | JSON Map | {"customKey1":"customValue1","customKey2":"customValue2"} | 发起对话时的附加消息，用于传入使用方的自定义数据，[查看对话详情](https://docs.coze.cn/developer_guides/retrieve_chat)时也会返回此附加消息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 |
| last_error | Object of [LastError](#lasterror) | \ | 对话运行异常时，此字段中返回详细的错误信息，包括： <br>  <br> * Code：错误码。Integer 类型。0 表示成功，其他值表示失败。 <br> * Msg：错误信息。String 类型。 |
| section_id | String | 737946218936519**** | 上下文片段 ID。每次调用[清除上下文](https://docs.coze.cn/developer_guides/clear_conversation_context) API 都会生成一个新的 section_id。 |
| required_action | Object of [RequiredAction](#requiredaction) | {"type":"submit_tool_outputs","submit_tool_outputs":{"tool_calls":[{"id":"738137187639794****","type":"function","function":{"name":"get_weather","arguments":"{\"city\":\"Beijing\"}"}}]}} | 当对话状态为 `requires_action` 时，此字段包含需要进一步处理的信息详情，用于继续对话。 |
| usage | Object of [Usage](#usage) |  | 预留字段，无需关注，具体消耗的 Token 请查看火山账单。 |
### LastError
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| msg | String | 详见响应示例 | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 |
| code | Integer | 0 | 状态码。 <br> 0 代表调用成功。 |
### RequiredAction
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| type | String | submit_tool_outputs | 额外操作的类型，枚举值： <br> `submit_tool_outputs`：需要提交工具输出以继续对话。 |
| submit_tool_outputs | Object of [SubmitToolOutputs](#submittooloutputs) | {"tool_calls":[{"id":"738137187639794****","type":"function","function":{"name":"get_weather","arguments":"{\"city\":\"Beijing\"}"}}]} | 当对话状态为 `requires_action`时，此字段包含需要提交的工具输出信息，用于继续对话。通常包含一个工具调用列表，每个工具调用包含工具类型和参数。 |
### SubmitToolOutputs
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| tool_calls | Array of [InterruptPlugin](#interruptplugin) | [{"id":"738137187639794****","type":"function","function":{"name":"get_weather","arguments":"{\"city\":\"Beijing\"}"}}] | 当对话状态为 `requires_action` 时，此字段包含需要提交的工具调用列表，每个工具调用包含工具类型和参数。 |
### InterruptPlugin
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 738137187639794**** | 上报运行结果的 ID。 |
| type | String | function | 工具类型，枚举值包括： <br>  <br> * function：待执行的方法，通常是端插件。触发端插件时会返回此枚举值。 <br> * reply_message：待回复的选项。触发工作流问答节点时会返回此枚举值。 |
| function | Object of [InterruptFunction](#interruptfunction) | {"name":"get_weather","arguments":"{\"city\":\"Beijing\"}"} | 当对话状态为 `requires_action`时，此字段表示需要调用的工具或函数的定义，包含函数名称和参数。通常用于指定工具的具体执行方法。 |
### InterruptFunction
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| name | String | get_weather | 当对话状态为 `requires_action` 时，此字段表示需要调用的工具或函数的名称，用于继续对话。通常与 `arguments`字段配合使用，指定工具的具体执行方法。 |
| arguments | String | {"city":"Beijing"} | 当对话状态为 `requires_action`时，此字段表示需要调用的工具或函数的参数，通常为 JSON 格式的字符串，用于指定工具的具体执行参数。 |
### Usage
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| input_count | Integer | 50 | 输入内容所消耗的 Token 数，包含对话上下文、系统提示词、用户当前输入等所有输入类的 Token 消耗。 |
| token_count | Integer | 150 | 本次 API 调用消耗的 Token 总量，包括输入和输出两部分的消耗。 |
| output_count | Integer | 100 | 大模型输出的内容所消耗的 Token 数。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
### 流式响应
在流式响应中，服务端不会一次性发送所有数据，而是以数据流的形式逐条发送数据给客户端，数据流中包含对话过程中触发的各种事件（event），直至处理完毕或处理中断。处理结束后，服务端会通过 conversation.message.completed 事件返回拼接后完整的模型回复信息。各个事件的说明可参考**流式响应事件**。
流式响应允许客户端在接收到完整的数据流之前就开始处理数据，例如在对话界面实时展示智能体的回复内容，减少客户端等待模型完整回复的时间。
流式响应的整体流程如下：
```JSON
######### 整体概览 （chat, MESSAGE 两级）
# chat - 开始
# chat - 处理中
#   MESSAGE - 知识库召回
#   MESSAGE - function_call
#   MESSAGE - tool_output
#   MESSAGE - answer is card
#   MESSAGE - answer is normal text
#   MESSAGE - 多 answer 的情况下，会继续有 message.delta
#   MESSAGE - verbose （多 answer、Multi-agent 跳转等场景）
#   MESSAGE - suggestion
# chat - 完成
# 流结束 event: done
#########
```


| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| event | String | conversation.chat.created | 流式响应事件，事件说明如下： <br>  <br> * conversation.chat.created：创建对话的事件，表示对话开始。 <br> * conversation.chat.in_progress：服务端正在处理对话。 <br> * conversation.message.delta：增量消息，通常是 type=answer 时的增量消息。 <br> * conversation.message.completed：message 已回复完成，此时流式包中带有所有 message.delta 的拼接结果，且每个消息均为 completed 状态。 <br> * conversation.chat.completed：对话完成。 <br> * conversation.chat.failed：此事件用于标识对话失败。 <br> * conversation.chat.requires_action：对话中断，需要使用方上报工具的执行结果。 <br> * error：流式响应过程中的错误事件，关于 code 和 msg 的详细说明，可参考错误码。 <br> * done：本次会话的流式返回正常结束。 |
| data | Object of [ChatV3MessageDetail](#chatv3messagedetail) |  | 消息内容。其中，chat 事件和 message 事件的格式不同。 <br>  <br> * chat 事件中，data 为 **Chat** **Object**。 <br> * message 事件中，data 为 **Message** **Object**。 |
### ChatV3MessageDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 738216762080970**** | 智能体回复的消息的 Message ID，即消息的唯一标识。 |
| conversation_id | String | 738147352534297**** | 此消息所在的会话 ID。 |
| role | String | assistant | 发送这条消息的实体。取值： <br>  <br> * user：代表该条消息内容是用户发送的。 <br> * assistant：代表该条消息内容是智能体发送的。 |
| type | String | verbose | 消息类型。 <br>  <br> * **question**：用户输入内容。 <br> * **answer**：智能体返回给用户的消息内容，支持增量返回。如果工作流绑定了 messge 节点，可能会存在多 answer 场景，此时可以用流式返回的结束标志来判断所有 answer 完成。 <br> * **function_call**：智能体对话过程中调用函数（function call）的中间结果。 <br> * **tool_output**：调用工具 （function call）后返回的结果。 <br> * **tool_response**：调用工具 （function call）后返回的结果。 <br> * **follow_up**：如果在智能体上配置打开了用户问题建议开关，则会返回推荐问题相关的回复内容。 <br> * **verbose**：多 answer 场景下，服务端会返回一个 verbose 包，对应的 content 为 JSON 格式，content.msg_type =generate_answer_finish 代表全部 answer 回复完成。 <br>  <br> 仅发起对话（v3）接口支持将此参数作为入参，且： <br>  <br> * 如果 autoSaveHistory=true，type 支持设置为 question 或 answer。 <br> * 如果 autoSaveHistory=false，type 支持设置为 question、answer、function_call、tool_output、tool_response。 <br>  <br> 其中，type=question 只能和 role=user 对应，即仅用户角色可以且只能发起 question 类型的消息。详细说明可参考[消息 type 说明](https://docs.coze.cn/developer_guides/message_type)。 <br>  |
| bot_id | String | 737946218936519**** | 编写此消息的智能体 ID。此参数仅在对话产生的消息中返回。 |
| chat_id | String | 747946218936519**** | Chat ID。此参数仅在对话产生的消息中返回。 |
| section_id | String | 767946218936519**** | 上下文片段 ID。每次调用[清除上下文](https://docs.coze.cn/developer_guides/clear_conversation_context) API 都会生成一个新的 section_id。 |
| content | String | {"msg_type":"generate_answer_finish","data":"","from_module":null,"from_unit":null} | 消息的内容，支持纯文本、多模态（文本、图片、文件混合输入）、卡片等多种类型的内容。 <br> * 当 `role` 为 user 时，支持返回多模态内容。 <br> * 当 `role` 为 assistant 时，只支持返回纯文本内容。 <br>  |
| meta_data | JSON Map | { "uuid": "newid1234" } | 创建消息时的附加消息，[查看消息列表](https://docs.coze.cn/developer_guides/list_message)时也会返回此附加消息。 |
| created_at | Long | 1718592898 | 消息的创建时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| updated_at | Long | 1718592898 | 消息的更新时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| content_type | String | text | 消息内容的类型，取值包括： <br>  <br> * text：文本。 <br> * object_string：多模态内容，即文本和文件的组合、文本和图片的组合。 <br> * card：卡片。此枚举值仅在接口响应中出现，不支持作为入参。 |
| reasoning_content | String | 好的，我现在需要给一个13岁的大学生提供学习建议。首先，我得考虑用户的情况…… | 模型的思维链（CoT），展示模型如何将复杂问题逐步分解为多个简单步骤并推导出最终答案。仅当模型支持深度思考、且智能体开启了深度思考时返回该字段，当前支持深度思考的模型请参考[模型能力差异](https://docs.coze.cn/developer_guides/model_api_param_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v3/chat/submit_tool_outputs?chat_id=738137187639794****&conversation_id=738136585609548****' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
{
    "tool_outputs":[
        {
            "tool_call_id":"BUJJF0dAQ0NAEBVeQkVKEV5HFURFXhFCEhFeFxdHShcSQEtFSxY****",
            "output":"Tokyo"
        }
    ]
}
```

### 返回示例

<div type="doc-tabs">
<div type="tab-item" title="非流式返回示例" key="sGGopSW5aOsB9AYHBcWsQ">

```JSON
{
  "data": {
    "id": "738137187639794****",
    "conversation_id": "748348012449138***",
    "bot_id": "22248012449138***",
    "created_at": 1710348675,
    "completed_at": 1710348675,
    "last_error": null,
    "meta_data": {},
    "status": "completed",
    "usage": {
      "token_count": 3397,
      "output_tokens": 1173,
      "input_tokens": 2224
    }
  },
  "code": 0,
  "msg": ""
}
```


</div>
<div type="tab-item" title="需要使用方额外处理的对话" key="Jd9gQ_RcKWtpMW4CCUdfV">

```JSON
{
    "code": 0,
    "data": {
        "bot_id": "737282596785517****",
        "completed_at": 1717513285,
        "conversation_id": "737666232053956****",
        "created_at": 1717513283,
        // 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。
        "id": "737666232053959****",
        "required_action": {
            "submit_tool_outputs": {
                "tool_calls": [
                    {
                        "function": {
                            "arguments": "{\"location\":\"南京\",\"type\":0}",
                            "name": "local_data_assistant"
                        },
                        "id": "BUJJF0dAQ0NAEBVeQkVKEV5HFURFXhFCEhFeFxdHShcSQEtFSxYRSUI=",
                        "type": "function"
                    }
                ]
            },
            "type": "submit_tool_outputs"
        },
        "status": "requires_action"
    },
    "msg": ""
}
```


</div>
<div type="tab-item" title="流式响应示例" key="69AZAe-Sl0V6fb0DfMsZg">

```JSON
# chat - 开始
event: conversation.chat.created
// 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。
data: {"id": "123", "conversation_id":"123", "bot_id":"222", "created_at":1710348675,compleated_at:null, "last_error": null, "meta_data": {}, "status": "created","usage":null}
# chat - 处理中
event: conversation.chat.in_progress
data: {"id": "123", "conversation_id":"123", "bot_id":"222", "created_at":1710348675, compleated_at: null, "last_error": null,"meta_data": {}, "status": "in_progress","usage":null}
# MESSAGE - 知识库召回
event: conversation.message.completed
data: {"id": "msg_001", "role":"assistant","type":"knowledge","content":"---\nrecall slice 1:xxxxxxx\n","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
# MESSAGE - function_call
event: conversation.message.completed
data: {"id": "msg_002", "role":"assistant","type":"function_call","content":"{\"name\":\"toutiaosousuo-search\",\"arguments\":{\"cursor\":0,\"input_query\":\"今天的体育新闻\",\"plugin_id\":7281192623887548473,\"api_id\":7288907006982012986,\"plugin_type\":1","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
# MESSAGE - toolOutput
event: conversation.message.completed
data: {"id": "msg_003", "role":"assistant","type":"tool_output","content":"........","content_type":"card","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
# MESSAGE - answer is card
event: conversation.message.completed
data: {"id": "msg_004", "role":"assistant","type":"answer","content":"{{card_json}}","content_type":"card","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
# MESSAGE - answer is normal text
event: conversation.message.delta
data:{"id": "msg_005", "role":"assistant","type":"answer","content":"以下","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
event: conversation.message.delta
data:{"id": "msg_005", "role":"assistant","type":"answer","content":"是","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
...... {{ N 个 delta 消息包}} ......
event: conversation.message.completed
data:{"id": "msg_005", "role":"assistant","type":"answer","content":"{{msg_005 完整的结果。即之前所有 msg_005 delta 内容拼接的结果}}","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
# MESSAGE - 多 answer 的情况,会继续有 message.delta
event: conversation.message.delta
data:{"id": "msg_006", "role":"assistant","type":"answer","content":"你好你好","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
...... {{ N 个 delta 消息包}} ......
event: conversation.message.completed
data:{"id": "msg_006", "role":"assistant","type":"answer","content":"{{msg_006 完整的结果。即之前所有 msg_006 delta 内容拼接的结果}}","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
# MESSAGE - Verbose （流式 plugin, 多 answer 结束，Multi-agent 跳转等场景）
event: conversation.message.completed
data:{"id": "msg_007", "role":"assistant","type":"verbose","content":"{\"msg_type\":\"generate_answer_finish\",\"data\":\"\"}","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
# MESSAGE - suggestion
event: conversation.message.completed
data: {"id": "msg_008", "role":"assistant","type":"follow_up","content":"朗尼克的报价是否会成功？","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
event: conversation.message.completed
data: {"id": "msg_009", "role":"assistant","type":"follow_up","content":"中国足球能否出现？","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
event: conversation.message.completed
data: {"id": "msg_010", "role":"assistant","type":"follow_up","content":"羽毛球种子选手都有谁？","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
# chat - 完成
event: conversation.chat.completed （chat完成）
data: {"id": "123", "chat_id": "123", "conversation_id":"123", "bot_id":"222", "created_at":1710348675, compleated_at:1710348675, "last_error":null, "meta_data": {}, "status": "compleated", "usage":{"token_count":3397,"output_tokens":1173,"input_tokens":2224}}
event: done （stream流结束）
data: [DONE]
# chat - 失败
event: conversation.chat.failed
data: {
    "code":701231,
    "msg":"error"
}
```


</div>
</div>
## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。