# 查询工作流异步运行结果
工作流异步运行后，查看执行结果。
## 接口说明
调用[执行工作流](https://docs.coze.cn/developer_guides/workflow_run)或[恢复运行工作流](https://docs.coze.cn/developer_guides/resume_workflow) API 时，如果选择异步运行工作流，响应信息中会返回 execute_id，开发者可以通过本 API 查询指定事件的执行结果。
## 限制说明

* 本 API 的流控限制请参见 [API 介绍](https://docs.coze.cn/developer_guides/coze_api_overview)。
* 工作流的**输出节点**的输出数据最多保存 24 小时，**结束节点**为 7 天。
* 输出节点的输出内容超过 1MB 时，无法保证返回内容的完整性。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflows/:workflow_id/run_histories/:execute_id <br> ``` <br>  |
| **权限** | `listRunHistory` <br> 确保调用该接口使用的访问令牌开通了 `listRunHistory` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 工作流异步运行后，查看执行结果。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workflow_id | String | 可选 | 73505836754923*** | 待执行的 Workflow ID，此工作流应已发布。 <br> 进入 Workflow 编排页面，在页面 URL 中，workflow 参数后的数字就是 Workflow ID。例如` https://www.coze.com/work_flow?space_id=42463***&workflow_id=73505836754923***`，Workflow ID 为 `73505836754923***`。 |
| execute_id | String | 可选 | 743104097880585**** | 工作流执行 ID。调用接口[执行工作流](https://docs.coze.cn/developer_guides/workflow_run)，如果选择异步执行工作流，响应信息中会返回 execute_id。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Array of [WorkflowExecuteHistory](#workflowexecutehistory) | \ | 异步工作流的执行结果。 <br> 每次只能查询一个异步事件的执行结果，所以此数组只有一个对象。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 本次请求的执行详情。 |
### WorkflowExecuteHistory
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| execute_id | String | 743104097880585**** | 工作流执行 ID。 |
| execute_status | String | Success | 执行状态。 <br>  <br> * Success：执行成功。 <br> * Running：执行中。 <br> * Fail：执行失败。 |
| bot_id | String | 75049216555930**** | 执行工作流时指定的 Agent ID。返回 0 表示未指定智能体 ID。 |
| connector_id | String | 1024 | 智能体的发布渠道 ID，默认仅显示 Agent as API 渠道，渠道 ID 为 1024。 |
| connector_uid | String | 123 | 用户 UID，执行工作流时通过 ext 字段指定的 user_id。如果未指定，则返回 Token 申请人的扣子用户 UID。 |
| run_mode | Integer | 0 | 工作流的运行方式： <br>  <br> * 0：同步运行。 <br> * 1：流式运行。 <br> * 2：异步运行。 |
| output | String | {\"Output\":\"{\\\"content_type\\\":1,\\\"data\\\":\\\"来找姐姐有什么事呀\\\",\\\"original_result\\\":null,\\\"type_for_model\\\":2}\"} | 工作流的输出，通常为 JSON 序列化字符串，也有可能是非 JSON 结构的字符串。 <br> 工作流输出的内容包括： <br>  <br> * 输出节点的输出。 <br> * 结束节点的输出。在扣子编程代码中，结束节点的输出是通过键（key）`Output` 来标识。 <br>  <br> 工作流输出的结构如下所示： <br> ```JSON <br> { <br>   "Output": "结束节点的输出内容", <br>   "输出节点_1": "输出节点_1的输出内容", <br>   "输出节点_2": "输出节点_2的输出内容" <br> } <br> ``` <br>  |
| create_time | Long | 1730174063 | 工作流运行开始时间，Unixtime 时间戳格式，单位为秒。 |
| update_time | Long | 1730174063 | 工作流的恢复运行时间，Unixtime 时间戳格式，单位为秒。 |
| node_execute_status | JSON Map | \ | 输出节点的运行情况。字段的格式为：`key:node_status,value:map[node_title]*nodeExecuteStatus{}`。 <br> `key`为节点的名称，如果节点运行了多次，则会随机生成节点名称。 <br> 当输出节点的输出内容超过 1MB 时，调用本 API 会导致返回内容不完整，建议通过[查询工作流节点的输出](https://docs.coze.cn/developer_guides/get_node_execute_history_response) API 逐一查询各节点的输出内容。 <br>  |
| error_code | String | "" | 执行失败调用状态码。0 表示调用成功。其他值表示调用失败。你可以通过 error_message 字段判断详细的错误原因。 |
| debug_url | String | https://www.coze.cn/work_flow?execute_id=743104097880585****&space_id=730976060439760****&workflow_id=742963539464539**** | 工作流试运行调试页面。访问此页面可查看每个工作流节点的运行结果、输入输出等信息。 <br> debug_url 的访问有效期为 7 天，过期后将无法访问。 <br>  |
| usage | Object of [Usage](#usage) | {"input_count":50,"token_count":150,"output_count":100} | 本次 API 调用消耗的 Token 数量。 <br> 此处大模型返回的消耗 Token 仅供参考，以[火山引擎账单](https://console.volcengine.com/finance/bill/detail)实际为准。 |
| is_output_trimmed | Boolean | false | 标识工作流的输出内容是否因过大而不完整。 <br>  <br> * `true`：输出内容因过大被截断。 <br> * `false`：输出内容完整。 |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
| error_message | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 |
| interrupt_data | Object of [Interrupt](#interrupt) | {"data":"{\"content_type\":\"text\",\"content\":\"请输入您的姓名\"}","type":2,"event_id":"740483198820252***","required_parameters":{"name":{"type":"string","required":true}}} | 中断事件的详细信息，包含中断控制内容、中断类型和事件 ID 等信息。 |
### NodeExecuteStatus
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| node_id | String | node_123 | 工作流中的节点 ID。 |
| is_finish | Boolean | true | 工作流中的节点是否已经运行结束。 |
| loop_index | Long | 2 | 当前节点在循环节点中的循环次数。 <br> 第一次循环时值为 `0`。 <br> 仅当节点为循环节点，且未嵌套子工作流时，才会返回该参数。 <br>  |
| batch_index | Long | 3 | 当前节点在批处理节点中的执行次数。 <br> 第一次执行时值为 `0`。 <br> 仅当节点为批处理节点，且未嵌套子工作流时，才会返回该参数。 <br>  |
| update_time | Long | 1730174063 | 工作流上次运行的时间，采用 Unix 时间戳格式，单位为秒。 |
| sub_execute_id | String | 743104097880585**** | 子工作流执行的 ID。 |
| node_execute_uuid | String | 78923456777***** | 节点每次执行的 ID，用于追踪和识别工作流中特定节点的单次执行情况。 |
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
curl --location --request GET 'https://api.coze.cn/v1/workflows/742963539464539****/run_histories/743104097880585****' \
--header 'Authorization: Bearer pat_xitq9LWlowpX3qGCih1lwpAdzvXNqgmpfhpV28HLWFypY37xR5Uaj2GioN****' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "detail": {
    "logid": "20241029152003BC531DC784F1897B****"
  },
  "code": 0,
  "msg": "",
  "data": [
    {
      "update_time": 1730174065,
      "usage": {
        "input_count": 50,
        "token_count": 150,
        "output_count": 100
      },
      "output": "{\"Output\":\"{\\\"content_type\\\":1,\\\"data\\\":\\\"来找姐姐有什么事呀\\\",\\\"original_result\\\":null,\\\"type_for_model\\\":2}\"}",
      "bot_id": "742963486232569****",
      "token": "0",
      "execute_status": "Success",
      "connector_uid": "223687073464****",
      "run_mode": 0,
      "connector_id": "1024",
      "logid": "20241029115423ED85C3401395715F726E",
      "debug_url": "https://www.coze.cn/work_flow?execute_id=743104097880585****&space_id=730976060439760****&workflow_id=742963539464539****",
      "error_code": "",
      "error_message": "",
      "execute_id": "743104097880585****",
      "create_time": 1730174063
    }
  ]
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
