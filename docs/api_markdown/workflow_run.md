# 执行工作流
执行已发布的工作流。
## 接口说明
此接口为非流式响应模式，如果需要采用流式输出，请参考[执行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_stream_run)。
调用此接口后，你可以从响应中获得 debug_url，访问链接即可通过可视化界面查看工作流的试运行过程，其中包含每个执行节点的输入输出等详细信息，帮助你在线调试或排障。
扣子个人付费版、企业版（企业标准版、企业旗舰版）用户调用此接口时，支持通过 `is_async` 参数异步运行工作流，适用于工作流执行耗时较长，导致运行超时的情况。异步运行后可通过本接口返回的 execute_id 调用[查询工作流异步执行结果](https://docs.coze.cn/developer_guides/workflow_history) API 获取工作流的执行结果。
## 限制说明
调用此非流式响应 API 时，若 API 在 90 秒内未收到响应，将因超时而断开连接。对于执行耗时较长的工作流，建议使用[执行工作流（流式响应）](https://docs.coze.cn/developer_guides/workflow_stream_run)API。

| **限制项** | **说明** |
| --- | --- |
| 工作流发布状态 |  必须为已发布。执行未发布的工作流会返回错误码 4200。 创建并发布工作流的操作可参考[使用工作流](https://docs.coze.cn/guides/use_workflow)。 |
| 节点限制 | 工作流中不能包含输出节点、开启了流式输出的结束节点。 |
| 关联智能体 | 调用此 API 之前，应先在扣子平台中试运行此工作流，如果试运行时需要关联智能体，则调用此 API 执行工作流时，也需要指定智能体ID。通常情况下，执行存在数据库节点、变量节点等节点的工作流需要关联智能体。 |
| 请求大小上限 |  20 MB，包括输入参数及运行期间产生的消息历史等所有相关数据。  |
| 超时时间  | * 未开启工作流异步运行时，工作流整体超时时间为 10 分钟，建议执行时间控制在 5 分钟以内，否则不保障执行结果的准确性。 详细说明可参考[工作流使用限制](https://docs.coze.cn/guides/workflow_limits)。 <br> * 开启工作流异步运行后，工作流整体超时时间为 24 小时。 |

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workflow/run <br> ``` <br>  |
| **权限** | `run` <br> 确保调用该接口使用的访问令牌开通了 `run` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 执行已发布的工作流。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子平台中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workflow_id | String | 必选 | 73664689170551***** | 待执行的 Workflow ID，此工作流应已发布。 <br> 进入 Workflow 编排页面，在页面 URL 中，`workflow` 参数后的数字就是 Workflow ID。例如 `https://www.coze.com/work_flow?space_id=42463***&workflow_id=73505836754923***`，Workflow ID 为 `73505836754923***`。 |
| parameters | String | 可选 | {"image": "{\"file_id\":\"1122334455\"}","user_name":"George"} | 工作流开始节点的输入参数及取值，以 JSON 序列化字符串形式传入。你可以在指定工作流的编排页面查看输入参数列表。 <br> 如果工作流输入参数为 Image 等类型的文件，你可以传入文件 URL 或调用[上传文件](https://docs.coze.cn/developer_guides/upload_files) API 获取 file_id 后传入 file_id。示例： <br>  <br> * 上传文件并传入 file_id： <br>    * 单个文件示例：`"parameters": { "image": "{\"file_id\":\"1122334455\"}" }` <br>    * 文件数组示例：`"parameters": { "image": [ "{\"file_id\":\"1122334455\"}" ] }`。 <br> * 传入文件 URL：`“parameters” :{"input":"请总结图片内容", "image": "https://example.com/tos-cn-i-mdko3gqilj/example.png" } ` |
| bot_id | String | 可选 | 73428668***** | 需要关联的智能体 ID。 部分工作流执行时需要指定关联的智能体，例如存在数据库节点、变量节点等节点的工作流。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/55746fa5540b488ea83a79064a223500~tplv-goo7wpa0wc-image.image) <br> 进入智能体的开发页面，开发页面 URL 中 bot 参数后的数字就是智能体t ID。例如 `https://www.coze.com/space/341****/bot/73428668*****`，智能体 ID 为 `73428668*****`。 <br> * 确保调用该接口使用的令牌开通了此智能体所在空间的权限。 <br> * 确保该智能体已发布为 API 服务。 <br>  |
| app_id | String | 可选 | 744208683** | 该工作流关联的扣子应用的 ID。 <br> 进入应用开发界面，开发页面 URL 中的 `project-ide` 参数后的数字就是 AppID，例如`https://www.coze.cn/space/74421656*****/project-ide/744208683**` ，扣子应用 ID 为`744208683**`**。** <br> 仅运行扣子应用中的工作流时，才需要设置 app_id。智能体绑定的工作流、空间资源库中的工作流无需设置 app_id。 <br>  |
| ext | JSON Map | 可选 | {"latitude": "39.9042", "longitude": "116.4074"} | 用于指定一些额外的字段，以 Map[String][String] 格式传入。例如某些插件会隐式用到的经纬度等字段。 <br> 目前仅支持以下字段： <br>  <br> * latitude：String 类型，表示纬度。 <br> * longitude：String 类型，表示经度。 <br> * user_id：String 类型，表示用户 ID。 |
| is_async | Boolean | 可选 | true | 是否异步运行。异步运行后可通过本接口返回的 execute_id 调用[查询工作流异步执行结果](https://docs.coze.cn/developer_guides/workflow_history)API 获取工作流的最终执行结果。 <br>  <br> * true：异步运行。 <br> * false：（默认）同步运行。 <br>  <br> 异步运行的参数 is_async 仅限扣子个人付费版、企业版（企业标准版、企业旗舰版）使用，否则调用此接口会报错 6003 Workflow execution with is_async=true is a premium feature available only to Coze Professional users <br>  |
| workflow_version | String | 可选 | v0.0.5 | 工作流的版本号，仅当运行的工作流属于资源库工作流时有效。未指定版本号时默认执行最新版本的工作流。 |
| connector_id | String | 可选 |  | 渠道 ID，用于配置该工作流在什么渠道执行。 <br> 当智能体或扣子应用发布到某个渠道后，可以通过该参数指定工作流在对应的渠道执行。 <br> 扣子的渠道 ID 包括： <br>  <br> * 1024（默认值）：API 渠道。 <br> * 999：Chat SDK。 <br> * 998：WebSDK。 <br> * 10000122：扣子商店。 <br> * 10000113：微信客服。 <br> * 10000120：微信服务号。 <br> * 10000121：微信订阅号。 <br> * 10000126：抖音小程序。 <br> * 10000127：微信小程序。 <br> * 10000011：飞书。 <br> * 10000128：飞书多维表格。 <br> * 10000117：掘金。 <br> * 自定义渠道 ID。自定义渠道 ID 的获取方式如下：在扣子左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 <br>  <br> 不同渠道的用户数据、会话记录等相互隔离，进行数据分析统计时，不支持跨渠道数据调用。 <br>  |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | String | \ | 工作流执行结果，通常为 JSON 序列化字符串，部分场景下可能返回非 JSON 结构的字符串。 |
| execute_id | String | 741364789030728**** | 异步执行的事件 ID。 |
| debug_url | String | https://www.coze.cn/work_flow?execute_id=741364789030728****&space_id=736142423532160****&workflow_id=738958910358870**** | 工作流试运行调试页面。访问此页面可查看每个工作流节点的运行结果、输入输出等信息。 <br> debug_url 的访问有效期为 7 天，过期后将无法访问。 <br>  |
| usage | Object of [Usage](#usage) | {"input_count":50,"token_count":150,"output_count":100} | 资源使用情况，包含本次 API 调用消耗的 Token 数量等信息。 <br> 此处大模型返回的消耗 Token 仅供参考，以[火山引擎账单](https://console.volcengine.com/finance/bill/detail)实际为准。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| code | Long | 0 | 调用状态码。 <br>  <br> * 0 表示调用成功。 <br> * 其他值表示调用失败。你可以通过 msg 字段判断详细的错误原因。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
| interrupt_data | Object of [Interrupt](#interrupt) | {"data":"{\"content_type\":\"text\",\"content\":\"请输入您的姓名\"}","type":2,"event_id":"740483198820252***","required_parameters":{"name":{"type":"string","required":true}}} | 中断事件的详细信息，包含中断控制内容、中断类型和事件 ID 等信息。 |
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
## 示例
### 请求示例

<div type="doc-tabs">
<div type="tab-item" title="同步执行工作流" key="_kYQh_2BaSMTjs7LwRYHS">

```JSON
curl --location --request POST 'https://api.coze.cn/v1/workflow/run' \
--header 'Authorization: Bearer pat_hfwkehfncaf****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "workflow_id": "73664689170551*****",
    "parameters": {
        "user_name":"George"
    }
}'
```


</div>
<div type="tab-item" title="异步执行工作流" key="jgPbGNPRlu5UsTjHbq03t">

```JSON
curl --location --request POST 'https://api.coze.cn/v1/workflow/run' \
--header 'Authorization: Bearer pat_hfwkehfncaf****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "workflow_id": "73664689170551*****",
    "parameters": {
        "user_name":"George"
    },
    "is_async": true
}'
```


</div>
</div>
### 返回示例

<div type="doc-tabs">
<div type="tab-item" title="同步执行工作流" key="6LlfytkgCs05JyHWI44I7">

```JSON
{
  "code": 0,
  "data": "{\"output\":\"北京的经度为116.4074°E，纬度为39.9042°N。\"}",
  "debug_url": "https://www.coze.cn/work_flow?execute_id=741364789030728****&space_id=736142423532160****&workflow_id=738958910358870****",
  "msg": "",
  "usage": {
    "input_count": 50,
    "token_count": 150,
    "output_count": 100
  }
}
```


</div>
<div type="tab-item" title="异步执行工作流" key="yNcpsneHcOmL9oV5ILWwF">

```JSON
{
  "debug_url": "https://www.coze.cn/work_flow?execute_id=742482313128840****&space_id=731375784444321****&workflow_id=74243949454920****",
  "execute_id": "74248231312884****",
  "msg": ""
}
```


</div>
</div>
## 错误码
如果成功调用扣子的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。