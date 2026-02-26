# 修改消息
修改一条消息，支持修改消息内容、附加内容和消息类型。
## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/conversation/message/modify <br> ``` <br>  |
| **权限** | `modifyMessage` <br> 确保调用该接口使用的访问令牌开通了 `modifyMessage` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 修改一条消息，支持修改消息内容、附加内容和消息类型。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| conversation_id | Integer | 必选 | 737475200011611**** | Conversation ID，即会话的唯一标识。可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口 Response 中查看 conversation_id 字段。 |
| message_id | Integer | 必选 | 737492830742885*** | Message ID，即消息的唯一标识。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| content | String | 可选 | 早上好，今天深圳天气怎么样？ | 消息的内容，支持纯文本、多模态（文本、图片、文件混合输入）、卡片等多种类型的内容。 <br> meta_data 和 content 不能同时为空。 <br>  |
| content_type | String | 可选 | text | 消息内容的类型。取值包括： <br>  <br> * text：文本 <br> * object_string：多模态内容，即文本和文件的组合、文本和图片的组合 <br>  <br> 传入 content 时，应同时指定 content_type。 <br>  |
| meta_data | JSON Map | 可选 |  {} | 创建消息时的附加消息，获取消息时也会返回此附加消息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 <br> meta_data 和 content 不能同时为空。 <br>  |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| message | Object of [OpenMessageApi](#openmessageapi) | { "bot_id": "", "chat_id": "", "content": "早上好，今天深圳天气怎么样？", "content_type": "text", "conversation_id": "737999610479815****", "created_at": 1718595518, "id": "738131140595115****", "meta_data": {}, "role": "user", "type": "", "updated_at": "1718595677" } | 修改后的消息详细信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFA***" } | 响应详情。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
### OpenMessageApi
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 738130009748252**** | Message ID，即消息的唯一标识。 |
| conversation_id | String | 737999610479815**** | 此消息所在的会话 ID。 |
| bot_id | String | 747363834493437**** | 编写此消息的智能体 ID。此参数仅在对话产生的消息中返回。 |
| chat_id | String | 757363834493437**** | Chat ID。此参数仅在对话产生的消息中返回。 <br> 不同的对话中，系统会生成新的`chat_id`。同一个用户在第一次对话和第二次对话时，`chat_id`不一样。 |
| meta_data | JSON Map | {"source":"mobile_app","location":"Beijing","custom_flag":"high_priority"} | 创建消息时的附加信息，以键值对形式存储，获取消息时会原样返回。可用于存储与业务相关的自定义数据。 |
| role | String | user | 发送这条消息的实体。取值： <br>  <br> * **user**：代表该条消息内容是用户发送的。 <br> * **assistant**：代表该条消息内容是 Bot 发送的。 |
| content | String | 早上好，今天星期几？ | 消息的内容，支持纯文本、多模态（文本、图片、文件混合输入）、卡片等多种类型的内容。 |
| content_type | String | text | 消息内容的类型，取值包括： <br>  <br> * text：文本。 <br> * object_string：多模态内容，即文本和文件的组合、文本和图片的组合。 <br> * card：卡片。此枚举值仅在接口响应中出现，不支持作为入参。 |
| created_at | Long | 1718592898 | 消息的创建时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| updated_at | Long | 1718592898 | 消息的更新时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| type | String | question | 消息类型。 <br>  <br> * **question**：用户输入内容。 <br> * **answer**：智能体返回给用户的消息内容，支持增量返回。如果工作流绑定了 messge 节点，可能会存在多 answer 场景，此时可以用流式返回的结束标志来判断所有 answer 完成。 <br> * **function_call**：智能体对话过程中调用函数（function call）的中间结果。 <br> * **tool_response**：调用工具 （function call）后返回的结果。 <br> * **follow_up**：如果在 Bot 上配置打开了用户问题建议开关，则会返回推荐问题相关的回复内容。 <br> * **verbose**：多 answer 场景下，服务端会返回一个 verbose 包，对应的 content 为 JSON 格式，`content.msg_type =generate_answer_finish` 代表全部 answer 回复完成。 <br>  <br> 仅发起对话（v3）接口支持将此参数作为入参，且： <br>  <br> * 如果 autoSaveHistory=true，type 支持设置为 question 或 answer。 <br> * 如果 autoSaveHistory=false，type 支持设置为 question、answer、function_call、tool_output、tool_response。 <br>  <br> 其中，type=question 只能和 role=user 对应，即仅用户角色可以且只能发起 question 类型的消息。详细说明可参考[消息 type 说明](https://docs.coze.cn/developer_guides/message_type)。 <br>  |
| section_id | String | 757363834493437**** | 上下文片段 ID。每次清除上下文都会生成一个新的 section_id。 |
| reasoning_content | String | 好的，我现在需要给一个13岁的大学生提供学习建议。首先，我得考虑用户的情况…… | 模型的思维链（CoT），展示模型如何将复杂问题逐步分解为多个简单步骤并推导出最终答案。仅当模型支持深度思考、且智能体开启了深度思考时返回该字段，当前支持深度思考的模型请参考[模型能力差异](https://docs.coze.cn/developer_guides/model_api_param_support)。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
POST https://api.coze.cn/v1/conversation/message/modify?conversation_id=737475200011611****&message_id=737492830742885****
{
    "content": "早上好，今天深圳天气怎么样？",
    "content_type": "text"
}
```

### 返回示例
```JSON
{
  "code": 0,
  "message": {
    "bot_id": "",
    "chat_id": "",
    "content": "早上好，今天深圳天气怎么样？",
    "content_type": "text",
    "conversation_id": "737999610479815****",
    "created_at": 1718595518,
    "id": "738131140595115****",
    "meta_data": {},
    "role": "user",
    "type": "",
    "updated_at": "1718595677"
  },
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
