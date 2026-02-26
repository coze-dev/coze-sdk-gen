# 语音合成
将指定文本合成为音频文件。
## 接口描述
此 API 用于将指定文本内容合成为自然流畅的音频，同步返回合成的音频文件，适用于有声书合成、智能客服语音、音视频配音等场景。合成音频文件之前，可以先调用[查看音色列表](https://docs.coze.cn/developer_guides/list_voices) API，查看所有可用音色。
调用语音合成 API 会产生**语音合成费用**，具体费用详情请参考[音视频费用](https://docs.coze.cn/coze_pro/asr_tts_fee#b7ced22d)。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/audio/speech <br> ``` <br>  |
| **权限** | `createSpeech` <br> 确保调用该接口使用的个人令牌开通了 `createSpeech` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 将指定文本合成为音频文件。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| input | String | 必选 | 今天天气怎么样 | 合成语音的文本，经由 UTF-8 编码。长度限制为 1024 字节。 |
| voice_id | String | 必选 | 742894******** | 音频文件使用的音色 ID。 <br> 调用[查看音色列表](https://docs.coze.cn/developer_guides/list_voices) API，可查看所有可用音色。 |
| emotion | String | 可选 | happy | 设置多情感音色的情感类型，仅当 `voice_id` 为多情感音色时才支持设置情感类型。不同音色支持的情感范围不同，可以通过[系统音色列表](https://docs.coze.cn/dev_how_to_guides/sys_voice)查看各音色支持的情感。默认为空。枚举值如下： <br>  <br> * `happy`：开心。 <br> * `sad`：悲伤。 <br> * `angry`：愤怒。 <br> * `surprised`：惊讶。 <br> * `fear`：恐惧。 <br> * `hate`：厌恶。 <br> * `excited`：兴奋。 <br> * `coldness`：冷漠。 <br> * `neutral`：中性。 |
| emotion_scale | Double | 可选 | 3 | 情感值用于量化情感的强度。数值越高，情感表达越强烈，例如： “开心” 的情感值 5 比 1 更显兴奋。 <br> 取值范围：1.0~5.0，默认值：4.0。 |
| response_format | String | 可选 | mp3 | 音频文件的编码格式，支持设置为： <br>  <br> * wav：返回二进制 wav 音频。 <br> * pcm：返回二进制 pcm 音频。 <br> * ogg_opus：返回多段含 opus 压缩分片音频。 <br> * mp3：（默认）返回二进制 mp3 音频。 |
| speed | Double | 可选 | 1 | 语速，大模型音色的取值范围为 0.5~2，小模型音色的取值范围为 0.2~3，通常保留一位小数即可。 <br> 其中 0.2 表示 0.2 倍速，3 表示 3 倍速。默认为 1，表示 1 倍速。 |
| sample_rate | Integer | 可选 | 24000 | 音频采样率，单位为 Hz。 <br>  <br> * 8000：8k <br> * 16000：16k <br> * 22050：22.05k <br> * 24000：（默认）24k <br> * 32000：32k <br> * 44100：44.1k <br> * 48000：48k |
| loudness_rate | Integer | 可选 | 30 | 音频输出音量的增益或衰减比例，以百分比形式表示。取值范围为 `-50` ~ `100`，默认值为 0，表示原始音量。 <br>  <br> * 负值表示衰减：`-50` 表示音量降低 50%（即 0.5 倍）。 <br> * 正值表示增益：`100`表示音量提升 100%（即 2 倍）。 |
| context_texts | String | 可选 | 用低沉沙哑的语气、带着沧桑与绝望地说 | 语音合成的辅助信息，用于控制合成语音的整体情绪（如悲伤、生气）、方言（如四川话、北京话）、语气（如撒娇、暧昧、吵架、夹子音）、语速（快慢）及音调（高低）等。 <br> 默认为空。 <br> * 仅当 `voice_id` 为豆包语音合成大模型 2.0 音色时才支持该参数，具体音色列表请参见[系统音色列表](https://docs.coze.cn/dev_how_to_guides/sys_voice)。 <br> * 更多关于豆包语音合成 2.0 的 `context_texts` 示例和效果可参考[语音指令-示例库](https://www.volcengine.com/docs/6561/1871062?lang=zh#_1-2-%F0%9F%92%A1%E8%AF%AD%E9%9F%B3%E6%8C%87%E4%BB%A4-%E7%A4%BA%E4%BE%8B%E5%BA%93)。 <br>  |

## 返回参数
如果成功调用此 API，接口会直接返回语音文件的内容。
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/audio/speech' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
-d '{
  "input": "你好呀",
  "voice_id": "742894*********",
  "response_format": "wav"
}' 
--output speech.wav
```

### 返回示例
无
## 错误码
如果调用此 API 失败，返回信息中包含 code 和 msg 字段。其中 code 为非 0 的错误码，表示接口调用失败；msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。