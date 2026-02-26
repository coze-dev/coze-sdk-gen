# 清除上下文
清除指定会话中的上下文。
## 接口说明
在智能体对话场景中，上下文指模型在处理当前输入时所能参考的历史消息、对话记录，它决定了模型对当前任务的理解深度和连贯性，直接影响输出结果的准确性和相关性。多轮对话中，如果需要开启新的话题，可以调用此 API 清除上下文。清除上下文后，对话中的历史消息不会作为上下文被输入给模型，后续对话不再受之前历史消息的影响，避免信息干扰，确保对话的准确性和连贯性。
会话中的消息存储在上下文段落（section）中，每次调用此 API 清除上下文时，系统会自动删除旧的上下文段落，并创建新的上下文段落用于存储新的消息。再次发起对话时，新分区中的消息会作为新的上下文传递给模型参考。
* 清除上下文 API 只是清空模型可见的消息历史，不会实际删除会话中的消息，通过[查看消息列表](https://docs.coze.cn/developer_guides/list_message)或[查看消息详情](https://docs.coze.cn/developer_guides/retrieve_message)等 API 仍能查看到消息内容。
* 仅支持清除本人创建的会话的上下文。
* 会话、对话、消息和上下文段落的术语解释请参见[基础概念](https://docs.coze.cn/developer_guides/coze_api_overview#fed4a54c)。

![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/e4b55816254c4446ae59bbca33ca8e1d~tplv-goo7wpa0wc-image.image)

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/conversations/:conversation_id/clear <br> ``` <br>  |
| **权限** | `editConversation` <br> 确保调用该接口使用的访问令牌开通了 `editConversation` 权限，详细信息参考[准备工作](https://www.coze.com/docs/developer_guides/preparation)。 |
| **接口说明** | 清除指定会话中的上下文。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| conversation_id | String | 必选 | 737989918257**** | 待清除上下文的会话 ID。你可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口的 Response 中通过 conversation_id 字段获取会话 ID。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 `msg` 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [Section](#section) | { "id": "75331946251175***", "conversation_id": "737989918257****" } | 上下文段落（section）的详细信息。 <br> Section 是一个独立的上下文段落，用于分隔不同的对话阶段或主题。Section 中包括上下文消息，当用户清除上下文时，系统会创建一个新的 Section，从而确保新的对话不受历史消息的影响。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 返回的详情信息，包含本次请求的日志 ID。如果遇到异常报错场景，可以根据此日志 ID 联系技术支持获取帮助。 |
### Section
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 75331946251175*** | 新的 Session ID，即清除上下文后新创建的上下文段落（section）的 ID。 <br> 每个上下文段落对应一批独立的上下文消息。每次清除上下文时，系统会新建一个上下文段落用于存储新的上下文消息。 |
| conversation_id | String | 737999610479815**** | Conversation ID，即会话的唯一标识。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/conversations/737989918257****/clear' \
--header 'Authorization: Bearer pat_xitq9LWlowpX3qGCih1lwpAdzvXNqgmpfhpV28HLWFypY37xR5Uaj2GioN****' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "id": "75331946251175***",
    "conversation_id": "737989918257****"
  },
  "detail": {
    "logid": "202411123456789****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。