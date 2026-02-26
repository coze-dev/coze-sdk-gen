# 查看音色列表
查看可用的音色列表，包括系统预置音色和自定义音色。
## 接口说明
调用此 API 可查看当前扣子用户可使用的音色列表，包括：

* 系统预置音色：扣子编程提供的默认音色。
* 自定义音色：当前扣子用户通过**复刻音色** API 复刻的音色、当前账号加入的所有工作空间中其他扣子用户复刻的音色。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/audio/voices <br> ``` <br>  |
| **权限** | `listVoice` <br> 确保调用该接口使用的个人令牌开通了 `listVoice` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查看可用的音色列表，包括系统预置音色和自定义音色。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| filter_system_voice | Boolean | 可选 | false | 查看音色列表时是否过滤掉系统音色。 <br>  <br> * true：过滤系统音色 <br> * false：（默认）不过滤系统音色 |
| model_type | String | 可选 | big | 音色模型的类型，如果不填，默认都返回。可选值包括： <br>  <br> * big：大模型 <br> * small：小模型 |
| voice_state | String | 可选 | cloned | 音色克隆状态，用于筛选特定状态的音色。可选值包括：  <br>  <br> * init：待克隆。  <br> * cloned：（默认值）已克隆。  <br> * all：全部。 |
| page_num | Integer | 可选 | 1 | 查询结果分页展示时，此参数用于设置查看的页码。最小值为 1，默认为 1。 |
| page_size | Integer | 可选 | 100 | 查询结果分页展示时，此参数用于设置每页返回的数据量。取值范围为 1~100，默认为 100。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码。 <br> `0`代表调用成功。 |
| data | Object of [ListVoiceData](#listvoicedata) | 参见响应参数 | 音色的详细信息。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "2024102916355736DC98FBC4D32FD7E59C" } | 本次请求的详细信息。 |
### ListVoiceData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| has_more | Boolean | false | 标识是否还有未返回的音色数据。 <br>  <br> *  `true` ：当前返回的音色列表未包含所有符合条件的音色。 <br> *  `false`：表示已返回所有符合条件的音色数据。 |
| voice_list | Array of [OpenAPIVoiceData](#openapivoicedata) | 参见响应参数 | 音色列表详情。 |
### OpenAPIVoiceData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| name | String | 开朗大男孩 | 音色的名称。 |
| state | String | cloned | 音色克隆状态。枚举值：  <br>  <br> * init：待克隆。  <br> * cloned：已克隆。  |
| voice_id | String | 734829333445931**** | 音色的 ID。 |
| model_name | String |  | 模型类型字符串 |
| model_type | String | big | 音色模型的类型，枚举值： <br>  <br> * big：大模型 <br> * small：小模型 |
| create_time | Integer | 1729686510 | 音色的创建时间，格式为 11 位的 Unixtime 时间戳。 |
| update_time | Integer | 1729686510 | 音色的更新时间，格式为 11 位的 Unixtime 时间戳。 |
| preview_text | String | 你好呀 | 此音色预览音频对应的文案。 |
| language_code | String | zh | 此音色的语种代号。 |
| language_name | String | 中文 | 此音色的语种名称。 |
| preview_audio | String | https://lf3-appstore-sign.oceancloudapi.com/ocean-cloud-tos/VolcanoUserVoice/xxxxxx.mp3?lk3s=da27ec82&x-expires=1730277357&x-signature=xu2O6Gp5RvTyJOawqjAfsJZvifc%3D | 此音色的预览音频。通常是一个公开可访问的网络地址。 |
| is_system_voice | Boolean | false | 标识当前音色是否为系统预置音色。 <br>  <br> * `true`：系统预置音色。 <br> * `false`：用户自定义音色。 |
| support_emotions | Array of [EmotionInfo](#emotioninfo) | [{"emotion":"happy","display_name":"开心","emotion_scale_interval":{"max":5,"min":1,"default":4}},{"emotion":"sad","display_name":"悲伤","emotion_scale_interval":{"max":5,"min":1,"default":3}}] | 音色支持的情感类型列表，仅当音色为多情感音色时返回。 |
| available_training_times | Integer | 6 | 当前音色还可训练的次数。包括首次复刻音色在内，每个自定义音色最多被训练 10 次。 |
### EmotionInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| emotion | String | happy | 音色支持的情感类型标识符，仅当音色为多情感音色时返回。枚举值如下： <br>  <br> * `happy`：开心。 <br> * `sad`：悲伤。 <br> * `angry`：愤怒。 <br> * `surprised`：惊讶。 <br> * `fear`：恐惧。 <br> * `hate`：厌恶。 <br> * `excited`：兴奋。 <br> * `coldness`：冷漠。 <br> * `neutral`：中性。 |
| display_name | String | 开心 | 音色支持的情感类型的中文显示名称，用于直观展示情感类型。 |
| emotion_scale_interval | Object of [Interval](#interval) | {"max":5,"min":1,"default":4} | 情感强度的取值范围，用于量化情感的强度。 |
### Interval
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| max | Double | 5 | 情感强度的最大值。 |
| min | Double | 1 | 情感强度的最小值。 |
| default | Double | 4 | 情感强度的默认值。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/audio/voices?filter_system_voice=false&model_type=big&voice_state=&page_num=1&page_size=100' \
--header 'Authorization: Bearer Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
    "detail": {
        "logid": "2024102916355736DC98FBC4D32F***"
    },
    "data": {
        "voice_list": [
            {
                "preview_audio": "https://lf3-appstore-sign.oceancloudapi.com/ocean-cloud-tos/VolcanoUserVoice/xxxxxx.mp3?lk3s=da27ec82&x-expires=1730277357&x-signature=xu2O6Gp5RvTyJOawqjAfsJZvifc%3D",
                "language_name": "中文",
                "is_system_voice": false,
                "preview_text": "你好，欢迎来到AI世界，我是你的专属AI克隆声音，希望未来可以一起好好相处。",
                "create_time": 1729686510,
                "update_time": 1729686510,
                "name": "jay", // 音色名称
                "language_code": "zh",
                "voice_id": "12344",
                "available_training_times": 6, // 当前音色还可以训练的次数
                "model_type": "big"
            }
        ],
        "has_more": false
    },
    "code": 0,
    "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
