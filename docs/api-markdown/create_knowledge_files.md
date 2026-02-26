# 创建知识库文件
调用此 API 向指定的扣子知识库上传文件。
## 接口说明

* 通过此 API，你可以向文本或图片知识库中上传文件。
* 暂不支持通过 API 创建表格知识库。
* 本 API 仅适用于扣子知识库的文件上传操作，不适用于火山知识库上传文件。火山知识库上传文件请参见[火山知识库上传文件](https://www.volcengine.com/docs/84313/1254624)。
* 上传图片到图片知识库时，可以通过 `caption_type` 参数设置系统标注或手动标注，如果选择手动标注，还需要调用 [更新知识库图片描述](https://docs.coze.cn/developer_guides/update_image_caption) API 手动设置标注。
* 支持的上传方式如下：

| **上传方式** | **文本知识库** | **图片知识库** |
| --- | --- | --- |
| 通过 Base 64 上传本地文件 | ✅ <br> 格式支持 pdf、txt、doc、docx 类型。 | ❌ |
| 上传在线网页 | ✅ | ❌ |
| 通过 [上传文件](https://docs.coze.cn/developer_guides/upload_files) 上传 | ❌ | ✅ |
## 注意事项

* 每次最多可上传 10 个文件。
* 必须上传和知识库类型匹配的文件，例如 txt 等文档类型的文件只能上传到文档知识库中，不可上传到表格知识库中。
* 每个请求只能选择一种上传方式，不支持同时上传在线网页和本地文档。
* 仅知识库的所有者可以向知识库中上传文件。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/open_api/knowledge/document/create <br> ``` <br>  |
| **权限** | `createDocument` <br> 确保调用该接口使用的服务令牌开通了 `createDocument` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用此接口向指定知识库中上传文件。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 请求正文的方式。 |
| Agw-Js-Conv | str | 防止丢失数字类型参数的精度。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| dataset_id | String | 必选 | 736356924530694**** | 扣子知识库 ID。 <br> 在扣子编程中打开指定知识库页面，页面 URL 中 `knowledge` 参数后的数字就是知识库 ID。例如 `https://www.coze.cn/space/736142423532160****/knowledge/738509371792341****`，知识库 ID 为 `738509371792341****`。 |
| document_bases | Array of [DocumentBase](#documentbase) | 必选 | [ { "name": "Coze.pdf", "source_info": { "file_base64": "5rWL6K+V5LiA5LiL5ZOm", "file_type": "pdf" } } ] | 待上传文件的元数据信息。数组最大长度为 10，即每次最多上传 10 个文件。 <br> 支持的上传方式如下： <br>  <br> * 文本知识库： <br>    * 通过 Base64 上传本地文件。 <br>    * 上传在线网页。 <br> * 图片知识库：通过[上传文件](https://docs.coze.cn/developer_guides/upload_files)API 获取的 file_id 上传图片。 |
| chunk_strategy | Object of [ChunkStrategy](#chunkstrategy) | 必选 | { <br> "separator": "\n\n", <br> "max_tokens": 800, <br> "remove_extra_spaces": false, <br> "remove_urls_emails": false, <br> "chunk_type": 1 <br> } | 分段规则。详见 [ChunkStrategy object](https://docs.coze.cn/developer_guides/create_knowledge_files#fb4d001d) <br> 每次都需要传入 <br>  |
| format_type | Integer | 必选 | 2 | 知识库类型。取值包括： <br>  <br> * 0：文本类型 <br> * 2：图片类型 |
### DocumentBase
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| name | String | 必选 | Coze.pdf | 文件名称。 |
| source_info | Object of [SourceInfo](#sourceinfo) | 必选 | { "file_base64": "5rWL6K+V5LiA5LiL5ZOm", "file_type": "pdf" } | 文件的元数据信息。 <br> 不同的上传方式需要指定不同参数： <br>  <br> * **本地文件**：Base64 形式上传，需要设置参数 file_base64、file_type。 <br> * **在线网页**：需要设置参数 web_url、document_source。 <br> * **file_id**：需要设置参数 source_file_id、document_source。其中 source_file_id 为 [上传文件](https://docs.coze.cn/developer_guides/upload_files)API 的响应参数 file_id。 |
| update_rule | Object of [UpdateRule](#updaterule) | 可选 | {"update_type": 0} | 在线网页的更新策略。默认不自动更新。仅上传在线网页时需要设置。 |
| caption | String | 可选 | 可爱的 | 当知识库类型为图片类型时，通过该字段设置图片描述。 |
### SourceInfo
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| file_base64 | String | 可选 | 5rWL6K+V5LiA5LiL5ZOm | 本地文件的 Base64 编码。 <br> 上传本地文件时必选。 <br>  |
| file_type | String | 可选 | pdf | 本地文件格式，即文件后缀，例如 txt。格式支持 pdf、txt、doc、docx 类型。 <br> 上传的文件类型应与知识库类型匹配，例如 txt 文件只能上传到文档类型的知识库中。 <br> 上传本地文件时必选。 <br>  |
| web_url | String | 可选 | https://docs.coze.cn/developer_guides/upload_files | 网页的 URL 地址。 <br> 上传在线网页时必选。 <br>  |
| document_source | Integer | 可选 | 1 | 文件的上传方式。支持设置为： <br>  <br> * 0：本地文件上传。 <br> * 1：上传在线网页。 <br> * 5：通过 [上传文件](https://docs.coze.cn/developer_guides/upload_files) 上传图片。 |
| source_file_id | String | 可选 | 736949598110202**** | 通过[上传文件](https://docs.coze.cn/developer_guides/upload_files)接口获取的文件 ID。 |
### UpdateRule
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| update_type | Integer | 可选 | 1 | 在线网页是否自动更新。取值包括： <br>  <br> * 0：（默认）不自动更新 <br> * 1：自动更新 |
| update_interval | Integer | 可选 | 24 | 在线网页自动更新的频率。单位为小时，最小值为 24。 |
### ChunkStrategy
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| chunk_type | Integer | 可选 | 0 | 分段设置。取值包括： <br>  <br> * 0：自动分段与清洗。采用扣子编程预置规则进行数据分段与处理。 <br> * 1：自定义。此时需要通过 separator、max_tokens、remove_extra_spaces 和 remove_urls_emails 分段规则细节。 |
| separator | String | 可选 | # | 分段标识符。 <br> 在 chunk_type=1 时必选。 |
| max_tokens | Long | 可选 | 800 | 最大分段长度，取值范围为 100~2000。 <br> 在 chunk_type=1 时必选。 |
| remove_extra_spaces | Boolean | 可选 | true | 是否自动过滤连续的空格、换行符和制表符。取值包括： <br>  <br> * true：自动过滤 <br> * false：（默认）不自动过滤 <br>  <br> 在 chunk_type=1 时生效。 |
| remove_urls_emails | Boolean | 可选 | true | 是否自动过滤所有 URL 和电子邮箱地址。取值包括： <br>  <br> * true：自动过滤 <br> * false：（默认）不自动过滤 <br>  <br> 在 chunk_type=1 时生效。 |
| caption_type | Integer | 可选 | 0 | 图片知识库的标注方式： <br>  <br> * 0：（默认）系统自动标注描述信息 <br> * 1：手工标注。上传图片后需要再次调用 API [更新知识库图片描述](https://docs.coze.cn/developer_guides/update_image_caption)来手动设置标注。 <br>  <br> 在空的知识库中首次上传图片时，需要手工设置 `caption_type` 参数的值，否则会报错。 <br>  |

## 返回参数

| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| document_infos | Array of [DocumentInfo](#documentinfo) | - | 已上传文件的基本信息。 |
| code | Long | 0 | 状态码。 <br> 0 代表调用成功。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFAD653960" } | 响应详情信息。 |
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

<div type="doc-tabs">
<div type="tab-item" title="上传本地文件" key="3BbTJs4uIluEViM2j5gZf">

```JSON
curl --location --request POST 'https://api.coze.cn/open_api/knowledge/document/create' \
--data-raw '{
    "dataset_id": "736356924530694****",
    "document_bases": [
        {
            "name": "Coze.pdf",
            "source_info": {
                "file_base64": "5rWL6K+V5LiA5LiL5ZOm",
                "file_type": "pdf"
            }
        }
    ],
    "chunk_strategy": {
        "separator": "\n\n",
        "max_tokens": 800,
        "remove_extra_spaces": false,
        "remove_urls_emails": false,
        "chunk_type": 1
    }
}'
```


</div>
<div type="tab-item" title="上传在线网页" key="Y4RCziW2E5RIUorsj53_j">

```JSON
curl --location --request POST 'https://api.coze.cn/open_api/knowledge/document/create' \
--data-raw '{
    "dataset_id": "7363569245306945580",
    "document_bases": [
        {
            "name": "直接传 url163",
            "source_info": {
                "web_url" : "163.com",
                "document_source" : 1 
            },
            "update_rule": {
                "update_type": 1, 
                "update_interval": 24 
            }
        }
    ],
     "chunk_strategy": {
        "separator": "\n\n", 
        "max_tokens": 800,
        "remove_extra_spaces": false
        "remove_urls_emails": false
        "chunk_type": 1 
    }
}'
```


</div>
<div type="tab-item" title="通过 file_id 上传图片" key="HHgbcQ3nuapaAQNqSTimV">

```JSON
curl --location --request POST 'https://api.coze.cn/open_api/knowledge/document/create' \
--data-raw '{
    "dataset_id": "744632265564659****",
    "format_type": 2,
    "chunk_strategy": {
        "caption_type": 0
    },
    "document_bases": [
        {
            "name": "test1",
            "source_info": {
                "source_file_id": 744667846938145****,
                "document_source": 5
            }
        },
        {
            "name": "test2",
            "source_info": {
                "source_file_id": 744667846938145****,
                "document_source": 5
            }
        }
    ]
}'
```


</div>
</div>
### 返回示例
```JSON
{
    "document_infos": [
        {
            "name": "Coze.pdf",
            "size": 14164,
            "type": "pdf",
            "status": 1,
            "tos_uri": "FileBizType.BIZ_BOT_DATASET/847077809337655_1727579972975689529_0ytrdq****.docx",
            "hit_count": 0,
            "char_count": 4,
            "create_time": 1719907964,
            "document_id": "738694205603010****",
            "format_type": 2,
            "slice_count": 1,
            "source_type": null,
            "update_time": 1719907969,
            "update_type": null,
            "chunk_strategy": {
                "chunk_type": 1,
                "max_tokens": 800,
                "remove_extra_spaces": false,
                "remove_urls_emails": false,
                "separator": "\n\n"
            },
            "update_interval": 0
        }
    ],
    "code": 0,
    "msg": "",
    "detail": {
        "logid": "20250106172024B5F607030EFFA***"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

