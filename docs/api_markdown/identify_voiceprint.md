# 声纹特征比对
声纹特征比对。
## 接口说明
声纹特征比对将产生声纹识别费用，详细费用说明可参考[音视频费用](https://docs.coze.cn/coze_pro/asr_tts_fee#b7ced22d)。

你可以调用本 API 识别说话者的身份。上传说话者的音频，扣子编程将根据该音频与指定声纹组中已有的声纹进行对比，计算相似度，并返回与各声纹的匹配度列表。
上传的音频文件的限制如下：
| **限制** | **说明** |
| --- | --- |
| 文件格式 | 支持的文件格式包括 wav 和 pcm。其中 pcm 仅支持 16kHz 采样率、单声道、16 位采样深度。 |
| 文件大小 | 每个音频文件最大为 4 MB。 |
| 文件录制 | * 录制环境：选择安静的空间，建议将麦克风放置在离说话人 50 厘米以内的位置，尽量保持自然的发声状态，避免刻意改变声线或呢喃，这样得到的声音会更加自然。尽量减少杂音、噪音和混响的干扰。 <br> * 音频质量：确保录音中只包含一个人的声音。 |

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/audio/voiceprint_groups/:group_id/speaker_identify <br> ``` <br>  |
| **权限** | `identifyInVoiceprintGroup` <br> 确保调用该接口使用的个人令牌开通了 `identifyInVoiceprintGroup` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 声纹特征比对。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | application/json | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | multipart/form-data | 文件类型。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| group_id | Integer | 可选 | 75143038117378*** | 声纹所属的声纹组 ID。你可以通过[查看声纹组列表](https://docs.coze.cn/developer_guides/list_voiceprint_group) API 查看声纹组 ID。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| file | File | 必选 | @/voice_test_data/test_1.wav;type=audio/wav | 上传需要进行声纹特征比对的音频文件。 |
| top_k | Interger | 可选 | 5 | 返回特征匹配度最高的前 k 项结果。若未指定，则默认返回全部匹配结果。 |
| sample_rate | Integer | 可选 | 16000  | 采样率，单位：Hz。当文件类型为 pcm 时需传入该参数。 <br> 默认的采样率为 16kHz。 |
| channel | Integer | 可选 | 1 | 声道数，当文件类型为 pcm 时需传入该参数。 <br> 默认为单声道。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [SpeakerIdentifyData](#speakeridentifydata) | [{"score":60,"feature_id":"751427391012944***","feature_desc":"妈妈的声音","feature_name":"妈妈"}] | 声纹匹配度列表，包含测试音频与声纹库中已有声纹的匹配信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "202410242028595CCF353CEC86A8*****" } | 本次请求返回的详细信息。 |
### SpeakerIdentifyData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| feature_score_list | Array of [FeatureScore](#featurescore) | {"score":60,"feature_id":"751427391012944***","feature_desc":"妈妈的声音","feature_name":"妈妈"} | 声纹匹配度列表，包含测试音频与声纹库中已有声纹的匹配信息。 |
### FeatureScore
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| score | Double | 60 | 声纹匹配度。表示测试音频与声纹库中已有声纹的匹配程度，数值范围为 0~100，数值越高表示匹配度越高。 |
| feature_id | String | 751427391012944*** | 与测试音频匹配的声纹的 ID。 |
| feature_desc | String | 妈妈的声音 | 声纹的描述。 |
| feature_name | String | 妈妈 | 声纹的名称。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/audio/voiceprint_groups/7513864977189*****/speaker_identify' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: multipart/form-data' \
-F 'file=@/voice_test_data/test_3.wav;type=audio/wav'
```

### 返回示例
```JSON
{
  "data": {
    "feature_score_list": [
      {
        "score": 23.139079321463843,
        "feature_name": "妈妈的声音",
        "feature_desc": "妈妈的声音",
        "feature_id": "7513865283312***"
      }
    ]
  },
  "code": 0,
  "msg": "",
  "detail": {
    "logid": "202506091647409DF126AF380****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://www.coze.cn/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。