# 查询声纹组列表
查询已创建的声纹组列表。
## 接口描述
支持通过名称模糊匹配、用户ID、声纹组 ID 进行精确匹配。返回结果包括声纹组的详细信息，包括声纹组的创建者、创建时间等信息。
你可以查询团队或企业中的全部声纹组，或查询本人创建的声纹组。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/audio/voiceprint_groups <br> ``` <br>  |
| **权限** | `listVoiceprintGroup` <br> 确保调用该接口使用的个人令牌开通了 `listVoiceprintGroup` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查询已创建的声纹组列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_num | Long | 可选 | 1 | 查询结果分页展示时，此参数用于设置查看的页码。最小值为 1，默认为 1。 |
| page_size | Long | 可选 | 20 | 查询结果分页展示时，此参数用于设置每页返回的数据量。取值范围为 1~200，默认为 10。 |
| name | String | 可选 | 声纹组 | 用于根据声纹组名称前缀进行模糊匹配筛选，支持输入部分名称前缀进行查询。 |
| user_id | String | 可选 | 757378*** | 声纹创建者的扣子用户 UID。 <br>  <br> * 若指定用户 ID，则只能填写本人的用户 ID，即仅能查询本人创建的声纹组。 <br> * 如果不填写，则查询团队或企业中的所有声纹组。 <br>  <br> 你可以单击扣子编程左下角的头像，选择**账号设置**，在页面底部查看扣子用户的 UID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/6123f0e7c9cf4f769464618aa7f57a70~tplv-goo7wpa0wc-image.image) |
| group_id | String | 可选 | 75143038117378*** | 声纹组 ID，用于指定查询特定的声纹组。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [GetVoicePrintGroupListData](#getvoiceprintgrouplistdata) | \ | 返回的声纹组列表及相关信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {logid": "202410242028595CCF353CEC86A8*****" } | 本次请求返回的详细信息。 |
### GetVoicePrintGroupListData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [VoicePrintGroup](#voiceprintgroup) | \ | 声纹组列表，包含多个声纹组的详细信息。 |
| total | Long | 5 | 声纹组数量。 |
### VoicePrintGroup
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 75143038117378*** | 声纹组的 ID。 |
| desc | String | 声纹组1 | 声纹组的描述。 |
| name | String | 声纹组1 | 声纹组名称。 |
| icon_url | String | https://example.com/avatar/751427391012944***.jpg | 声纹组的头像 URL。 |
| user_info | Object of [UserInfo](#userinfo) | {"id":"324566624","name":"小王","nickname":"小王","avatar_url":"https://example.com/avatar/751427391012944***.jpg"} | 声纹组创建者的信息。 |
| created_at | Long | 1714567890123 | 声纹组的创建时间，以 Unix 时间戳格式表示，单位为秒。 |
| updated_at | Long | 1714567890223 | 声纹组的更新时间，以 Unix 时间戳格式表示，单位为秒。 |
| feature_count | Integer | 3 | 声纹组中包含的声纹数量。 |
### UserInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 324566624 | 声纹组创建者的用户 ID。 |
| name | String | 小王 | 声纹组创建者的用户名。 |
| nickname | String | 小王 | 声纹组创建者的用户昵称。 |
| avatar_url | String | https://example.com/avatar/751427391012944***.jpg | 声纹组创建者的用户头像。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
GET 'https://api.coze.cn/v1/audio/voiceprint_groups?page_num=1&page_size=10&group_id=7511988408546279468' \
--header 'Authorization: Bearer pat_gJUE27xgjJxfeH4ZtgrwrJHDnHPD5VTHfmA9IpBMfImgdcAQCzNLXDwm***' \
--header 'Content-Type: application/json'
```

### 返回示例
```JSON
{
  "detail": {
    "logid": "2025060415553824EF7ACB121****"
  },
  "data": {
    "total": 1,
    "items": [
      {
        "feature_count": 0,
        "desc": "This is a test group",
        "name": "test_group",
        "icon_url": "https://example.***/obj/ocean-cloud-tos/voiceprint/default_voiceprint_group_icon.png",
        "user_info": {
          "avatar_url": "https://p3-passport.byteacctimg.com/img/user-avatar/assets/e7b19241fb224cea967dfa****_1080_1080.png~0x0.image",
          "id": 14891859095***,
          "name": "testcase1",
          "nickname": "5707"
        },
        "id": "751198840854***",
        "created_at": 1749021940,
        "updated_at": 1749021940
      }
    ]
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://www.coze.cn/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。
