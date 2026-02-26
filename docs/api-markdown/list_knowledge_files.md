# 查看知识库文件列表
调用接口查看指定扣子知识库的文件列表，即文档、表格或图像列表。
仅支持通过本 API 查看扣子知识库中的文件列表，不支持查看火山知识库中的文件列表。查看火山知识库中的文件列表请参见[查看火山知识库的文件列表](https://www.volcengine.com/docs/84313/1254621)。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/open_api/knowledge/document/list <br> ``` <br>  |
| **权限** | `listDocument` <br> 确保调用该接口使用的个人令牌开通了 `listDocument` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用接口查看指定扣子知识库的内容列表，即文件、表格或图像列表。 |
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
| dataset_id | String | 必选 | 75034727177617*** | 待查看文件的扣子知识库 ID。 <br> 仅支持扣子知识库，不支持火山知识库。 <br> 在扣子编程中打开指定知识库页面，页面 URL 中 `knowledge` 参数后的数字就是知识库 ID。例如 `https://www.coze.cn/space/736142423532160****/knowledge/738509371792341****`，知识库 ID 为 `738509371792341****`。 |
| page | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即从第一页数据开始返回。 |
| size | Integer | 可选 | 20 | 分页大小。默认为 10，即每页返回 10 条数据。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码。 <br> 0 代表调用成功。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| document_infos | Array of [DocumentInfo](#documentinfo) | [ ] | 知识库文件列表。详细说明可参考 [DocumentInfo object](https://docs.coze.cn/developer_guides/list_knowledge_files#e638fe05)。 |
| total | Integer | 20 | 指定知识库中的文件总数。 |
| detail | Object of [ResponseDetail](#responsedetail) | "20241210152726467C48D89D6DB2****" | 返回详情，包含本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 `logid`及错误码联系扣子团队获取帮助。 |
### DocumentInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| char_count | Integer | 4 | 文件内容的总字符数量。 |
| chunk_strategy | Object of [ChunkStrategy](#chunkstrategy) | { "chunk_type": 1, "max_tokens": 800, "remove_extra_spaces": false, "remove_urls_emails": false, "separator": "#" } | 分段规则。 |
| create_time | Integer | 1719907964 | 文件的上传时间，格式为 10 位的 Unixtime 时间戳。 |
| document_id | String | 738694205603010**** | 文件的 ID。 |
| format_type | Integer | 0 | 文件的格式类型。取值包括： <br>  <br> * 0：文档类型，例如 txt 、pdf 、在线网页等格式均属于文档类型。 <br> * 1：表格类型，例如 xls 表格等格式属于表格类型。 <br> * 2：照片类型，例如 png 图片等格式属于照片类型。 |
| hit_count | Integer | 0 | 被对话命中的次数。 |
| name | String | Coze.pdf | 文件的名称。 |
| size | Integer | 14164 | 文件的大小，单位为字节。 |
| slice_count | Integer | 1 | 文件的分段数量。 |
| source_type | Integer | 0 | 文件的上传方式。取值包括： <br>  <br> * 0：上传本地文件。 <br> * 1：上传在线网页。 |
| status | Integer | 1 | 文件的处理状态。取值包括： <br>  <br> * 0：处理中 <br> * 1：处理完毕 <br> * 9：处理失败，建议重新上传 |
| type | String | pdf | 本地文件格式，即文件后缀，例如 txt。格式支持 pdf、txt、doc、docx 类型。 |
| update_interval | Integer | 0 | 在线网页自动更新的频率。单位为小时。 |
| update_time | Integer | 1719907969 | 文件的最近一次修改时间，格式为 10 位的 Unixtime 时间戳。 |
| update_type | Integer | 0 | 在线网页是否自动更新。取值包括： <br>  <br> * 0：不自动更新 <br> * 1：自动更新 |
| tos_uri | String | FileBizType.BIZ_BOT_DATASET/847077809337655_1727579972975689529_0ytrdq****.docx | 上传的本地文档的唯一标识。 |
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
curl --location --request POST 'https://api.coze.cn/open_api/knowledge/document/list' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
--header 'Agw-Js-Conv: str' \
--data-raw '{
    "dataset_id": "736356924530694****",
    "page": 0,
    "size": 10
}'
```

### 返回示例
```JSON
{
  "code": 0,
  "document_infos": [
    {
      "char_count": 4,
      "chunk_strategy": {
        "chunk_type": 0,
        "max_tokens": 0,
        "remove_extra_spaces": false,
        "remove_urls_emails": false,
        "separator": ""
      },
      "create_time": 1719476392,
      "document_id": "738508308097900****",
      "format_type": 0,
      "hit_count": 0,
      "name": "小猫的阳光午睡.pdf.pdf",
      "size": 30142,
      "slice_count": 1,
      "source_type": 0,
      "status": 1,
      "type": "pdf",
      "update_interval": 0,
      "update_time": 1719476430,
      "update_type": 0
    }
  ],
  "msg": "",
  "total": 1
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
