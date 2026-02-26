# 修改知识库信息
调用此接口修改扣子知识库信息。
* 此接口会全量刷新知识库的 name、file_id 和 description 配置，如果未设置这些参数，参数将恢复默认配置。
* 知识库分为扣子知识库和火山知识库，该 API 仅用于修改扣子知识库，不支持修改火山知识库，如果需要修改火山知识库的信息，请参见[修改火山知识库信息 API 文档](https://www.volcengine.com/docs/84313/1254592)。
* 仅支持修改本人为所有者的知识库信息，包括知识库名称、图标、描述等信息。
* 如需修改知识库图标，需要先调用 API [上传文件](https://docs.coze.cn/developer_guides/upload_files)，将图片文件上传至扣子编程。

## 基础信息
| **请求方式** | PUT |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/datasets/:dataset_id <br> ``` <br>  |
| **权限** | `update` <br> 确保调用该接口使用的个人令牌开通了 `update` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用此接口修改知识库信息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| dataset_id | String | 必选 | 744632974166804**** | 知识库 ID。 <br> 在扣子编程中打开指定知识库页面，页面 URL 中 `knowledge` 参数后的数字就是知识库 ID。例如 `https://www.coze.cn/space/736142423532160****/knowledge/738509371792341****`，知识库 ID 为 `738509371792341****`。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| name | String | 必选 | knowledge | 知识库名称，长度不超过 100 个字符。 |
| file_id | String | 可选 | 744667846938145**** | 知识库图标，应传入[上传文件](https://docs.coze.cn/developer_guides/upload_files)接口中获取的 file_id。 |
| description | String | 可选 | description of knowledge | 知识库描述信息。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 4000 | 状态码。 <br> 0 代表调用成功。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFAD653960" } | 响应详情信息。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request PUT 'https://api.coze.cn/v1/datasets/:dataset_id' \
--header 'Authorization: Bearer pat_qad2rHYNqKnCmRYZW4PhVRibaS***' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "AI知识库",
    "file_id": "74466784693814***",
    "description": "这是一个关于人工智能的知识库，包含机器学习、深度学习等内容。"
}'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20241210191248C8EF7607554A****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://www.coze.cn/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。