# 获取用户变量的值
获取智能体或应用中设置的用户变量的值。
## 接口说明
通过[设置用户变量的值](https://docs.coze.cn/developer_guides/update_variable) API 设置变量值之后，可以通过本 API 查看智能体或应用中设置的用户变量的值。调用此 API 时可以查看指定变量的值，也可以将 `keywords` 指定为空，查看智能体或应用下的所有用户变量的值。如果输入的 `keywords` 在智能体或应用中不存在，扣子编程不会报错，但返回结果中不会包含相应的用户变量。
## 限制说明

* 仅支持获取已发布 API、ChatSDK 的智能体或应用中的用户变量的值。
* 用户变量按照 `user_id` + `connector_uid` + `connector_id` + `bot_id` 的组合进行隔离，因此在扣子编程站内设置的用户变量，在 API 渠道可能无法获取对应的值。不同渠道用户标识的规则存在差异，具体如下表所示。

| **参数** | **扣子编程站内** | **API 渠道** |
| --- | --- | --- |
| `user_id ` | 当前使用者的扣子用户 ID。 | API Token 生成者的 ID。 |
| `connector_uid` | 当前使用者的扣子用户 ID。 | 用户在[发起对话](https://docs.coze.cn/developer_guides/chat_v3) 等 API 中输入的 `user_id`。 |
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/variables <br> ``` <br>  |
| **权限** | `readVariable` <br> 确保调用该 API 使用的访问令牌开通了 `readVariable` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 获取智能体或应用中设置的用户变量的值。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| app_id | String | 可选 | 744885747763*** | 如果需要获取应用中设置的用户变量的值，填入对应的应用ID。 <br> 你可以通过应用的业务编排页面 URL 中获取应用 ID，也就是 URL 中 project-ide 参数后的一串字符，例如 `https://www.coze.cn/space/739174157340921****/project-ide/743996105122521****/workflow/744102227704147****` 中，应用的 ID 为 `743996105122521****`。 <br> `app_id` 和 `bot_id` 应至少填写一个，否则会报错。 <br>  |
| bot_id | String | 可选 | 749315102740*** | 如果需要获取智能体中设置的用户变量的值，填入对应的应用ID。 <br> 你可以通过智能体的开发页面获取智能体 ID，开发页面 URL 中 bot 参数后的数字就是智能体 ID。例如`https://www.coze.com/space/341****/bot/73428668*****`，bot ID 为`73428668*****`。 |
| connector_id | String | 可选 | 1024 | 智能体或应用的发布渠道 ID 列表。目前支持如下渠道： <br>  <br> * API：1024 <br> * ChatSDK：999 |
| connector_uid | String | 必选 | openAPI | 查看指定用户 ID 的变量值。用户 ID 对应[执行工作流](https://docs.coze.cn/developer_guides/workflow_run) API 中 `ext` 字段中的 user_id。 |
| keywords | Array of String | 可选 | name,age | 变量名称，多个变量用英文逗号分隔。 <br> 当 `keywords` 为空时，将返回用户在智能体或应用下的所有用户变量的值。如果输入的 `keywords` 在智能体或应用中不存在，则返回结果中不会包含相应的用户变量。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [GetVariableData](#getvariabledata) | - | 用户变量数组。 |
| detail | Object of [ResponseDetail](#responsedetail) | - | 返回详情。 |
### GetVariableData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [KVItem](#kvitem) | [{ "value": "小王", "create_time": 0, "update_time": 0, "keyword": "name" }] | 用户变量数组。 |
### KVItem
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| value | String | 18 | 用户变量的值。 |
| keyword | String | age | 用户变量的名称。 <br> keyword 必须为智能体或应用中已创建并开启的用户变量，不能设置为系统变量。 <br>  |
| create_time | Long | 1744637812 | 首次设置变量值的时间。如果变量值为默认值，则 `create_time` 的值为 0。 <br> Unixtime 时间戳格式，单位为秒。 |
| update_time | Long | 1744637812 | 更新时间，Unixtime 时间戳格式，单位为秒。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/variables?app_id=7448857477636685850&bot_id=74931510***9&connector_id=1024' \
--header 'Authorization: Bearer pat_Q8LmOJU9mRCUzNCdnq6DMakFsm*****'
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "items": [
      {
        "value": "小王",
        "create_time": 0,
        "update_time": 0,
        "keyword": "name"
      },
      {
        "update_time": 1744637812,
        "keyword": "age",
        "value": "18",
        "create_time": 1744637812
      }
    ]
  },
  "detail": {
    "logid": "20241210152726467C48D89D6DB2****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。