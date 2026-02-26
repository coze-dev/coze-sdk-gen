# 修改工作空间基本信息
修改指定工作空间的基本信息，包括空间名称和描述。
## 接口限制

* **个人空间**不支持修改空间基本信息。
* 工作空间的名称和描述需要符合内容安全规范，不得包含涉政、涉黄等违规内容，否则扣子编程会提示 4014 错误。

## 基础信息
| **请求方式** | PUT |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workspaces/:workspace_id <br> ``` <br>  |
| **权限** | `Workspace.update` <br> 确保调用该 API 使用的访问令牌开通了 `Workspace.update` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 修改工作空间基本信息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workspace_id | String | 必选 | 736163827687053**** | 需要修改工作空间信息的工作空间 ID。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，workspace_id 为 `75814654762959***`。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| name | String | 可选 | 文档组的工作空间 | 修改后的工作空间的名称。默认为空，表示不修改。 <br> 最大长度为 50 个字符。 |
| description | String | 可选 | 文档组内部使用的工作空间。 | 修改后的工作空间的描述。默认为空，表示不修改。 <br> 最大长度为 2000 个字符。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的日志信息，用于问题排查和技术支持。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request PUT 'https://api.coze.cn/v1/workspaces/736163827687053****' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "文档组的工作空间",
    "description": "文档组内部使用的工作空间。"
}'
```

### 返回示例
```JSON
{
    "code": 0,
    "msg": "",
    "detail": {
        "logid": "20241210152726467C48D89D6DB2****"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
