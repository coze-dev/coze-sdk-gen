# 查看文件详情
查看已上传的文件详情。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/files/retrieve <br> ``` <br>  |
| **权限** | `retrieveFile` <br> 确保调用该接口使用的个人令牌开通了 `retrieveFile` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查看已上传的文件详情。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.com/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| file_id | String | 必选 | 7369495981102022700 | 已上传的文件 ID。你可以通过[上传文件](https://docs.coze.cn/developer_guides/upload_files) API 的返回信息中查看文件 ID。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [File](#file) | {"id":"736949598110202****","bytes":152236,"file_name":"1120.jpeg","created_at":1715847583} | 已上传的文件信息，包含文件 ID、文件大小、文件名和上传时间等详细信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFAD653960" } | 响应详情信息。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
### File
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 736949598110202**** | 已上传的文件 ID。 |
| bytes | Long | 152236 | 文件的总字节数。 |
| file_name | String | 1120.jpeg | 文件名称。 |
| created_at | Long | 1715847583 | 文件的上传时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/files/retrieve?file_id=7369495981102022700' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
    "code": 0,
    "data": {
        "bytes": 152236,
        "created_at": 1715847583,
        "file_name": "1120.jpeg",
        "id": "736949598110202****"
    },
    "msg": ""
}
{
  "code": 0,
  "data": {
    "bytes": 152236,
    "created_at": 1715847583,
    "file_name": "1120.jpeg",
    "id": "736949598110202****"
  },
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。