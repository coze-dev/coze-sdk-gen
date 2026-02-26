# 取消进行中的对话
调用此接口取消进行中的对话。
调用[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口时，如果对话触发了复杂的工作流、图像流，或模型处理数据量大时，对话可能耗时较久。对话进行中时，用户无法在此会话中发起新的对话。此时可以调用此接口取消正在进行中的对话，取消后，对话转为已取消状态（status=canceled），你可以在会话中创建新的对话。
## 注意事项

* 调用取消对话 API 仅切换对话状态，不会中断 chat API 的流式回复，同时根据完整回复内容来计算消耗的模型 Token。如需中断流式回复、停止打印机效果，可以在调用此 API 之后主动中断客户端连接，例如调用浏览器 Web API `AbortController`。
* 取消对话后，本轮用户的 Query 和智能体的回复不会记录为对话的上下文。
* 不支持取消以下状态的对话。你可以通过[查看对话详情](https://docs.coze.cn/developer_guides/retrieve_chat)接口的 status 字段查看对话状态。
   * completed：会话结束。 
   * failed：会话失败。 
   * requires_action：会话中断。

对话过程中的状态流转：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/d7c3509b18454159ab3cfb89e7e7b222~tplv-goo7wpa0wc-image.image)

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v3/chat/cancel <br> ``` <br>  |
| **权限** | `cancelChat` <br> 确保调用该接口使用的访问令牌开通了 `cancelChat` 权限，详细信息可参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用此接口取消进行中的对话。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer ***$Access_Token*** | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| Content-Type | application/json | 请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| chat_id | String | 必选 | 7398048669188**** | 对话 ID，即 Chat ID。可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口 Response 中查看 id 字段，如果是流式响应，则在 Response 的 chat 事件中查看 id 字段。 |
| conversation_id | String | 必选 | 7397787494399**** | 会话 ID，即 Conversation ID。可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口 Response 中查看 conversation_id 字段。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [ChatV3ChatDetail](#chatv3chatdetail) | 详见响应示例 | 被取消对话的详细信息。详细说明可参考 **ChatV3ChatDetail**。 |
| detail | Object of [ResponseDetail](#responsedetail) | 详见响应示例 | 详细信息。 |
| code | Long | 0 | 调用状态码。 <br>  <br> * 0 表示调用成功。 <br> * 其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
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

## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v3/chat/cancel' \
--header 'Authorization: Bearer pat_hfwkehfncaf****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "chat_id": "73980486691***",
    "conversation_id": "739778749439***"
}'
```

### 返回示例
```JSON
{
  "data": {
    "id": "738137187639794****",
    "usage": {},
    "bot_id": "737946218936519****",
    "status": "completed",
    "failed_at": null,
    "meta_data": {},
    "created_at": 1718609571,
    "last_error": {
      "msg": "",
      "code": null
    },
    "section_id": "",
    "completed_at": 1718609575,
    "conversation_id": "739778749439***",
    "required_action": {
      "type": "",
      "submit_tool_outputs": {
        "tool_calls": [
          {
            "id": "",
            "type": "",
            "function": {
              "name": "",
              "arguments": ""
            }
          }
        ]
      }
    }
  },
  "detail": {
    "logid": "20241210152726467C48D89D6DB2F37A23"
  },
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

