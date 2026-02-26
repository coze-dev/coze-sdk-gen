# 提交消息评价
用户可以对智能体或应用回复的消息提交评价，包括点赞、点踩和反馈详细的评论。
## 接口描述
无论是“一问一答”还是“一问多答”场景，每条消息的评价都是独立的，不会相互覆盖或影响。如果用户对同一条消息多次提交评价，系统将仅保留最后一次的评价内容。
开发者可以在扣子罗盘的消息日志中查看消息的评价数据，以便改进智能体或应用。
## 接口限制

* 仅会话创建者能评价对应会话中的消息。
* 仅支持评价以下来源的文本消息:
   * 通过发起对话 API 生成的 **type=answer** 类型的文本消息。
   * 通过执行对话流 API 返回的文本消息。
* 仅支持评价 2025 年 7 月 11 日 0 时之后的消息。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/conversations/:conversation_id/messages/:message_id/feedback <br> ``` <br>  |
| **权限** | `feedback` <br> 确保调用该接口使用的访问令牌开通了 `feedback` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 对智能体或应用回复的消息提交评价。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| conversation_id | String | 必选 | 737363834493434**** | Conversation ID，即会话的唯一标识。你可以在[发起对话](https://docs.coze.cn/developer_guides/chat_v3) API 返回的 Response 中查看 conversation_id 字段。 |
| message_id | String | 必选 | 74076856761178**** | 针对指定的消息 ID 提交评价。你可以通过[查看对话消息详情](https://docs.coze.cn/developer_guides/list_chat_messages) API 返回的 Response 中查看消息 ID。 <br>  <br> * 此消息必须在 conversation_id 指定的会话中。 <br> * 仅支持评价以下来源的文本消息： <br>    * 通过发起对话 API 生成的 **type=answer** 类型的文本消息。 <br>    * 通过执行对话流 API 返回的文本消息。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| feedback_type | String | 必选 | like | 对智能体或应用回复的消息提交评价，包括以下枚举值： <br>  <br> * `like`：点赞。 <br> * `unlike`：点踩。 |
| reason_types | Array of String | 可选 | ["事实性错误","逻辑矛盾","格式混乱"] | 用户自定义的反馈标签列表，用于分类标识对智能体回复的具体不满类型。每个标签应为简明扼要的分类描述，如问题类型等。 <br> 最多可添加 10 个标签，单个标签长度不超过 30 个字符。 |
| comment | String | 可选 | 智能体关于产品功能的回答存在错误，实际参数应为 5.0 版本而非 4.0 版本。 | 用户对智能体回复消息的具体评价，最多可输入 250 个字符。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细日志信息，用于问题排查。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/conversations/7515364893***/messages/752173729302***/feedback' \
--header 'Authorization: Bearer pat_xFWpGsNio4S7sfAzpu02vHCkAdL38VnSsTOIu8CkySdY9Z2xmeM8jjn***' \
--header 'Content-Type: application/json'
{
    "feedback_type": "unlike",
    "reason_types": [
        "内容有误",
        "不够详细"
    ],
    "comment": "实际参数应为 5.0 版本而非 4.0 版本。"
}
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20241210191248C8EF760***"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。