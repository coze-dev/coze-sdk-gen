# 恢复运行工作流
恢复运行工作流。
## 接口说明
当调用[执行工作流](https://docs.coze.cn/developer_guides/workflow_run) API 执行包含问答节点、输入节点的工作流时，智能体会以指定问题向用户提问，并等待用户回答。此时工作流为中断状态，开发者需要调用此 API 回答问题，并恢复运行工作流。如果问答节点勾选了**从回复中提取字段**，当用户的响应和智能体预期提取的信息不匹配，例如缺少必选的字段，或字段数据类型不一致，工作流会再次中断并追问。
如果[执行工作流](https://docs.coze.cn/developer_guides/workflow_run)  API 为同步运行，则恢复运行工作流也为同步运行。如果[执行工作流](https://docs.coze.cn/developer_guides/workflow_run)  API 为异步运行，则恢复运行工作流也为异步运行，你还需要调用[查询异步执行结果](https://docs.coze.cn/developer_guides/workflow_history) API 查询执行结果。
[恢复运行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_resume)和[恢复运行工作流](https://docs.coze.cn/developer_guides/resume_workflow) 的区别如下：

* 如果调用[执行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_stream_run)API 时工作流中断，恢复时需要使用[恢复运行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_resume) API，该 API 通过流式返回执行结果。
* 如果调用[执行工作流](https://docs.coze.cn/developer_guides/workflow_run) API 时工作流中断，恢复时需要使用[恢复运行工作流](https://docs.coze.cn/developer_guides/resume_workflow)  API，该 API 支持同步运行或异步运行返回执行结果。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflows/resume <br> ``` <br>  |
| **权限** | `run` <br> 确保调用该接口使用的访问令牌开通了 `run` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 恢复运行工作流。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workflow_id | String | 必选 | 73505836754923*** | 待执行的 Workflow ID，此工作流应已发布。 <br> 进入 Workflow 编排页面，在页面 URL 中，workflow 参数后的数字就是 Workflow ID。例如 `https://www.coze.com/work_flow?space_id=42463***&workflow_id=73505836754923***`，Workflow ID 为 `73505836754923***`。 |
| event_id | String | 必选 | 74048319882025*** | 工作流执行中断事件 ID。你可以从[执行工作流](https://docs.coze.cn/developer_guides/workflow_run) 的响应信息中获得中断事件 ID。 |
| resume_data | String | 必选 | 杭州，2024-08-20 | 恢复执行时，用户对智能体指定问题的回复。 <br> 如果是问答节点或输入节点导致的中断，回复中应包含对应节点中的必选参数，否则工作流会再次中断并提问。 <br> 如果要传入 Image 等类型的文件，可以调用[上传文件](https://docs.coze.cn/developer_guides/upload_files)API 获取 file_id，在调用此 API 时在 resume_data 中以序列化之后的 JSON 格式传入 file_id。例如 `“resume_data” : "{\"file_id\": \"1456234***\"}"`。 |
| interrupt_type | Integer | 必选 | 2 | 工作流执行中断的类型，用于标识导致工作流中断的具体原因，你可以从[执行工作流](https://docs.coze.cn/developer_guides/workflow_run) 的响应信息中获得中断事件的中断类型。枚举值： <br>  <br> * `6`：端插件触发中断。 <br> * `2`：问答节点触发中断。 <br> * `5`：输入节点触发中断。 <br> * `7`：OAuth 插件触发中断。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | String | {\"output\":\"杭州当天的天气为小雨。\"} | 同步运行时返回的工作流执行结果，通常为 JSON 序列化字符串，部分场景下可能返回非 JSON 结构的字符串。 |
| debug_url | String | https://www.coze.cn/work_flow?execute_id=743104097880585****&space_id=730976060439760****&workflow_id=742963539464539**** | 工作流试运行调试页面。访问此页面可查看每个工作流节点的运行结果、输入输出等信息。 <br> debug_url 的访问有效期为 7 天，过期后将无法访问。 <br>  |
| usage | Object of [Usage](#usage) | {"input_count":50,"token_count":150,"output_count":100} | 本次 API 调用消耗的 Token 数量。 <br> 此处大模型返回的消耗 Token 仅供参考，以[火山引擎账单](https://console.volcengine.com/finance/bill/detail)实际为准。 |
| interrupt_data | Object of [Interrupt](#interrupt) | {"data":"{\"content_type\":\"text\",\"content\":\"请输入您的姓名\"}","type":2,"event_id":"740483198820252***","required_parameters":{"name":{"type":"string","required":true}}} | 中断事件的详细信息，包含中断控制内容、中断类型和事件 ID 等信息。 |
| execute_id | String | 743104097880585**** | 异步运行时的事件 ID。如果[执行工作流](https://docs.coze.cn/developer_guides/workflow_run)为异步运行时，本 API 也为异步运行，响应信息中会返回 execute_id。 <br> 调用[查询异步执行结果](https://docs.coze.cn/developer_guides/workflow_history) API 时需要传入此 ID，查询工作流的执行状态或结果。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的日志信息，用于问题排查和技术支持。 |
### Usage
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| input_count | Integer | 50 | 输入内容所消耗的 Token 数，包含对话上下文、系统提示词、用户当前输入等所有输入类的 Token 消耗。 |
| output_count | Integer | 100 | 大模型输出的内容所消耗的 Token 数。 |
| token_count | Integer | 150 | 本次 API 调用消耗的 Token 总量，包括输入和输出两部分的消耗。 |
### Interrupt
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | String | {\"content_type\":\"text\",\"content\":\"[{\\\"type\\\":\\\"string\\\",\\\"name\\\":\\\"img\\\",\\\"required\\\":true,\\\"assistType\\\":2}]\"} | 中断控制内容，用于在工作流中断时传递控制信息。当工作流需要用户输入或执行特定操作时，通过此字段传递相关信息。 |
| type | Integer | 2 | 工作流中断类型，调用[恢复运行工作流](https://docs.coze.cn/developer_guides/resume_workflow) API 恢复运行时应回传此字段。 <br> 枚举值： <br>  <br> * `6`：端插件触发中断。 <br> * `2`：问答节点触发中断。 <br> * `5`：输入节点触发中断。 <br> * `7`：OAuth 插件触发中断。 |
| event_id | String | 740483198820252*** | 工作流中断事件 ID，调用[恢复运行工作流](https://docs.coze.cn/developer_guides/resume_workflow) API 恢复运行时应回传此字段。 |
| required_parameters | JSON Map | { "img": { "required": true, "type": "image" } } | 工作流中断时需要补充的参数信息，采用键值对（key-value）结构，其中 key 为输入节点对应的参数名称，value 为该参数的定义信息（包含类型、是否必填等属性）。 |
### RequiredParameters
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| type | String | string | 该参数的类型。 |
| items | Object of [OpenAPIParameter](#openapiparameter) | {"type":"image"} | 当参数类型为 `array` 时，该字段用于定义数组元素的子类型。 |
| required | Boolean | true | 标识输入参数是否为必填项。 <br>  <br> * `true`：该参数为必填项。 <br> * `false`：该参数为可选项。 |
| properties | JSON Map | { "arr_obj_num": { "required": false, "type": "number" } } | 当参数类型为 `object`时，该字段用于定义对象类型的子参数信息，采用键值对（key-value）结构，其中 key 为子参数的名称，value 为该子参数的定义信息（包含类型、是否必填等属性）。 |
| description | String | 上传的图片。 | 该参数的描述信息。 |
| default_value | String | - | 该参数配置的默认值。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/workflows/resume' \
--header 'Authorization: Bearer pat_I6w8e****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "event_id": "75683404684*****/25042098367****",
    "interrupt_type": 2,
    "resume_data": "{\"file_id\": \"1456234***\"}",
    "workflow_id": "73505836754923***"
}'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "debug_url": "https://api.coze.cn/work_flow?execute_id=756949870377***&space_id=75687968944***&workflow_id=75688742653***&execute_mode=2",
  "interrupt_data": {
    "data": "{\"content_type\":\"text\",\"content\":\"[{\\\"type\\\":\\\"string\\\",\\\"name\\\":\\\"img\\\",\\\"required\\\":true,\\\"assistType\\\":2}]\"}",
    "required_parameters": {
      "img": {
        "required": true,
        "type": "image"
      }
    },
    "event_id": "75694987037***/2691977632***",
    "type": 5
  },
  "execute_id": "7569498703774***",
  "detail": {
    "logid": "202511061720005429856FA5***"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://www.coze.cn/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。
