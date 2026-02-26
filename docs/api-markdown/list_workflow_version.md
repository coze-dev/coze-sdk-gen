# 查询工作流的版本列表
查询工作流的版本列表。
你可以通过本 API 查询某个工作流的所有历史版本记录，包括版本号、版本描述、操作者和发布时间等。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflows/:workflow_id/versions <br> ``` <br>  |
| **权限** | `listVersion` <br> 确保调用该接口使用的访问令牌开通了 `listVersion` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询工作流的版本列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workflow_id | String | 必选 | 73505836754923*** | 待查询的 Workflow ID。 <br> 进入 Workflow 编排页面，在页面 URL 中，workflow 参数后的数字就是 Workflow ID。例如 https://www.coze.com/work_flow?space_id=42463***&workflow_id=73505836754923***，Workflow ID 为 73505836754923***。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| publish_status | String | 可选 | all | 工作流的发布状态，用于筛选不同发布状态的版本。枚举值： <br>  <br> * all ：所有状态。 <br> * published_online ：（默认值）已发布的正式版。 <br> * unpublished_draft ：草稿状态。 |
| page_size | Integer | 可选 | 20 | 查询结果分页展示时，此参数用于设置每页返回的数据量。取值范围为 1 ~ 30，默认为 10。 |
| page_token | String | 可选 | 1752666079000 | 分页查询时的翻页标识，表示下一页的起始位置。默认为 `""`，即从第一页数据开始返回。如果要查询下一页，需要使用上一次返回的 `next_page_token` 作为这次请求的入参。 |
| include_input_output | Boolean | 可选 | true | 是否在返回结果中返回输入和输出参数的结构体。 <br>  <br> * `true`：返回输入输出参数结构体 <br> * `false`：不返回输入输出参数结构体 <br>  <br> 默认值为 `false`。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [OpenAPIListVersionData](#openapilistversiondata) | {"items":[{"creator":{"id":"2478774393***","name":"user41833***"},"version":"v1.0.0","created_at":"1700000000","updated_at":"1701234567","description":"工作流测试","workflow_id":"73505836754923***"}],"has_more":false,"next_page_token":"73505836754923***"} | 包含工作流列表及其相关信息的对象，用于返回查询结果。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### OpenAPIListVersionData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [OpenAPIVersionMetaInfo](#openapiversionmetainfo) | \ | 工作流的历史版本列表。 |
| has_more | Boolean | false | 标识当前返回的版本列表是否还有更多数据未返回。 <br>  <br> * `true` ：还有更多未返回的回调应用。 <br> * `false`：已返回所有数据。 |
| next_page_token | String | 73505836754923*** | 翻页标识，表示下一页的起始位置。当 `has_more` 为 `true` 时，表示还有更多数据未返回，可以通过此令牌获取下一页数据。首次请求不填或置空，后续翻页需使用上一次返回的 `next_page_token`。 |
### OpenAPIVersionMetaInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| input | Object of [OpenAPIWorkflowInput](#openapiworkflowinput) | { "parameters": { "USER_INPUT": { "type": "string", "required": false }, "CONVERSATION_NAME": { "type": "string", "required": false, "description": "", "default_value": "" } } } | 工作流开始节点的输入参数的结构体。 |
| output | Object of [OpenAPIWorkflowOutput](#openapiworkflowoutput) | { "parameters": { "arr_obj_num": { "required": false, "type": "number" } }, "terminate_plan": "use_answer_content", "content": "输出为{{output}}" } | 工作流结束节点的输出参数的结构体。 |
| creator | Object of [OpenAPIUserInfo](#openapiuserinfo) | {"id":"2478774393***","name":"user41833***"} | 工作流操作者的信息，包含操作者的用户 UID 和用户名。 |
| version | String | v0.0.2 | 工作流的版本号，用于标识工作流的不同版本。 |
| created_at | String | 1700000000 | 工作流的创建时间，以 Unix 时间戳表示，单位为秒。 |
| updated_at | String | 1701234567 | 工作流的最后更新时间，以 Unix 时间戳表示，单位为秒。 |
| description | String | 工作流测试 | 工作流的描述。 |
| workflow_id | String | 73505836754923*** | 工作流 ID。 |
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
### OpenAPIUserInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 2478774393*** | 工作流操作者的扣子用户 UID。 |
| name | String | user41833*** | 工作流操作者的扣子用户名。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET  'https://api.coze.cn/v1/workflows/74800714******84300/versions?page_size=1&page_token=175266607***0&include_input_output=true' \
--header 'authorization: Bearer pat_A*******' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "has_more": false,
    "next_page_token": "17552****000",
    "items": [
      {
        "workflow_id": "7538347*****8076",
        "version": "v0.0.1",
        "description": "统计图片数量",
        "created_at": "1755251584",
        "updated_at": "1755****12",
        "creator": {
          "id": "3361****58211",
          "name": "张三"
        },
        "input": {
          "parameters": {
            "USER_INPUT": {
              "type": "string",
              "required": false
            },
            "CONVERSATION_NAME": {
              "type": "string",
              "required": false,
              "description": "本次请求绑定的会话，会自动写入消息、会从该会话读对话历史。",
              "default_value": ""
            }
          }
        },
        "output": {
          "terminate_plan": "use_answer_content",
          "content": "数量为{{obj_num}}",
          "parameters": {
            "obj_num": {
              "type": "number",
              "required": false
            }
          }
        }
      }
    ]
  },
  "detail": {
    "logid": "202508221521****5C3B159FCD2"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
