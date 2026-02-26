# 查询已订阅的事件
查询回调应用中已订阅的事件。
## 接口限制

* 扣子个人版中，仅回调应用的创建者可以查看已订阅的事件。
* 扣子企业版中，仅超级管理员和管理员可以查看回调应用中已订阅的事件。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/api_apps/:api_app_id/events <br> ``` <br>  |
| **权限** | `listApiApp` <br> 确保调用该接口使用的服务令牌开通了 `listApiApp` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询回调应用中已订阅的事件。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| api_app_id | String | 可选 | 7512045450401153075*** | 回调应用的 ID。你可以通过[查询回调应用列表](https://docs.coze.cn/developer_guides/list_callback_app) API 获取回调应用的 ID。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_token | String | 可选 | \ | 分页查询时的翻页标识，表示下一页的起始位置。默认为 `""`，即从第一页数据开始返回。如果要查询下一页，需要使用上一次返回的 `next_page_token` 作为这次请求的入参。 |
| page_size | Integer | 可选 | 20 | 每页返回的数据条数，用于分页查询。默认值：50。 <br> 取值范围：1 ~ 50。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [ListSubscribedApiAppEventOpenRespData](#listsubscribedapiappeventopenrespdata) | {"items":[{"name":"智能客服机器人回调","api_app_id":"752690322532938****","event_type":"bot.published","description":"智能客服机器人回调应用，用于处理用户与机器人的交互事件。"}],"has_more":true,"next_page_token":"eyJwYWdlX3Rva2VuIjoiMTIzNDU****"} | 包含回调应用的信息列表及分页信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### ListSubscribedApiAppEventOpenRespData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [ApiAppEventOpen](#apiappeventopen) | [{"name":"智能客服机器人回调","api_app_id":"cli_a1b2c3d4e5f6g7h8","event_type":"bot.published","description":"智能客服机器人回调应用，用于处理用户与机器人的交互事件。"}] | 回调应用的信息列表，包含每个回调应用的名称、ID、事件类型和描述。 |
| has_more | Boolean | true | 标识当前返回的回调应用列表是否还有更多数据未返回。 <br>  <br> * `true` ：还有更多未返回的回调应用。 <br> * `false`：已返回所有数据。 |
| next_page_token | String | eyJwYWdlX3Rva2VuIjoiMTIzNDU**** | 翻页标识，表示下一页的起始位置。当 `has_more` 为 `true` 时，表示还有更多数据未返回，可以通过此令牌获取下一页数据。首次请求不填或置空，后续翻页需使用上一次返回的 `next_page_token`。 |
### ApiAppEventOpen
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| name | String | 智能客服机器人回调应用 | 回调应用的名称。 |
| api_app_id | String | 752690322532938**** | 回调应用的 ID。 |
| event_type | String | bot.published | 回调事件类型，目前支持的事件类型包括： <br>  <br> * `bot.published`：智能体发布事件。 <br> * `bot.deleted`：智能体删除事件。 <br> * `bot.unpublished`：智能体下架事件。 <br> * `benefit.usage`：账单推送回调。 <br> * `benefit.plugin.scale.requested`：申请插件扩容事件。 <br> * `benefit.plugin.scale.expired`：插件扩容到期事件。 |
| description | String | 智能客服机器人回调应用，用于处理用户与机器人的交互事件。 | 回调应用的描述信息，用于说明回调应用的功能或用途。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
GET 'https://api.coze.cn/v1/api_apps/752690322532938****/events' \
--header 'Authorization: Bearer czs_l8wA17hHrFts8ebCe9RxrjDBJvJJhXyxHQK1I8hEcu8cGsQm8LPI6Xs4LHyj*****' \
--header 'Content-Type: application/json'
{"event_types": ["bot.published", "bot.deleted"]}
```

### 返回示例
```JSON
{
  "data": {
    "has_more": false,
    "items": [
      {
        "description": "智能体已被所有者删除",
        "event_type": "bot.deleted",
        "api_app_id": "752690322532938****",
        "name": "智能音箱回调应用"
      }
    ]
  },
  "detail": {
    "logid": "202507141951091846A5E****"
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。