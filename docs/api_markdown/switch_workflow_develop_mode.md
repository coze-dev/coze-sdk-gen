# 开启或关闭工作流多人协作
开启或关闭工作流、对话流多人协作。
开启多人协作后，你才能调用[添加工作流的协作者](https://docs.coze.cn/developer_guides/add_workflow_collaborator) API 为工作流添加协作者。
## 接口限制

* **套餐限制**：扣子企业版（企业标准版、企业旗舰版）。
* 只有资源库中的工作流和对话流支持添加多人协作，扣子应用中的工作流不支持开启多人协作模式，否则会报 4000 错误。
* 不支持渠道类型 OAuth 应用。使用 OAuth JWT 应用和服务访问令牌时，只需要有对应权限点即可。其余认证方式，只有**工作流所有者**能开启或关闭工作流的多人协作模式。
* 关闭工作流多人协作前，需要先调用[删除工作流协作者](https://docs.coze.cn/developer_guides/remove_workflow_collaborator) API 删除所有协作者。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflows/:workflow_id/collaboration_mode <br> ``` <br>  |
| **权限** | `Workflow.switchDevelopMode` <br> 确保调用该接口使用的访问令牌开通了 `Workflow.switchDevelopMode` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 开启或关闭工作流多人协作。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workflow_id | String | 可选 | 73505836754923*** | 需要设置协作模式的工作流 ID。 <br> 进入工作流编排页面，在页面 URL 中，workflow 参数后的数字就是 Workflow ID。例如 `https://www.coze.com/work_flow?space_id=42463***&workflow_id=73505836754923***`，Workflow ID 为 `73505836754923***`。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| collaboration_mode | String | 可选 | collaboration | 工作流的协作模式，枚举值： <br>  <br> * `single`：单用户模式。 <br> * `collaboration`：多人协作模式。 <br>  <br> 如需将工作流多人协作模式变更为单用户模式，需要先调用[删除工作流协作者](https://docs.coze.cn/developer_guides/remove_workflow_collaborator) API 删除所有协作者。 <br>  |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的日志信息，用于问题排查和技术支持。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/workflows/73505836754923***/collaboration_mode' \
--header 'Authorization : Bearer pat_Osa******' \
--header 'Content-Type : application/json' \
--data-raw '{
    "collaboration_mode": "collaboration"
}'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20241210152726467C48D89D6DB2****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://www.coze.cn/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。