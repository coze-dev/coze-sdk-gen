# 删除回调应用
删除回调应用。
如果不再需要接收扣子编程回调，可以删除回调应用。一旦删除，该回调应用所订阅的所有回调事件触发时，将不再向该回调应用推送通知。
## 接口限制

* 扣子个人版中，仅支持删除本人创建的回调应用。
* 扣子企业版中，仅超级管理员和管理员可删除回调应用。

## 基础信息
| **请求方式** | DELETE |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/api_apps/:api_app_id <br> ``` <br>  |
| **权限** | `deleteApiApp` <br> 确保调用该接口使用的访问令牌开通了 `deleteApiApp` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 删除回调应用。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| api_app_id | String | 必选 | 75268815369567***** | 待删除的回调应用的 ID。你可以通过[查询回调应用列表](https://docs.coze.cn/developer_guides/list_callback_app) API 获取回调应用的 ID。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
DELETE 'https://api.coze.cn/v1/api_apps/75268815369567*****' \
--header 'Authorization: Bearer pat_Elfvzxu8OofVLUS8tHwBDd4ZCfj7N3IAc1hr6tRdnqiDthUJpu2m1EeitXq*****' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20250714194722A3C36B7A****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。