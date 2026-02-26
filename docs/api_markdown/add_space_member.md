# 批量邀请用户加入空间
邀请用户加入工作空间，支持批量邀请，每次最多添加 20 个成员。
## 接口描述
通过本 API 批量邀请用户加入工作空间，并设置用户是空间管理员或空间成员。

* 在企业版（企业标准版、企业旗舰版）中，添加用户时会直接将用户加入工作空间，无需被添加者确认。
* 在个人版中，添加用户时会向用户发送邀请，用户同意后才能加入工作空间。
* 如果被添加者的主账号设置了**不允许加入外部空间**，添加该用户时会报错。

## 接口限制

* 在企业版（企业标准版、企业旗舰版）中，仅支持邀请已加入企业的用户加入工作空间，否则调用 API 时会执行失败，并提示错误码 702042162。
* 添加成员后，工作空间成员总数不能超过该工作空间的成员数量上限，否则调用 API 时会执行失败，并提示错误码 702042018。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workspaces/:workspace_id/members <br> ``` <br>  |
| **权限** | `addMember` <br> 确保调用该接口使用的访问令牌开通了 `addMember` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 批量邀请用户加入工作空间。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workspace_id | String | 必选 | 736163827687053**** | 需要添加用户的工作空间 ID。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，workspace_id 为 `75814654762959***`。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| users | Array of [OpenSpaceMember](#openspacemember) | 可选 | ["206972012541****","552425858****"] | 待邀请加入工作空间的用户列表，单次最多添加 20 个成员。 |
### OpenSpaceMember
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| user_id | String | 必选 | 2069720456***** | 待邀请加入工作空间的用户的扣子用户 ID。 <br> 你可以单击扣子编程左下角的头像，选择**账号设置**，在页面底部查看扣子用户的 UID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/c77171908fad4b69a328dfd3210b3576~tplv-goo7wpa0wc-image.image) |
| role_type | String | 必选 | member | 该用户在工作空间中的角色： <br>  <br> * admin：工作空间管理员。 <br> * member：工作空间成员。 <br>  <br> 不支持设置为空间所有者。 <br>  |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [OpenAddSpaceMemberData](#openaddspacememberdata) | \ | 添加成员的结果信息。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFAD6*****" } | 响应详情信息。 |
### OpenAddSpaceMemberData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| not_exist_user_ids | Array of String | ["21357147977***"] | 因用户 ID 不存在而导致添加失败的用户 ID 列表。 |
| added_success_user_ids | Array of String | ["21399947977***"] | 企业标准版或企业旗舰版成功添加的用户 ID 列表。 |
| already_joined_user_ids | Array of String | ["2177747977***"] | 用户在该工作空间中已经存在，不重复添加。 |
| already_invited_user_ids | Array of String | ["21888147977***"] | 已经发起邀请但用户还未同意加入的用户 ID 列表。 <br> 仅个人版的工作空间添加用户时需要发出邀请，用户同意后才会加入工作空间。 |
| invited_success_user_ids | Array of String | ["2136666147977***"] | 个人版中，发起邀请且用户同意加入的用户 ID 列表。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
POST 'https://https://api.coze.cn/v1/workspaces/7515267805***/members' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer pat_FFTwAe4KB9GpaXCGoCjKbztWig4MT1SRB6yjP4n0PJ1AE3ukqw5yGj****'
{
    "users": [
        {
            "role_type": "member",
            "user_id": "21357147977***"
        },
        {
            "role_type": "member",
            "user_id": "55242585801***"
        }
    ]
}
```

### 返回示例
```JSON
{
  "data": {
    "already_joined_user_ids": [
      "2135714797****"
    ],
    "already_invited_user_ids": [],
    "added_success_user_ids": [
      "55242585****"
    ],
    "invited_success_user_ids": [],
    "not_exist_user_ids": []
  },
  "msg": "",
  "detail": {
    "logid": "20250616153813E8D5D5F993****"
  },
  "code": 0
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。