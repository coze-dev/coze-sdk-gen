# 更新会话名称
会话创建者可以更新指定会话的会话名，以便更好地识别和管理会话。
在创建会话时，扣子编程会默认将用户发送的第一个消息内容作为会话名称。用户也可以根据会话的实际内容或主题对会话名称进行更新，从而更直观地区分不同的会话，提升管理效率。
## 基础信息
| **请求方式** | PUT |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/conversations/:conversation_id <br> ``` <br>  |
| **权限** | `editConversation` <br> 确保调用该接口使用的访问令牌开通了 `editConversation` 权限，详细信息参考[准备工作](https://www.coze.com/docs/developer_guides/preparation)。 |
| **接口说明** | 更新指定会话的会话名称。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| conversation_id | String | 必选 | 737989918257**** | 待更新的会话 ID。你可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口的 Response 中通过 `conversation_id` 字段获取会话 ID。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| name | String | 必选 | 推荐杭州美食 | 更新后的会话名称，最多支持 100 个字符。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [ConversationData](#conversationdata) | { "creator_id": "14264794847***", "connector_id": "1024", "created_at": 1749439208, "id": "7513776737285***", "meta_data": {}, "updated_at": 1753326053, "last_section_id": "7513776737285570569", "name": "推荐杭州美食" } | 包含会话详细信息的对象。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### ConversationData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 737999610479815**** | Conversation ID，即会话的唯一标识。 |
| name | String | 推荐杭州美食 | 会话的名称，用于标识和区分不同的会话。 |
| meta_data | JSON Map | { "uuid": "newid1234" } | 附加信息，通常用于封装一些业务相关的字段。[查看会话信息](https://docs.coze.cn/developer_guides/retrieve_conversation)时，扣子编程会透传此附加信息，[查看消息列表](https://docs.coze.cn/developer_guides/list_message)时不会返回该附加信息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 |
| creator_id | String | 247877439325**** | 会话创建者的扣子 UID，用于标识创建该会话的用户。 |
| created_at | Long | 1718289297 | 会话创建的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| updated_at | Long | 1718289297 | 会话的最后更新时间，格式为 10 位的 Unix 时间戳，单位为秒。 |
| last_section_id | String | 749566434761695*** | 会话中最新的一个上下文片段 ID。 |
| connector_id | String | 1024 | 该会话在哪个渠道创建。目前支持如下渠道： <br>  <br> * API：1024 <br> * ChatSDK：999 <br> * 自定义渠道：自定义渠道 ID。自定义渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request PUT 'https://api.coze.cn/v1/conversations/737999610479815****' \
--header 'Authorization: Bearer pat_xitq9LWlowpX3qGCih1lwpAdzvXNqgmpfhpV28HLWFypY37xR5Uaj2GioN****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "推荐杭州美食"
}'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "202507311334280ADC900DF8FCF***"
  },
  "data": {
    "creator_id": "14264794847***",
    "connector_id": "1024",
    "created_at": 1749439208,
    "id": "7513776737285***",
    "meta_data": {},
    "updated_at": 1753326053,
    "last_section_id": "751377673728557***",
    "name": "推荐杭州美食"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

