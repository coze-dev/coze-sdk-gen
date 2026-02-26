# 恢复运行工作流（流式响应）
流式恢复运行已中断的工作流。
## 接口说明
执行包含问答节点的工作流时，智能体会以指定问题向用户提问，并等待用户回答。此时工作流为中断状态，开发者需要调用此接口回答问题，并恢复运行工作流。如果用户的响应和智能体预期提取的信息不匹配，例如缺少必选的字段，或字段数据类型不一致，工作流会再次中断并追问。如果询问 3 次仍未收到符合预期的回复，则判定为工作流执行失败。
[恢复运行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_resume)和[恢复运行工作流](https://docs.coze.cn/developer_guides/resume_workflow) 的区别如下：

* 如果调用[执行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_stream_run)API，中断恢复时需要使用[恢复运行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_resume) API，该 API 通过流式返回执行结果。
* 如果调用[执行工作流](https://docs.coze.cn/developer_guides/workflow_run) API，中断恢复时需要使用[恢复运行工作流](https://docs.coze.cn/developer_guides/resume_workflow)  API，该 API 支持同步运行或异步运行返回执行结果。

## 限制说明

* 最多调用此接口恢复运行 3 次，如果第三次恢复运行时智能体仍未收到符合预期的回复，则判定为工作流执行失败。
* 恢复运行后，index 和节点 index 都会重置。
* 恢复运行工作流也会产生 token 消耗，且与[执行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_stream_run)时消耗的 token 数量相同。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflow/stream_resume <br> ``` <br>  |
| **权限** | run <br> 确保调用该接口使用的访问令牌开通了 run 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 恢复运行已中断的工作流。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workflow_id | String | 必选 | 73505836754923*** | 待执行的 Workflow ID，此工作流应已发布。 <br> 进入 Workflow 编排页面，在页面 URL 中，workflow 参数后的数字就是 Workflow ID。例如 https://www.coze.com/work_flow?space_id=42463***&workflow_id=73505836754923***，Workflow ID 为 73505836754923***。 |
| event_id | String | 必选 | 74048319882025*** | 工作流执行中断事件 ID。你可以从[执行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_stream_run)的响应信息中获得中断事件 ID。 |
| interrupt_type | Integer | 必选 | 2 | 工作流执行中断的类型，用于标识导致工作流中断的具体原因，你可以从[执行工作流](https://docs.coze.cn/developer_guides/workflow_run) 的响应信息中获得中断事件的中断类型。枚举值： <br>  <br> * `6`：端插件触发中断。 <br> * `2`：问答节点触发中断。 <br> * `5`：输入节点触发中断。 <br> * `7`：OAuth 插件触发中断。 |
| resume_data | String | 必选 | 杭州，2024-08-20 | 恢复执行时，用户对智能体指定问题的回复。 <br> 如果是问答节点导致的中断，回复中应包含问答节点中的必选参数，否则工作流会再次中断并提问。 <br> 如果要传入 Image 等类型的文件，可以调用[上传文件](https://docs.coze.cn/developer_guides/upload_files)[上传文件](https://docs.coze.cn/api/open/docs/developer_guides/upload_files)API 获取 file_id，在调用此 API 时在 resume_data 中以序列化之后的 JSON 格式传入 file_id。例如 `“resume_data” : "{\"file_id\": \"1456234***\"}"`。 |
## 返回参数

在流式响应中，开发者需要注意是否存在丢包现象。
事件 ID（id）默认从 0 开始计数，以包含 `event: Done` 的事件为结束标志。开发者应根据 id 确认响应消息整体无丢包现象。
Message 事件的消息 ID 默认从 0 开始计数，以包含 `node_is_finish : true` 的事件为结束标志。开发者应根据 node_seq_id 确认 Message 事件中每条消息均完整返回，无丢包现象。

| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 0 | 此消息在接口响应中的事件 ID。以 0 为开始。 |
| event | String | Message | 当前流式返回的数据包事件。包括以下类型： <br>  <br> * Message：工作流节点输出消息，例如输出节点、结束节点的输出消息。可以在 data 中查看具体的消息内容。 <br> * Error：报错。可以在 data 中查看 error_code 和 error_message，排查问题。 <br> * Done：结束。表示工作流执行结束，此时 data 中包含。 <br> * Interrupt：中断。表示工作流中断，此时 data 字段中包含具体的中断信息。 |
| node_seq_id | String | 0 | 此消息在节点中的消息 ID，从 0 开始计数，例如输出节点的第 5 条消息。 |
| node_title | String | End | 输出消息的节点名称，例如输出节点、结束节点。 |
| content | String | 请问你想查看哪个城市、哪一天的天气呢 | 流式输出的消息内容。 |
| node_is_finish | Boolean | true | 当前消息是否为此节点的最后一个数据包。 |
| interrupt_data | Object of [Interrupt](#interrupt) | {"data":"","event_id":"7404831988202520614/6302059919516746633","type":2} | content type为interrupt时传输，中断协议 |
| cost | String | 0 | 预留字段，无需关注。 |
| error_code | Long | 0 | 调用状态码。  <br>  <br> * 0 表示调用成功。  <br> * 其他值表示调用失败。你可以通过 error_message 字段判断详细的错误原因。 |
| error_message | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 |
| usage | Object of [Usage](#usage) |  | 资源使用情况。 |
| detail | Object of [ResponseDetail](#responsedetail) | 20241210152726467C48D89D6DB2**** | 响应详情。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
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
### Usage
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| input_count | Integer | 50 | 输入内容所消耗的 Token 数，包含对话上下文、系统提示词、用户当前输入等所有输入类的 Token 消耗。 |
| output_count | Integer | 100 | 大模型输出的内容所消耗的 Token 数。 |
| token_count | Integer | 150 | 本次 API 调用消耗的 Token 总量，包括输入和输出两部分的消耗。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location 'https://api.coze.cn/v1/workflow/stream_resume' \
--header 'Authorization: Bearer pat_vTG1****' \
--header 'Content-Type: application/json' \
--data '{
    "event_id":"740483727529459****/433802199567434****",
    "interrupt_type":2,
    "resume_data":"杭州，2024-08-20",
    "workflow_id":"739739507914235****"
}'
```

### 返回示例
```JSON
id: 0
event: Message
data: {"content":"{\"output\":[{\"condition\":\"中到大雨\",\"humidity\":72,\"predict_date\":\"2024-08-20\",\"temp_high\":35,\"temp_low\":26,\"weather_day\":\"中到大雨\",\"wind_dir_day\":\"西风\",\"wind_dir_night\":\"西风\",\"wind_level_day\":\"3\",\"wind_level_night\":\"3\"}]}","content_type":"text","cost":"0","node_is_finish":true,"node_seq_id":"0","node_title":"End","token":386}
id: 1
event: Done
data: {}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。