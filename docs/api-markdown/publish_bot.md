# 发布智能体
将指定智能体发布到 API、Chat SDK 或者自定义渠道。
## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/bot/publish <br> ``` <br>  |
| **权限** | `publish` <br> 确保调用该接口使用的访问令牌开通了 `publish` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用接口将指定智能体发布到 API、Chat SDK 或者自定义渠道。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| bot_id | String | 必选 | 73428668***** | 要发布的智能体 ID。 <br> 进入智能体的开发页面，开发页面 URL 中 bot 参数后的数字就是智能体 ID。例如`https://www.coze.com/space/341****/bot/73428668*****`，bot ID 为`73428668*****`。 |
| connector_ids | Array of String | 必选 | ["1024"] | 智能体的发布渠道 ID 列表。目前支持通过此 API 将智能体发布为 API、Chat SDK 以及自定义渠道。 <br>  <br> * API: 1024 <br> * ChatSDK: 999 <br> * 自定义渠道: 自定义渠道 ID。自定义渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 |
| connectors | JSON Map | 可选 | {"7470849514748***":{"device_id":"[\"51237825***\"]"}} | 智能体发布自定义渠道时，你可以通过该参数传递自定义参数给渠道。例如，设置将智能体发布到指定设备。格式：`{"渠道 ID":{"key":"value"}}`，其中 `key` 为[绑定设备 API](https://docs.coze.cn/developer_guides/bind_connector_config) 中设置的 key 值，`value` 需要经 JSON 序列化。 <br> 智能体发布自定义渠道时，需要同时发布 API 渠道。 <br>  |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。 <br>  <br> * 0 表示调用成功。 <br> * 其他值表示调用失败。你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [PublishDraftBotData](#publishdraftbotdata) | { "bot_id": "743961547827****", "version": "1732190531***" } | 接口响应的业务信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"1234567890****"} | 包含请求的详细日志信息，用于问题排查和调试。 |
### PublishDraftBotData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| bot_id | String | 73428668***** | 发布的智能体的 ID。 |
| version | String | 1753871809679 | 智能体的版本号，用于标识智能体的当前版本号。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例

<div type="doc-tabs">
<div type="tab-item" title="发布官方渠道" key="W1oT_pjzMY1GkOPc0CSvN">

```JSON
curl --location --request POST 'https://api.coze.cn/v1/bot/publish' \
--header 'Authorization: Bearer pat_x*******' \
--header 'Content-Type: application/json' \
--data-raw '{
    "bot_id": "73428668*****",
    "connector_ids": [
        "1024"
    ]
}'
```


</div>
<div type="tab-item" title="发布自定义渠道" key="REcI8dc3HIlBFe519J2Ji">

```JSON
curl --location --request POST 'https://api.coze.cn/v1/bot/publish' \
--header 'Authorization: Bearer pat_xitq****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "bot_id": "73428668*****",
    "connector_ids": [
        "1024",
        "74708495147488***"
    ],
    "connectors": {
        "74708495147488***": {
            "device_id": "[\"51237825***\"]"
        }
    }
}'
```


</div>
</div>
### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "bot_id": "743961547827****",
    "version": "1732190531***"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

