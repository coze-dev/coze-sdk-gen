# 创建房间
创建房间，并将智能体加入房间。
## 接口描述
* 调用创建房间 API 之后，智能体随即加入房间并开始计费，包括智能体的**对话式 Al-音频费用**和**语音通话费用**，具体费用详情请参考[音视频费用](https://docs.coze.cn/coze_pro/asr_tts_fee#f4e51f74)。**为避免不必要的费用产生，请在调用接口前仔细了解费用详情，并合理控制创建房间接口的调用频率。**
* 用户未加入房间与智能体进行对话时，智能体会在房间等待用户 3 分钟，这期间会产生 3 分钟的最低费用。

扣子智能语音功能通过 RTC 技术实现用户和智能体的实时语音通话。在 RTC 领域中，房间是一个虚拟的概念，类似逻辑上的分组，同一个房间内的用户才能互相接收和交换音视频数据、实现音视频通话。
此 API 用于创建一个房间，并将智能体加入房间，然后才能调用 RTC SDK 和智能体开始语音通话。
## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/audio/rooms <br> ``` <br>  |
| **权限** | `createRoom` <br> 确保调用该接口使用的令牌开通了 `createRoom` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 创建房间，并将智能体加入房间。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| bot_id | String | 必选 | 73428668***** | 智能体 ID。 <br> 进入智能体的开发页面，开发页面 URL 中 bot 参数后的数字就是智能体 ID。例如`https://www.coze.com/space/341****/bot/73428668*****`，bot ID 为`73428668*****`。 |
| conversation_id | String | 可选 | 734829333445931**** | 会话 ID。后续调用发起对话 API 产生的消息记录都会保存在此对话中。 <br> 调用[创建会话](https://docs.coze.cn/developer_guides/create_conversation) API 可以创建一个会话。若未指定会话 ID，扣子编程会默认创建一个新的会话。 |
| voice_id | String | 可选 | 734829333445931**** | 智能体使用的音色 ID，默认为`柔美女友`音色。 <br> 你可以通过[查看音色列表](https://docs.coze.cn/developer_guides/list_voices) API 获取音色 ID。 |
| config | Object of [RoomConfig](#roomconfig) | 可选 | - | 房间内的音视频参数配置。 |
| uid | String | 可选 | uid_123 | 标识当前与智能体对话的用户，由使用方自行定义、生成与维护。uid 用于标识对话中的不同用户，不同的 uid，其对话的数据库等对话记忆数据互相隔离。如果不需要用户数据隔离，可以不传此参数。 |
### RoomConfig
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| room_mode | String | 可选 | default | 房间模式，默认值为 `default`。枚举值： <br>  <br> * `default`：对话模式。 <br> * `podcast`：播客模式。 |
| audio_config | Object of [AudioConfig](#audioconfig) | 可选 | \ | 房间音频配置。 |
| video_config | Object of [VideoConfig](#videoconfig) | 可选 | \ | 房间视频配置。 |
| turn_detection | Object of [TurnDetectionConfig](#turndetectionconfig) | 可选 | {"type":"server_vad"} | 语音检测配置，用于控制语音交互的检测方式。 |
| prologue_content | String | 可选 | \ | 自定义开场白。 |
| prologue_delay_duration_ms | Integer | 可选 | 100 | 在进房后等待多长时间播放开场白，单位：ms。 <br> 默认为 500ms，取值范围为 0~5000。 |
### AudioConfig
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| codec | String | 可选 | OPUS | 房间音频编码格式，支持设置为： <br>  <br> * AACLC：AAC-LC 编码格式。 <br> * G711A：G711A 编码格式。 <br> * OPUS：（默认）Opus 编码格式。 <br> * G722：G.722 编码格式。 |
### VideoConfig
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| codec | String | 可选 | H264 | 房间视频编码格式，支持设置为： <br>  <br> * H264：（默认）H264 编码格式。 <br> * BYTEVC1：火山引擎自研的视频编码格式。 |
| video_frame_rate | Integer | 可选 | 15 | 每秒抽帧数，在视频通话过程中，摄像头/屏幕共享捕捉画面的频率。捕捉到的画面会作为视觉模型的输入。 <br> 默认值为 `1`，取值范围为 1~24。 |
| stream_video_type | String | 可选 | main | 房间视频流类型, 支持 main 和 screen。 <br>  <br> * main：主流，包括通过摄像头/麦克风的内部采集机制获取的流，以及通过自定义采集方式获取的流。 <br> * screen：屏幕流，用于屏幕共享或屏幕录制的视频流。 |
| video_frame_expire_duration | Integer | 可选 | 5 | 用户开始说话前，抽取多少秒画面。主要是兼容连贯动作的场景。用于帮模型理解用户没开始说话前在做什么。 <br> 单位为秒，默认值为 `1`，取值范围为 1~10。 |
### TurnDetectionConfig
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| type | String | 可选 | server_vad | 语音检测类型，用于控制语音交互的检测方式，默认值为 `server_vad`。枚举值： <br>  <br> * `server_vad`：自由对话模式，扣子编程云端通过语音活动检测（VAD）自动判断语音的开始和结束，实现无缝的自然对话体验。 <br> * `client_interrupt`：按键说话模式，由客户端控制语音的开始与结束，需配合 Realtime 上行事件 `input_audio_buffer.start` 和 `input_audio_buffer.complete`使用。 <br> * `semantic_vad`：采用语义判停的自由对话模式（**此功能仅对企业旗舰版用户开放**），由服务端识别语义来判断是否停止说话。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码。 <br> `0`代表调用成功。 |
| data | Object of [CreateRoomData](#createroomdata) | - | 接口返回的业务数据。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "202410291302044CD1CC3B4AE0897***" } | 本次请求的详细信息。 |
### CreateRoomData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| uid | String | uid_123 | 进入房间的用户 ID，和入参中 uid 一致。 |
| token | String | token123 | 房间密钥，用户加入房间和智能体对话时需要通过 token 进行身份认证和鉴权。 |
| app_id | String | app_id | RTC 应用 ID。 |
| room_id | String | room_id_123 | 已创建的 RTC 房间的房间 ID。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/audio/rooms' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "bot_id": "734829333445931****"
}'
```

### 返回示例
```JSON
{
    "detail": {
        "logid": "202410291302044CD1CC3B4AE0897***"
    },
    "data": {
        "room_id": "room_id_743105798342***",   // 房间 id
        "app_id": "6705332c79516e01****",       // app_id
        "token": "0016705*****NzkxMxcANTE1MjkFAAAAAAAAAAEAAAAAACAA58QEAvxd****/VB4b1xEFA=",
        "uid": "uid_74310579834270***"
    },
    "code": 0,
    "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。

