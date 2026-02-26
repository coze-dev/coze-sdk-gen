# 创建会话
创建一个会话。
## 接口描述
会话是智能体和用户之间的基于一个或多个主题的问答交互，一个会话包含一条或多条消息。创建会话时，扣子编程会同时在会话中创建一个空白的上下文片段，用于存储某个主题的消息。后续发起对话时，上下文片段中的消息是模型可见的消息历史。
你可以在创建会话时同步写入消息，或者创建会话后另外调用 [创建消息 API](https://docs.coze.cn/developer_guides/create_message) 写入消息。消息默认写入到最新的一个上下文片段中，对话时将作为上下文传递给模型。
会话、对话、消息和上下文片段的概念说明，可参考[基础概念](https://docs.coze.cn/developer_guides/coze_api_overview#4a288f73)。
如果需要将不同业务侧用户的会话互相隔离，请参见[如何实现会话隔离](https://docs.coze.cn/developer_guides/session_isolation)。
## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/conversation/create <br> ``` <br>  |
| **权限** | `createConversation` <br> 确保调用该接口使用的访问令牌开通了 `createConversation` 权限，详细信息参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| **接口说明** | 创建一个会话。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| bot_id | String | 可选 | 730454116184516* | 会话对应的智能体 ID。 <br> 创建会话时指定智能体 ID，后续才能通过 [查看会话列表](https://docs.coze.cn/developer_guides/list_conversations) 查看指定智能体 ID 对应的会话列表。 <br> 智能体 ID，获取方法如下： <br> 进入智能体的 开发页面，开发页面 URL 中 `bot` 参数后的数字就是智能体 ID。例如`https://www.coze.cn/space/341****/bot/73428668*****`，智能体 ID 为`73428668*****`。 |
| name | String | 可选 | 推荐杭州美食 | 会话的名称。默认为空。为了方便区分不同会话，建议设置会话名称。最多支持 100 个字符。 |
| messages | Array of [EnterMessage](#entermessage) | 可选 | [ { "role": "user", "content":"[{"type":"text","text":"你好，这是我的图片"},{"type":"image","file_id":"145657573***"}]", "content_type":"object_string" }, { "role": "assistant", "content": "你好我是一个bot", "content_type":"text" } ] | 会话中的消息内容。详细说明可参考 [EnterMessage object](https://docs.coze.cn/developer_guides/create_conversation#cde8cc95)。 |
| connector_id | String | 可选 | 1024 | 该会话在哪个渠道创建。目前支持如下渠道： <br>  <br> * API：（默认）1024 <br> * ChatSDK：999 <br> * 自定义渠道：自定义渠道 ID。自定义渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 |
| meta_data | JSON Map | 可选 | { "uuid": "newid1234" } | 附加信息，通常用于封装一些业务相关的字段。[查看会话信息](https://docs.coze.cn/developer_guides/retrieve_conversation)时，扣子编程会透传此附加信息，[查看消息列表](https://docs.coze.cn/developer_guides/list_message)时不会返回该附加信息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 |
### EnterMessage
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| role | String | 必选 | user | 发送这条消息的实体。取值： <br>  <br> * **user**：代表该条消息内容是用户发送的。 <br> * **assistant**：代表该条消息内容是 Bot 发送的。 |
| type | String | 可选 | question | 消息类型。默认为 **question。** <br>  <br> * **question**：用户输入内容。 <br> * **answer**：Bot 返回给用户的消息内容，支持增量返回。如果工作流绑定了输出节点，可能会存在多 answer 场景，此时可以用流式返回的结束标志来判断所有 answer 完成。 <br> * **function_call**：Bot 对话过程中调用函数（function call）的中间结果。 <br> * **tool_response**：调用工具 （function call）后返回的结果。 <br> * **follow_up**：如果在 Bot 上配置打开了用户问题建议开关，则会返回推荐问题相关的回复内容。不支持在请求中作为入参。 <br> * **verbose**：多 answer 场景下，服务端会返回一个 verbose 包，对应的 content 为 JSON 格式，`content.msg_type =generate_answer_finish` 代表全部 answer 回复完成。不支持在请求中作为入参。 <br>  <br> 仅发起会话（v3）接口支持将此参数作为入参，且： <br>  <br> * 如果 autoSaveHistory=true，type 支持设置为 question 或 answer。 <br> * 如果 autoSaveHistory=false，type 支持设置为 question、answer、function_call、tool_output/tool_response。 <br>  <br> 其中，type=question 只能和 role=user 对应，即仅用户角色可以且只能发起 question 类型的消息。详细说明可参考[消息 type 说明](https://docs.coze.cn/developer_guides/message_type)。 <br>  |
| content | String | 可选 | [{"type":"text","text":"你好，这是我的图片"},{"type":"image","file_id":"145657573***"}] | 消息的内容，支持纯文本、多模态（文本、图片、文件混合输入）、卡片等多种类型的内容。 <br>  <br> * content_type 为 object_string 时，content 为 **object_string object** 数组序列化之后的 JSON String，详细说明可参考 **object_string object。** <br> * 当 content_type **=** text 时，content 为普通文本，例如 `"content" :"Hello!"`。 |
| meta_data | JSON Map | 可选 | { "uuid": "newid1234" } | 创建消息时的附加消息，获取消息时也会返回此附加消息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 <br> content 不为空时，此参数为必选。 <br>  |
| content_type | String | 可选 | text | 消息内容的类型，支持设置为： <br>  <br> * text：文本。 <br> * object_string：多模态内容，即文本和文件的组合、文本和图片的组合。 <br> * card：卡片。此枚举值仅在接口响应中出现，不支持作为入参。 <br>  <br> content 不为空时，此参数为必选。 <br>  <br>  |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [ConversationData](#conversationdata) | \ | 会话的基础信息，包含会话的唯一标识、名称、创建时间等详细信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
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

<div type="doc-tabs">
<div type="tab-item" title="创建空会话" key="FaeXPDQ891fBoStvaL8-I">

```JSON
curl --location --request POST 'https://api.coze.cn/v1/conversation/create'\
--header 'Authorization: Bearer pat_xitq9LWlowpX3qGCih1lwpAdzvXNqgmpfhpV28HLWFypY37xR5Uaj2GioN****' \
--header 'Content-Type: application/json' \
```


</div>
<div type="tab-item" title="创建会话和消息" key="JkiHBUOY_qA1KVPRW6GN1">

创建会话，并在会话中携带消息内容。
```JSON
curl --location --request POST 'https://api.coze.cn/v1/conversation/create' \
--header 'Authorization: Bearer pat_xitq9LWlowpX3qGCih1lwpAdzvXNqgmpfhpV28HLWFypY37xR5Uaj2GioN****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "meta_data": {
        "uuid": "newid1234"
    },
    "messages": [
        {
            "role": "user",
            "type": "question",
            "content": "[{\"type\":\"text\",\"text\":\"你好，这是我的图片\"},{\"type\":\"image\",\"file_id\":\"145657573***\"}]",
            "meta_data": {
                "uuid": "newid1234"
            },
            "content_type": "text"
        }
    ],
    "bot_id": "730454116184516*",
    "connector_id": "1024",
    "name": "推荐杭州美食"
}'
```


</div>
</div>
### 返回示例
```JSON
{
  "data": {
    "id": "737999610479815****",
    "name": "推荐杭州美食",
    "meta_data": {
      "uuid": "newid1234"
    },
    "created_at": 1718289297,
    "creator_id": "247877439325****",
    "connector_id": "1024",
    "updated_at": 1718289297,
    "last_section_id": "7495664347616952360"
  },
  "detail": {
    "logid": "20241210152726467C48D89D6DB2****"
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。