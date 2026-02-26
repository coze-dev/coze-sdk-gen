# 查询回调应用列表
查询回调应用列表。
## 接口限制

* 扣子个人版中，仅支持查看本人创建的回调应用。
* 扣子企业版中，仅超级管理员和管理员可以查看企业的回调应用列表。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/api_apps <br> ``` <br>  |
| **权限** | `listApiApp` <br> 确保调用该接口使用的服务令牌开通了 `listApiApp` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询回调应用列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_token | String | 可选 | "" | 分页查询时的翻页标识，表示下一页的起始位置。默认为空字符串 `""`，即从第一页数据开始返回。如果要查询下一页，需要使用上一次返回的 `next_page_token` 作为这次请求的入参。 |
| page_size | Integer | 可选 | 20 | 每页返回的数据条数，用于分页查询。默认值：50。 <br> 取值范围：1 ~ 50。 |
| app_type | String | 可选 | normal | 回调应用的类型。默认为空，查看所有类型的回调应用。枚举值如下： <br>  <br> * `normal`：普通回调应用。 <br> * `connector`：渠道回调应用。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [GetApiAppListOpenRespData](#getapiapplistopenrespdata) | {"items":[{"id":"74876004423701****","name":"智能客服机器人回调","app_type":"normal","callback_url":"https://example.com/api/callback","verify_token":"OYDacMzM3WyOWV3Dtj2bHRMymzxP****"}],"has_more":true,"next_page_token":"next_page_token_value"} | 返回的回调应用列表数据，包含分页信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的日志信息，用于问题排查。 |
### GetApiAppListOpenRespData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [ApiApp](#apiapp) | [{"id":"74876004423701****","name":"智能客服机器人回调","app_type":"normal","callback_url":"https://example.com/api/callback","verify_token":"OYDacMzM3WyOWV3Dtj2bHRMymzxP****"}] | 回调应用列表，包含每个回调应用的详细信息。 |
| has_more | Boolean | true | 标识当前返回的回调应用列表是否还有更多数据未返回。 <br>  <br> * `true` ：还有更多未返回的回调应用。 <br> * `false`：已返回所有数据。 |
| next_page_token | String | 1752492196 | 分页 token，用于获取下一页数据。第一页请求时传空，后续请求通过上一个响应返回的 `next_page_token`获取下一页数据。 |
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
GET https://api.coze.cn/v1/api_apps \
--header 'Authorization: Bearer czs_hh19tcCx3qn05X7bG71AGzyD2DjrunJqaAeHORjqZCUydGLRqFhSXWr*******' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "data": {
    "has_more": false,
    "items": [
      {
        "name": "api app 1",
        "verify_token": "cZTAkNZMhfapeskz3beBaOvuxcSLqM5o2uQwgIHTgumCUFmDi54WRfIsB1****",
        "callback_url": "https://callback.example.com/callback",
        "id": "751230611020066****",
        "app_type": "normal"
      },
      {
        "verify_token": "qrwk0rqd843N4SKvyxoHY4yHSsvde0p8IWOH2ef0cR2FWahyxB1Evhl5y****",
        "callback_url": "https://callback.example.com/callback",
        "id": "751230611020070****",
        "app_type": "normal",
        "name": "api app 2"
      }
    ]
  },
  "detail": {
    "logid": "20250714193443C60E2B008BF****"
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

