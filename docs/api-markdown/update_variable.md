# 设置用户变量的值
为智能体或应用中的用户变量设置变量值。
## 接口说明
当智能体或应用中已创建并开启用户变量时，你可以通过该 API 设置变量值，未开启的用户变量无法设置变量值。
设置用户变量值后，你可以调用[获取用户变量值](https://docs.coze.cn/developer_guides/read_variable) API 查询是否设置成功。
## 限制说明
仅支持为已发布 API、ChatSDK 的智能体或应用设置用户变量的值。
## 基础信息
| **请求方式** | PUT |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/variables <br> ``` <br>  |
| **权限** | `updateVariable` <br> 确保调用该 API 使用的访问令牌开通了 `updateVariable` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 为智能体或应用中的用户变量设置变量值。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| app_id | String | 可选 | 7448857477636685850 | 如果需要为应用设置用户变量的值时，填入对应的应用ID。 <br> 你可以通过应用的业务编排页面 URL 中获取应用 ID，也就是 URL 中 project-ide 参数后的一串字符，例如 `https://www.coze.cn/space/739174157340921****/project-ide/743996105122521****/workflow/744102227704147****` 中，应用的 ID 为 `743996105122521****`。 <br> `app_id` 和 `bot_id` 应至少填写一个，否则会报错。 <br>  |
| bot_id | String | 可选 | 7493151027400097829 | 需要为智能体设置用户设置变量的值时，填入对应的智能体 ID。 <br> 你可以通过智能体的编排页面获取智能体 ID，开发页面 URL 中 bot 参数后的数字就是智能体 ID。例如`https://www.coze.com/space/341****/bot/73428668*****`，bot ID 为`73428668*****`。 |
| connector_id | String | 可选 | 1024 | 智能体或应用的发布渠道 ID 列表。目前支持如下渠道： <br>  <br> * API：（默认）1024 <br> * ChatSDK：999 |
| connector_uid | String | 必选 | 7432567 | 用户 ID，用于为指定的用户设置变量的值。用户 ID 对应[执行工作流](https://docs.coze.cn/developer_guides/workflow_run) API 中 `ext` 字段指定的 user_id 或发起对话 API 中的 user_id。 |
| data | Array of [KVItem](#kvitem) | 必选 | [{"keyword":"name","value":""},{"keyword":"age","value":""}] | 用户变量的数组，不能为空。 |
### KVItem
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| value | String | 必选 | 18 | 用户变量的值。 |
| keyword | String | 必选 | age | 用户变量的名称。 <br> keyword 必须为智能体或应用中已创建并开启的用户变量，不能设置为系统变量。 <br>  |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250416125552EE59A23A87A***" } | 返回详情。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request PUT 'https://api.coze.cn/v1/variables' \
--header 'Authorization: Bearer pat_Q8LmOJU9mRCUzNCdnq6DMakFsmr***' \
--header 'Content-Type: application/json' \
--data '{
    "bot_id":"749315102740009***",
    "connector_id":"1024",
    "connector_uid":"7432567",
    "data":[
        {
            "keyword":"name",
            "value":"小王"
        },
        {
            "keyword":"age",
            "value":"18"
        }
    ]

}'
```

### 返回示例
```JSON
{
    "code": 0,
    "msg": "",
    "detail": {
        "logid": "20250416125552EE59A23***"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示 API 调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

