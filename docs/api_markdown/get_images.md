# 查看知识库图片列表
调用此接口查看图片类知识库中图片的详细信息。
查看图片时，支持通过图片的标注进行筛选。
## 接口限制
此 API 仅支持查看扣子知识库中的图片详细信息，不适用于火山知识库。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/datasets/:dataset_id/images <br> ``` <br>  |
| **权限** | `listPhoto` <br> 确保调用该接口使用的访问令牌开通了 `listPhoto` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用此接口查看图片类知识库中图片的详细信息。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| dataset_id | String | 必选 | 744632974166804**** | 知识库 ID。 <br> 在扣子编程中打开指定知识库页面，页面 URL 中 `knowledge` 参数后的数字就是知识库 ID。例如 `https://www.coze.cn/space/736142423532160****/knowledge/738509371792341****`，知识库 ID 为 `738509371792341****`。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_num | Integer | 可选 | 1 | 查询结果分页展示时，此参数用于设置查看的页码。最小值为 1，默认为 1。 |
| page_size | Integer | 可选 | 5 | 查询结果分页展示时，此参数用于设置每页返回的数据量。取值范围为 1~299，默认为 10。 |
| keyword | String | 可选 | 小猫 | 对图片描述进行搜索时，搜索的关键字。 |
| has_caption | Boolean | 可选 | true | 图片是否已设置了描述信息。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [ListPhotoOpenApiData](#listphotoopenapidata) | 参考返回示例 | 返回内容 |
| code | Long | 0 | 状态码。 <br> 0 代表调用成功。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 `logid`及错误码联系扣子团队获取帮助。 |
### ListPhotoOpenApiData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| photo_infos | Array of [PhotoInfo](#photoinfo) | 参考返回示例 | 图片的详细信息。 |
| total_count | Integer | 1 | 符合查询条件的图片总数量。 |
### PhotoInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| url | String | https://lf26-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_DATASET/217526895615****.jpg?lk3s=5ec9c6e9&x-expires=1733821017&x-signature=KJbBgbcwSZYYh5Uf5gISU5HAPB**** | 图片链接。 |
| name | String | test2 | 图片名。 |
| size | Integer | 171340 | 图片大小，单位为字节。 |
| type | String | jpg | 文件格式，即文件后缀，例如 jpg、png。 |
| status | Integer | 1 | 文件的状态。取值包括： <br>  <br> * 0：处理中 <br> * 1：处理完毕 <br> * 9：处理失败，建议重新上传 |
| caption | String | 小猫 | 图片描述信息。 |
| creator_id | String | 217526895615**** | 创建人的扣子 UID。 |
| create_time | Integer | 1733815232 | 图片的上传时间，格式为 10 位的 Unixtime 时间戳。 |
| document_id | String | 744667080521911**** | 图片的 ID。 |
| source_type | Integer | 5 | 上传方式。取值包括： <br>  <br> * 0：上传本地文件。 <br> * 1：上传在线网页。 <br> * 5：上传 file_id。 |
| update_time | Integer | 1733817093 | 更新时间，格式为 10 位的 Unixtime 时间戳。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/datasets/744632974166804****/images?page_num=1&page_size=5&keyword=小猫&has_caption=true' \
--header 'Authorization: Bearer pat_hfwkehfncaf****' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "detail": {
    "logid": "2024121015565741D31B341C9C0***"
  },
  "data": {
    "total_count": 1,
    "photo_infos": [
      {
        "name": "test2",
        "status": 1,
        "creator_id": "217526895615****",
        "create_time": 1733815232,
        "update_time": 1733817093,
        "size": 171340,
        "url": "https://lf26-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_DATASET/217526895615****.jpg?lk3s=5ec9c6e9&x-expires=1733821017&x-signature=KJbBgbcwSZYYh5Uf5gISU5HAPB****",
        "caption": "小猫",
        "type": "jpg",
        "source_type": 5,
        "document_id": "744667080521911****"
      }
    ]
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://www.coze.cn/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。