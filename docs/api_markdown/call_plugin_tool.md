# 调用插件工具
调用指定付费插件中的工具。
## 插件限制

* 套餐限制：仅付费版本支持此能力。
* 调用额度与 QPS 限制：请参见[通过 MCP 方式调用付费插件](https://docs.coze.cn/guides/call_plugin_mcp)。
* 插件类型：仅支持调用付费插件，包括扣子编程官方插件和三方插件。
* Token 限制：不支持渠道 Token 调用该 API。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/plugins/:plugin_id/tools/call <br> ``` <br>  |
| **权限** | `Plugin.callTool` <br> 确保调用该接口使用的访问令牌开通了 `Plugin.callTool` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用付费插件工具。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| plugin_id | String | 必选 | 75458732632938*** | 待执行的插件 ID。你可以调用[查询插件列表](https://docs.coze.cn/developer_guides/list_plugin) API 获取对应的插件 ID，即  `entity_id` 的值。 |
## Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| tool_name | String | 必选 | getToutiaoNews | 插件工具的名称。你可以通过[查询插件详情](https://docs.coze.cn/developer_guides/get_plugin) API 获取。 |
| arguments | Map[String, Any] | 必选 | { "q": "杭州秋假" } | 请根据对应插件工具的输入参数结构，输入相应参数的键值对。 <br> 你可以通过[查询插件详情](https://docs.coze.cn/developer_guides/get_plugin) API 获取插件工具的输入参数结构体（`inputSchema`）。 |
## 返回参数

|  **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data.result | map | \ | 插件工具的返回结果，具体由插件的返回结构体定义。 |
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
curl --location --request POST 'https://api.coze.cn/v1/plugins/7545873263293***/tools/call' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer pat_3UXDHN82MgQANqK****' \
--data-raw '{
    "tool_name": "getToutiaoNews",
    "arguments": {
        "q": "杭州秋假"
    }
}'
```

### 返回示例
```JSON
{
    "code": 0,
    "msg": "",
    "data": {
        "result": {
            "news": [
                {
                    "time": "2025-09-19 15:02",
                    "url": "https://toutiao.com/group/75516909358***/",
                    "summary": "2025 年秋假安排：杭州市多城区 2025 年中小学秋假时间定为 9 月 28 日 - 30 日共 3 天，9 月 27 日上课。加上国庆、中秋假期，最长可连休 11 天",
                    "media_name": "长江日报",
                    "categories": [
                        "news_society/news_law/statute",
                        "news_society/negative_energy_society",
                        "news_society/news_law",
                        "legal",
                        "news_society"
                    ],
                    "title": "杭州中小学秋假相关信息",
                    "cover": ""
                }
            ]
        }
    },
    "detail": {
        "logid": "20250926192926C176DEC5A***"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。