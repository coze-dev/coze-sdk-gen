# 复制资源
复制智能体、应用和工作流。
## 接口描述

* 本 API 支持**同工作空间复制**或**跨工作空间复制**。如果你是空间成员，你可以调用本 API 将工作空间内任意智能体、应用和工作流，复制到本空间或同一企业/团队下的其他已加入的工作空间。
* 跨空间复制智能体或应用成功后，为保证正常运行，智能体和应用所使用的工作流、插件等资源也将同时复制到新空间。
* 该 API 分为同步执行和异步执行，扣子编程根据不同资源复制类型自动采用对应的执行方式，具体说明如下：

| **执行模式** | **适用场景** | **结果返回方式** |
| --- | --- | --- |
| 同步执行 | 同一工作空间内复制**智能体**。 | API 直接返回复制后的智能体 ID。 |
| 异步执行 | * 同一工作空间内复制**应用**、**工作流**。 <br> * 跨工作空间复制**智能体**、**应用**、**工作流**。 | API 返回任务 ID（`task_id`），需要通过[查询资源复制的结果](https://docs.coze.cn/developer_guides/query_resource_copy_execution_result)  API 查询执行结果。 |
## 接口限制

* 不支持将个人工作空间中的资源复制到企业或团队工作空间中，个人版工作空间中的资源只能复制到本空间。
* 操作者需要加入源工作空间和目标工作空间。
* 复制后的智能体或应用为草稿状态。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/entities/copy_tasks <br> ``` <br>  |
| **权限** | * 复制智能体：`Bot.duplicate`、`Workspace.createBot` <br> * 复制扣子应用：`Project.copy`、`Workspace.createProject` <br> * 复制工作流：`Workflow.copy`、`Workspace.createWorkflow` <br>  <br> 根据复制资源的类型授予对应的权限，确保调用该接口使用的访问令牌已开通相应权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 复制智能体、应用和工作流。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| entity_id | String | 可选 | 753093880875067*** | 待复制资源的 ID，需要和 `entity_type` 配套使用。 <br> 例如，当 `entity_type` 为 `bot`时，此字段应填写智能体的 ID。 |
| entity_type | String | 可选 | bot | 待复制资源的类型。枚举值： <br>  <br> * `app`：应用。 <br> * `bot`：智能体。 <br> * `workflow`：工作流。 |
| to_workspace_id | String | 可选 | 7498204883280*** | 目标工作空间 ID。默认为空，即同工作空间复制。 <br> 如果是跨工作空间复制，需要指定资源复制到哪个工作空间。只能填写本企业或团队下的工作空间。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，工作空间 ID 为 `75814654762959***`。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [OpenDuplicateDraftEntityData](#openduplicatedraftentitydata) | {"copied_entity_id":"75309388087500****"} | 资源复制的返回结果。 <br>  <br> * 如果任务是异步执行的，则返回 `task_id` 用于查询任务状态。 <br> * 如果任务直接完成，则返回 `copied_entity_id`表示复制后的实体 ID。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 本次请求的详细日志信息，用于问题排查。 |
### OpenDuplicateDraftEntityData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| task_id | String | 123e456700*** | 当资源复制任务为异步执行时，API 返回的任务 ID，你可以根据该 `task_id`，通过[查询资源复制的结果](https://docs.coze.cn/developer_guides/query_resource_copy_execution_result)  API 查询执行结果。  <br> 以下场景中 API 为异步执行： <br>  <br> * 同工作空间复制应用、工作流。 <br> * 跨工作空间复制智能体、应用、工作流。 |
| copied_entity_id | String | 75309388087500**** | 同工作空间复制智能体时，API 会直接返回复制后的智能体 ID。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/entities/copy_tasks' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type: application/json' \
--data-raw '{
    "entity_id": "753093880875067***",
    "entity_type": "bot",
    "to_workspace_id": "7498204883280***"
}'
```

### 返回示例
```JSON
{
  "data": {
    "copied_entity_id": "75309388087500****"
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