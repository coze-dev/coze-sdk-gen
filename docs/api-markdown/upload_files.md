# 上传文件
调用接口上传文件到扣子编程。
## 接口说明
消息中无法直接使用本地文件，创建消息或对话前，需要先调用此接口上传本地文件到扣子编程。上传文件后，你可以在消息中通过指定 file_id 的方式在多模态内容中直接使用此文件。此接口上传的文件可用于[发起对话](https://docs.coze.cn/developer_guides/chat_v3)等 API 中传入文件等多模态内容。使用方式可参考 [object_string object](https://docs.coze.cn/developer_guides/chat_v3#426b92e7) 。
### 使用限制
| **限制** | **说明** |
| --- | --- |
| 文件大小 | 该 API 允许上传的最大文件大小为 512 MB。然而，在与智能体对话时，实际可使用的文件大小取决于智能体的模型版本。 |
| 上传方式 | 必须使用 multipart/form-data 方式上传文件。 |
| 有效期 | * 普通上传的文件将保存在扣子编程服务端，有效期为 3 个月。 <br> * 若上传的文件被用作扣子头像，则永久有效。 |
| 使用限制 | * 上传到扣子编程的文件仅限本账号查看或使用。 <br> * 调用此接口上传文件只能获得文件的 **file_id**，如需获取文件的 URL，建议将文件上传到专业的存储工具中。 <br> * 不支持下载已上传的文件。用户仅可在对话、工作流、端插件、RTC 和 WebSocket 中通过 `file.id` 访问和使用文件。 |
| API 流控 | * 个人版：10 QPS。 <br> * 企业版：20 QPS。 |
### 支持的文件格式
| **文件类型** | **支持的格式** |
| --- | --- |
| 文档 | DOC、DOCX、XLS、XLSX、PPT、PPTX、PDF、Numbers、CSV |
| 文本文件 | CPP、PY、JAVA、C  |
| 图片 | JPG、JPG2、PNG、GIF、WEBP、HEIC、HEIF、BMP、PCD、TIFF |
| 音频 | WAV、MP3、FLAC、M4A、AAC、OGG、WMA、MIDI |
| 视频 | MP4、AVI、MOV、3GP、3GPP、FLV、WEBM、WMV、RMVB、M4V、MKV  |
| 压缩文件 | RAR、ZIP、7Z、GZ、GZIP、BZ2 |

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/files/upload <br> ``` <br>  |
| **权限** | `uploadFile` <br> 确保调用该接口使用的访问令牌开通了 `uploadFile` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 调用接口上传文件到扣子编程。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.com/docs/developer_guides/preparation)。 |
| Content-Type | multipart/form-data | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| Data | File | 必选 | @"/test/1120.jpeg"' | 待上传文件的二进制数据。 <br> 开发者需要在前端应用中实现文件选择功能，获取用户选择的文件后，将文件的二进制内容直接设置为 HTTP 请求体。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [File](#file) | { "bytes": 152236, "created_at": 1715847583, "file_name": "1120.jpeg", "id": "736949598110202****" } | 已上传的文件信息。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
### File
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 736949598110202**** | 已上传的文件 ID。 |
| bytes | Long | 152236 | 文件的总字节数。 |
| file_name | String | 1120.jpeg | 文件名称。 |
| created_at | Long | 1715847583 | 文件的上传时间，格式为 10 位的 Unixtime 时间戳，单位为秒（s）。 |

## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/files/upload' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type: multipart/form-data' \
--form 'file=@"/test/1120.jpeg"'
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
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

