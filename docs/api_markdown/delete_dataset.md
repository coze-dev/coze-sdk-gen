# 删除知识库
调用此接口删除扣子知识库。
* 空间管理员可以删除团队中所有知识库，其他成员只能删除本人为所有者的知识库，
* 删除知识库时，会同时删除知识库中已上传的所有文件，且绑定此知识库的智能体会自动解绑。
* 此 API 仅支持删除扣子知识库，不适用于火山知识库。如需删除火山知识库，请参考[删除火山知识库](https://www.volcengine.com/docs/84313/1254607)。

## 基础信息
| **请求方式** | DELETE |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/datasets/:dataset_id <br> ``` <br>  |
| **权限** | `delete` <br> 确保调用该接口使用的个人令牌开通了 `delete` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用此接口以删除本人持有的知识库信息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| dataset_id | String | 必选 | 744632974166804**** | 扣子知识库 ID。 <br> 在扣子编程中打开指定知识库页面，页面 URL 中 `knowledge` 参数后的数字就是知识库 ID。例如 `https://www.coze.cn/space/736142423532160****/knowledge/738509371792341****`，知识库 ID 为 `738509371792341****`。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码。 <br> 0 代表调用成功。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | - | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request DELETE  'https://api.coze.cn/v1/datasets/744668935865830****' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "20241210191248C8EF7607554A***"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://www.coze.cn/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。