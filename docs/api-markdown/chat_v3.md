# 发起对话
调用此接口发起一次对话，支持添加上下文和流式响应。
会话、对话和消息的概念说明，可参考[基础概念](https://docs.coze.cn/developer_guides/coze_api_overview#4a288f73)。 
## 接口说明 
发起对话接口用于向指定智能体发起一次对话，支持在对话时添加对话的上下文消息，以便智能体基于历史消息做出合理的回复。开发者可以按需选择响应方式，即流式或非流式响应，响应方式决定了开发者获取智能体回复的方式。关于获取智能体回复的详细说明可参考[通过对话接口获取智能体回复](https://docs.coze.cn/developer_guides/get_chat_response)。 

* **流式响应**：智能体在生成回复的同时，将回复消息以数据流的形式逐条发送给客户端。处理结束后，服务端会返回一条完整的智能体回复。详细说明可参考[流式响应](https://docs.coze.cn/developer_guides/chat_v3#AJThpr1GJe)。 
* **非流式响应**：无论对话是否处理完毕，立即发送响应消息。开发者可以通过接口[查看对话详情](https://docs.coze.cn/developer_guides/retrieve_chat)确认本次对话处理结束后，再调用[查看对话消息详情](https://docs.coze.cn/developer_guides/list_chat_messages)接口查看模型回复等完整响应内容。详细说明可参考[非流式响应](https://docs.coze.cn/developer_guides/chat_v3#337f3d53)。

**创建会话** API 和**发起对话** API 的区别如下：

* 创建会话：
   * 主要用于初始化一个新的会话环境。
   * 一个会话是Bot和用户之间的一段问答交互，可以包含多条消息。
   * 创建会话时，可以选择携带初始的消息内容。
* 发起对话：
   * 用于在已经存在的会话中发起一次对话。
   * 支持添加上下文和流式响应。
   * 可以基于历史消息进行上下文关联，提供更符合语境的回复。

# 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br>  https://api.coze.cn/v3/chat <br> ``` <br>  |
| **权限** | `chat` <br> 确保调用该接口使用的个人令牌开通了 `chat` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/api/open/docs/docs/developer_guides/authentication)。 |
| **接口说明** | 调用此接口发起一次对话，支持添加上下文和流式响应。 |
# Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer <span style="color: #D83931"><em>$Access_Token</em></span> | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/api/open/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |


# Query
| **参数** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| conversation_id | String | 可选 | 标识对话发生在哪一次会话中。 <br> 会话是智能体和用户之间的一段问答交互。一个会话包含一条或多条消息。对话是会话中对智能体的一次调用，智能体会将对话中产生的消息添加到会话中。 <br>  <br> * 可以使用已创建的会话，会话中已存在的消息将作为上下文传递给模型。创建会话的方式可参考[创建会话](https://docs.coze.cn/api/open/docs/developer_guides/create_conversation)。 <br> * 对于一问一答等不需要区分 conversation 的场合可不传该参数，系统会自动生成一个会话。  <br>  <br> 一个会话中，只能有一个进行中的对话，否则调用此接口时会报错 4016。 <br>  |
# Body
| **参数** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| bot_id | String | 必选 | 要进行会话聊天的智能体ID。 <br> 进入智能体的 开发页面，开发页面 URL 中 `bot` 参数后的数字就是智能体ID。例如`https://www.coze.cn/space/341****/bot/73428668*****`，智能体ID 为`73428668*****`。 |
| user_id | String | 必选 | 标识当前与智能体对话的用户，由使用方自行定义、生成与维护。user_id 用于标识对话中的不同用户，不同的 user_id，其对话的上下文消息、数据库等对话记忆数据互相隔离。如果不需要用户数据隔离，可将此参数固定为一个任意字符串，例如 `123`，`abc` 等。 <br> 出于数据隐私及信息安全等方面的考虑，不建议使用业务系统中定义的用户 ID。 <br>  |
| additional_messages <br>  <br>  | Array of object <br>  | 可选 <br>  | 对话的附加信息。你可以通过此字段传入历史消息和本次对话中用户的问题。数组长度限制为 100，即最多传入 100 条消息。 <br>  <br> * 若未设置 additional_messages，智能体收到的消息只有会话中已有的消息内容，其中最后一条作为本次对话的用户输入，其他内容均为本次对话的上下文。 <br> * 若设置了 additional_messages，智能体收到的消息包括会话中已有的消息和 additional_messages 中添加的消息，其中 additional_messages 最后一条消息会作为本次对话的用户输入，其他内容均为本次对话的上下文。 <br>  <br> 消息结构可参考[Message Object](https://docs.coze.cn/api/open/docs/developer_guides/chat_v3#67ff5aad)，具体示例可参考[携带上下文](https://docs.coze.cn/api/open/docs/developer_guides/chat_v3#ccb24bc0)。 <br> * 会话或 additional_messages 中最后一条消息应为 role=user 的记录，以免影响模型效果。 <br> * 如果本次对话未指定会话或指定的会话中无消息时，必须通过此参数传入智能体用户的问题。 <br>  |
| stream <br>  | Boolean <br>  | 可选 <br>  | 是否启用流式返回。 <br>  <br> * **true**：采用流式响应。 “流式响应”将模型的实时响应提供给客户端，类似打字机效果。你可以实时获取服务端返回的对话、消息事件，并在客户端中同步处理、实时展示，也可以直接在 completed 事件中获取智能体最终的回复。 <br> * **false**：（默认）采用非流式响应。 “非流式响应”是指响应中仅包含本次对话的状态等元数据。此时应同时开启 auto_save_history，在本次对话处理结束后再查看模型回复等完整响应内容。可以参考以下业务流程： <br>    1. 调用`发起对话`接口，并设置 stream = false，auto_save_history=true，表示使用非流式响应，并记录历史消息。 <br>       你需要记录会话的 Conversation ID 和 Chat ID，用于后续查看详细信息。 <br>    2. 定期轮询[查看对话详情](https://docs.coze.cn/api/open/docs/developer_guides/retrieve_chat)接口，建议每次间隔 1 秒以上，直到会话状态流转为终态，即 status 为 completed、required_action、canceled 或 failed。 <br>    3. 调用[查看对话消息详情](https://docs.coze.cn/api/open/docs/developer_guides/list_chat_messages)接口，查询大模型生成的最终结果。 |
| custom_variables | Map<String, String> | 可选 | 智能体提示词中定义的变量。在智能体 prompt 中设置变量 {{key}} 后，可以通过该参数传入变量值，同时支持 Jinja2 语法。详细说明可参考[Prompt 变量](https://docs.coze.cn/api/open/docs/developer_guides/chat_v3#4698e92c)。 <br> * 仅适用于智能体提示词中定义的变量，不支持用于智能体的变量，或者传入到工作流中。 <br> * 变量名只支持英文字母和下划线。 <br>  |
| auto_save_history <br>  | Boolean <br>  | 可选 | 是否保存本次对话记录。 <br>  <br> * true：（默认）会话中保存本次对话记录，包括 additional_messages 中指定的所有消息、本次对话的模型回复结果、模型执行中间结果。 <br> * false：会话中不保存本次对话记录，后续也无法通过任何方式查看本次对话信息、消息详情。在同一个会话中再次发起对话时，本次会话也不会作为上下文传递给模型。 <br>  <br> * 非流式响应下（stream=false），此参数必须设置为 true，即保存本次对话记录，否则无法查看对话状态和模型回复。 <br> * 调用端插件时，此参数必须设置为 true，即保存本次对话记录，否则调用[提交工具执行结果](https://docs.coze.cn/api/open/docs/developer_guides/chat_submit_tool_outputs)时会提示 5000 错误，端插件的详细 API 使用示例请参见[通过 API 使用端插件](https://docs.coze.cn/api/open/docs/guides/use_local_plugin)。 <br>  |
| meta_data | Map | 可选 | 附加信息，通常用于封装一些业务相关的字段。[查看对话详情](https://www.coze.cn/open/docs/developer_guides/retrieve_chat)时，扣子编程会透传此附加信息，[查看消息列表](https://www.coze.cn/open/docs/developer_guides/list_message)时不会返回该附加信息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 |
| extra_params | Map<String, String> | 可选 | 附加参数，通常用于特殊场景下指定一些必要参数供模型判断，例如指定经纬度，并询问智能体此位置的天气。该参数不会传给工作流。 <br> 自定义键值对格式，其中键（key）仅支持设置为： <br>  <br> * latitude：纬度，此时值（Value）为纬度值，例如 39.9800718。 <br> * longitude：经度，此时值（Value）为经度值，例如 116.309314。 |
| shortcut_command | Object | 可选 | 快捷指令信息。你可以通过此参数指定此次对话执行的快捷指令，必须是智能体已绑定的快捷指令。 <br> 消息结构可参考 **ShortcutCommandDetail Object**。 <br> 调用快捷指令，会自动根据快捷指令配置信息生成本次对话中的用户问题，并放入 additional_messages 最后一条消息作为本次对话的用户输入。 <br>  |
| parameters | Map[String, Any] | 可选 | 给自定义参数赋值并传给对话流。你可以根据实际业务需求，在对话流开始节点的输入参数中设置自定义参数，调用本接口发起对话时，可以通过`parameters` 参数传入自定义参数的值并传给对话流。具体示例代码请参见[给自定义参数赋值](https://docs.coze.cn/api/open/docs/developer_guides/chat_v3#022e2f67)。 <br> 仅支持为已发布 API、ChatSDK 的单 Agent（对话流模式）的智能体设置该参数。 <br>  |
| enable_card <br>  <br>  | Boolean <br>  | 可选 <br>  | 设置问答节点返回的内容是否为卡片形式。默认为 `false`。 <br>  <br> * `true`：问答节点返回卡片形式的内容。 <br>    * API 渠道暂时不支持直接渲染卡片交互形式。仅在 Chat SDK 中支持呈现智能体的卡片交互。 <br>    * 如果需要实现卡片内容展示，你可以在该 API 响应中获取卡片数据，在 Card SDK 的 `runtimeOptions.dsl` 中引用 API 返回的卡片 data 字段，通过 Card SDK 进行解析并将卡片数据转换为可视化界面，具体请参见[安装并使用 Card SDK](https://docs.coze.cn/api/open/docs/developer_guides/card_sdk)。 <br>  <br> * `false`：问答节点返回普通文本形式的内容。 |
| publish_status | String | 可选 | 智能体的发布状态，用于指定与已发布版本或草稿版本的智能体对话。默认值为 `published_online`。枚举值： <br>  <br> * `published_online`：与已发布的线上版本的智能体对话。 <br> * `unpublished_draft`：与草稿版本的智能体对话。 |
| bot_version | String | 可选 | 指定智能体的版本号，用于与历史版本的智能体进行对话。默认与最新版本的智能体对话。 <br> 你可以通过[查看智能体版本列表](https://www.coze.cn/open/docs/developer_guides/list_bot_versions) API 查看智能体的版本号。 <br> 当 `publish_status` 设置为 `unpublished_draft`时，填写此参数会提示 4000 错误。 |
## EnterMessage Object
| **参数** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| role | String | 必选 | 发送这条消息的实体。取值： <br>  <br> * **user**：代表该条消息内容是用户发送的。 <br> * **assistant**：代表该条消息内容是智能体发送的。 |
| type <br>  | String | 可选 <br>  | 消息类型。默认为 **question。** <br>  <br> * **question**：用户输入内容。 <br> * **answer**：智能体返回给用户的消息内容，支持增量返回。如果工作流绑定了输出节点，可能会存在多 answer 场景，此时可以用流式返回的结束标志来判断所有 answer 完成。 <br> * **function_call**：智能体对话过程中调用函数（function call）的中间结果。  <br> * **tool_response**：调用工具 （function call）后返回的结果。 <br> * **follow_up**：如果在 智能体上配置打开了用户问题建议开关，则会返回推荐问题相关的回复内容。不支持在请求中作为入参。 <br> * **verbose**：多 answer 场景下，服务端会返回一个 verbose 包，对应的 content 为 JSON 格式，`content.msg_type =generate_answer_finish` 代表全部 answer 回复完成。不支持在请求中作为入参。 <br>  <br> 仅发起会话（v3）接口支持将此参数作为入参，且： <br>  <br> * 如果 autoSaveHistory=true，type 支持设置为 question 或 answer。 <br> * 如果 autoSaveHistory=false，type 支持设置为 question、answer、function_call、tool_output/tool_response。 <br>  <br> 其中，type=question 只能和 role=user 对应，即仅用户角色可以且只能发起 question 类型的消息。详细说明可参考[消息 type 说明](https://docs.coze.cn/api/open/docs/developer_guides/message_type)。 <br>  |
| content | String | 必选 | 消息的内容，支持纯文本、多模态（文本、图片、文件混合输入）、卡片等多种类型的内容。 <br>  <br> * content_type 为 object_string 时，content 为 **object_string object** 数组序列化之后的 JSON String，详细说明可参考 **object_string object。** <br> * 当 content_type **=** text **** 时，content 为普通文本，例如 `"content" :"Hello!"`。 |
| content_type | String | 必选 | 消息内容的类型，支持设置为： <br>  <br> * text：文本。 <br> * object_string：多模态内容，即文本和文件的组合、文本和图片的组合。 <br> * card：卡片。此枚举值仅在接口响应中出现，不支持作为入参。 <br>  <br> content 不为空时，此参数为必选。 <br>  |
|  meta_data | Map  | 可选 | 创建消息时的附加消息，[查看消息列表](https://www.coze.cn/open/docs/developer_guides/list_message)时也会返回此附加消息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 |
#### object_string object
| **参数** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| type | String | 必选 | 多模态消息内容类型，支持设置为： <br>  <br> * text：文本类型。 <br> * file：文件类型。 <br> * image：图片类型。 <br> * audio：音频类型。 |
| text | String | 可选 | 文本内容。 |
| file_id | String | 可选 | 文件、图片、音频内容的 ID。 <br> * 必须是当前账号上传的文件 ID，上传方式可参考[上传文件](https://docs.coze.cn/api/open/docs/developer_guides/upload_files)。 <br> * 在 type 为 file、image 或 audio 时，file_id 和 file_url 应至少指定一个。 <br>  |
| file_url | String | 可选 | 文件、图片或语音文件的在线地址。必须是可公共访问的有效地址。 <br> 在 type 为 file、image 或 audio 时，file_id 和 file_url 应至少指定一个。 |
* 一个 object_string 数组中最多包含一条 `text` 类型消息，但可以包含多个 `file`、`image` 类型的消息。
* 当 object_string 数组中存在 `text` 类型消息时，必须同时存在至少 1 条 `file` 或 `image` 消息，纯文本消息（仅包含 `text` 类型）需要使用 `content_type: text` 字段直接指定，不能使用 `object_string` 数组。
* 支持发送纯图片或纯文件消息，但每条纯图片或纯文件消息的前一条或后一条消息中，必须包含一条 `content_type: text` 的纯文本消息，作为用户查询的上下文。例如， `"content": "[{\"type\":\"image\",\"file_id\":\"112233***\"}]"` 为一条纯图片消息，该纯图片消息的前一条或后一条消息必须是一条纯文本消息，否则接口会报 4000 参数错误。

例如，以下数组是一个完整的多模态内容：

<div style="display: flex;">
<div style="flex-shrink: 0;width: calc((100% - 16px) * 0.5000);">

序列化前：
```JSON
[
    {
        "type": "text",
        "text": "你好我有一个帽衫，我想问问它好看么，你帮我看看"
    }, {
        "type": "image",
        "file_id": "112233***"
    }, {
        "type": "file",
        "file_id": "144423***"
    },
        {
        "type": "file",
        "file_url": "https://example.com/tos-cn-i-mdko3gqilj/example.png"
    }
]
```



</div>
<div style="flex-shrink: 0;width: calc((100% - 16px) * 0.5000);margin-left: 16px;">

序列化后：
```JSON
"[{\"type\":\"text\",\"text\":\"你好我有一个帽衫，我想问问它好看么，你帮我看看\"},{\"type\":\"image\",\"file_id\":\"112233***\"},{\"type\":\"file\",\"file_id\":\"144423***\"},{\"type\":\"file\",\"file_url\":\"https://example.com/tos-cn-i-mdko3gqilj/example.png\"}]"
```




</div>
</div>

消息结构示例：

<div type="doc-tabs">
<div type="tab-item" title="文本消息" key="Agz7LSaSNn">

文本消息的 content_type 为 text，消息结构示例如下。
```JSON
{
    "role": "user",
    "content": "搜几个最新的军事新闻",
    "content_type": "text"
}
```


</div>
<div type="tab-item" title="多模态消息" key="wAOVPS08BR">

多模态消息的 content_type 为 object_string，消息结构示例如下。
```JSON
{
    "role": "user",
    "content": "[{\"type\":\"text\",\"text\":\"你好我有一个帽衫，我想问问它好看么，你帮我看看\"},{\"type\":\"image\",\"file_id\":\"112233***\"},{\"type\":\"file\",\"file_id\":\"144423***\"}]",
    "content_type": "object_string"
}
```


</div>
</div>
#### ShortcutCommandDetail Object
| **参数** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| command_id | String | 必选 | 指定对话要执行的快捷指令 ID，必须是智能体已绑定的快捷指令。 <br> 若对话无需执行快捷指令时，无需设置此参数。 <br> 你可以通过[查看智能体配置](https://docs.coze.cn/api/open/docs/developer_guides/get_metadata_draft_published) API 中的[ShortcutCommandInfo](https://docs.coze.cn/api/open/docs/developer_guides/get_metadata_draft_published#shortcutcommandinfo)查看快捷指令 ID。 |
| parameters | Map<String, String> | 可选 | 用户输入的快捷指令组件参数信息。 <br> 自定义键值对，其中键（key）为快捷指令组件的名称，值（value）为组件对应的用户输入，为 **object_string object** 数组序列化之后的 JSON String，详细说明可参考 **object_string object。** |
# 返回结果
此接口通过请求 Body 参数 stream 为 true 或 false 来指定 Response 为流式或非流式响应。你可以根据以下步骤判断当前业务场景适合的响应模式。

<img src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB2ZXJzaW9uPSIxLjEiIHdpZHRoPSIzNzZweCIgaGVpZ2h0PSI1MDZweCIgdmlld0JveD0iLTAuNSAtMC41IDM3NiA1MDYiPjxkZWZzLz48Zz48cGF0aCBkPSJNIDgzIDgzIEwgODMgMTM2LjYzIiBmaWxsPSJub25lIiBzdHJva2U9IiMwMDAwMDAiIHN0cm9rZS1taXRlcmxpbWl0PSIxMCIgcG9pbnRlci1ldmVudHM9InN0cm9rZSIvPjxwYXRoIGQ9Ik0gODMgMTQxLjg4IEwgNzkuNSAxMzQuODggTCA4MyAxMzYuNjMgTCA4Ni41IDEzNC44OCBaIiBmaWxsPSIjMDAwMDAwIiBzdHJva2U9IiMwMDAwMDAiIHN0cm9rZS1taXRlcmxpbWl0PSIxMCIgcG9pbnRlci1ldmVudHM9ImFsbCIvPjxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKC0wLjUgLTAuNSkiPjxmb3JlaWduT2JqZWN0IHN0eWxlPSJvdmVyZmxvdzogdmlzaWJsZTsgdGV4dC1hbGlnbjogbGVmdDsiIHBvaW50ZXItZXZlbnRzPSJub25lIiB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIj48ZGl2IHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hodG1sIiBzdHlsZT0iZGlzcGxheTogZmxleDsgYWxpZ24taXRlbXM6IHVuc2FmZSBjZW50ZXI7IGp1c3RpZnktY29udGVudDogdW5zYWZlIGNlbnRlcjsgd2lkdGg6IDFweDsgaGVpZ2h0OiAxcHg7IHBhZGRpbmctdG9wOiAxMDdweDsgbWFyZ2luLWxlZnQ6IDc0cHg7Ij48ZGl2IHN0eWxlPSJib3gtc2l6aW5nOiBib3JkZXItYm94OyBmb250LXNpemU6IDA7IHRleHQtYWxpZ246IGNlbnRlcjsgIj48ZGl2IHN0eWxlPSJkaXNwbGF5OiBpbmxpbmUtYmxvY2s7IGZvbnQtc2l6ZTogMTJweDsgZm9udC1mYW1pbHk6IEhlbHZldGljYTsgY29sb3I6ICMwMDAwMDA7IGxpbmUtaGVpZ2h0OiAxLjI7IHBvaW50ZXItZXZlbnRzOiBhbGw7IHdoaXRlLXNwYWNlOiBub3dyYXA7ICI+Tm88L2Rpdj48L2Rpdj48L2Rpdj48L2ZvcmVpZ25PYmplY3Q+PC9nPjxwYXRoIGQ9Ik0gMTYzIDQzIEwgMzMzIDQzIEwgMzMzIDEzNi42MyIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjMDAwMDAwIiBzdHJva2UtbWl0ZXJsaW1pdD0iMTAiIHBvaW50ZXItZXZlbnRzPSJzdHJva2UiLz48cGF0aCBkPSJNIDMzMyAxNDEuODggTCAzMjkuNSAxMzQuODggTCAzMzMgMTM2LjYzIEwgMzM2LjUgMTM0Ljg4IFoiIGZpbGw9IiMwMDAwMDAiIHN0cm9rZT0iIzAwMDAwMCIgc3Ryb2tlLW1pdGVybGltaXQ9IjEwIiBwb2ludGVyLWV2ZW50cz0iYWxsIi8+PHBhdGggZD0iTSA4MyAzIEwgMTYzIDQzIEwgODMgODMgTCAzIDQzIFoiIGZpbGw9IiNmZmZmZmYiIHN0cm9rZT0iIzAwMDAwMCIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbWl0ZXJsaW1pdD0iMTAiIHBvaW50ZXItZXZlbnRzPSJhbGwiLz48ZyB0cmFuc2Zvcm09InRyYW5zbGF0ZSgtMC41IC0wLjUpIj48Zm9yZWlnbk9iamVjdCBzdHlsZT0ib3ZlcmZsb3c6IHZpc2libGU7IHRleHQtYWxpZ246IGxlZnQ7IiBwb2ludGVyLWV2ZW50cz0ibm9uZSIgd2lkdGg9IjEwMCUiIGhlaWdodD0iMTAwJSI+PGRpdiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94aHRtbCIgc3R5bGU9ImRpc3BsYXk6IGZsZXg7IGFsaWduLWl0ZW1zOiB1bnNhZmUgY2VudGVyOyBqdXN0aWZ5LWNvbnRlbnQ6IHVuc2FmZSBjZW50ZXI7IHdpZHRoOiAxNThweDsgaGVpZ2h0OiAxcHg7IHBhZGRpbmctdG9wOiA0M3B4OyBtYXJnaW4tbGVmdDogNHB4OyI+PGRpdiBzdHlsZT0iYm94LXNpemluZzogYm9yZGVyLWJveDsgZm9udC1zaXplOiAwOyB0ZXh0LWFsaWduOiBjZW50ZXI7ICI+PGRpdiBzdHlsZT0iZGlzcGxheTogaW5saW5lLWJsb2NrOyBmb250LXNpemU6IDEycHg7IGZvbnQtZmFtaWx5OiBIZWx2ZXRpY2E7IGNvbG9yOiAjMDAwMDAwOyBsaW5lLWhlaWdodDogMS4yOyBwb2ludGVyLWV2ZW50czogYWxsOyB3aGl0ZS1zcGFjZTogbm9ybWFsOyB3b3JkLXdyYXA6IG5vcm1hbDsgIj7mmK/lkKbpnIDopoHmiZPlrZfmnLrmlYjmnpw8YnIgLz7mmL7npLpSZXNwb25zZTwvZGl2PjwvZGl2PjwvZGl2PjwvZm9yZWlnbk9iamVjdD48L2c+PHBhdGggZD0iTSA4MyAzNjMgTCA4MyA0MTYuNjMiIGZpbGw9Im5vbmUiIHN0cm9rZT0iIzAwMDAwMCIgc3Ryb2tlLW1pdGVybGltaXQ9IjEwIiBwb2ludGVyLWV2ZW50cz0ic3Ryb2tlIi8+PHBhdGggZD0iTSA4MyA0MjEuODggTCA3OS41IDQxNC44OCBMIDgzIDQxNi42MyBMIDg2LjUgNDE0Ljg4IFoiIGZpbGw9IiMwMDAwMDAiIHN0cm9rZT0iIzAwMDAwMCIgc3Ryb2tlLW1pdGVybGltaXQ9IjEwIiBwb2ludGVyLWV2ZW50cz0iYWxsIi8+PHBhdGggZD0iTSAxNjMgMzIzIEwgMzMzIDMyMyBMIDMzMyAyMjkuMzciIGZpbGw9Im5vbmUiIHN0cm9rZT0iIzAwMDAwMCIgc3Ryb2tlLW1pdGVybGltaXQ9IjEwIiBwb2ludGVyLWV2ZW50cz0ic3Ryb2tlIi8+PHBhdGggZD0iTSAzMzMgMjI0LjEyIEwgMzM2LjUgMjMxLjEyIEwgMzMzIDIyOS4zNyBMIDMyOS41IDIzMS4xMiBaIiBmaWxsPSIjMDAwMDAwIiBzdHJva2U9IiMwMDAwMDAiIHN0cm9rZS1taXRlcmxpbWl0PSIxMCIgcG9pbnRlci1ldmVudHM9ImFsbCIvPjxwYXRoIGQ9Ik0gODMgMjgzIEwgMTYzIDMyMyBMIDgzIDM2MyBMIDMgMzIzIFoiIGZpbGw9IiNmZmZmZmYiIHN0cm9rZT0iIzAwMDAwMCIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbWl0ZXJsaW1pdD0iMTAiIHBvaW50ZXItZXZlbnRzPSJhbGwiLz48ZyB0cmFuc2Zvcm09InRyYW5zbGF0ZSgtMC41IC0wLjUpIj48Zm9yZWlnbk9iamVjdCBzdHlsZT0ib3ZlcmZsb3c6IHZpc2libGU7IHRleHQtYWxpZ246IGxlZnQ7IiBwb2ludGVyLWV2ZW50cz0ibm9uZSIgd2lkdGg9IjEwMCUiIGhlaWdodD0iMTAwJSI+PGRpdiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94aHRtbCIgc3R5bGU9ImRpc3BsYXk6IGZsZXg7IGFsaWduLWl0ZW1zOiB1bnNhZmUgY2VudGVyOyBqdXN0aWZ5LWNvbnRlbnQ6IHVuc2FmZSBjZW50ZXI7IHdpZHRoOiAxNThweDsgaGVpZ2h0OiAxcHg7IHBhZGRpbmctdG9wOiAzMjNweDsgbWFyZ2luLWxlZnQ6IDRweDsiPjxkaXYgc3R5bGU9ImJveC1zaXppbmc6IGJvcmRlci1ib3g7IGZvbnQtc2l6ZTogMDsgdGV4dC1hbGlnbjogY2VudGVyOyAiPjxkaXYgc3R5bGU9ImRpc3BsYXk6IGlubGluZS1ibG9jazsgZm9udC1zaXplOiAxMnB4OyBmb250LWZhbWlseTogSGVsdmV0aWNhOyBjb2xvcjogIzAwMDAwMDsgbGluZS1oZWlnaHQ6IDEuMjsgcG9pbnRlci1ldmVudHM6IGFsbDsgd2hpdGUtc3BhY2U6IG5vcm1hbDsgd29yZC13cmFwOiBub3JtYWw7ICI+5piv5ZCm6ZyA6KaBPGJyIC8+5ZCM5q2l5aSE55CGIFJlc3BvbnNlPC9kaXY+PC9kaXY+PC9kaXY+PC9mb3JlaWduT2JqZWN0PjwvZz48cmVjdCB4PSIyOTMiIHk9IjE0MyIgd2lkdGg9IjgwIiBoZWlnaHQ9IjgwIiByeD0iNC44IiByeT0iNC44IiBmaWxsPSIjZmZmZmZmIiBzdHJva2U9IiMwMDAwMDAiIHN0cm9rZS13aWR0aD0iMiIgcG9pbnRlci1ldmVudHM9ImFsbCIvPjxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKC0wLjUgLTAuNSkiPjxmb3JlaWduT2JqZWN0IHN0eWxlPSJvdmVyZmxvdzogdmlzaWJsZTsgdGV4dC1hbGlnbjogbGVmdDsiIHBvaW50ZXItZXZlbnRzPSJub25lIiB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIj48ZGl2IHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hodG1sIiBzdHlsZT0iZGlzcGxheTogZmxleDsgYWxpZ24taXRlbXM6IHVuc2FmZSBjZW50ZXI7IGp1c3RpZnktY29udGVudDogdW5zYWZlIGNlbnRlcjsgd2lkdGg6IDc4cHg7IGhlaWdodDogMXB4OyBwYWRkaW5nLXRvcDogMTgzcHg7IG1hcmdpbi1sZWZ0OiAyOTRweDsiPjxkaXYgc3R5bGU9ImJveC1zaXppbmc6IGJvcmRlci1ib3g7IGZvbnQtc2l6ZTogMDsgdGV4dC1hbGlnbjogY2VudGVyOyAiPjxkaXYgc3R5bGU9ImRpc3BsYXk6IGlubGluZS1ibG9jazsgZm9udC1zaXplOiAxMnB4OyBmb250LWZhbWlseTogSGVsdmV0aWNhOyBjb2xvcjogIzAwMDAwMDsgbGluZS1oZWlnaHQ6IDEuMjsgcG9pbnRlci1ldmVudHM6IGFsbDsgd2hpdGUtc3BhY2U6IG5vcm1hbDsgd29yZC13cmFwOiBub3JtYWw7ICI+5rWB5byP5ZON5bqUPGJyIC8+c3RyZWFtPXRydWU8L2Rpdj48L2Rpdj48L2Rpdj48L2ZvcmVpZ25PYmplY3Q+PC9nPjxyZWN0IHg9IjQzIiB5PSI0MjMiIHdpZHRoPSI4MCIgaGVpZ2h0PSI4MCIgcng9IjQuOCIgcnk9IjQuOCIgZmlsbD0iI2ZmZmZmZiIgc3Ryb2tlPSIjMDAwMDAwIiBzdHJva2Utd2lkdGg9IjIiIHBvaW50ZXItZXZlbnRzPSJhbGwiLz48ZyB0cmFuc2Zvcm09InRyYW5zbGF0ZSgtMC41IC0wLjUpIj48Zm9yZWlnbk9iamVjdCBzdHlsZT0ib3ZlcmZsb3c6IHZpc2libGU7IHRleHQtYWxpZ246IGxlZnQ7IiBwb2ludGVyLWV2ZW50cz0ibm9uZSIgd2lkdGg9IjEwMCUiIGhlaWdodD0iMTAwJSI+PGRpdiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94aHRtbCIgc3R5bGU9ImRpc3BsYXk6IGZsZXg7IGFsaWduLWl0ZW1zOiB1bnNhZmUgY2VudGVyOyBqdXN0aWZ5LWNvbnRlbnQ6IHVuc2FmZSBjZW50ZXI7IHdpZHRoOiA3OHB4OyBoZWlnaHQ6IDFweDsgcGFkZGluZy10b3A6IDQ2M3B4OyBtYXJnaW4tbGVmdDogNDRweDsiPjxkaXYgc3R5bGU9ImJveC1zaXppbmc6IGJvcmRlci1ib3g7IGZvbnQtc2l6ZTogMDsgdGV4dC1hbGlnbjogY2VudGVyOyAiPjxkaXYgc3R5bGU9ImRpc3BsYXk6IGlubGluZS1ibG9jazsgZm9udC1zaXplOiAxMnB4OyBmb250LWZhbWlseTogSGVsdmV0aWNhOyBjb2xvcjogIzAwMDAwMDsgbGluZS1oZWlnaHQ6IDEuMjsgcG9pbnRlci1ldmVudHM6IGFsbDsgd2hpdGUtc3BhY2U6IG5vcm1hbDsgd29yZC13cmFwOiBub3JtYWw7ICI+6Z2e5rWB5byP5ZON5bqUPGJyIC8+c3RyZWFtPWZhbHNlPC9kaXY+PC9kaXY+PC9kaXY+PC9mb3JlaWduT2JqZWN0PjwvZz48cGF0aCBkPSJNIDgzIDIyMyBMIDgzIDI3Ni42MyIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjMDAwMDAwIiBzdHJva2UtbWl0ZXJsaW1pdD0iMTAiIHBvaW50ZXItZXZlbnRzPSJzdHJva2UiLz48cGF0aCBkPSJNIDgzIDI4MS44OCBMIDc5LjUgMjc0Ljg4IEwgODMgMjc2LjYzIEwgODYuNSAyNzQuODggWiIgZmlsbD0iIzAwMDAwMCIgc3Ryb2tlPSIjMDAwMDAwIiBzdHJva2UtbWl0ZXJsaW1pdD0iMTAiIHBvaW50ZXItZXZlbnRzPSJhbGwiLz48cGF0aCBkPSJNIDE2MyAxODMgTCAyODYuNjMgMTgzIiBmaWxsPSJub25lIiBzdHJva2U9IiMwMDAwMDAiIHN0cm9rZS1taXRlcmxpbWl0PSIxMCIgcG9pbnRlci1ldmVudHM9InN0cm9rZSIvPjxwYXRoIGQ9Ik0gMjkxLjg4IDE4MyBMIDI4NC44OCAxODYuNSBMIDI4Ni42MyAxODMgTCAyODQuODggMTc5LjUgWiIgZmlsbD0iIzAwMDAwMCIgc3Ryb2tlPSIjMDAwMDAwIiBzdHJva2UtbWl0ZXJsaW1pdD0iMTAiIHBvaW50ZXItZXZlbnRzPSJhbGwiLz48cGF0aCBkPSJNIDgzIDE0MyBMIDE2MyAxODMgTCA4MyAyMjMgTCAzIDE4MyBaIiBmaWxsPSIjZmZmZmZmIiBzdHJva2U9IiMwMDAwMDAiIHN0cm9rZS13aWR0aD0iMiIgc3Ryb2tlLW1pdGVybGltaXQ9IjEwIiBwb2ludGVyLWV2ZW50cz0iYWxsIi8+PGcgdHJhbnNmb3JtPSJ0cmFuc2xhdGUoLTAuNSAtMC41KSI+PGZvcmVpZ25PYmplY3Qgc3R5bGU9Im92ZXJmbG93OiB2aXNpYmxlOyB0ZXh0LWFsaWduOiBsZWZ0OyIgcG9pbnRlci1ldmVudHM9Im5vbmUiIHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiPjxkaXYgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGh0bWwiIHN0eWxlPSJkaXNwbGF5OiBmbGV4OyBhbGlnbi1pdGVtczogdW5zYWZlIGNlbnRlcjsganVzdGlmeS1jb250ZW50OiB1bnNhZmUgY2VudGVyOyB3aWR0aDogMTU4cHg7IGhlaWdodDogMXB4OyBwYWRkaW5nLXRvcDogMTgzcHg7IG1hcmdpbi1sZWZ0OiA0cHg7Ij48ZGl2IHN0eWxlPSJib3gtc2l6aW5nOiBib3JkZXItYm94OyBmb250LXNpemU6IDA7IHRleHQtYWxpZ246IGNlbnRlcjsgIj48ZGl2IHN0eWxlPSJkaXNwbGF5OiBpbmxpbmUtYmxvY2s7IGZvbnQtc2l6ZTogMTJweDsgZm9udC1mYW1pbHk6IEhlbHZldGljYTsgY29sb3I6ICMwMDAwMDA7IGxpbmUtaGVpZ2h0OiAxLjI7IHBvaW50ZXItZXZlbnRzOiBhbGw7IHdoaXRlLXNwYWNlOiBub3JtYWw7IHdvcmQtd3JhcDogbm9ybWFsOyAiPuaYr+WQpumcgOimgTxiciAvPuWNs+aXtuafpeeciyBCb3Qg5Zue5aSNPC9kaXY+PC9kaXY+PC9kaXY+PC9mb3JlaWduT2JqZWN0PjwvZz48cmVjdCB4PSIyMTYiIHk9IjI4IiB3aWR0aD0iNDAiIGhlaWdodD0iMjAiIGZpbGw9Im5vbmUiIHN0cm9rZT0ibm9uZSIgcG9pbnRlci1ldmVudHM9ImFsbCIvPjxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKC0wLjUgLTAuNSkiPjxmb3JlaWduT2JqZWN0IHN0eWxlPSJvdmVyZmxvdzogdmlzaWJsZTsgdGV4dC1hbGlnbjogbGVmdDsiIHBvaW50ZXItZXZlbnRzPSJub25lIiB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIj48ZGl2IHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hodG1sIiBzdHlsZT0iZGlzcGxheTogZmxleDsgYWxpZ24taXRlbXM6IHVuc2FmZSBjZW50ZXI7IGp1c3RpZnktY29udGVudDogdW5zYWZlIGNlbnRlcjsgd2lkdGg6IDFweDsgaGVpZ2h0OiAxcHg7IHBhZGRpbmctdG9wOiAzOHB4OyBtYXJnaW4tbGVmdDogMjM2cHg7Ij48ZGl2IHN0eWxlPSJib3gtc2l6aW5nOiBib3JkZXItYm94OyBmb250LXNpemU6IDA7IHRleHQtYWxpZ246IGNlbnRlcjsgIj48ZGl2IHN0eWxlPSJkaXNwbGF5OiBpbmxpbmUtYmxvY2s7IGZvbnQtc2l6ZTogMTJweDsgZm9udC1mYW1pbHk6IEhlbHZldGljYTsgY29sb3I6ICMwMDAwMDA7IGxpbmUtaGVpZ2h0OiAxLjI7IHBvaW50ZXItZXZlbnRzOiBhbGw7IHdoaXRlLXNwYWNlOiBub3dyYXA7ICI+WUVTPC9kaXY+PC9kaXY+PC9kaXY+PC9mb3JlaWduT2JqZWN0PjwvZz48cmVjdCB4PSIyMTYiIHk9IjE2MyIgd2lkdGg9IjQwIiBoZWlnaHQ9IjIwIiBmaWxsPSJub25lIiBzdHJva2U9Im5vbmUiIHBvaW50ZXItZXZlbnRzPSJhbGwiLz48ZyB0cmFuc2Zvcm09InRyYW5zbGF0ZSgtMC41IC0wLjUpIj48Zm9yZWlnbk9iamVjdCBzdHlsZT0ib3ZlcmZsb3c6IHZpc2libGU7IHRleHQtYWxpZ246IGxlZnQ7IiBwb2ludGVyLWV2ZW50cz0ibm9uZSIgd2lkdGg9IjEwMCUiIGhlaWdodD0iMTAwJSI+PGRpdiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94aHRtbCIgc3R5bGU9ImRpc3BsYXk6IGZsZXg7IGFsaWduLWl0ZW1zOiB1bnNhZmUgY2VudGVyOyBqdXN0aWZ5LWNvbnRlbnQ6IHVuc2FmZSBjZW50ZXI7IHdpZHRoOiAxcHg7IGhlaWdodDogMXB4OyBwYWRkaW5nLXRvcDogMTczcHg7IG1hcmdpbi1sZWZ0OiAyMzZweDsiPjxkaXYgc3R5bGU9ImJveC1zaXppbmc6IGJvcmRlci1ib3g7IGZvbnQtc2l6ZTogMDsgdGV4dC1hbGlnbjogY2VudGVyOyAiPjxkaXYgc3R5bGU9ImRpc3BsYXk6IGlubGluZS1ibG9jazsgZm9udC1zaXplOiAxMnB4OyBmb250LWZhbWlseTogSGVsdmV0aWNhOyBjb2xvcjogIzAwMDAwMDsgbGluZS1oZWlnaHQ6IDEuMjsgcG9pbnRlci1ldmVudHM6IGFsbDsgd2hpdGUtc3BhY2U6IG5vd3JhcDsgIj5ZRVM8L2Rpdj48L2Rpdj48L2Rpdj48L2ZvcmVpZ25PYmplY3Q+PC9nPjxyZWN0IHg9IjIyMyIgeT0iMzAzIiB3aWR0aD0iNDAiIGhlaWdodD0iMjAiIGZpbGw9Im5vbmUiIHN0cm9rZT0ibm9uZSIgcG9pbnRlci1ldmVudHM9ImFsbCIvPjxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKC0wLjUgLTAuNSkiPjxmb3JlaWduT2JqZWN0IHN0eWxlPSJvdmVyZmxvdzogdmlzaWJsZTsgdGV4dC1hbGlnbjogbGVmdDsiIHBvaW50ZXItZXZlbnRzPSJub25lIiB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIj48ZGl2IHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hodG1sIiBzdHlsZT0iZGlzcGxheTogZmxleDsgYWxpZ24taXRlbXM6IHVuc2FmZSBjZW50ZXI7IGp1c3RpZnktY29udGVudDogdW5zYWZlIGNlbnRlcjsgd2lkdGg6IDFweDsgaGVpZ2h0OiAxcHg7IHBhZGRpbmctdG9wOiAzMTNweDsgbWFyZ2luLWxlZnQ6IDI0M3B4OyI+PGRpdiBzdHlsZT0iYm94LXNpemluZzogYm9yZGVyLWJveDsgZm9udC1zaXplOiAwOyB0ZXh0LWFsaWduOiBjZW50ZXI7ICI+PGRpdiBzdHlsZT0iZGlzcGxheTogaW5saW5lLWJsb2NrOyBmb250LXNpemU6IDEycHg7IGZvbnQtZmFtaWx5OiBIZWx2ZXRpY2E7IGNvbG9yOiAjMDAwMDAwOyBsaW5lLWhlaWdodDogMS4yOyBwb2ludGVyLWV2ZW50czogYWxsOyB3aGl0ZS1zcGFjZTogbm93cmFwOyAiPllFUzwvZGl2PjwvZGl2PjwvZGl2PjwvZm9yZWlnbk9iamVjdD48L2c+PGcgdHJhbnNmb3JtPSJ0cmFuc2xhdGUoLTAuNSAtMC41KSI+PGZvcmVpZ25PYmplY3Qgc3R5bGU9Im92ZXJmbG93OiB2aXNpYmxlOyB0ZXh0LWFsaWduOiBsZWZ0OyIgcG9pbnRlci1ldmVudHM9Im5vbmUiIHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiPjxkaXYgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGh0bWwiIHN0eWxlPSJkaXNwbGF5OiBmbGV4OyBhbGlnbi1pdGVtczogdW5zYWZlIGNlbnRlcjsganVzdGlmeS1jb250ZW50OiB1bnNhZmUgY2VudGVyOyB3aWR0aDogMXB4OyBoZWlnaHQ6IDFweDsgcGFkZGluZy10b3A6IDI1NHB4OyBtYXJnaW4tbGVmdDogNjhweDsiPjxkaXYgc3R5bGU9ImJveC1zaXppbmc6IGJvcmRlci1ib3g7IGZvbnQtc2l6ZTogMDsgdGV4dC1hbGlnbjogY2VudGVyOyAiPjxkaXYgc3R5bGU9ImRpc3BsYXk6IGlubGluZS1ibG9jazsgZm9udC1zaXplOiAxMnB4OyBmb250LWZhbWlseTogSGVsdmV0aWNhOyBjb2xvcjogIzAwMDAwMDsgbGluZS1oZWlnaHQ6IDEuMjsgcG9pbnRlci1ldmVudHM6IGFsbDsgd2hpdGUtc3BhY2U6IG5vd3JhcDsgIj5ObzwvZGl2PjwvZGl2PjwvZGl2PjwvZm9yZWlnbk9iamVjdD48L2c+PGcgdHJhbnNmb3JtPSJ0cmFuc2xhdGUoLTAuNSAtMC41KSI+PGZvcmVpZ25PYmplY3Qgc3R5bGU9Im92ZXJmbG93OiB2aXNpYmxlOyB0ZXh0LWFsaWduOiBsZWZ0OyIgcG9pbnRlci1ldmVudHM9Im5vbmUiIHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiPjxkaXYgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGh0bWwiIHN0eWxlPSJkaXNwbGF5OiBmbGV4OyBhbGlnbi1pdGVtczogdW5zYWZlIGNlbnRlcjsganVzdGlmeS1jb250ZW50OiB1bnNhZmUgY2VudGVyOyB3aWR0aDogMXB4OyBoZWlnaHQ6IDFweDsgcGFkZGluZy10b3A6IDM5MXB4OyBtYXJnaW4tbGVmdDogNjZweDsiPjxkaXYgc3R5bGU9ImJveC1zaXppbmc6IGJvcmRlci1ib3g7IGZvbnQtc2l6ZTogMDsgdGV4dC1hbGlnbjogY2VudGVyOyAiPjxkaXYgc3R5bGU9ImRpc3BsYXk6IGlubGluZS1ibG9jazsgZm9udC1zaXplOiAxMnB4OyBmb250LWZhbWlseTogSGVsdmV0aWNhOyBjb2xvcjogIzAwMDAwMDsgbGluZS1oZWlnaHQ6IDEuMjsgcG9pbnRlci1ldmVudHM6IGFsbDsgd2hpdGUtc3BhY2U6IG5vd3JhcDsgIj5ObzwvZGl2PjwvZGl2PjwvZGl2PjwvZm9yZWlnbk9iamVjdD48L2c+PC9nPjwvc3ZnPg==" from="flow-chart" payload="{&quot;data&quot;:{&quot;mxGraphModel&quot;:{&quot;dx&quot;:&quot;1426&quot;,&quot;dy&quot;:&quot;744&quot;,&quot;grid&quot;:&quot;1&quot;,&quot;gridSize&quot;:&quot;10&quot;,&quot;guides&quot;:&quot;1&quot;,&quot;tooltips&quot;:&quot;1&quot;,&quot;connect&quot;:&quot;1&quot;,&quot;arrows&quot;:&quot;1&quot;,&quot;fold&quot;:&quot;1&quot;,&quot;page&quot;:&quot;1&quot;,&quot;pageScale&quot;:&quot;1&quot;,&quot;pageWidth&quot;:&quot;1169&quot;,&quot;pageHeight&quot;:&quot;827&quot;},&quot;mxCellMap&quot;:{&quot;v6dX7BR7&quot;:{&quot;id&quot;:&quot;v6dX7BR7&quot;},&quot;B9iU9UZj&quot;:{&quot;id&quot;:&quot;B9iU9UZj&quot;,&quot;parent&quot;:&quot;v6dX7BR7&quot;},&quot;YmZCqelD&quot;:{&quot;id&quot;:&quot;YmZCqelD&quot;,&quot;style&quot;:&quot;edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;entryX=0.5;entryY=0;entryDx=0;entryDy=0;entryPerimeter=0;&quot;,&quot;edge&quot;:&quot;1&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;source&quot;:&quot;RiTMeqgl&quot;,&quot;target&quot;:&quot;7fIGoKmF&quot;,&quot;-0-mxGeometry&quot;:{&quot;relative&quot;:&quot;1&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;YoOyvSGd&quot;:{&quot;id&quot;:&quot;YoOyvSGd&quot;,&quot;value&quot;:&quot;No&quot;,&quot;style&quot;:&quot;edgeLabel;html=1;align=center;verticalAlign=middle;resizable=0;points=[];&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;connectable&quot;:&quot;0&quot;,&quot;parent&quot;:&quot;YmZCqelD&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;-0.2333&quot;,&quot;y&quot;:&quot;-1&quot;,&quot;relative&quot;:&quot;1&quot;,&quot;as&quot;:&quot;geometry&quot;,&quot;-0-mxPoint&quot;:{&quot;x&quot;:&quot;-9&quot;,&quot;as&quot;:&quot;offset&quot;}}},&quot;zkrCXwEg&quot;:{&quot;id&quot;:&quot;zkrCXwEg&quot;,&quot;style&quot;:&quot;edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;&quot;,&quot;edge&quot;:&quot;1&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;source&quot;:&quot;RiTMeqgl&quot;,&quot;target&quot;:&quot;HOUJ8rxm&quot;,&quot;-0-mxGeometry&quot;:{&quot;relative&quot;:&quot;1&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;RiTMeqgl&quot;:{&quot;id&quot;:&quot;RiTMeqgl&quot;,&quot;value&quot;:&quot;是否需要打字机效果<br />显示Response&quot;,&quot;style&quot;:&quot;shape=mxgraph.flowchart.decision;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#000000;strokeWidth=2&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;diagramName&quot;:&quot;Decision&quot;,&quot;diagramCategory&quot;:&quot;Flowchart&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;50&quot;,&quot;y&quot;:&quot;80&quot;,&quot;width&quot;:&quot;160&quot;,&quot;height&quot;:&quot;80&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;myx2SUiP&quot;:{&quot;id&quot;:&quot;myx2SUiP&quot;,&quot;style&quot;:&quot;edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;entryX=0.5;entryY=0;entryDx=0;entryDy=0;entryPerimeter=0;&quot;,&quot;edge&quot;:&quot;1&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;source&quot;:&quot;8YOj2FLv&quot;,&quot;target&quot;:&quot;8VEvPeng&quot;,&quot;-0-mxGeometry&quot;:{&quot;relative&quot;:&quot;1&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;nHpb1eol&quot;:{&quot;id&quot;:&quot;nHpb1eol&quot;,&quot;style&quot;:&quot;edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;entryX=0.5;entryY=1;entryDx=0;entryDy=0;entryPerimeter=0;&quot;,&quot;edge&quot;:&quot;1&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;source&quot;:&quot;8YOj2FLv&quot;,&quot;target&quot;:&quot;HOUJ8rxm&quot;,&quot;-0-mxGeometry&quot;:{&quot;relative&quot;:&quot;1&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;8YOj2FLv&quot;:{&quot;id&quot;:&quot;8YOj2FLv&quot;,&quot;value&quot;:&quot;是否需要<br />同步处理 Response&quot;,&quot;style&quot;:&quot;shape=mxgraph.flowchart.decision;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#000000;strokeWidth=2&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;diagramName&quot;:&quot;Decision&quot;,&quot;diagramCategory&quot;:&quot;Flowchart&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;50&quot;,&quot;y&quot;:&quot;360&quot;,&quot;width&quot;:&quot;160&quot;,&quot;height&quot;:&quot;80&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;HOUJ8rxm&quot;:{&quot;id&quot;:&quot;HOUJ8rxm&quot;,&quot;value&quot;:&quot;流式响应<br />stream=true&quot;,&quot;style&quot;:&quot;shape=mxgraph.flowchart.process;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#000000;strokeWidth=2&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;diagramName&quot;:&quot;Process&quot;,&quot;diagramCategory&quot;:&quot;Flowchart&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;340&quot;,&quot;y&quot;:&quot;220&quot;,&quot;width&quot;:&quot;80&quot;,&quot;height&quot;:&quot;80&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;8VEvPeng&quot;:{&quot;id&quot;:&quot;8VEvPeng&quot;,&quot;value&quot;:&quot;非流式响应<br />stream=false&quot;,&quot;style&quot;:&quot;shape=mxgraph.flowchart.process;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#000000;strokeWidth=2&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;diagramName&quot;:&quot;Process&quot;,&quot;diagramCategory&quot;:&quot;Flowchart&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;90&quot;,&quot;y&quot;:&quot;500&quot;,&quot;width&quot;:&quot;80&quot;,&quot;height&quot;:&quot;80&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;ochkKORB&quot;:{&quot;id&quot;:&quot;ochkKORB&quot;,&quot;style&quot;:&quot;edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;entryX=0.5;entryY=0;entryDx=0;entryDy=0;entryPerimeter=0;&quot;,&quot;edge&quot;:&quot;1&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;source&quot;:&quot;7fIGoKmF&quot;,&quot;target&quot;:&quot;8YOj2FLv&quot;,&quot;-0-mxGeometry&quot;:{&quot;relative&quot;:&quot;1&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;BHIsH2r7&quot;:{&quot;id&quot;:&quot;BHIsH2r7&quot;,&quot;value&quot;:&quot;&quot;,&quot;style&quot;:&quot;edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;entryX=0;entryY=0.5;entryDx=0;entryDy=0;entryPerimeter=0;&quot;,&quot;edge&quot;:&quot;1&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;source&quot;:&quot;7fIGoKmF&quot;,&quot;target&quot;:&quot;HOUJ8rxm&quot;,&quot;-0-mxGeometry&quot;:{&quot;relative&quot;:&quot;1&quot;,&quot;as&quot;:&quot;geometry&quot;,&quot;-0-mxPoint&quot;:{&quot;x&quot;:&quot;290&quot;,&quot;y&quot;:&quot;260&quot;,&quot;as&quot;:&quot;targetPoint&quot;}}},&quot;7fIGoKmF&quot;:{&quot;id&quot;:&quot;7fIGoKmF&quot;,&quot;value&quot;:&quot;是否需要<br />即时查看 Bot 回复&quot;,&quot;style&quot;:&quot;shape=mxgraph.flowchart.decision;whiteSpace=wrap;html=1;fillColor=#ffffff;strokeColor=#000000;strokeWidth=2&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;diagramName&quot;:&quot;Decision&quot;,&quot;diagramCategory&quot;:&quot;Flowchart&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;50&quot;,&quot;y&quot;:&quot;220&quot;,&quot;width&quot;:&quot;160&quot;,&quot;height&quot;:&quot;80&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;zXOnerCh&quot;:{&quot;id&quot;:&quot;zXOnerCh&quot;,&quot;value&quot;:&quot;YES&quot;,&quot;style&quot;:&quot;text;html=1;align=center;verticalAlign=middle;resizable=0;points=[];autosize=1;&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;263&quot;,&quot;y&quot;:&quot;105&quot;,&quot;width&quot;:&quot;40&quot;,&quot;height&quot;:&quot;20&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;P07BDIgi&quot;:{&quot;id&quot;:&quot;P07BDIgi&quot;,&quot;value&quot;:&quot;YES&quot;,&quot;style&quot;:&quot;text;html=1;align=center;verticalAlign=middle;resizable=0;points=[];autosize=1;&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;263&quot;,&quot;y&quot;:&quot;240&quot;,&quot;width&quot;:&quot;40&quot;,&quot;height&quot;:&quot;20&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;3M7TtFTy&quot;:{&quot;id&quot;:&quot;3M7TtFTy&quot;,&quot;value&quot;:&quot;YES&quot;,&quot;style&quot;:&quot;text;html=1;align=center;verticalAlign=middle;resizable=0;points=[];autosize=1;&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;270&quot;,&quot;y&quot;:&quot;380&quot;,&quot;width&quot;:&quot;40&quot;,&quot;height&quot;:&quot;20&quot;,&quot;as&quot;:&quot;geometry&quot;}},&quot;UjbGJGzQ&quot;:{&quot;id&quot;:&quot;UjbGJGzQ&quot;,&quot;value&quot;:&quot;No&quot;,&quot;style&quot;:&quot;edgeLabel;html=1;align=center;verticalAlign=middle;resizable=0;points=[];&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;connectable&quot;:&quot;0&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;130&quot;,&quot;y&quot;:&quot;193&quot;,&quot;as&quot;:&quot;geometry&quot;,&quot;-0-mxPoint&quot;:{&quot;x&quot;:&quot;-16&quot;,&quot;y&quot;:&quot;137&quot;,&quot;as&quot;:&quot;offset&quot;}}},&quot;g36rpM8s&quot;:{&quot;id&quot;:&quot;g36rpM8s&quot;,&quot;value&quot;:&quot;No&quot;,&quot;style&quot;:&quot;edgeLabel;html=1;align=center;verticalAlign=middle;resizable=0;points=[];&quot;,&quot;vertex&quot;:&quot;1&quot;,&quot;connectable&quot;:&quot;0&quot;,&quot;parent&quot;:&quot;B9iU9UZj&quot;,&quot;-0-mxGeometry&quot;:{&quot;x&quot;:&quot;130&quot;,&quot;y&quot;:&quot;200&quot;,&quot;as&quot;:&quot;geometry&quot;,&quot;-0-mxPoint&quot;:{&quot;x&quot;:&quot;-18&quot;,&quot;y&quot;:&quot;267&quot;,&quot;as&quot;:&quot;offset&quot;}}}},&quot;mxCellList&quot;:[&quot;v6dX7BR7&quot;,&quot;B9iU9UZj&quot;,&quot;YmZCqelD&quot;,&quot;YoOyvSGd&quot;,&quot;zkrCXwEg&quot;,&quot;RiTMeqgl&quot;,&quot;myx2SUiP&quot;,&quot;nHpb1eol&quot;,&quot;8YOj2FLv&quot;,&quot;HOUJ8rxm&quot;,&quot;8VEvPeng&quot;,&quot;ochkKORB&quot;,&quot;BHIsH2r7&quot;,&quot;7fIGoKmF&quot;,&quot;zXOnerCh&quot;,&quot;P07BDIgi&quot;,&quot;3M7TtFTy&quot;,&quot;UjbGJGzQ&quot;,&quot;g36rpM8s&quot;]},&quot;lastEditTime&quot;:0,&quot;snapshot&quot;:&quot;&quot;}" />

## 流式响应
在流式响应中，服务端不会一次性发送所有数据，而是以数据流的形式逐条发送数据给客户端，数据流中包含对话过程中触发的各种事件（event），直至处理完毕或处理中断。处理结束后，服务端会通过 conversation.message.completed 事件返回拼接后完整的模型回复信息。各个事件的说明可参考**流式响应事件**。
流式响应允许客户端在接收到完整的数据流之前就开始处理数据，例如在对话界面实时展示智能体的回复内容，减少客户端等待模型完整回复的时间。
流式响应的整体流程如下：

<div type="doc-tabs">
<div type="tab-item" title="流式响应流程" key="k55EkHyV7d">

```JSON
######### 整体概览 （chat, MESSAGE 两级）
# chat - 开始
# chat - 处理中
#   MESSAGE - 知识库召回
#   MESSAGE - function_call
#   MESSAGE - tool_output
#   MESSAGE - answer is card
#   MESSAGE - answer is normal text
#   MESSAGE - 多 answer 的情况下，会继续有 message.delta
#   MESSAGE - verbose （多 answer、Multi-agent 跳转等场景）
#   MESSAGE - suggestion
# chat - 完成
# 流结束 event: done
#########
```


</div>
<div type="tab-item" title="流式响应示例" key="qnSIxFYtuw">

```JSON
# chat - 开始
event: conversation.chat.created
// 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。
data: {"id": "123", "conversation_id":"123", "bot_id":"222", "created_at":1710348675,compleated_at:null, "last_error": null, "meta_data": {}, "status": "created","usage":null}

# chat - 处理中
event: conversation.chat.in_progress
data: {"id": "123", "conversation_id":"123", "bot_id":"222", "created_at":1710348675, compleated_at: null, "last_error": null,"meta_data": {}, "status": "in_progress","usage":null}

# MESSAGE - 知识库召回
event: conversation.message.completed
data: {"id": "msg_001", "role":"assistant","type":"knowledge","content":"---\nrecall slice 1:xxxxxxx\n","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

# MESSAGE - function_call
event: conversation.message.completed
data: {"id": "msg_002", "role":"assistant","type":"function_call","content":"{\"name\":\"toutiaosousuo-search\",\"arguments\":{\"cursor\":0,\"input_query\":\"今天的体育新闻\",\"plugin_id\":7281192623887548473,\"api_id\":7288907006982012986,\"plugin_type\":1","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

# MESSAGE - toolOutput
event: conversation.message.completed
data: {"id": "msg_003", "role":"assistant","type":"tool_output","content":"........","content_type":"card","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

# MESSAGE - answer is card
event: conversation.message.completed
data: {"id": "msg_004", "role":"assistant","type":"answer","content":"{{card_json}}","content_type":"card","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

# MESSAGE - answer is normal text
event: conversation.message.delta
data:{"id": "msg_005", "role":"assistant","type":"answer","content":"以下","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

event: conversation.message.delta
data:{"id": "msg_005", "role":"assistant","type":"answer","content":"是","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

...... {{ N 个 delta 消息包}} ......

event: conversation.message.completed
data:{"id": "msg_005", "role":"assistant","type":"answer","content":"{{msg_005 完整的结果。即之前所有 msg_005 delta 内容拼接的结果}}","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}


# MESSAGE - 多 answer 的情况,会继续有 message.delta
event: conversation.message.delta
data:{"id": "msg_006", "role":"assistant","type":"answer","content":"你好你好","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

...... {{ N 个 delta 消息包}} ......

event: conversation.message.completed
data:{"id": "msg_006", "role":"assistant","type":"answer","content":"{{msg_006 完整的结果。即之前所有 msg_006 delta 内容拼接的结果}}","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

# MESSAGE - Verbose （流式 plugin, 多 answer 结束，Multi-agent 跳转等场景）
event: conversation.message.completed
data:{"id": "msg_007", "role":"assistant","type":"verbose","content":"{\"msg_type\":\"generate_answer_finish\",\"data\":\"\"}","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

# MESSAGE - suggestion
event: conversation.message.completed
data: {"id": "msg_008", "role":"assistant","type":"follow_up","content":"朗尼克的报价是否会成功？","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
event: conversation.message.completed
data: {"id": "msg_009", "role":"assistant","type":"follow_up","content":"中国足球能否出现？","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}
event: conversation.message.completed
data: {"id": "msg_010", "role":"assistant","type":"follow_up","content":"羽毛球种子选手都有谁？","content_type":"text","chat_id": "123", "conversation_id":"123", "bot_id":"222"}

# chat - 完成
event: conversation.chat.completed （chat完成）
data: {"id": "123", "chat_id": "123", "conversation_id":"123", "bot_id":"222", "created_at":1710348675, compleated_at:1710348675, "last_error":null, "meta_data": {}, "status": "compleated", "usage":{"token_count":3397,"output_tokens":1173,"input_tokens":2224}}

event: done （stream流结束）
data: [DONE]

# chat - 失败
event: conversation.chat.failed
data: {
    "code":701231,
    "msg":"error"
}
```


</div>
</div>
返回的事件消息体结构如下：
| **参数** | **类型** | **说明** |
| --- | --- | --- |
| event | String | 当前流式返回的数据包事件。详细说明可参考 **流式响应事件**。 |
| data | Object | 消息内容。其中，chat 事件和 message 事件的格式不同。 <br>  <br> * chat 事件中，data 为 **Chat** **Object**。 <br> * message、audio 事件中，data 为 **Message** **Object**。 |
## **流式响应事件**
| **事件（event）名称** | **说明** |
| --- | --- |
| conversation.chat.created | 创建对话的事件，表示对话开始。 |
| conversation.chat.in_progress | 服务端正在处理对话。 |
| conversation.message.delta | 增量消息，通常是 type=answer 时的增量消息。 |
| conversation.audio.delta | 增量语音消息，通常是 type=answer 时的增量消息。只有输入中包含语音消息时，才会返回 audio.delta。 |
| conversation.message.completed | message 已回复完成。此时流式包中带有所有 message.delta 的拼接结果，且每个消息均为 completed 状态。 |
| conversation.chat.completed | 对话完成。 |
| conversation.chat.failed | 此事件用于标识对话失败。 |
| conversation.chat.requires_action | 对话中断，需要使用方上报工具的执行结果。 |
| error | 流式响应过程中的错误事件。关于 code 和 msg 的详细说明，可参考[错误码](https://docs.coze.cn/api/open/docs/developer_guides/coze_error_codes)。 |
| done | 本次会话的流式返回正常结束。 |
## 非流式响应
在非流式响应中，无论服务端是否处理完毕，立即发送响应消息。其中包括本次对话的 chat_id、状态等元数据信息，但不包括模型处理的最终结果。
非流式响应不需要维持长连接，在场景实现上更简单，但通常需要客户端主动查询对话状态和消息详情才能得到完整的数据。你可以通过接口[查看对话详情](https://docs.coze.cn/api/open/docs/developer_guides/retrieve_chat)确认本次对话处理结束后，再调用[查看对话消息详情](https://docs.coze.cn/api/open/docs/developer_guides/list_chat_messages)接口查看模型回复等完整响应内容。流程如下：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/79f9ce1cefc94f4daf20cad241382cdf~tplv-goo7wpa0wc-image.image)
非流式响应的结构如下：
| **参数** | **类型** | **说明** |
| --- | --- | --- |
| data | Object | 本次对话的基本信息。详细说明可参考 **Chat** **Object**。 |
| code | Integer | 状态码。 <br> `0` 代表调用成功。 |
| msg | String | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 |
## 响应参数
#### Chat Object
| 参数 | 类型 | 是否可选 | 说明 |
| --- | --- | --- | --- |
| id | String | 必填 | 对话 ID，即对话的唯一标识。 |
| conversation_id | String | 必填 | 会话 ID，即会话的唯一标识。 |
| bot_id | String | 必填 | 要进行会话聊天的智能体 ID。 |
| created_at | Integer | 选填 | 对话创建的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| completed_at | Integer | 选填 | 对话结束的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| failed_at | Integer | 选填 | 对话失败的时间。格式为 10 位的 Unixtime 时间戳，单位为秒。 |
| meta_data | Map<String,String> | 选填 | 发起对话时的附加消息，用于传入使用方的自定义数据，[查看对话详情](https://www.coze.cn/open/docs/developer_guides/retrieve_chat)时也会返回此附加消息。 <br> 自定义键值对，应指定为 Map 对象格式。长度为 16 对键值对，其中键（key）的长度范围为 1～64 个字符，值（value）的长度范围为 1～512 个字符。 |
| last_error <br>  | Object | 选填 | 对话运行异常时，此字段中返回详细的错误信息，包括： <br>  <br> * Code：错误码。Integer 类型。0 表示成功，其他值表示失败。 <br> * Msg：错误信息。String 类型。 <br>  <br> * 对话正常运行时，此字段返回 null。 <br> * suggestion 失败不会被标记为运行异常，不计入 last_error。 <br>  |
| status <br>  | String | 必填 | 对话的运行状态。取值为： <br>  <br> * created：对话已创建。 <br> * in_progress：智能体正在处理中。 <br> * completed：智能体已完成处理，本次对话结束。 <br> * failed：对话失败。 <br> * requires_action：对话中断，需要进一步处理。 <br> * canceled：对话已取消。 |
| required_action | Object | 选填 | 需要运行的信息详情。 |
| » type | String | 选填 | 额外操作的类型，枚举值为 submit_tool_outputs。 |
| »submit_tool_outputs | Object | 选填 | 需要提交的结果详情，通过提交接口上传，并可以继续聊天 |
| »» tool_calls | Array of Object | 选填 | 具体上报信息详情。 |
| »»» id | String | 选填 | 上报运行结果的 ID。 |
| »»» type | String | 选填 | 工具类型，枚举值包括： <br>  <br> * function：待执行的方法，通常是端插件。触发端插件时会返回此枚举值。 <br> * reply_message：待回复的选项。触发工作流问答节点时会返回此枚举值。 |
| »»» function | Object | 选填 | 执行方法 function 的定义。 |
| »»»» name | String | 选填 | 方法名。 |
| »»»» arguments | String | 选填 | 方法参数。 |
| usage | Object | 选填 | Token 消耗的详细信息。实际的 Token 消耗以对话结束后返回的值为准。 |
| »token_count | Integer | 选填 | 本次对话消耗的 Token 总数，包括 input 和 output 部分的消耗。 |
| »output_count | Integer | 选填 | output 部分消耗的 Token 总数。 |
| »input_count | Integer | 选填 | input 部分消耗的 Token 总数。 |
Chat Object 的示例如下：

<div type="doc-tabs">
<div type="tab-item" title="状态正常的对话" key="JBNDckznoe">

```JSON
{
// 在 chat 事件里，data 字段中的 id 为 Chat ID，即对话 ID。
    "id": "737662389258662****",
    "conversation_id": "737554565555041****",
    "bot_id": "736661612448078****",
    "completed_at": 1717508113,
    "last_error": {
        "code": 0,
        "msg": ""
    },
    "status": "completed",
    "usage": {
        "token_count": 6644,
        "output_count": 766,
        "input_count": 5878
    }
}
```


</div>
<div type="tab-item" title="需要使用方额外处理的对话" key="n2AU9663Ib">

```JSON
{
    "code": 0,
    "data": {
        "bot_id": "737282596785517****",
        "completed_at": 1717513285,
        "conversation_id": "737666232053956****",
        "created_at": 1717513283,
        // 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。
        "id": "737666232053959****",
        "required_action": {
            "submit_tool_outputs": {
                "tool_calls": [
                    {
                        "function": {
                            "arguments": "{\"location\":\"南京\",\"type\":0}",
                            "name": "local_data_assistant"
                        },
                        "id": "BUJJF0dAQ0NAEBVeQkVKEV5HFURFXhFCEhFeFxdHShcSQEtFSxYRSUI=",
                        "type": "function"
                    }
                ]
            },
            "type": "submit_tool_outputs"
        },
        "status": "requires_action"
    },
    "msg": ""
}
```


</div>
</div>



#### Message Object
| **参数** | **类型** | **说明** |
| --- | --- | --- |
| id | String | 智能体回复的消息的 Message ID，即消息的唯一标识。 |
| conversation_id | String | 此消息所在的会话 ID。 |
| bot_id | String | 编写此消息的智能体ID。此参数仅在对话产生的消息中返回。 |
| chat_id | String | Chat ID。此参数仅在对话产生的消息中返回。 |
| meta_data | Map | 创建消息时的附加消息，[查看消息列表](https://www.coze.cn/open/docs/developer_guides/list_message)时也会返回此附加消息。 |
| role | String | 发送这条消息的实体。取值： <br>  <br> * **user**：代表该条消息内容是用户发送的。 <br> * **assistant**：代表该条消息内容是智能体发送的。 |
| content | String <br>  | 消息的内容，支持纯文本、多模态（文本、图片、文件混合输入）、卡片等多种类型的内容。 <br> * 当 `role` 为 user 时，支持返回多模态内容。 <br> * 当 `role` 为 assistant 时，只支持返回纯文本内容。 <br>  |
| content_type | String | 消息内容的类型，取值包括： <br>  <br> * text：文本。 <br> * object_string：多模态内容，即文本和文件的组合、文本和图片的组合。 <br> * card：卡片。此枚举值仅在接口响应中出现，不支持作为入参。 <br> * audio：音频。此枚举值仅在接口响应中出现，不支持作为入参。仅当输入有 audio 文件时，才会返回此类型。当 content_type 为 audio 时，content 为 base64 后的音频数据。音频的编码根据输入的 audio 文件的不同而不同： <br>    * 输入为 wav 格式音频时，content 为**采样率 24kHz，raw 16 bit, 1 channel, little-endian 的 pcm 音频片段 base64 后的字符串** <br>    * 输入为 ogg_opus 格式音频时，content 为**采样率 48kHz，1 channel，10ms 帧长的 opus 格式音频片段base64 后的字符串** |
| created_at | Integer | 消息的创建时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| updated_at | Integer | 消息的更新时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| type | String | 消息类型。 <br>  <br> * **question**：用户输入内容。 <br> * **answer**：智能体返回给用户的消息内容，支持增量返回。如果工作流绑定了 messge 节点，可能会存在多 answer 场景，此时可以用流式返回的结束标志来判断所有 answer 完成。 <br> * **function_call**：智能体对话过程中调用函数（function call）的中间结果。 <br> * **tool_response**：调用工具 （function call）后返回的结果。 <br> * **follow_up**：如果在智能体上配置打开了用户问题建议开关，则会返回推荐问题相关的回复内容。 <br> * **verbose**：多 answer 场景下，服务端会返回一个 verbose 包，对应的 content 为 JSON 格式，`content.msg_type =generate_answer_finish` 代表全部 answer 回复完成。 <br>  <br> 仅发起会话（v3）接口支持将此参数作为入参，且： <br>  <br> * 如果 autoSaveHistory=true，type 支持设置为 question 或 answer。 <br> * 如果 autoSaveHistory=false，type 支持设置为 question、answer、function_call、tool_response。 <br>  <br> 其中，type=question 只能和 role=user 对应，即仅用户角色可以且只能发起 question 类型的消息。详细说明可参考[消息 type 说明](https://docs.coze.cn/api/open/docs/developer_guides/message_type)。 <br>  |
| section_id | String | 上下文片段 ID。每次清除上下文都会生成一个新的 section_id。 |
| reasoning_content | String | 模型的思维链（CoT），展示模型如何将复杂问题逐步分解为多个简单步骤并推导出最终答案。仅当模型支持深度思考、且智能体开启了深度思考时返回该字段，当前支持深度思考的模型如下： <br>  <br> * 豆包·1.6·自动深度思考·多模态模型 <br> * 豆包·1.6·极致速度·多模态模型 <br> * 豆包·1.5·Pro·视觉深度思考 <br> * 豆包·GUI·Agent模型 <br> * DeepSeek-R1 |
# 示例
## 流式响应

<div type="doc-tabs">
<div type="tab-item" title="基础问答" key="Mml7aHdE4M">

* **Request**
   ```Shell
   curl --location --request POST 'https://api.coze.cn/v3/chat?conversation_id=7374752000116113452' \
   --header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
   --header 'Content-Type: application/json' \
   --data-raw '{
       "bot_id": "734829333445931****",
       "user_id": "123456789",
       "stream": true,
       "auto_save_history":true,
       "additional_messages":[
           {
               "role":"user",
               "content":"2024年10月1日是星期几",
               "content_type":"text"
           }
       ]
   }'
   ```

* **Response**
   ```JSON
   event:conversation.chat.created
   // 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。
   data:{"id":"7382159487131697202","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","completed_at":1718792949,"last_error":{"code":0,"msg":""},"status":"created","usage":{"token_count":0,"output_count":0,"input_count":0}}
   
   event:conversation.chat.in_progress
   data:{"id":"7382159487131697202","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","completed_at":1718792949,"last_error":{"code":0,"msg":""},"status":"in_progress","usage":{"token_count":0,"output_count":0,"input_count":0}}
   
   event:conversation.message.delta
   data:{"id":"7382159494123470858","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"2","content_type":"text","chat_id":"7382159487131697202"}
   
   event:conversation.message.delta
   data:{"id":"7382159494123470858","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"0","content_type":"text","chat_id":"7382159487131697202"}
   
   //省略模型回复的部分中间事件event:conversation.message.delta
   ......
   
   event:conversation.message.delta
   data:{"id":"7382159494123470858","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"星期三","content_type":"text","chat_id":"7382159487131697202"}
   
   event:conversation.message.delta
   data:{"id":"7382159494123470858","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"。","content_type":"text","chat_id":"7382159487131697202"}
   
   event:conversation.message.completed
   data:{"id":"7382159494123470858","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"2024 年 10 月 1 日是星期三。","content_type":"text","chat_id":"7382159487131697202"}
   
   event:conversation.message.completed
   data:{"id":"7382159494123552778","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"verbose","content":"{\"msg_type\":\"generate_answer_finish\",\"data\":\"\",\"from_module\":null,\"from_unit\":null}","content_type":"text","chat_id":"7382159487131697202"}
   
   event:conversation.chat.completed
   data:{"id":"7382159487131697202","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","completed_at":1718792949,"last_error":{"code":0,"msg":""},"status":"completed","usage":{"token_count":633,"output_count":19,"input_count":614}}
   
   event:done
   data:"[DONE]"
   ```


</div>
<div type="tab-item" title="图文问答" key="azV1Tv9j4o">

* Request
   ```Shell
   curl --location --request POST 'https://api.coze.cn/v3/chat?conversation_id=7374752000116113452' \
   --header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
   --header 'Content-Type: application/json' \
   --data-raw '{
       "bot_id": "737946218936519****",
       "user_id": "123456789",
       "stream": true,
       "auto_save_history":true,
       "additional_messages":[
           {
               "role":"user",
               "content":"[{\"type\":\"image\",\"file_url\":\"https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/4ca71a5f55d54efc95ed9c06e019ff4b.png\"},{\"type\":\"text\",\"text\":\"帮我看看这张图片里都有什么\"}]",
               "content_type":"object_string"
           }
       ]
   }'
   ```

* Response
   ```JSON
   event:conversation.chat.created
   // 在 chat 事件里，data 字段中的 id 为 Chat ID，即会话 ID。
   data:{"id":"7382158397837344768","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","completed_at":1718792697,"last_error":{"code":0,"msg":""},"status":"created","usage":{"token_count":0,"output_count":0,"input_count":0}}
   
   event:conversation.chat.in_progress
   data:{"id":"7382158397837344768","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","completed_at":1718792697,"last_error":{"code":0,"msg":""},"status":"in_progress","usage":{"token_count":0,"output_count":0,"input_count":0}}
   
   event:conversation.message.completed
   data:{"id":"7382158491307212815","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"function_call","content":"{\"name\":\"tupianlijie-imgUnderstand\",\"arguments\":{\"text\":\"描述图片里有什么\",\"url\":\"https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/4ca71a5f55d54efc95ed9c06e019ff4b.png\"},\"plugin_id\":7379227414322217010,\"api_id\":7379227414322233394,\"plugin_type\":1,\"thought\":\"需求为描述图片（https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/4ca71a5f55d54efc95ed9c06e019ff4b.png）里都有什么，需要调用tupianlijie-imgUnderstand工具理解图片\"}","content_type":"text","chat_id":"7382158397837344768"}
   
   event:conversation.message.completed
   data:{"id":"7382158637826998312","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"tool_response","content":"{\"content_type\":1,\"response_for_model\":\" 这幅图像描绘了一片宁静而神秘的森林景象。画面中心是一条蜿蜒流淌的小溪，溪水清澈见底，溪床上散布着几块大小不一的石头。溪流两岸覆盖着厚厚的青苔，与周围的树木形成鲜明对比。树木高大挺拔，树干粗壮，树皮呈深褐色，树枝上长满了翠绿的针叶，阳光透过树叶的缝隙洒在地面上，形成斑驳的光影。整个场景被一层淡淡的雾气笼罩，增添了一丝神秘和幽静的氛围。画面远处的树木逐渐变得模糊，与天空的灰白色融为一体，整个画面色彩以绿色和棕色为主，营造出一种深邃而古老的感觉。整体上，这幅图像传达了一种与自然和谐共处的宁静与平和。\",\"type_for_model\":1}","content_type":"text","chat_id":"7382158397837344768"}
   
   event:conversation.message.delta
   data:{"id":"7382158479043379234","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"这","content_type":"text","chat_id":"7382158397837344768"}
   
   event:conversation.message.delta
   data:{"id":"7382158479043379234","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"是","content_type":"text","chat_id":"7382158397837344768"}
   
   //省略模型回复的部分中间事件event:conversation.message.delta
   ......
   
   event:conversation.message.delta
   data:{"id":"7382158479043379234","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"树木","content_type":"text","chat_id":"7382158397837344768"}
   
   event:conversation.message.delta
   data:{"id":"7382158479043379234","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"。","content_type":"text","chat_id":"7382158397837344768"}
   
   event:conversation.message.completed
   data:{"id":"7382158479043379234","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"answer","content":"这是一幅非常漂亮的森林图片，里面有小溪、石头和青苔覆盖的树木。","content_type":"text","chat_id":"7382158397837344768"}
   
   event:conversation.message.completed
   data:{"id":"7382158652519645218","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","role":"assistant","type":"verbose","content":"{\"msg_type\":\"generate_answer_finish\",\"data\":\"\",\"from_module\":null,\"from_unit\":null}","content_type":"text","chat_id":"7382158397837344768"}
   
   event:conversation.chat.completed
   data:{"id":"7382158397837344768","conversation_id":"7381473525342978089","bot_id":"7379462189365198898","completed_at":1718792697,"last_error":{"code":0,"msg":""},"status":"completed","usage":{"token_count":2308,"output_count":111,"input_count":2197}}
   
   event:done
   data:"[DONE]"
   ```


</div>
</div>
## 非流式响应

<div type="doc-tabs">
<div type="tab-item" title="基础问答" key="GsU2PlCeyJ">

* **Request**
   ```Shell
   curl --location --request POST 'https://api.coze.cn/v3/chat?conversation_id=737475200011611****' \
   --header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
   --header 'Content-Type: application/json' \
   --data-raw '{
       "bot_id": "734829333445931****",
       "user_id": "123456789",
       "stream": false,
       "auto_save_history":true,
       "additional_messages":[
           {
               "role":"user",
               "content":"今天杭州天气如何",
               "content_type":"text"
           }
       ]
   }'
   ```

* **Response**
   ```JSON
   {
       "data":{
       // data 字段中的 id 为 Chat ID，即会话 ID。
           "id": "123",
           "conversation_id": "123456",
           "bot_id": "222",
           "created_at": 1710348675,
           "completed_at": 1710348675,
           "last_error": null,
           "meta_data": {},
           "status": "completed",
           "usage": {
               "token_count": 3397,
               "output_count": 1173,
               "input_count": 2224
           }
       },
       "code":0,
       "msg":""
   }
   ```


</div>
<div type="tab-item" title="图文问答" key="bDGoCM7iwU">

* Request
   ```Shell
   curl --location --request POST 'https://api.coze.cn/v3/chat?conversation_id=737475200011611****' \
   --header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
   --header 'Content-Type: application/json' \
   --data-raw '{
       "bot_id": "737946218936519****",
       "user_id": "123456789",
       "stream": false,
       "auto_save_history":true,
       "additional_messages":[
           {
               "role":"user",
               "content":"[{\"type\":\"image\",\"file_url\":\"https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/4ca71a5f55d54efc95ed9c06e019ff4b.png\"},{\"type\":\"text\",\"text\":\"帮我看看这张图片里都有什么\"}]",
               "content_type":"object_string"
           }
       ]
   }'
   ```

* Response
   ```JSON
   {
       "data":{
       // data 字段中的 id 为 Chat ID，即会话 ID。
           "id": "123",
           "conversation_id": "123456",
           "bot_id": "222",
           "created_at": 1710348675,
           "completed_at": 1710348675,
           "last_error": null,
           "meta_data": {},
           "status": "compleated",
           "usage": {
               "token_count": 3397,
               "output_count": 1173,
               "input_count": 2224
           }
       },
       "code":0,
       "msg":""
   }
   ```


</div>
<div type="tab-item" title="快捷指令" key="BazQzexecW">

* Request
   ```Shell
   curl --location --request POST 'https://api.coze.cn/v3/chat?conversation_id=737475200011611****' \
   --header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
   --header 'Content-Type: application/json' \
   --data-raw '{
       "bot_id": "737946218936519****",
       "user_id": "123456789",
       "stream": false,
       "auto_save_history":true,
       "additional_messages":[],
       "shortcut_command":{
           "command_id":"745701083352557****",
           "parameters":{
               "news_query":"[{\"type\":\"text\",\"text\":\"娱乐新闻\"}]"
           }
       }
   }'
   ```

* Response
   ```JSON
   {
       "data":{
       // data 字段中的 id 为 Chat ID，即会话 ID。
           "id": "123",
           "conversation_id": "123456",
           "bot_id": "222",
           "created_at": 1710348675,
           "completed_at": 1710348675,
           "last_error": null,
           "meta_data": {},
           "status": "compleated",
           "usage": {
               "token_count": 3397,
               "output_count": 1173,
               "input_count": 2224
           }
       },
       "code":0,
       "msg":""
   }
   ```


</div>
</div>
## 给自定义参数赋值
自定义参数用于在智能体交互中存储和管理每个用户的特定信息，例如用户ID、地理位置等，以便实现个性化处理和差异化响应。你可以对话流中输入自定义参数，并在用户与智能体对话时动态更新和读取变量值。

1. 在对话流开始节点的输入参数中设置自定义用户变量。本文以自定义参数 `user` 为例，你可以根据实际业务场景设置其他参数。
   ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/47ae41fd2e864943adee2a872eaec23b~tplv-goo7wpa0wc-image.image)
2. 在调用发起对话接口时，通过 `parameters` 参数传入变量的值。例如给自定义参数 `user` 赋值，API 调用的示例代码如下：
   ```Shell
   
   curl --location --request POST 'https://api.coze.cn/v3/chat?conversation_id=7374752000116113452' \
        --header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
        --header 'Content-Type: application/json' \
        --data-raw '{
       "bot_id": "734829333445931****",
       "user_id": "123456789",
       "stream": true,
       "auto_save_history": true,
       "parameters": {
           "user": [
               {
                   "user_id": "123456",
                   "user_name": "user"
               }
           ]
       },
       "additional_messages": [
           {
               "role": "user",
               "content": "2024年10月1日是星期几",
               "content_type": "text"
           }
       ]
   }'
   ```


## Prompt 变量
例如在智能体的 Prompt 中定义了一个 {{bot_name}} 的变量，在调用接口时，可以通过 **custom_variables** 参数传入变量值。

<div style="display: flex;">
<div style="flex-shrink: 0;width: calc((100% - 16px) * 0.5000);">

智能体Prompt 配置示例：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/23fa1031548148bd969b3462f5b4e436~tplv-goo7wpa0wc-image.image)


</div>
<div style="flex-shrink: 0;width: calc((100% - 16px) * 0.5000);margin-left: 16px;">

API 调用示例：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/9097bfa34a06460cabf3682f815e54a3~tplv-goo7wpa0wc-image.image)


</div>
</div>

扣子编程也支持 Jinja2 语法。在下面这个模板中，prompt1 将在 key 变量存在时使用，而 prompt2 将在 key 变量不存在时使用。通过在 custom_variables 中传递 key 的值，你可以控制智能体的响应。
```Python
{% if key -%}
prompt1
{%- else %}
prompt2
{% endif %}
```


<div style="display: flex;">
<div style="flex-shrink: 0;width: calc((100% - 16px) * 0.5000);">

智能体Prompt 配置示例：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/a137259f46954467879145d31999e4ee~tplv-goo7wpa0wc-image.image)


</div>
<div style="flex-shrink: 0;width: calc((100% - 16px) * 0.5000);margin-left: 16px;">

API 调用示例：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/a2e07f36ca9943949889200bc51de7ec~tplv-goo7wpa0wc-image.image)


</div>
</div>

**custom_variables** 参数仅适用于智能体提示词中定义的变量，用于提示词范围内动态注入值，他和智能体或工作流的用户变量不同，不支持在工作流中调用提示词变量，工作流中的用户变量需要通过 `parameters` 参数传递给工作流。

## 携带上下文
你可以在发起对话时把多条消息作为上下文一起上传，模型会参考上下文消息，对用户 Query 进行针对性回复。在发起对话时，扣子编程会将以下内容作为上下文传递给模型。

* **会话中的消息**：调用`发起对话`接口时，如果指定了会话 ID，会话中已有的消息会作为上下文传递给模型。
* **additional_messages 中的消息**：如果 additional_messages 中有多条消息，则最后一条会作为本次用户 Query，其他消息为上下文。

扣子编程推荐你通过以下方式在对话中指定上下文：
| 方式 | 说明 |
| --- | --- |
| 方式一：通过会话传递历史消息，通过 additional_messages 指定用户 Query | 适用于在已有会话中再次发起对话的场景，会话中通常已经存在部分历史消息，开发者也可以手动插入一些消息作为上下文。 |
| 方式二：通过 additional_messages 指定历史消息和用户 Query | 此方式无需提前创建会话，通过`发起对话`一个接口即可完成一次携带上下文的对话，更适用于一问一答的场景，使用方式更加简便。 |
以方式一为例，在对话中携带上下文的操作步骤如下：

1. 准备上下文消息。
   准备上下文消息时应注意：
   
   * 应包含用户询问和模型返回两部分消息数据。详情可参考返回参数内容中 Message 消息结构的具体说明。
   * 上下文消息列表按时间递增排序，即最近一条 message 在列表的最后一位。
   * 只需传入用户输入内容及模型返回内容即可，即 `role=user` 和 `role=assistant; type=answer`。

   以下消息列表是一个完整的上下文消息。其中：
   * 第 2 行是用户传入的历史消息
   * 第 4 行是模型返回的历史消息
   ```JSON
   [
   { "role": "user", "content_type":"text", "content": "你可以读懂图片中的内容吗" }
   
   {"role":"assistant","type":"answer","content":"没问题！你想查看什么图片呢？"，"content_type":"text"}
   ]
   ```

2. 调用[创建会话](https://docs.coze.cn/api/open/docs/developer_guides/create_conversation)接口创建一个会话，其中包含以上两条消息，并记录会话 ID。
   请求示例如下：
   ```Shell
   curl --location --request POST 'https://api.coze.cn/v1/conversation/create' \
   --header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
   --header 'Content-Type: application/json' \
   --data-raw '{
       "meta_data": {
           "uuid": "newid1234"
       },
      "messages": [
           {
               "role": "user",
               "content":"你可以读懂图片中的内容吗",
               "content_type":"text"
           },
           {
               "role": "assistant",
               "type":"answer",
               "content": "没问题！你想查看什么图片呢？",
               "content_type":"text"
           }
       ]
   }'
   ```

3. 调用发起对话（V3）接口，并指定会话 ID。
   在对话中可以通过 additional_messages 增加本次对话的 query。这些消息会和对话中已有的消息一起作为上下文被传递给大模型。
   ```Shell
   curl --location --request POST 'https://api.coze.cn/v3/chat?conversation_id=737363834493434****' \
   --header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
   --header 'Content-Type: application/json' \
   --data-raw '{
       "bot_id": "734829333445931****",
       "user_id": "123456789",
       "stream": false,
       "auto_save_history":true,
       "additional_messages":[
           {
               "role":"user",
               "content":"[{\"type\":\"image\",\"file_url\":\"https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fci.xiaohongshu.com%2Fe7368218-8a64-bda3-56ad-5672b2a113b2%3FimageView2%2F2%2Fw%2F1080%2Fformat%2Fjpg&refer=http%3A%2F%2Fci.xiaohongshu.com&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=auto?sec=1720005307&t=1acd734e6e8937c2d77d625bcdb0dc57\"},{\"type\":\"text\",\"text\":\"这张可以吗\"}]",
               "content_type":"object_string"
           }
       ]
   }'
   ```

4. 调用接口[查看对话消息详情](https://docs.coze.cn/api/open/docs/developer_guides/list_chat_messages)查看模型回复。
   你可以从智能体的回复中看出这一次会话是符合上下文语境的。响应信息如下：
   ```JSON
   {
       "code": 0,
       "data": [
           {
               "bot_id": "737946218936519****",
               "content": "{\"name\":\"tupianlijie-imgUnderstand\",\"arguments\":{\"text\":\"图中是什么内容\",\"url\":\"https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/4ca71a5f55d54efc95ed9c06e019ff4b.png\"},\"plugin_id\":7379227414322217010,\"api_id\":7379227414322233394,\"plugin_type\":1,\"thought\":\"需求为识别图中（https://lf-bot-studio-plugin-resource.coze.cn/obj/bot-studio-platform-plugin-tos/artist/image/4ca71a5f55d54efc95ed9c06e019ff4b.png）的内容，需要调用tupianlijie-imgUnderstand进行识别\"}",
               "content_type": "text",
               "conversation_id": "738147352534297****",
               "id": "7381473945440239668",
               "role": "assistant",
               "type": "function_call"
           },
           {
               "bot_id": "7379462189365198898",
               "content": "{\"content_type\":1,\"response_for_model\":\"图中展示的是一片茂密的树林。\",\"type_for_model\":1}",
               "content_type": "text",
               "conversation_id": "738147352534297****",
               "id": "7381473964905807872",
               "role": "assistant",
               "type": "tool_response"
           },
           {
               "bot_id": "7379462189365198898",
               "content": "{\"msg_type\":\"generate_answer_finish\",\"data\":\"\",\"from_module\":null,\"from_unit\":null}",
               "content_type": "text",
               "conversation_id": "738147352534297****",
               "id": "7381473964905906176",
               "role": "assistant",
               "type": "verbose"
           },
           {
               "bot_id": "7379462189365198898",
               "content": "这幅图展示的是一片茂密的树林。",
               "content_type": "text",
               "conversation_id": "738147352534297****",
               "id": "7381473945440223284",
               "role": "assistant",
               "type": "answer"
           }
       ],
       "msg": ""
   }
   ```





