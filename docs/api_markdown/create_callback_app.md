# 创建回调应用
本 API 用于创建回调应用，支持创建普通回调应用和渠道回调应用。订阅扣子编程回调功能时需要创建回调应用。
## 接口说明
[订阅回调](https://docs.coze.cn/dev_how_to_guides/add_callback)功能支持开发者通过配置回调应用实时获取扣子编程的事件通知。当**智能体发布**、**智能体删除**、**账单生成**等关键业务事件被触发时，扣子编程将向开发者指定的服务器地址发送回调消息。
回调分为普通回调和渠道回调，具体说明如下：

* 普通回调应用：开发者在扣子编程中创建回调应用，用于接收扣子编程触发的事件通知。当订阅的事件被触发时，扣子编程会向该回调地址推送回调消息。
* 渠道回调应用：当渠道入驻扣子编程后，开发者可以在该渠道中创建回调应用，用于接收该渠道中触发的事件通知。当订阅的事件被触发时，扣子编程会向渠道指定的回调地址推送回调消息。

## 接口限制

* 扣子个人版中，任何用户均可以创建普通回调应用。仅渠道创建者支持创建对应渠道的回调应用，统一接收该渠道中的回调消息。
* 扣子企业版（企业标准版、企业旗舰版）中，仅超级管理员和管理员可创建回调应用。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/api_apps <br> ``` <br>  |
| **权限** | `createApiApp` <br> 确保调用该接口使用的访问令牌开通了 `createApiApp` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 创建回调应用。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| app_type | String | 必选 | normal | 回调应用的类型。枚举值如下： <br>  <br> * `normal`：普通回调应用。 <br> * `connector`：渠道回调应用。 |
| name | String | 可选 | 订单回调应用 | 回调应用的名称，当 `app_type` 为 `normal`时需要传入该值。最多 128 个字符。 <br> 当 `app_type` 为 `connector`时无需填写，扣子编程默认使用渠道名称作为渠道回调应用的名称。 |
| connector_id | String | 可选 | 1056899*** | 渠道的 ID，当 `app_type` 为 `connector`时需要传入该值。 <br> 渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [ApiApp](#apiapp) | {"id":"74876004423701****","name":"智能客服机器人回调","app_type":"normal","callback_url":"https://example.com/api/callback","connector_id":"1056899***","verify_token":"abc123def456ghi789"} | 返回扣子编程应用的详细信息，包括应用 ID、名称、类型、回调 URL 和验证令牌等。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的详细日志信息，用于问题排查。如果遇到异常报错场景，且反复重试仍然报错，可以根据此日志 ID 联系扣子团队获取帮助。 |
### ApiApp
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 7512045450401153075 | 回调应用的 ID。 |
| name | String | 智能客服机器人回调 | 回调应用的名称。 |
| app_type | String | normal | 回调应用的类型。枚举值如下： <br>  <br> * `normal`：普通回调应用。 <br> * `connector`：渠道回调应用。 |
| callback_url | String | https://example.com/api/callback | 回调地址。后续该回调应用订阅的所有回调，均会在触发时向该回调地址发送回调数据。 |
| connector_id | String | 1056899*** | 自定义的渠道 ID。仅渠道回调应用会返回该参数。 |
| verify_token | String | OYDacMzM3WyOWV3Dtj2bHRMymzxP**** | 扣子编程会为每个回调应用自动生成一个 Token，不支持手动修改或删除。 <br> 当扣子编程推送回调时，会携带 Token 签发的签名，用于验证推送的回调是否属于扣子编程推送的合法回调。开发者验证签名的具体操作请参见[接收并处理回调](https://docs.coze.cn/dev_how_to_guides/receive_handle_callbacks)。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
POST 'https://api.coze.cn/v1/api_apps' \
--header 'Authorization: Bearer czs_hh19tcCx3qn05X7bG71AGzyD2DjrunJqaAeHORjqZCUydGLRqFhSXWr0SJa******' \
--header 'Content-Type: application/json'
{"app_type": "normal", "name": "sample api app"}
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "id": "752195368553526****",
    "name": "sample api app",
    "app_type": "normal",
    "verify_token": "XzHka769anqJdSfyndM5VM6jPTg2Lcel5YVBHMjfczpQhOR5e6taOt2y12z*****"
  },
  "detail": {
    "logid": "20250714193129A38DEB6D6****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。