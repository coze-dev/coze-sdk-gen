# 创建声纹
创建声纹。
## 接口描述
创建声纹组后，你可以在对应的声纹组中创建声纹，上传本地预先录制好的音频文件进行创建声纹，扣子将通过该音频文件创建声纹。
## 接口限制

* 每个声纹组中最多可创建 10 个声纹。
* 声纹文件的具体限制如下：
   | **限制** | **说明** |
   | --- | --- |
   | 文件格式 | 支持的文件格式包括 wav 和 pcm。其中 pcm 仅支持 16kHz 采样率、单声道、16 位采样深度。 |
   | 文件大小 | 每个音频文件最大为 4 MB。 |
   | 文件录制 | * 录制环境：选择安静的空间，建议将麦克风放置在离说话人 50 厘米以内的位置，尽量保持自然的发声状态，避免刻意改变声线或呢喃，这样得到的声音会更加自然。尽量减少杂音、噪音和混响的干扰。 <br> * 音频质量：确保录音中只包含一个人的声音。 |


## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/audio/voiceprint_groups/:group_id/features <br> ``` <br>  |
| **权限** | `createVoiceprintFeature` <br> 确保调用该接口使用的个人令牌开通了 `createVoiceprintFeature` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 创建声纹。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | multipart/form-data | 文件类型。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| group_id | Integer | 必选 | 75143038117378*** | 声纹所属的声纹组 ID。你可以通过[查看声纹组列表](https://docs.coze.cn/developer_guides/list_voiceprint_group) API 查看声纹组 ID。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| file | File | 必选 | 无 | 上传本地预先录制好的音频文件。 |
| name | String | 必选 | 妈妈 | 声纹的名称。 |
| desc | String | 可选 | 妈妈的声音 | 声纹的描述。 |
| sample_rate | Integer | 可选 | 16000  | 采样率，单位：Hz。当文件类型为 pcm 时需传入该参数。 <br> 默认为 16kHz 采样率。 |
| channel | Integer | 可选 | 1 | 声道数，当文件类型为 pcm 时需传入该参数。 <br> 默认为单声道。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [CreateVoicePrintGroupFeatureData](#createvoiceprintgroupfeaturedata) | {"id":"751427391012944***"} | 返回的声纹信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "202410242028595CCF353CEC86A8*****" } | 本次请求返回的详细信息。 |
### CreateVoicePrintGroupFeatureData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 751427391012944*** | 声纹的 ID。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
POST https://api.coze.cn/v1/audio/voiceprint_groups/75143038117378***/features \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: multipart/form-data' \
-F 'file=@/voice_test_data/test_1.wav;type=audio/wav' \
-F 'name=妈妈的声音' \
-F 'desc=妈妈的声音'
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "202506091643272FB1EDC9CA****"
  },
  "data": {
    "id": "75138652833128***"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。