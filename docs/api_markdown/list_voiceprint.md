# 查询声纹列表
查询指定声纹组中的声纹列表。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/audio/voiceprint_groups/:group_id/features <br> ``` <br>  |
| **权限** | `listVoiceprintFeature` <br> 确保调用该接口使用的个人令牌开通了 `listVoiceprintFeature` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询声纹组中的声纹列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| group_id | Integer | 必选 | 75143038117378*** | 声纹组 ID。你可以通过[查看声纹组列表](https://docs.coze.cn/developer_guides/list_voiceprint_group) API 查看声纹组 ID。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_num | Long | 可选 | 1 | 查询结果分页展示时，此参数用于设置查看的页码。最小值为 1，默认为 1。 |
| page_size | Long | 可选 | 5 | 查询结果分页展示时，此参数用于设置每页返回的数据量。取值范围为 1~200，默认为 10。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [GetVoicePrintGroupFeatureListData](#getvoiceprintgroupfeaturelistdata) | \ | 返回的声纹列表及相关信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### GetVoicePrintGroupFeatureListData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [VoicePrintGroupFeature](#voiceprintgroupfeature) | \ | 声纹列表，包含多个声纹的详细信息。 |
| total | Long | 5 | 声纹数量。 |
### VoicePrintGroupFeature
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 751427391012944*** | 声纹的 ID。 |
| desc | String | 妈妈的声纹 | 声纹的描述。 |
| name | String | 妈妈 | 声纹的名称。 |
| group_id | String | 75143038117378*** | 声纹组的 ID。 |
| icon_url | String | https://example.com/voiceprint/default_voiceprint_group_feature_icon.png | 声纹的头像 URL。 |
| audio_url | String | https://example.com/audio/751427391012944***.pcm | 声纹原始音频的 URL 地址，用于访问或下载该声纹对应的音频文件。 |
| user_info | Object of [UserInfo](#userinfo) | {"id":"324566624","name":"小王","nickname":"小王","avatar_url":"https://example.com/avatar/751427391012944***.image"} | 声纹创建者的用户信息。 |
| created_at | Long | 1714567890123 | 声纹的创建时间，以 Unix 时间戳格式表示，单位为秒。 |
| updated_at | Long | 1714567890223 | 声纹的更新时间，以 Unix 时间戳格式表示，单位为秒。 |
### UserInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 14891859095*** | 声纹创建者的扣子用户的 UID。 |
| name | String | 小王 | 声纹创建者的扣子用户名。 |
| nickname | String | 小王 | 声纹创建者的扣子用户昵称。 |
| avatar_url | String | https://example.com/e7b19241fb224cea967df****_1080_1080.png~0x0.image | 声纹创建者的扣子用户头像。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl -X POST 'https://api.coze.cn/v1/audio/voiceprint_groups/7511988408546279468/features?page_num=1&page_size=10' \
--header 'Authorization: Bearer pat_aqtInOKBi7ThceVOrVvm51kguh39iXlvHxvjr9NtZ8pe20xh2qLniSZWeq***' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
  "detail": {
    "logid": "20250604160710FB81ACE4CFC****"
  },
  "data": {
    "total": 1,
    "items": [
      {
        "id": "7511995644152****",
        "name": "test_feature",
        "user_info": {
          "id": 14891859095***,
          "name": "testcase1",
          "nickname": "5707",
          "avatar_url": "https://example.com/img/user-avatar/assets/e7b19241fb224cea967dfa****_1080_1080.png~0x0.image"
        },
        "audio_url": "https://example.com/voiceprint/9a71240d-f316-4bb6-8a5f-8a****",
        "desc": "This is a test feature",
        "icon_url": "https://example.com/voiceprint/default_voiceprint_group_feature_icon.png",
        "group_id": "75119884085462***",
        "created_at": 1749024396,
        "updated_at": 1749024396
      }
    ]
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。