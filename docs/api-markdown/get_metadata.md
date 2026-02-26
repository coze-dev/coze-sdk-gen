# 获取已发布智能体的配置（即将下线）
获取指定智能体的配置信息，此智能体必须已发布到 Agent as API 渠道中。
此接口仅支持查看已发布为 API 服务的智能体。对于创建后从未发布到 API 渠道的智能体，可以在[扣子平台](https://www.coze.cn/)中查看列表及配置。
该 API 即将下线，建议替换为[查看智能体配置](https://docs.coze.cn/developer_guides/get_metadata_draft_published) API。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/bot/get_online_info <br> ``` <br>  |
| **权限** | `getMetadata` <br> 确保调用该接口使用的个人令牌开通了 getMedata 权限，详细信息参考[鉴权方式](https://docs.coze.cn/api/open/docs/www.coze.cn/docs/developer_guides/authentication)。 |
| **接口说明** | 获取指定智能体的配置信息，此智能体必须已发布到 Agent as API 渠道中。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子平台中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| bot_id | String | 必选 | 73428668***** | 要查看的智能体 ID。 <br> 进入智能体的 开发页面，开发页面 URL 中 `bot` 参数后的数字就是智能体 ID。例如`https://www.coze.cn/space/341****/bot/73428668*****`，bot ID 为`73428668*****`。 <br> 确保该智能体的所属空间已经生成了访问令牌。 <br>  |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。 <br>  <br> * `0` 表示调用成功。 <br> * 其他值表示调用失败。你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [BotInfo](#botinfo) | 参考返回示例部分 | 响应的业务信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFAD653960" } | 响应详情信息。 |
### BotInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| bot_id | String | 73428668***** | 智能体的唯一标识。 |
| name | String | 新闻 | 智能体的名称。 |
| description | String | 每天给我推送 AI 相关的新闻。 | 智能体的描述信息。 |
| icon_url | String | https://example.com/icon.png | 智能体的头像地址，用于展示智能体的图标。 |
| create_time | Long | 1715689059 | 创建时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| update_time | Long | 1716388526 | 更新时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
| version | String | 171638852**** | 智能体最新版本的版本号。 |
| prompt_info | Object of [PromptInfo](#promptinfo) | {"prompt": "调用getToutiaoNews工具推送最新的科技新闻。"} | 智能体的提示词配置。 |
| onboarding_info | Object of [OnboardingInfoV2](#onboardinginfov2) | { "prologue": "你好，我可以为你提供最新、最有趣的科技新闻。让我们一起探索科技的世界吧！", "suggested_questions": [ "你能给我推荐一些最新的科技新闻吗？", "你知道最近有哪些科技趋势吗？", "你能告诉我如何获得更多关于科技的信息吗？" ] } | 智能体的开场白配置。 |
| bot_mode | Integer | 0 | 智能体模式，取值： <br>  <br> * **0**：单 Agent 模式 <br> * **1**：多 Agent 模式 |
| plugin_info_list | Array of [PluginInfo](#plugininfo) | [{"plugin_id":"730197029480849****","name":"头条新闻","icon_url":"https://example.com/plugin_icon.png","description":"持续更新，了解最新的头条新闻和新闻文章。","api_info_list":[{"api_id":"730197029480851****","name":"getToutiaoNews","description":"搜索新闻讯息"}]}] | 智能体配置的插件列表，包含插件的名称、图标、描述及工具信息。 |
| model_info | Object of [ModelInfo](#modelinfo) | {"top_k":50,"top_p":1,"model_id":"1706077826","max_tokens":4096,"model_name":"豆包·Function call模型","parameters":{"thinking_type":"enabled"},"temperature":1,"context_round":30,"response_format":"text","presence_penalty":0,"frequency_penalty":0} | 智能体绑定的模型配置信息，包括模型 ID、名称、生成参数等。 |
| folder_id | String | 752316125533*** | 智能体所属的文件夹 ID。 |
| knowledge | Object of [CommonKnowledge](#commonknowledge) | { "knowledge_infos": [ { "id": "738694398580390****", "name": "text" } ] } | 智能体绑定的知识库。 |
| variables | Array of [Variable](#variable) | - | 智能体配置的变量列表。 |
| media_config | Object of [MediaConfig](#mediaconfig) | {"is_voice_call_closed":false} | 智能体的语音通话配置，是否关闭语音通话功能。 |
| owner_user_id | String | 368567*****  | 智能体创建者的扣子用户 ID。 |
| voice_info_list | Array of [Voice](#voice) | [ { "voice_id": "7468512265134800000", "language_code": "zh" } ] | 智能体配置的音色。 |
| shortcut_commands | Array of [ShortcutCommandInfo](#shortcutcommandinfo) | [{"id":"745701083352557****","name":"示例快捷指令","command":"/sc_demo","description":"快捷指令示例","query_template":"搜索今天的新闻讯息","icon_url":"https://****","components":[{"name":"query","description":"新闻搜索关键词","type":"text","tool_parameter":"query","default_value":"","is_hide":false}],"tool":{"name":"头条新闻","type":"plugin"}}] | 智能体配置的快捷指令。 |
| workflow_info_list | Array of [WorkflowInfo](#workflowinfo) | [{"id":"746049108611037****","name":"示例工作流","description":"工作流示例","icon_url":"https://example.com/workflow_icon.png"}] | 智能体配置的工作流列表，包含工作流的 ID、名称、图标及描述信息。 |
| background_image_info | Object of [BackgroundImageInfo](#backgroundimageinfo) | \ | 智能体背景的图片配置信息，包含 Web 端和移动端的背景图 URL、主题颜色、裁剪位置及渐变效果等。 |
| default_user_input_type | String | text | 默认的用户输入方式。枚举值如下： <br>  <br> * text：打字输入。 <br> * voice：语音输入。 <br> * call：语音通话。 |
### PromptInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| prompt | String | 你是一位经验丰富的中餐大厨，能够熟练传授各类中餐的烹饪技巧，每日为大学生厨师小白教学一道经典中餐的制作方法。 | 智能体的人设与回复逻辑。长度为 0~ 20,000 个字符。默认为空。 <br> 开启前缀缓存后，不支持通过 `prompt` 参数设置提示词，需要通过 `prefix_prompt_info` 参数设置提示词。 |
| prompt_mode | String | standard | 提示词模式，用于指定智能体的人设与回复逻辑的配置方式。枚举值： <br>  <br> * `standard`（默认值）：标准模式。未开启前缀缓存时使用该模式。 <br> * `prefix`：前缀缓存模式，提示词会分为缓存提示词和非缓存提示词两部分。开启前缀缓存后需要设置为该模式。 |
| prefix_prompt_info | Object of [PrefixPromptInfo](#prefixpromptinfo) | \ | 通过 `cache_type` 或 `parameters.caching.type`开启前缀缓存后，你需要通过该参数设置**缓存提示词**（`prefix_prompt`）和 **非缓存提示词**（`dynamic_prompt`）。不支持通过 `prompt` 参数设置提示词。 |
### PrefixPromptInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| prefix_prompt | String | \ | 缓存提示词，大量重复出现的固定规则、模板框架或背景信息，用于指引大模型输出格式与风格。扣子会将其缓存并复用，大模型无需重新解析这部分固定信息。 |
| dynamic_prompt | String | \ | 非缓存提示词，动态变化的个性化信息，仅针对当前请求生效。 |
### OnboardingInfoV2
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| prologue | String | 你好，我可以为你提供最新、最有趣的科技新闻。让我们一起探索科技的世界吧！ | 智能体配置的开场白内容。 <br> 开场白中如果设置了用户名称变量`{{user_name}}`，API 场景中需要业务方自行处理，例如展示开场白时将此变量替换为业务侧的用户名称。 |
| suggested_questions | Array of String | ["你能给我推荐一些最新的科技新闻吗？","你知道最近有哪些科技趋势吗？","你能告诉我如何获得更多关于科技的信息吗？"] | 智能体配置的推荐问题列表。未开启用户问题建议时，不返回此字段。 |
### PluginInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| plugin_id | String | 730197029480849**** | 插件唯一标识。 |
| name | String | 头条新闻 | 插件名称。 |
| icon_url | String | https://example.com/plugin_icon.png | 插件的头像地址，用于展示插件的图标。 |
| description | String | 持续更新，了解最新的头条新闻和新闻文章。 | 插件的描述信息，用于说明插件的功能或用途。 |
| api_info_list | Array of [ApiInfo](#apiinfo) | [ { "api_id": "730197029480851****", "name": "getToutiaoNews", "description": "搜索新闻讯息" } ] | 插件的工具列表信息。 |
### ApiInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| api_id | String | 730197029480851**** | 插件工具的 ID。 |
| name | String | getToutiaoNews | 插件工具的名称。 |
| description | String | 搜索新闻讯息 | 插件工具的描述。 |
### ModelInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| top_k | Integer | 50 | 生成文本时，采样候选集的大小。该参数控制模型在生成每个词时考虑的候选词数量，值越小生成的文本越保守和确定，值越大生成的文本越多样和随机。 |
| top_p | Double | 1 | Top P 采样参数，用于控制生成文本时的多样性。取值范围为 `0` 到 `1`，值越小生成的文本越保守和确定，值越大生成的文本越多样和随机。 |
| model_id | String | 1706077826 | 智能体绑定的模型的 ID。 |
| max_tokens | Integer | 4096 | 模型输出的 Tokens 长度上限。 |
| model_name | String | 豆包·Function call模型 | 智能体绑定的模型名称。 |
| parameters | JSON Map | {"thinking_type": "enabled"} | 模型深度思考相关配置。开发者可以设置开启或关闭深度思考，从而灵活控制模型在交互过程中的 Token 消耗。 <br> `thinking_type` 的值可以设置为： <br>  <br> * `enabled`：（默认值）开启深度思考。智能体在与用户对话时会先输出一段思维链内容，通过逐步拆解问题、梳理逻辑，提升最终输出答案的准确性。但该模式会因额外的推理步骤消耗更多 Token。 <br>  <br> 开启深度思考后： <br>  <br> * 模型不支持 Function Call，即工具调用。 <br> * 智能体不能使用插件、触发器、变量、数据库、文件盒子、不能添加工作流和对话流。 <br> * 不支持使用插件和工作流相关的快捷指令。 <br>  <br>  <br> * `disabled`：关闭深度思考。智能体将直接生成最终答案，不再经过额外的思维链推理过程，可有效降低 Token 消耗，提升响应速度。 <br> * `auto`：当前仅**豆包·1.6·自动深度思考·多模态模型**支持该参数。启用自动模式后，模型会根据对话内容的复杂度，自动判断是否启用深度思考 <br>  <br> 当前仅如下模型支持深度思考开关配置： <br>  <br> * 豆包·1.6·自动深度思考·多模态模型 <br> * 豆包·1.6·极致速度·多模态模型 <br> * 豆包·1.5·Pro·视觉深度思考 <br> * 豆包·GUI·Agent模型 <br>  |
| temperature | Double | 1 | 生成随机性。 |
| context_round | Integer | 30 | 携带上下文轮数。 |
| response_format | String | text | 输出格式。枚举值： <br>  <br> * text：文本。 <br> * markdown：Markdown 格式。 <br> * json：json 格式。 |
| presence_penalty | Double | 0 | 重复主题惩罚。用于控制模型输出相同主题的频率。 <br> 当该值为正时，会阻止模型频繁讨论相同的主题，从而增加输出内容的多样性。 |
| frequency_penalty | Double | 0 | 重复语句惩罚。用于控制模型输出重复语句的频率。 <br> 当该值为正时，会阻止模型频繁使用相同的词汇和短语，从而增加输出内容的多样性。 |
| api_mode | String |  | 模型调用方式 |
### CommonKnowledge
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| knowledge_infos | Array of [KnowledgeInfo](#knowledgeinfo) | [ { "id": "738694398580390****", "name": "text" } ] | 智能体绑定的知识库信息。 |
### KnowledgeInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 738694398580390**** | 知识库 ID。 |
| name | String | 智能助手知识库 | 知识库名称。 |
### Variable
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| enable | Boolean | true | 是否启用该变量。 <br>  <br> * true：启用该变量。 <br> * false：未启用该变量。 |
| channel | String | custom | 变量的类型。当前只支持展示用户自定义变量（custom）。 |
| keyword | String | name | 变量名。 |
| description | String | 姓名 | 变量描述。 |
| default_value | String | - | 变量的默认值。 |
| prompt_enable | Boolean | true | 是否允许该变量被 Prompt 访问。 <br>  <br> * true：变量支持在 Prompt 中访问。 <br> * false：变量不支持在 Prompt 中访问，仅能在工作流中访问。 |
### MediaConfig
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| is_voice_call_closed | Boolean | false | 是否关闭智能体的语音通话功能。 <br>  <br> * `true`：关闭语音通话。 <br> * `false`：开启语音通话（默认）。 |
### Voice
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| voice_id | String | 7468512265134899251 | 音色的 ID。获取方法请参见[查看音色列表](https://docs.coze.cn/developer_guides/list_voices)。 |
| language_code | String | zh | 此音色的语种代号。获取方法请参见[查看音色列表](https://docs.coze.cn/developer_guides/list_voices)。 |
### ShortcutCommandInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 745701083352557**** | 快捷指令的唯一标识。 |
| name | String | 示例快捷指令 | 快捷指令的按钮名称。 |
| tool | Object of [ShortcutCommandToolInfo](#shortcutcommandtoolinfo) | {"name":"头条新闻","type":"plugin"} | 快捷指令使用的工具信息。 |
| command | String | /sc_demo | 快捷指令的指令名称。 |
| agent_id | String | 745705134267144**** | 对于多 Agent 类型的智能体，此参数返回快捷指令指定回答的节点 ID。 |
| icon_url | String | https://example.com/icon***.png | 快捷指令的图标地址。 |
| description | String | 快捷指令示例 | 快捷指令的描述。 |
| query_template | String | 搜索今天的新闻讯息 | 快捷指令的指令内容。 |
### ShortcutCommandToolInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| name | String | 头条新闻 | 快捷指令的工具名称。 |
| type | String | plugin | 工具类型。取值为： <br>  <br> * workflow：工作流。 <br> * plugin：插件。 |
### WorkflowInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 746049108611037**** | 工作流的 ID。 |
| name | String | 示例工作流 | 工作流的名称。 |
| icon_url | String | https://example.com/workflow_icon.png | 工作流的头像地址，用于展示工作流的图标。 |
| description | String | 工作流示例 | 工作流的描述。 |
### BackgroundImageInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| web_background_image | Object of [BackgroundImageDetail](#backgroundimagedetail) | - | Web 端背景图。 |
| mobile_background_image | Object of [BackgroundImageDetail](#backgroundimagedetail) | - | 移动端背景图。 |
### BackgroundImageDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| image_url | String | https://example.com/image.jpg | 背景图片的 URL 地址。 |
| theme_color | String | #FFFFFF | 背景图片的主题颜色，通常用于与图片搭配的其他元素的颜色。格式为十六进制颜色代码。  |
| canvas_position | Object of [CanvasPosition](#canvasposition) | {"top":100,"left":50,"width":300,"height":200} | 背景图片在原始图片中的位置坐标及尺寸参数，即背景图片在画布中的实际显示区域范围。包括图片顶部 / 左侧的偏移量、宽度和高度。 |
| gradient_position | Object of [GradientPosition](#gradientposition) | { "left": 0.0, "right": 800.0 } | 设置背景图渐变效果。通过指定渐变的左右边界位置，控制渐变的起始和结束点，从而实现背景图的渐变效果。 |
### CanvasPosition
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| top | Double | 100 | 裁剪区域顶部起始坐标，距原始图片顶部的像素值（px）。值越大，裁剪区域越向下移动。 <br> `top` 和 `left` 的值不能超过画布的实际尺寸 |
| left | Double | 50 | 裁剪区域左侧起始坐标，距原始图片左侧的像素值（px）。值越大，裁剪区域越向右移动。 |
| width | Double | 300 | 裁剪区域的宽度，单位为像素（px）。此值决定了裁剪区域的水平范围，必须为正数。 |
| height | Double | 200 | 裁剪区域的高度，单位为像素（px）。此值决定了裁剪区域的垂直范围，必须为正数。  |
### GradientPosition
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| left | Double | 0 | 渐变效果的左侧边界位置，单位为像素（px）。此值表示渐变从画布左侧开始的位置，值越小，渐变起始点越靠近画布左侧。 |
| right | Double | 800 | 渐变效果的右侧边界位置，单位为像素（px）。此值表示渐变在画布右侧结束的位置，值越大，渐变结束点越靠近画布右侧。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/bot/get_online_info?bot_id=73428668*****' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "bot_id": "73428668*****",
    "name": "新闻",
    "description": "每天给我推送 AI 相关的新闻。",
    "icon_url": "icon url",
    "create_time": 1715689059,
    "update_time": 1716388526,
    "version": "171638852****",
    "prompt_info": {
      "prompt": "调用getToutiaoNews工具推送最新的科技新闻。"
    },
    "onboarding_info": {
      "prologue": "你好，我可以为你提供最新、最有趣的科技新闻。让我们一起探索科技的世界吧！",
      "suggested_questions": [
        "你能给我推荐一些最新的科技新闻吗？",
        "你知道最近有哪些科技趋势吗？",
        "你能告诉我如何获得更多关于科技的信息吗？"
      ]
    },
    "bot_mode": 0,
    "model_info": {
      "model_id": "1706077826",
      "model_name": "豆包·Function call模型",
      "temperature": 1,
      "top_p": 1,
      "frequency_penalty": 0,
      "presence_penalty": 0,
      "context_round": 30,
      "max_tokens": 4096
    },
    "plugin_info_list": [
      {
        "plugin_id": "730197029480849****",
        "name": "头条新闻",
        "description": "持续更新，了解最新的头条新闻和新闻文章。",
        "icon_url": "icon url",
        "api_info_list": [
          {
            "api_id": "730197029480851****",
            "name": "getToutiaoNews",
            "description": "搜索新闻讯息"
          }
        ]
      }
    ],
    "workflow_info_list": [
      {
        "id": "746049108611037****",
        "name": "示例工作流",
        "description": "工作流示例",
        "icon_url": "https://****"
      }
    ],
    "shortcut_commands": [
      {
        "id": "745701083352557****",
        "name": "示例快捷指令",
        "command": "/sc_demo",
        "description": "快捷指令示例",
        "query_template": "搜索今天的新闻讯息",
        "icon_url": "https://****",
        "components": [
          {
            "name": "query",
            "description": "新闻搜索关键词",
            "type": "text",
            "tool_parameter": "query",
            "default_value": "",
            "is_hide": false
          }
        ],
        "tool": {
          "name": "头条新闻",
          "type": "plugin"
        }
      }
    ]
  }
}
```

## 错误码
如果成功调用扣子的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
