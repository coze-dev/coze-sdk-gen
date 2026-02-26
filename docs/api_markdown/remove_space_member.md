# 批量移除空间中的用户
批量移除工作空间中的用户。每次最多移除 5 个成员。
## 接口描述
调用本 API 将用户从指定的工作空间中移除。只能移除空间管理员或成员，不支持移除空间所有者。批量移除用户时，你可以在返回结果中查看移除失败的用户以及具体原因。
移除成员后，该成员拥有的资源将自动转移给工作空间所有者。
## 基础信息
| **请求方式** | DELETE |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workspaces/:workspace_id/members <br> ``` <br>  |
| **权限** | `removeMember` <br> 确保调用该接口使用的访问令牌开通了 `removeMember` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 批量移除空间中的用户。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workspace_id | String | 必选 | 736163827687053**** | 需要移除用户的空间 ID。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，workspace_id 为 `75814654762959***`。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| user_ids | Array of String | 必选 | ["206972012541****","552425858****"] | 要移除的成员，单次最多移除 5 个成员。 |

## 返回参数

| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [OpenRemoveSpaceMemberData](#openremovespacememberdata) | \ | 移除空间成员的结果信息。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "20250106172024B5F607030EFFAD6*****" } | 响应详情信息。 |
### OpenRemoveSpaceMemberData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| removed_success_user_ids | Array of String | ["21399947977***"] | 成功移除的成员列表。 |
| not_in_workspace_user_ids | Array of String | ["21399947666***"] | 不在当前空间中的用户 ID 列表，这些用户不会被处理。 |
| owner_not_support_remove_user_ids | Array of String | ["2177747977***"] | 移除失败，该用户为空间所有者。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |


## 示例
### 请求示例
```JSON
curl --location --request DELETE 'https://api.coze.cn/v1/workspaces/751526780526****/members' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer pat_FFTwA******' \
{
    "user_ids": ["206972012541****", "2135714797****", "552425858****"]
```

### 返回示例
```JSON
{
  "msg": "",
  "detail": {
    "logid": "20250616153946D01251E104****"
  },
  "code": 0,
  "data": {
    "removed_success_user_ids": [
      "552425858****"
    ],
    "not_in_space_user_ids": [
      "2069720125****"
    ],
    "owner_not_support_remove_user_ids": [
      "21357147977***"
    ]
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。