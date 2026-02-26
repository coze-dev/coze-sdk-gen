# 更新审核结果
调用该接口更新智能体发布到自定义渠道的审核结果。
用户将智能体发布到自定义渠道后，如果触发了自定义渠道的审核流程，自定义渠道会通过回调消息返回审核结果。如果审核耗时较长，自定义渠道可以先在回调中返回审核中状态，审核结束后，再通过此 API 通知扣子此智能体的最终审核结果。
## 基础信息
| **请求方式** | PUT |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/connectors/:connector_id/bots/:bot_id <br> ``` <br>  |
| **权限** | `Connector.botUpdateProfile` <br> 确保调用该接口使用的访问令牌开通了`Connector.botUpdateProfile`权限，详细信息参考[鉴权](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 更新智能体发布到自定义渠道的审核结果。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| bot_id | String | 可选 | 73428668***** | 要更新的智能体 ID。 <br> 进入智能体的编排页面，编排页面 URL 中 bot 参数后的数字就是智能体 ID。例如`https://www.coze.cn/space/341****/bot/73428668*****`，智能体 ID 为`73428668*****`。 |
| connector_id | String | 可选 | 74202797****** | 渠道 ID。需要跟 token 的渠道一致。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| audit_status | Integer | 可选 | 2 | 渠道审核状态。支持设置为： <br>  <br> * 1: 审核中 <br> * 2: 审核通过 <br> * 3: 审核不通过 |
| reason | String | 可选 | 头像无法通过审核 | 审核不通过的具体原因。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。 <br>  <br> * 0 表示调用成功。 <br> * 其他值表示调用失败。你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request PUT 'https://api.coze.cn/v1/connectors/74202797******/bots/73428668*****' \
--header 'Authorization : Bearer pat_Osa******' \
--header 'Content-Type : application/json' \
--data-raw '{
    "audit_status": 2,
    "reason": "头像无法通过审核"
}'
```

### 返回示例
```JSON
{
  "detail": {
    "logid": "20250106172024B5F607030EFFAD***"
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。