# 查询工作流基本信息
查询工作流的基本信息，包括工作流名称、描述、创建者信息等。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflows/:workflow_id <br> ``` <br>  |
| **权限** | `getMetaData` <br> 确保调用该接口使用的访问令牌开通了 `getMetaData` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询工作流基本信息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workflow_id | String | 必选 | 73505836754923*** | 工作流 ID。 <br> 进入工作流编排页面，在页面 URL 中，workflow 参数后的数字就是 Workflow ID。例如 https://www.coze.com/work_flow?space_id=42463***&workflow_id=73505836754923***，Workflow ID 为 73505836754923***。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| include_input_output | Boolean | 可选 | true | 是否在返回结果中返回输入和输出参数的结构体。 <br>  <br> * `true`：返回输入输出参数结构体 <br> * `false`：不返回输入输出参数结构体 <br>  <br> 默认值为 `false`。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [WorkflowInfo](#workflowinfo) | \ | 工作流的详细信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的日志信息，用于问题排查和技术支持。 |
### WorkflowInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| input | Object of [OpenAPIWorkflowInput](#openapiworkflowinput) | \ | 工作流开始节点的输入参数的结构体。 |
| output | Object of [OpenAPIWorkflowOutput](#openapiworkflowoutput) | { "parameters": { "arr_obj_num": { "required": false, "type": "number" } }, "terminate_plan": "use_answer_content", "content": "输出为{{output}}" } | 工作流结束节点的输出参数的结构体。 |
| workflow_detail | Object of [OpenAPIWorkflowBasic](#openapiworkflowbasic) | {"app_id":"744208683**","creator":{"id":"2478774393***","name":"user41833***"},"icon_url":"https://example.com/icon/workflow_123.png","created_at":1700000000,"updated_at":"1701234567","description":"工作流测试","workflow_id":"73505836754923***","workflow_name":"workflow_example"} | 工作流的详细信息。 |
### OpenAPIWorkflowInput
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| parameters | JSON Map | { "CONVERSATION_NAME": { "required": false, "description": "", "default_value": "", "type": "string" }} | 开始节点的输入参数结构体。 |
### Parameters
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| type | String | string | 该参数的类型。 |
| items | Object of [OpenAPIParameter](#openapiparameter) | {"type":"image"} | 当参数类型为 `array` 时，该字段用于定义数组元素的子类型。 |
| required | Boolean | true | 标识输入参数是否为必填项。 <br>  <br> * `true`：该参数为必填项。 <br> * `false`：该参数为可选项。 |
| properties | JSON Map | { "arr_obj_num": { "required": false, "type": "number" } } | 当参数类型为 `object`时，该字段用于定义对象类型的子参数信息，采用键值对（key-value）结构，其中 key 为子参数的名称，value 为该子参数的定义信息（包含类型、是否必填等属性）。 |
| description | String | 上传的图片。 | 该参数的描述信息。 |
| default_value | String | - | 该参数配置的默认值。 |
### OpenAPIWorkflowOutput
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| content | String | 数量为{{arr_obj_num}} | 工作流结束节点返回文本时，智能体回复内容的结构。仅当 `terminate_plan` 为 `use_answer_content` 时会返回。 |
| parameters | JSON Map | { "output": { "type": "string" } } | 工作流结束节点输出变量的数组。以键值对形式存储，格式为` { "变量名称": { "type": "变量类型" } }`。 |
| terminate_plan | String | return_variables | 结束节点的返回类型，枚举值： <br>  <br> * `return_variables`：返回变量。 <br> * `use_answer_content`：返回文本。 |
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
curl --location --request GET 'https://api.coze.cn/v1/workflows/7493*****453978?include_input_output=true' \
--header 'Authorization: Bearer pat_h******' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "data": {
    "workflow_detail": {
      "workflow_name": "workflow_image",
      "app_id": "7493481****8294",
      "creator": {
        "id": "44246792*****",
        "name": "张三"
      },
      "updated_at": 1747983675,
      "workflow_id": "750712******91219",
      "description": "统计图片数量",
      "icon_url": "https://*****pi.com/ocean-cloud-tos/plugin_icon/chatflow-icon.png?****",
      "created_at": 1747888248
    },
    "input": {
      "parameters": {
        "CONVERSATION_NAME": {
          "required": false,
          "description": "",
          "default_value": "",
          "type": "string"
        },
        "USER_INPUT": {
          "type": "string",
          "required": false
        }
      }
    },
    "output": {
      "parameters": {
        "arr_obj_num": {
          "type": "number"
        }
      },
      "terminate_plan": "use_answer_content",
      "content": "数量为{{arr_obj_num}}"
    }
  },
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "2025071621341*******4D"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://www.coze.cn/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。
