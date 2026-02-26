# 语音识别
将音频文件转录为文本。
## 接口描述
此 API 用于将指定音频文件转录为文本。
调用语音识别 API 会产生**语音识别费用**，具体费用详情请参考[音视频费用](https://docs.coze.cn/coze_pro/asr_tts_fee#d4ab71d9)。

语音文件的具体限制如下：
| **限制** | **说明** |
| --- | --- |
| 文件格式 | * 大模型语音识别支持的文件格式：opus、ogg、mp3、wav、m4a、mp4、pcm、raw、spx、aac、amr。 <br> * 小模型语音识别支持的文件格式：ogg、mp3、wav、mp4、m4a。 |
| 文件大小 | 每个音频文件最大为 10 MB，并且时长需小于 30 分钟。 |
* 上传的音频文件的采样率和码率等参数无限制。
* 如果语音文件过大，建议调用 WebSocket 的[双向流式语音识别 API](https://docs.coze.cn/developer_guides/asr_api) 分片上传文件。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/audio/transcriptions <br> ``` <br>  |
| **权限** | `createTranscription` <br> 确保调用该接口使用的个人令牌开通了 `createTranscription` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 将音频文件转录为文本。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | multipart/form-data | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| body | File | 必选 | - | 需要语音识别的音频文件的二进制数据，具体文件格式和大小限制请参考接口说明部分。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码。 <br> `0`代表调用成功。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [AudioTranscriptionsData](#audiotranscriptionsdata) | {"text":"你好，请问有什么可以帮您的吗？"} | 语音文件对应的文本内容，即语音识别后转换成的文字结果。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### AudioTranscriptionsData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| text | String | 你好，请问有什么可以帮您的吗？ | 语音文件对应的文本内容，即语音识别后转换成的文字结果。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/audio/transcriptions' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: multipart/form-data' \
--form 'file=@"/xx/xx/xx/jay.MP3"'
```

### 返回示例
```JSON
{
  "detail": {
    "logid": "202410291302044CD1CC3B4AE0897***"
  },
  "data": {
    "text": "你好呀"
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。