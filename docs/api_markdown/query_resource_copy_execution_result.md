# 查询资源复制的结果
查询资源复制的结果。
## 接口说明
调用[复制资源](https://docs.coze.cn/developer_guides/copy_resource_task) API 时，以下场景中 API 为异步执行，响应信息中会返回 task_id，开发者可以通过本 API 查询指定事件的执行结果。

* 同工作空间复制应用、工作流。
* 跨工作空间复制智能体、应用、工作流。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/entities/copy_tasks/:task_id <br> ``` <br>  |
| **权限** | 无 |
| **接口说明** | 查询资源复制异步执行的结果。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| task_id | String | 可选 | 567456700*** | 资源复制的任务 ID。调用[复制资源](https://docs.coze.cn/developer_guides/copy_resource_task) API 时，如果 API 为异步执行，响应信息中会返回 task_id。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [OpenCopyTaskInfoData](#opencopytaskinfodata) | {"task_id":"123e456700***","entity_type":"bot","task_status":"successful","source_entity_id":"7530938808750080***","target_entity_id":"7530938808750067***","target_workspace_id":"74982048832804***"} | 资源复制任务的结果数据，包含任务 ID、资源类型、任务状态、失败原因、原始资源 ID、复制后的资源 ID 以及目标工作空间 ID 等信息。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 本次请求的详细日志信息，用于问题排查。 |
### OpenCopyTaskInfoData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| task_id | String | 123e456700*** | 资源复制的任务 ID。 |
| entity_type | String | bot | 待复制资源的类型。枚举值： <br>  <br> * `app`：应用。 <br> * `bot`：智能体。 <br> * `workflow`：工作流。 |
| task_status | String | successful | 任务的状态。枚举值： <br>  <br> * `in_progress`：进行中。 <br> * `successful`：成功。 <br> * `failed`：失败。 |
| failed_reasons | Array of [TaskFailedReason](#taskfailedreason) | \ | 当资源复制任务失败时，该字段汇总所有失败的具体原因。每个失败原因包含资源 ID、名称、类型及详细错误描述。 |
| source_entity_id | String | 7530938808750080*** | 源资源 ID。 |
| target_entity_id | String | 7530938808750067*** | 复制后的资源 ID。 |
| target_workspace_id | String | 74982048832804*** | 资源复制后的目标工作空间 ID，表示资源被复制到的工作空间标识。 |
### TaskFailedReason
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| entity_id | String | 7530938808750080*** | 在资源复制任务失败时，标识具体失败的资源 ID。 |
| entity_name | String | 单词学习助手 | 资源的名称。 |
| entity_type | String | bot | 在资源复制任务失败时，标识具体失败资源的类型。支持的取值如下所示。 <br>  <br> * `app`：应用。 <br> * `bot`：智能体。 <br> * `workflow`：工作流。 <br> * `plugin`：插件。 <br> * `ui`：应用的用户界面。 <br> * `knowledge`：知识库。 <br> * `database`：数据库。 <br> * `variable`：变量。 <br> * `trigger`：触发器。 <br> * `prompt`：提示词。 <br> * `shortcut`：快捷指令。 |
| failed_reason | String | 资源配额不足，无法完成复制 | 资源复制任务失败的具体原因描述，API 调用失败时可通过此字段查看详细错误信息。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/entities/copy_tasks/567456700***' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
  "data": {
    "task_id": "123e456700***",
    "entity_type": "bot",
    "task_status": "successful",
    "source_entity_id": "7530938808750080***",
    "target_entity_id": "7530938808750067***",
    "target_workspace_id": "74982048832804***"
  },
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20241210152726467C48D89D6DB2****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。