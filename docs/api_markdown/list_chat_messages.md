# 查看对话消息详情
查看指定对话中除 Query 以外的其他消息，包括模型回复、智能体执行的中间结果等消息。
## 接口描述
**查看消息列表** API 与**查看对话消息详情** API 的区别在于：

* **查看消息列表** API 用于查询指定会话（conversation）中的消息记录，不仅包括开发者在会话中手动插入的每一条消息和用户发送的 Query，也包括调用**发起对话** API 得到的 type=answer 的智能体回复，但不包括 type=function_call、tool_response 和 follow-up 类型的对话中间态消息。
* **查看对话消息详情** API 通常用于非流式对话场景中，查看某次对话（chat）中 type=answer 的智能体回复及 type=function_call、tool_response 和 follow-up 类型类型的对话中间态消息。不包括用户发送的 Query。

调用此 API 之前，建议先以每秒最多 1 次的频率轮询 [查看对话详情](https://docs.coze.cn/developer_guides/retrieve_chat) API 确认本轮对话已结束（status=completed），否则调用此 API 时获取到的消息内容可能不完整。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v3/chat/message/list <br> ``` <br>  |
| **权限** | `chat`、`listMessage` <br> 确保调用该接口使用的访问令牌开通了 `chat` 和`listMessage`权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查看指定对话中除 Query 以外的其他消息，包括模型回复、智能体执行的中间结果等消息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| conversation_id | String | 必选 | 738216760624714**** | Conversation ID，即会话的唯一标识。可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口 Response 中查看 conversation_id 字段。 |
| chat_id | String | 必选 | 738147352534297**** | Chat ID，即对话的唯一标识。可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口 Response 中查看 id 字段，如果是流式响应，则在 Response 的 chat 事件中查看 id 字段。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Array of [ChatV3MessageDetail](#chatv3messagedetail) | [ { "bot_id": "737946218936519****", "content": "{\"msg_type\":\"generate_answer_finish\",\"data\":\"\",\"from_module\":null,\"from_unit\":null}", "content_type": "text", "conversation_id": "738147352534297****", "id": "738216762080970****", "role": "assistant", "type": "verbose" }] | 指定对话中除 Query 以外的其他消息，包括模型回复、智能体执行的中间结果等消息。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFAD653960" } | 响应详情信息。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
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
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v3/chat/message/list?conversation_id=738216760624714****&chat_id=738147352534297****' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
```

### 返回示例

<div type="doc-tabs">
<div type="tab-item" title="基础对话" key="Jaat7kEPjOnvOkdRVg5zt">

```JSON
{
  "code": 0,
  "data": [
    {
      "bot_id": "737946218936519****",
      "content": "{\"msg_type\":\"generate_answer_finish\",\"data\":\"\",\"from_module\":null,\"from_unit\":null}",
      "content_type": "text",
      "conversation_id": "738147352534297****",
      "id": "738216762080970****",
      "role": "assistant",
      "type": "verbose"
    },
    {
      "bot_id": "7379462189365198898",
      "content": "2024 年 10 月 1 日是星期二。您可以通过日历或者相关的日期查询工具来核实确认。 ",
      "content_type": "text",
      "conversation_id": "738147352534297****",
      "id": "738216760624724****",
      "role": "assistant",
      "type": "answer"
    }
  ],
  "msg": "",
  "detail": {
    "logid": "20250106172024B5F607030EFF***"
  }
}
```


</div>
<div type="tab-item" title="图文对话" key="mxtukDred8X5AdSHUP9kf">

```JSON
{
  "code": 0,
  "data": [
    {
      "bot_id": "737946218936519****",
      "content": "{\"name\":\"ts-byteartist-text2image\",\"arguments\":{\"prompt\":\"美丽的风景照\"},\"plugin_id\":7257418203524284472,\"api_id\":7288904268684378171,\"plugin_type\":1,\"thought\":\"需求为生成一张美丽的风景照，需要调用ts-byteartist-text2image工具进行生成\"}",
      "content_type": "text",
      "conversation_id": "738147352534297****",
      "id": "738137425881312****",
      "role": "assistant",
      "type": "function_call"
    },
    {
      "bot_id": "737946218936519****",
      "content": "{\"code\":0,\"msg\":\"success\",\"data\":{\"images\":[{\"image_url\":\"https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/e08b786baa4c4889ad225bc40244c25c.png\"}]},\"log_id\":\"20240617154200D83175443F722F124A9C\"}",
      "content_type": "text",
      "conversation_id": "738136585609548****",
      "id": "738137428990112****",
      "role": "assistant",
      "type": "tool_response"
    },
    {
      "bot_id": "737946218936519****",
      "content": "{\"msg_type\":\"generate_answer_finish\",\"data\":\"\",\"from_module\":null,\"from_unit\":null}",
      "content_type": "text",
      "conversation_id": "738147352534297****",
      "id": "738137431378355****",
      "role": "assistant",
      "type": "verbose"
    },
    {
      "bot_id": "7379462189365198898",
      "content": "为您生成了一张美丽的风景照：[图片链接](https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/e08b786baa4c4889ad225bc40244c25c.png)",
      "content_type": "text",
      "conversation_id": "738147352534297****",
      "id": "738137425881310****",
      "role": "assistant",
      "type": "answer"
    }
  ],
  "msg": "",
  "detail": {
    "logid": "20250106172024B5F607030EFFAD653960"
  }
}
```


</div>
</div>
## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。