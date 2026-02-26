# 删除消息评价
删除指定消息的评价。
## 接口限制
仅会话创建者能删除对应会话中消息的评价。
## 基础信息
| **请求方式** | DELETE |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/conversations/:conversation_id/messages/:message_id/feedback <br> ``` <br>  |
| **权限** | `feedback` <br> 确保调用该接口使用的访问令牌开通了 `feedback` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 删除指定消息的评价。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| conversation_id | String | 必选 | 737363834493434**** | Conversation ID，即会话的唯一标识。可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3)接口 Response 中查看 conversation_id 字段。 |
| message_id | String | 必选 | 74076856761178**** | 待删除评价的消息 ID。你可以通过[查看对话消息详情](https://docs.coze.cn/developer_guides/list_chat_messages) API 返回的 Response 中查看消息 ID。 <br>  <br> * 此消息必须在 conversation_id 指定的会话中。 <br> * 仅支持评价以下来源的文本消息： <br>    * 通过发起对话 API 生成的 **type=answer** 类型的文本消息。 <br>    * 通过执行对话流 API 返回的文本消息。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细日志信息，用于问题排查和技术支持。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
DELETE https://api.coze.cn/v1/conversations/751536489399***/messages/751903981556***/feedback \
--header 'Authorization: Bearer pat_xFWpGsNio4S7sfAzpu02vHCkAdL38VnSsTOIu8CkySdY9Z2xmeM8jjn****' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20241210191248C8EF76****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
