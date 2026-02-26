# 修改回调应用
修改回调应用的名称和回调地址。
## 接口限制

* 扣子个人版中，仅支持修改本人创建的回调应用。
* 扣子企业版中，仅超级管理员和管理员可修改回调应用。

## 基础信息
| **请求方式** | PUT |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/api_apps/:api_app_id <br> ``` <br>  |
| **权限** | `updateApiApp` <br> 确保调用该接口使用的访问令牌开通了 `updateApiApp` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 修改回调应用。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| api_app_id | String | 必选 | 752688153695671**** | 待修改的回调应用的 ID。你可以通过[查询回调应用列表](https://docs.coze.cn/developer_guides/list_callback_app) API 获取回调应用的 ID。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| name | String | 可选 | 智能客服回调应用 | 修改后的回调应用名称，用于标识修改后的回调应用身份或用途。 |
| callback_url | String | 可选 | https://api.example.com/callback/updated | 修改后的回调地址，用于接收扣子编程发送的事件通知。 <br> 扣子编程会向该地址发送包含 `challenge` 的 POST 请求进行校验，服务器需在 3 秒内原样返回 `challenge` 值以验证地址有效性。详情请参考[订阅回调](https://docs.coze.cn/dev_how_to_guides/add_callback)。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息，主要用于问题排查。如果遇到异常报错场景，且反复重试仍然报错，可以根据此字段中的 `logid`联系扣子团队获取帮助。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
PUT 'https://api.coze.cn/v1/api_apps/752688153695671****' \
--header 'Authorization: Bearer czs_lvJ1wCpwGsRYM0pxzCDae0dnvoNTJTXWhivKUHN1xQ9ay8m8ob2frbNx5hbS*****' \
--header 'Content-Type: application/json'
{ "name": "智能客服回调应用"}
```

### 返回示例
```JSON
{
  "detail": {
    "logid": "2025071419450725A01236138****"
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

