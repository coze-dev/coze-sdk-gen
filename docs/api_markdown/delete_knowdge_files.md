# 删除知识库文件
调用此接口删除扣子知识库中的文本、图片、表格等文件，支持批量删除。
* 仅知识库的所有者可以删除知识库文件。
* 知识库分为扣子知识库和火山知识库，该 API 仅用于删除扣子知识库中的文件，不支持删除火山知识库中的文件，如果需要删除火山知识库中的文件，请参见[删除火山知识库文件](https://www.volcengine.com/docs/84313/1254608)。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/open_api/knowledge/document/delete <br> ``` <br>  |
| **权限** | `deleteDocument` <br> 确保调用该接口使用的访问令牌开通了 `deleteDocument` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用此接口删除知识库中的文本、图片、表格等文件，支持批量删除。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 请求正文的方式。 |
| Agw-Js-Conv | str | 防止丢失数字类型参数的精度。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| document_ids | Array of String | 必选 | ["123"] | 待删除的知识库文件列表。数组最大长度为 100，即一次性最多可删除 100 个文件。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码。 <br> 0 代表调用成功。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息，主要用于调试和问题排查。如果遇到异常报错场景，且反复重试仍然报错，可以根据此字段中的 `logid`联系扣子团队获取帮助。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/open_api/knowledge/document/delete' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
--header 'Agw-Js-Conv: str' \
--data-raw '{
    "document_ids": [
        "123"
    ]
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