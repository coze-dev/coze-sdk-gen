# 查看空间成员列表
查看指定空间的成员列表。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workspaces/:workspace_id/members <br> ``` <br>  |
| **权限** | `readMember` <br> 确保调用该接口使用的访问令牌开通了 `readMember` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查看指定空间的成员列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workspace_id | String | 必选 | 736163827687053**** | 需要查看成员列表的空间 ID。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，workspace_id 为 `75814654762959***`。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_num | Integer | 可选 | 1 | 查询结果分页展示时，此参数用于设置查看的页码。最小值为 1，默认为 1。 |
| page_size | Integer | 可选 | 10 | 查询结果分页展示时，此参数用于设置每页返回的数据量。取值范围为 1~50，默认为 20。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| data | Object of [OpenSpaceMemberListData](#openspacememberlistdata) | \ | 返回的数据对象，包含空间成员列表和总数。 |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {logid": "202410242028595CCF353CEC86A8*****" } | 本次请求返回的详细信息。 |
### OpenSpaceMemberListData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [OpenSpaceMember](#openspacemember) | [ { "role_type": "owner", "user_id": "2135714797***", "user_nickname": "RootUser_210202***", "user_unique_name": "kou_testing_001", "avatar_url": "example.com/cloudidentity.9f6***.jpg" }, { "avatar_url": "https://example.com/cloudidentity.9f6***.jpg", "role_type": "member", "user_id": "402688082103***", "user_nickname": "test_user_01", "user_unique_name": "user170383***6" }] | 返回的成员列表。 |
| total_count | Long | 5 | 空间中的成员总数。 |
### OpenSpaceMember
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| user_id | String | 2069720456***** | 用户的扣子用户 ID。 |
| role_type | String | member | 该用户在空间中的角色： <br>  <br> * admin：空间管理员。 <br> * member：空间成员。 <br> * owner：空间所有者。 |
| user_nickname | String | 小王 | 用户的昵称。 |
| user_unique_name | String | 小王 | 用户的名称。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
GET 'https://api.coze.cn/v1/workspaces/7515267805****/members' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer pat_FFTwAe4KB9GpaXCGoCjKbztWig4MT1SRB6yjP4n0PJ1AE3ukqw5yGj****'
```

### 返回示例
```JSON
{
  "code": 0,
  "data": {
    "total_count": 4,
    "items": [
      {
        "role_type": "owner",
        "user_id": "2135714797***",
        "user_nickname": "RootUser_210202***",
        "user_unique_name": "kou_testing_001",
        "avatar_url": "example.com/cloudidentity.9f6***.jpg"
      },
      {
        "avatar_url": "https://example.com/cloudidentity.9f6***.jpg",
        "role_type": "member",
        "user_id": "402688082103***",
        "user_nickname": "test_user_01",
        "user_unique_name": "user170383***6"
      },
      {
        "user_id": "260191467***",
        "user_nickname": "test_user_03",
        "user_unique_name": "user127938***",
        "avatar_url": "https://example.com/cloudidentity.9f6***.jpg",
        "role_type": "member"
      },
      {
        "user_id": "55242585801***",
        "user_nickname": "test_003",
        "user_unique_name": "",
        "avatar_url": "https://example.com/cloudidentity.9f6***.jpg",
        "role_type": "member"
      }
    ]
  },
  "msg": "",
  "detail": {
    "logid": "20250616153822E8D5D5F****"
  }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。