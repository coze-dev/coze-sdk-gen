# 批量查询插件信息
批量查询插件的详细信息。
## 接口描述
你可以调用本 API 批量查询插件的详细信息，支持查询官方插件、三方插件和资源库插件。返回信息包括插件工具列表（工具名称、描述、输入参数和返回参数），付费插件还支持返回 MCP Server 的配置信息，以便在 Trae 等外部平台配置 MCP 服务，并调用插件工具。
## 接口限制

* 单次请求最多支持查询 20 个插件。
* 不支持通过渠道 Token 调用该 API。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/plugins/mget <br> ``` <br>  |
| **权限** | `Plugin.getPlugin` <br> 确保调用该接口使用的访问令牌开通了 `Plugin.getPlugin` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 批量查询插件信息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| ids | Array of String | 必选 | 7506817938062819***,742665585406735*** | 待查询插件详情的插件 ID 列表，多个 ID 需以英文逗号分隔。支持查询官方插件、三方插件和资源库插件。 <br> 获取插件 ID 的方法如下： <br>  <br> * 调用[查询插件列表](https://docs.coze.cn/developer_guides/list_plugin) API 获取对应的插件 ID，即 `entity_id` 的值。 <br> * 资源库自定义插件：在插件编辑页面的 URL 中获取插件 ID。例如在 URL `https://www.coze.cn/space/74982048832804***/plugin/75331489586134***`中，`plugin` 后面的数字即为插件 ID。 <br>  <br> * 单次请求最多支持输入 20 个插件ID。 <br> * 不支持通过插件商店 URL 中的插件商品 ID 查询插件详情。 <br>  |
## 返回参数

| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data.tools | Array of [Tools](#Tools) | \ | 插件工具的详细信息，包括名称、描述、输入参数和返回参数。 |
| data.mcp_json | String | { "mcpServers": { "coze_plugin_***": { "url": "https://mcp.coze.cn/v1/plugins/75458732632941***", "headers": { "Authorization": "Bearer ${COZE_API_TOKEN}" } } } } | MCP Server 的配置信息，包含 MCP 服务器信息、插件 URL 和访问令牌。 <br> `Authorization` 中的访问令牌默认生成的是临时访问凭证，有效期为 1 天。 <br> 如需使用长期凭证，在 Trae 等支持 MCP Server 的平台配置 MCP 信息时，需要手动替换为你在扣子编程中生成的访问令牌。支持个人访问令牌（PAT）、服务访问令牌（SAT）和 OAuth 鉴权，具体参考[获取访问令牌](https://docs.coze.cn/developer_guides/preparation#dbadd15c)。 <br>  |
| data.name_for_model | String | yuyinhecheng | 插件在模型内部使用的唯一标识符，用于AI模型在调用插件时快速识别和定位目标插件。 |
| data.icon_url | String | https://example.com/ocean-cloud-tos/plugin_icon/example.jpeg?lk3s=cd508e2b&x-expires=1763780763&x-signature=W0osLsZ*** | 插件图标的 URL 链接。 |
| data.is_call_available | Boolean | true | 标识该插件当前是否可被调用。 <br>  <br> * `true`：插件可正常调用。 <br> * `false`：暂时无法调用，三方付费插件未签署协议。 |
| data.created_at | Number | 1759044695 | 插件创建时间，以 Unix 时间戳格式表示，单位为秒。 |
| data.updated_at | Number | 1761188383 | 插件最后更新时间，以 Unix 时间戳格式表示，单位为秒。 |
### Tools
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| name | String | asr_lite | 插件工具的名称。 |
| description | String | 准确地将语音内容转写成文字。 | 插件工具的描述。 |
| inputSchema | Object | { "required": [ "audio_url" ], "properties": { "audio_url": { "description": "音频文件地址，插件会尝试识别文件格式，对于无后缀的 URL 可能识别不到。文件类型支持 raw / wav / mp3 / ogg", "type": "string" } }, "type": "object" } | 由插件开发者定义的插件工具的输入参数的数据结构，包含参数名称、参数说明、必填项、格式约束等。 |
| outputSchema | Object | { "properties": { "code": { "description": "运行状态。0 为成功，否则为报错码", "type": "number" }, "data": { "properties": { "text": { "description": "识别出的文本", "type": "string" } }, "type": "object" }, "log_id": { "type": "string" }, "msg": { "type": "string" } }, "type": "object" } | 由插件开发者定义的插件工具的输出参数的数据结构，用于描述当前插件工具调用成功后，返回数据的结构，如字段名称、类型、嵌套关系等。 |

| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/plugins/mget?ids=7506817938062819***,742665585406735***' \
--header 'Authorization: Bearer pat_O******' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
    "code": 0,
    "msg": "",
    "data": {
        "items": [
            {
                "tools": [
                    {
                        "description": "相比传统模型有更高的准确率；对多语种多方言/口音/噪声和背景人声更准确；能根据上下文、用户输入、背景信息输入有更多理解。按照时长计费，对应计费项“大模型录音文件识别时长”。计费项&免费额度说明：https://docs.coze.cn/coze_pro/asr_tts_fee",
                        "inputSchema": {
                            "required": [
                                "audio_url"
                            ],
                            "properties": {
                                "audio_url": {
                                    "description": "音频文件地址，插件会尝试识别文件格式，对于无后缀的 URL 可能识别不到。文件类型支持 raw / wav / mp3 / ogg",
                                    "type": "string"
                                }
                            },
                            "type": "object"
                        },
                        "name": "asr_llm",
                        "outputSchema": {
                            "properties": {
                                "code": {
                                    "description": "运行状态。0 为成功，否则为报错码",
                                    "type": "number"
                                },
                                "data": {
                                    "properties": {
                                        "text": {
                                            "description": "识别出的文本",
                                            "type": "string"
                                        }
                                    },
                                    "type": "object"
                                },
                                "log_id": {
                                    "type": "string"
                                },
                                "msg": {
                                    "type": "string"
                                }
                            },
                            "type": "object"
                        },
                        "tool_id": "75068179380628***"
                    }
                ],
                "mcp_json": "{\"mcpServers\":{\"coze_plugin_damoxingyuyinshibie\":{\"url\":\"https://mcp.coze.cn/v1/plugins/7506817938062819355\",\"headers\":{\"Authorization\":\"Bearer ${COZE_API_TOKEN}\"}}}}",
                "plugin_id": "75068179380628***",
                "name": "大模型语音识别",
                "name_for_model": "damoxingyuyinshibie",
                "description": "相比传统模型有更高的准确率；对多语种多方言/口音/噪声和背景人声更准确；能根据上下文、用户输入、背景信息输入有更多理解。按照时长计费，对应计费项“大模型录音文件识别时长”。计费项&免费额度说明：https://docs.coze.cn/coze_pro/asr_tts_fee",
                "icon_url": "https://example.com/ocean-cloud-tos/plugin_icon/example.jpeg?lk3s=cd508e2b&x-expires=1763780763&x-signature=W0osLsZ***",
                "is_call_available": true,
                "created_at": 1754644595,
                "updated_at": 1761197219
            },
            {
                "tools": [
                    {
                        "description": "根据音色和文本合成音频。按照字符数计费，计费项&免费额度说明：https://docs.coze.cn/coze_pro/asr_tts_fee。使用资源库音色时，计费项为复刻音色文字转语音字数；使用预设音色时，模型为大模型对应系统音色文字转语音字数，小模型对应小模型合成次数 。",
                        "inputSchema": {
                            "required": [
                                "text"
                            ],
                            "properties": {
                                "voice_id": {
                                    "description": "扣子编程音色 ID，支持选择扣子编程系统预置的音色或资源库中复刻的音色。 可以在操作页面直接选择音色，或通过系统音色列表https://docs.coze.cn/dev_how_to_guides/sys_voice查看音色 ID。",
                                    "format": "voice_id",
                                    "type": "string"
                                },
                                "emotion": {
                                    "description": "语音情感",
                                    "type": "string"
                                },
                                "emotion_scale": {
                                    "description": "调用emotion设置情感参数后可使用emotion_scale进一步设置情绪值，范围1~5，不设置时默认值为4。 注：理论上情绪值越大，情感越明显。但情绪值1~5实际为非线性增长，可能存在超过某个值后，情绪增加不明显，例如设置3和5时情绪值可能接近。",
                                    "type": "number"
                                },
                                "language": {
                                    "description": "音色的语种，非必填，所有中文音色支持中英文混合场景。可参考系统音色列表https://docs.coze.cn/dev_how_to_guides/sys_voice查看各音色支持的语种。",
                                    "type": "string"
                                },
                                "speaker_id": {
                                    "description": "音色ID，默认为爽快思思/Skye。详细音色列表参考 https://docs.coze.cn/guides/text_to_speech_plugin, default value is 爽快思思/Skye",
                                    "type": "string"
                                },
                                "speed_ratio": {
                                    "description": "语速，范围是[0.2,3]，默认为1，通常保留一位小数即可, default value is 1",
                                    "type": "number"
                                },
                                "text": {
                                    "description": "要合成音频的文本内容",
                                    "type": "string"
                                }
                            },
                            "type": "object"
                        },
                        "name": "speech_synthesis",
                        "outputSchema": {
                            "properties": {
                                "msg": {
                                    "type": "string"
                                },
                                "code": {
                                    "type": "number"
                                },
                                "data": {
                                    "properties": {
                                        "duration": {
                                            "description": "音频时长，单位是s",
                                            "type": "number"
                                        },
                                        "link": {
                                            "type": "string"
                                        }
                                    },
                                    "type": "object"
                                },
                                "log_id": {
                                    "type": "string"
                                }
                            },
                            "type": "object"
                        },
                        "tool_id": "7426655854067***"
                    }
                ],
                "mcp_json": "{\"mcpServers\":{\"coze_plugin_yuyinhecheng\":{\"url\":\"https://mcp.coze.cn/v1/plugins/7426655854067351562\",\"headers\":{\"Authorization\":\"Bearer ${COZE_API_TOKEN}\"}}}}",
                "plugin_id": "742665585406***",
                "name": "语音合成",
                "name_for_model": "yuyinhecheng",
                "description": "根据音色和文本合成音频。按照字符数计费，计费项&免费额度说明：https://docs.coze.cn/coze_pro/asr_tts_fee。使用资源库音色时，计费项为复刻音色文字转语音字数；使用预设音色时，模型为大模型对应系统音色文字转语音字数，小模型对应小模型合成次数 。",
                "icon_url": "https://example.com/ocean-cloud-tos/plugin_icon/example.jpeg?lk3s=cd508e2b&x-expires=1763780763&x-signature=W0osLsZ***",
                "is_call_available": true,
                "created_at": 1759044695,
                "updated_at": 1761199709
            }
        ]
    },
    "detail": {
        "logid": "202510231410590AC026226***"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
