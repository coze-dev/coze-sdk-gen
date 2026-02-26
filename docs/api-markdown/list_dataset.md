# 查看知识库列表
调用此接口查看指定空间资源库下的全部知识库。
此接口可查看工作空间下，空间资源库中的全部知识库，包括扣子知识库和火山知识库，无论知识库是否归本人所有。
* 暂不支持通过 API 查看低代码应用中的知识库。
* 暂不支持通过该 API 查看火山知识库中的文件列表等详细信息。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/datasets <br> ``` <br>  |
| **权限** | `listKnowledge` <br> 确保调用该接口使用的个人令牌开通了 `listKnowledge` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用此接口查看指定空间下的全部知识库。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| space_id | String | 必选 | 731121948439879**** | 知识库所在的空间的 Space ID。Space ID 是空间的唯一标识。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，工作空间 ID 为 `75814654762959***`。 |
| name | String | 可选 | 知识库 | 知识库名称，支持模糊搜索。 |
| format_type | Integer | 可选 | 2 | 类型 |
| page_num | Integer | 可选 | 1 | 查询结果分页展示时，此参数用于设置查看的页码。最小值为 1，默认为 1。 |
| page_size | Integer | 可选 | 5 | 查询结果分页展示时，此参数用于设置每页返回的数据量。取值范围为 1~300，默认为 10。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [ListDatasetOpenApiData](#listdatasetopenapidata) | 参考请求示例部分 | 接口响应的业务信息。 |
| code | Long | 4000 | 状态码。 <br> 0 代表调用成功。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFAD653960" } | 响应详情信息。 |
### ListDatasetOpenApiData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| total_count | Integer | 1 | 空间中的知识库总数量。 |
| dataset_list | Array of [Dataset](#dataset) | 参考请求示例部分 | 知识库详情。 |
### Dataset
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| name | String | openapi | 知识库名称。 |
| status | Integer | 1 | 知识库状态，取值包括： <br>  <br> * 1：启用中 <br> * 3：未启用 |
| can_edit | Boolean | true | 当前用户是否为该知识库的所有者。 |
| icon_uri | String | FileBizType.BIZ_DATASET_ICON/217526895615****.jpg | 知识库图标的 uri。 |
| icon_url | String | https://lf3-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_DATASET_ICON/217526895615****.jpg?lk3s=5ec9c6e9&x-expires=1733821937&x-signature=U6X%2BhLXnRk8%2FHr1xP7wiMJ3IE**** | 知识库图标的 URL。 |
| space_id | String | 731121948439879**** | 知识库所在空间的空间 ID。 |
| doc_count | Integer | 2 | 知识库中的文件数量。 |
| file_list | Array of String | ["8f6a939c9aad434fa81624556abf8a62****.png","9k7a939c9aad434fa81624556abf8a62****.png"] | 知识库中的文件列表。 |
| hit_count | Integer | 0 | 知识库命中总次数。 |
| avatar_url | String | https://p6-passport.byteacctimg.com/img/user-avatar/assets/e7b19241fb224cea967****.png~300x300.image | 知识库创建者的头像 url。 |
| creator_id | String | 217526895615**** | 知识库创建者的扣子用户 UID。 |
| dataset_id | String | 744668935865830**** | 知识库 ID。 |
| create_time | Integer | 1733817948 | 知识库创建时间，秒级时间戳。 |
| description | String | description | 知识库描述信息。 |
| format_type | Integer | 2 | 知识库类型，包括： <br>  <br> * 0：文本类型 <br> * 1：表格类型 <br> * 2：图片类型 |
| slice_count | Integer | 1 | 知识库分段总数。 |
| update_time | Integer | 1733817948 | 知识库的更新时间，秒级时间戳。 |
| creator_name | String | your name | 知识库创建者的用户名。 |
| all_file_size | Long | 0 | 知识库中已存文件的总大小。 |
| bot_used_count | Integer | 0 | 知识库已绑定的智能体数量。 |
| chunk_strategy | Object of [ChunkStrategy](#chunkstrategy) | {“chunk_type”: 0} | 知识库的切片规则。 |
| failed_file_list | Array of String | ["8f6a939c9aad434fa81624556abf8a62****.png","9k7a939c9aad434fa81624556abf8a62****.png"] | 处理失败的文件列表。 |
| processing_file_list | Array of String | ["8f6a939c9aad434fa81624556abf8a62****.png","9k7a939c9aad434fa81624556abf8a62****.png"] | 处理中的文件名。 |
| processing_file_id_list | Array of String | ["744779282568390****"] | 处理中的文件 ID。 |
### ChunkStrategy
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| chunk_type | Integer | 0 | 分段设置。取值包括： <br>  <br> * 0：自动分段与清洗。采用扣子编程预置规则进行数据分段与处理。 <br> * 1：自定义。此时需要通过 separator、max_tokens、remove_extra_spaces 和 remove_urls_emails 分段规则细节。 |
| separator | String | # | 分段标识符。 <br> 在 chunk_type=1 时必选。 |
| max_tokens | Long | 800 | 最大分段长度，取值范围为 100~2000。 <br> 在 chunk_type=1 时必选。 |
| remove_extra_spaces | Boolean | true | 是否自动过滤连续的空格、换行符和制表符。取值包括： <br>  <br> * true：自动过滤 <br> * false：（默认）不自动过滤 <br>  <br> 在 chunk_type=1 时生效。 |
| remove_urls_emails | Boolean | true | 是否自动过滤所有 URL 和电子邮箱地址。取值包括： <br>  <br> * true：自动过滤 <br> * false：（默认）不自动过滤 <br>  <br> 在 chunk_type=1 时生效。 |
| caption_type | Integer | 0 | 图片知识库的标注方式： <br>  <br> * 0：（默认）系统自动标注描述信息 <br> * 1：手工标注。上传图片后需要再次调用 API [更新知识库图片描述](https://docs.coze.cn/developer_guides/update_image_caption)来手动设置标注。 <br>  <br> 在空的知识库中首次上传图片时，需要手工设置 `caption_type` 参数的值，否则会报错。 <br>  |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/datasets?space_id=731121948439879****&name=知识库&format_type=&page_num=1&page_size=5' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "code": 0,
  "data": {
    "total_count": 1,
    "dataset_list": [
      {
        "hit_count": 0,
        "doc_count": 0,
        "status": 1,
        "icon_url": "https://lf3-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_DATASET_ICON/217526895615****.jpg?lk3s=5ec9c6e9&x-expires=1733821937&x-signature=U6X%2BhLXnRk8%2FHr1xP7wiMJ3IE****",
        "creator_name": "xxx",
        "avatar_url": "https://p6-passport.byteacctimg.com/img/user-avatar/assets/e7b19241fb224cea967****.png~300x300.image",
        "can_edit": true,
        "space_id": "731121948439879****",
        "failed_file_list": [],
        "processing_file_id_list": [],
        "description": "openapi",
        "chunk_strategy": {},
        "create_time": 1733817948,
        "slice_count": 0,
        "name": "openapi_img3",
        "format_type": 2,
        "project_id": "",
        "all_file_size": "0",
        "update_time": 1733817948,
        "creator_id": "217526895615****",
        "file_list": [],
        "processing_file_list": [],
        "icon_uri": "FileBizType.BIZ_DATASET_ICON/217526895615****.jpg",
        "dataset_id": "744668935865830****",
        "bot_used_count": 0
      }
    ]
  },
  "msg": "",
  "detail": {
    "logid": "20241210161217C90C9ABB86428***"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
