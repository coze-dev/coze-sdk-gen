# 查看知识库文件上传进度
调用此接口获取扣子知识库文件的上传进度。

* 此接口支持查看所有类型知识库文件的上传进度，例如文本、图片、表格。
* 支持批量查看多个文件的进度，多个文件必须位于同一个知识库中。
* 该 API 仅支持查看扣子知识库中的文件上传进度，不支持查看火山知识库中的文件上传进度。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/datasets/:dataset_id/process <br> ``` <br>  |
| **权限** | `readDocumentProgress` <br> 确保调用该接口使用的访问令牌开通了 `readDocumentProgress` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用此接口获取知识库文件的上传进度。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| dataset_id | String | 必选 | 744258581358768**** | 扣子知识库 ID。不能填写火山知识库 ID。 <br> 在扣子编程中打开指定扣子知识库页面，页面 URL 中 `knowledge` 参数后的数字就是扣子知识库 ID。例如 `https://www.coze.cn/space/736142423532160****/knowledge/738509371792341****`，扣子知识库 ID 为 `738509371792341****`。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| document_ids | Array of String | 必选 | ["744258581358768****", "744258581358768****"] | 需要获取上传进度的文件 ID 列表。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [GetDocumentProgressOpenApiData](#getdocumentprogressopenapidata) | 参考返回示例 | 接口返回的业务信息。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### GetDocumentProgressOpenApiData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Array of [DocumentProgress](#documentprogress) | 参考返回示例 | 文件的上传进度详情。 |
### DocumentProgress
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| url | String | https://example.com/ocean-cloud-tos/FileBizType.BIZ_BOT_DATASET/217526895615***.jpg?lk3s=5ec9c6e9&x-expires=1733819246&x-signature=G9VVp2ObiCwUourS7rv44YLdh**** | 文件地址。 |
| size | Long | 171340 | 文件的大小，单位为字节。 |
| type | String | jpg | 本地文件格式，即文件后缀，例如 txt。格式支持 pdf、txt、doc、docx 类型。 |
| status | Integer | 1 | 文件的处理状态。取值包括： <br>  <br> * 0：处理中 <br> * 1：处理完毕 <br> * 9：处理失败，建议重新上传 |
| progress | Integer | 100 | 文件上传的进度。单位为百分比。 |
| document_id | String | 744667080521909*** | 文件的 ID。 |
| update_type | Integer | 0 | 在线网页是否自动更新。取值包括： <br>  <br> * 0：不自动更新 <br> * 1：自动更新 |
| document_name | String | test | 文件名称。 |
| remaining_time | Long | 0 | 预期剩余时间，单位为秒。 |
| status_descript | String |  | 失败状态的详细描述，例如切片失败时返回失败信息。 <br> 仅文档处理失败时会返回此参数。 |
| update_interval | Integer | 0 | 在线网页自动更新的频率。单位为小时。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/datasets/:dataset_id/process' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type: application/json' \
--data-raw '{
    "document_ids": [
        "7442585813587***",
        "7442585813668***"
    ]
}'
```

### 返回示例
```JSON
{
  "code": 0,
  "data": {
    "data": [
      {
        "progress": 100,
        "update_interval": 0,
        "status": 1,
        "type": "jpg",
        "update_type": 0,
        "document_id": "744667080521909***",
        "remaining_time": 0,
        "size": 171340,
        "document_name": "test1",
        "url": "https://lf3-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_DATASET/217526895615***.jpg?lk3s=5ec9c6e9&x-expires=1733819246&x-signature=G9VVp2ObiCwUourS7rv44YLdh****"
      }
    ]
  },
  "msg": "",
  "detail": {
    "logid": "20241210152726467C48D89D6***"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
