# 执行工作流（流式响应）
执行已发布的工作流，响应方式为流式响应。
## 接口说明
调用 API 执行工作流时，对于支持流式输出的工作流，往往需要使用流式响应方式接收响应数据，例如实时展示工作流的输出信息、呈现打字机效果等。
在流式响应中，服务端不会一次性发送所有数据，而是以数据流的形式逐条发送数据给客户端，数据流中包含工作流执行过程中触发的各种事件（event），直至处理完毕或处理中断。处理结束后，服务端会通过 `event: Done` 事件提示工作流执行完毕。各个事件的说明可参考[返回结果](https://docs.coze.cn/api/open/docs/developer_guides/workflow_stream_run#970775c1)。
目前支持流式响应的工作流节点包括**输出节点**、**问答节点**和**开启了流式输出的结束节点**。对于不包含这些节点的工作流，可以使用[执行工作流](https://docs.coze.cn/api/open/docs/developer_guides/workflow_run)接口一次性接收响应数据。

## 限制说明

* 通过 API 方式执行工作流前，应确认此工作流已发布，执行从未发布过的工作流时会返回错误码 4200。创建并发布工作流的操作可参考[使用低代码工作流](https://docs.coze.cn/api/open/docs/guides/use_workflow)。
* 调用此 API 之前，应先在扣子编程中试运行此工作流。
   * 如果试运行时需要关联智能体，则调用此 API 执行工作流时，也需要指定 bot_id。通常情况下，执行存在数据库节点、变量节点等节点的工作流需要关联智能体。
   * 执行应用中的工作流时，需要指定 app_id。
   * 请勿同时指定 bot_id 和 app_id，否则 API 会报错 4000，表示请求参数错误。
* 此接口为同步接口，如果工作流整体或某些节点运行超时，智能体可能无法提供符合预期的回复，建议将工作流的执行时间控制在 5 分钟以内。同步执行时，工作流整体超时时间限制可参考[低代码工作流使用限制](https://docs.coze.cn/api/open/docs/guides/workflow_limits)。
* 工作流支持的请求大小上限为 20MB，包括输入参数以及运行期间产生的消息历史等所有相关数据。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflow/stream_run <br> ``` <br>  |
| **权限** | `run` <br> 确保调用该接口使用的访问令牌开通了 `run` 权限，详细信息参考[鉴权方式概述](https://docs.coze.cn/api/open/docs/developer_guides/authentication)。 |
| **接口说明** | 执行已发布的工作流，响应方式为流式响应。 |
## Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer <span style="color: #D83931"><em>$Access_Token</em></span> | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/api/open/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |


## Body
| **参数** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| workflow_id | String  | 必选 | 待执行的 Workflow ID，此工作流应已发布。 <br> 进入 Workflow 编排页面，在页面 URL 中，`workflow` 参数后的数字就是 Workflow ID。例如 `https://www.coze.com/work_flow?space_id=42463***&workflow_id=73505836754923***`，Workflow ID 为 `73505836754923***`。 |
| parameters | map[String]Any | 可选 | 工作流开始节点的输入参数及取值，你可以在指定工作流的编排页面查看参数列表。 <br> 如果工作流输入参数为 Image 等类型的文件，你可以传入文件 URL 或调用[上传文件](https://www.coze.cn/open/docs/developer_guides/upload_files) API 获取 file_id 后传入 file_id。示例： <br>  <br> * 上传文件并传入 file_id： <br>    * 单个文件示例：`"parameters": { "image": "{\"file_id\":\"1122334455\"}" }` <br>    * 文件数组示例：`"parameters": { "image": [ "{\"file_id\":\"1122334455\"}" ] }`。 <br> * 传入文件 URL： <br>  <br> `“parameters” :{"input":"请总结图片内容", "image": "https://example.com/tos-cn-i-mdko3gqilj/example.png" } ` |
| bot_id <br>  | String  <br>  | 可选 <br>  | 需要关联的智能体ID。 部分工作流执行时需要指定关联的智能体，例如存在数据库节点、变量节点等节点的工作流。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/55746fa5540b488ea83a79064a223500~tplv-goo7wpa0wc-image.image) <br> 进入智能体的开发页面，开发页面 URL 中 `bot` 参数后的数字就是智能体ID。例如 `https://www.coze.com/space/341****/bot/73428668*****`，Bot ID 为 `73428668*****`。  <br> * 确保调用该接口使用的令牌开通了此智能体所在空间的权限。 <br> * 确保该智能体已发布为 API 服务。 <br>  |
| ext | Map[String][String] <br>  | 可选 | 用于指定一些额外的字段，以 Map[String][String] 格式传入。例如某些插件会隐式用到的经纬度等字段。 <br> 目前仅支持以下字段： <br>  <br> * latitude：String 类型，表示纬度。 <br> * longitude：String 类型，表示经度。 <br> * user_id：String 类型，表示用户 ID。 |
| app_id | String | 可选 | 工作流所在的应用 ID。 <br> 你可以通过应用的业务编排页面 URL 中获取应用 ID，也就是 URL 中 project-ide 参数后的一串字符，例如 `https://www.coze.cn/space/739174157340921****/project-ide/743996105122521****/workflow/744102227704147****` 中，应用的 ID 为 `743996105122521****`。 <br> 仅运行扣子应用中的工作流时，才需要设置 app_id。智能体绑定的工作流、空间资源库中的工作流无需设置 app_id。 <br>  |
| workflow_version <br>  <br>  | String <br>  | 可选 <br>  | 工作流的版本号，仅当运行的工作流属于资源库工作流时有效。未指定版本号时默认执行最新版本的工作流。 |
| connector_id | String <br>  | 可选 <br>  | 渠道 ID，用于配置该工作流在什么渠道执行。 <br> 当智能体或扣子应用发布到某个渠道后，可以通过该参数指定工作流在对应的渠道执行。 <br> 扣子编程的渠道 ID 包括： <br>  <br> * 1024（默认值）：API 渠道。 <br> * 999：Chat SDK。 <br> * 998：WebSDK。 <br> * 10000122：扣子商店。 <br> * 10000113：微信客服。 <br> * 10000120：微信服务号。 <br> * 10000121：微信订阅号。 <br> * 10000126：抖音小程序。 <br> * 10000127：微信小程序。 <br> * 10000011：飞书。 <br> * 10000128：飞书多维表格。 <br> * 10000117：掘金。 <br> * 自定义渠道 ID。自定义渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 <br>  <br> 不同渠道的用户数据、会话记录等相互隔离，进行数据分析统计时，不支持跨渠道数据调用。 <br>  |
## 返回结果

在流式响应中，开发者需要注意是否存在丢包现象。

* 事件 ID（id）默认从 0 开始计数，以包含 `event: Done` 的事件为结束标志。开发者应根据 id 确认响应消息整体无丢包现象。
* Message 事件的消息 ID 默认从 0 开始计数，以包含 `node_is_finish : true` 的事件为结束标志。开发者应根据 node_seq_id 确认 Message 事件中每条消息均完整返回，无丢包现象。

| **参数名** | **参数类型** | **参数描述** |
| --- | --- | --- |
| id | Integer | 此消息在接口响应中的事件 ID。以 0 为开始。 |
| event | String  | 当前流式返回的数据包事件。包括以下类型： <br>  <br> * Message：工作流节点输出消息，例如输出节点、结束节点的输出消息。可以在 data 中查看具体的消息内容。 <br> * Error：报错。可以在 data 中查看 error_code 和 error_message，排查问题。 <br> * Done：结束。表示工作流执行结束，此时 data 中包含 debug URL。 <br> * Interrupt：中断。表示工作流中断，此时 data 字段中包含具体的中断信息。 <br> * PING：心跳信号。表示工作流执行中，消息内容为空，用于维持连接。 |
| data | Object | 事件内容。各个 event 类型的事件内容格式不同。 |
### Message 事件
Message 事件中，data 的结构如下：
| **参数名** | **参数类型** | **参数描述** |
| --- | --- | --- |
| content | String  | 流式输出的消息内容。 |
| node_title | String | 输出消息的节点名称，例如输出节点、结束节点。 |
| node_seq_id | String | 此消息在节点中的消息 ID，从 0 开始计数，例如输出节点的第 5 条消息。 |
| node_is_finish | Boolean | 当前消息是否为此节点的最后一个数据包。 |
| ext | Map[String]String | 额外字段。 |
| usage | Object of [Usage](https://docs.coze.cn/api/open/docs/developer_guides/workflow_run#1fc85191) | 资源使用情况，包含本次 API 调用消耗的 Token 数量等信息。 |
| node_id | String | 输出消息的节点 ID。 |
| node_execute_uuid | String | 节点每次执行的 ID，用于追踪和识别工作流中特定节点的单次执行情况。 |
### usage
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| input_count <br>  <br>  | Integer <br>  | 50 <br>  | 输入内容所消耗的 Token 数，包含对话上下文、系统提示词、用户当前输入等所有输入类的 Token 消耗。 |
| output_count <br>  <br>  | Integer <br>  | 100 <br>  | 大模型输出的内容所消耗的 Token 数。 <br>  <br>  |
| token_count <br>  <br>  | Integer <br>  | 150 <br>  | 本次 API 调用消耗的 Token 总量，包括输入和输出两部分的消耗。 <br>  <br>  |
### Interrupt 事件
Interrupt 事件中，data 的结构如下：
| **参数名** | **参数类型** | **参数描述** |
| --- | --- | --- |
| interrupt_data | Object | 中断控制内容。 |
| interrupt_data.event_id | String | 工作流中断事件 ID，恢复运行时应回传此字段，具体请参见[恢复运行工作流（流式响应）](https://docs.coze.cn/api/open/docs/developer_guides/workflow_resume)。 |
| interrupt_data.type | Integer | 工作流中断类型，恢复运行时应回传此字段，具体请参见[恢复运行工作流（流式响应）](https://docs.coze.cn/api/open/docs/developer_guides/workflow_resume)。 |
| interrupt_data.required_parameters | map<String,WorkflowParameter> | 工作流中断时需要补充的参数信息，采用键值对（key-value）结构：key为参数名称，value为该参数的值。 <br>  |
| node_title | String | 输出消息的节点名称，例如“问答”。 |
### Error 事件
Error 事件中，data 的结构如下：
| **参数名** | **参数类型** | **参数描述** |
| --- | --- | --- |
| error_code | Integer | 调用状态码。  <br>  <br> * 0 表示调用成功。  <br> * 其他值表示调用失败。你可以通过 error_message 字段判断详细的错误原因。 |
| error_message | String  | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 |
## 

## 示例
### 请求示例
```Shell
curl --location --request POST 'https://api.coze.cn/v1/workflow/stream_run' \
--header 'Authorization: Bearer pat_fhwefweuk****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "workflow_id": "73664689170551*****",
    "parameters": {
        "user_name":"George"
    }
}'
```

### 响应示例

<div type="doc-tabs">
<div type="tab-item" title="Message 事件" key="jXIyUj55Xb">

```JSON
id: 0
event: Message
data: {"content":"msg","node_is_finish":false,"node_seq_id":"0","node_title":"Message"}

id: 1
event: Message
data: {"content":"为","node_is_finish":false,"node_seq_id":"1","node_title":"Message"}

id: 2
event: Message
data: {"content":"什么小明要带一把尺子去看电影？\n因","node_is_finish":false,"node_seq_id":"2","node_title":"Message"}

id: 3
event: Message
data: {"content":"为他听说电影很长，怕","node_is_finish":false,"node_seq_id":"3","node_title":"Message"}

id: 4
event: Message
data: {"content":"坐不下！","node_is_finish":true,"node_seq_id":"4","node_title":"Message"}

id: 5
event: Message
data: {"content":"{\"output\":\"为什么小明要带一把尺子去看电影？\\n因为他听说电影很长，怕坐不下！\"}","cost":"0.00","node_is_finish":true,"node_seq_id":"0","node_title":"","token":0}

id: 6
event: Done
data: {}
```


</div>
<div type="tab-item" title="Error 事件" key="D5C15jhYFQ">

```JSON
id: 0
event: Error
data: {"error_code":4000,"error_message":"Request parameter error"}
```


</div>
<div type="tab-item" title="Interrupt 事件" key="brFfAaCBdU">

```JSON
// 流式执行工作流，触发问答节点，Bot提出问题
id: 0
event: Message
data: {"content":"请问你想查看哪个城市、哪一天的天气呢","content_type":"text","node_is_finish":true,"node_seq_id":"0","node_title":"问答"}

id: 1
event: Interrupt
data: {"interrupt_data":{"data":"","event_id":"7404830425073352713/2769808280134765896","type":2},"node_title":"问答"}
```


</div>
</div>
## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/api/open/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。