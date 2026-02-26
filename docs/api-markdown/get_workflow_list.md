# 查询工作流列表
查询指定工作空间中的工作流列表及其基本信息。
你可以查询某个工作空间中的所有工作流或对话流、某个应用关联的工作流或对话流。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflows <br> ``` <br>  |
| **权限** | `listWorkflow` <br> 确保调用该接口使用的访问令牌开通了 `listWorkflow` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询指定工作空间中的工作流列表及其基本信息。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workspace_id | String | 必选 | 736163827687053**** | 工作空间 ID，用于指定要查询的工作空间。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，工作空间 ID 为 `75814654762959***`。 |
| page_num | Integer | 必选 | 1 | 查询结果分页展示时，此参数用于设置查看的页码。最小值为 1。 |
| page_size | Integer | 可选 | 20 | 查询结果分页展示时，此参数用于设置每页返回的数据量。取值范围为 1 ~ 30，默认为 10。 |
| workflow_mode | String | 可选 | workflow | 工作流类型，默认为空，即查询所有工作流类型。枚举值： <br>  <br> * `workflow`：工作流。 <br> * `chatflow`：对话流。 |
| app_id | String | 可选 | 744208683** | 扣子应用 ID，用于查询指定应用关联的工作流。默认为空，即不指定应用。 <br> 进入应用开发界面，开发页面 URL 中的 `project-ide` 参数后的数字就是 AppID，例如`https://www.coze.cn/space/74421656*****/project-ide/744208683**` ，扣子应用 ID 为`744208683**`**。** |
| publish_status | String | 可选 | all | 工作流的发布状态，用于筛选不同发布状态的版本。枚举值： <br>  <br> * all ：所有状态。 <br> * published_online ：（默认值）已发布的正式版。 <br> * unpublished_draft ：草稿状态。 |

## 返回参数

| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [OpenAPIWorkflowList](#openapiworkflowlist) | {"items":[{"app_id":"744208683**","creator":{"id":"2478774393***","name":"user41833***"},"icon_url":"https://example.com/icon/workflow_123.png","created_at":"1700000000","updated_at":"1701234567","description":"工作流测试","workflow_id":"73505836754923***","workflow_name":"workflow_example"}],"has_more":false} | 工作流列表及其详细信息。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### OpenAPIWorkflowList
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [OpenAPIWorkflowBasic](#openapiworkflowbasic) | [{"app_id":"744208683**","creator":{"id":"2478774393***","name":"user41833***"},"icon_url":"https://example.com/icon/workflow_123.png","created_at":"1700000000","updated_at":"1701234567","description":"工作流测试","workflow_id":"73505836754923***","workflow_name":"workflow_example"}] | 工作流的基础信息。 |
| has_more | Boolean | false | 标识当前返回的工作流列表是否还有更多数据未返回。 <br>  <br> * `true` ：还有更多未返回的回调应用。 <br> * `false`：已返回所有数据。 |
### OpenAPIWorkflowBasic
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| app_id | String | 744208683** | 工作流关联的应用 ID。若工作流未关联任何应用，则该字段值为 `0`。 |
| creator | Object of [OpenAPIUserInfo](#openapiuserinfo) | {"id":"2478774393***","name":"user41833***"} | 工作流创建者的信息，包含创建者的用户 ID 和用户名。 |
| icon_url | String | https://example.com/icon/workflow_123.png | 工作流图标的 URL 地址。 |
| created_at | String | 1752060786 | 工作流的创建时间，以 Unix 时间戳表示，单位为秒。 |
| updated_at | String | 1752060827 | 工作流的最后更新时间，以 Unix 时间戳表示，单位为秒。 |
| description | String | 生成音乐的工作流 | 工作流的描述。 |
| workflow_id | String | 73505836754923*** | 工作流 ID。 |
| workflow_name | String | workflow_example | 工作流名称。 |
### OpenAPIUserInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 2478774393*** | 工作流创建者的扣子用户 UID。 |
| name | String | user41833*** | 工作流创建者的扣子用户名。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |


## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/workflows?page_num=1&workspace_id=74496******2844844' \
--header 'authorization: Bearer pat_hBHquB***' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "data": {
    "has_more": true,
    "items": [
      {
        "updated_at": "1752060827",
        "workflow_id": "7524966*******",
        "description": "Description",
        "app_id": "0",
        "created_at": "1752060786",
        "icon_url": "https://tosv.****/workflow.png",
        "creator": {
          "id": "26150****",
          "name": "alan"
        },
        "workflow_name": "workflow_Name"
      }
    ]
  },
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20250716215828******"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

