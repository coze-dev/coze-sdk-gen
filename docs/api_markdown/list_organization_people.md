# 查询组织成员列表
查询指定组织中的成员列表。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/organizations/:organization_id/members <br> ``` <br>  |
| **权限** | `Account.listOrganizationPeople` <br> 确保调用该接口使用的访问令牌开通了**企业特权应用**中的 `Account.listOrganizationPeople` 权限，详细信息参考[OAuth JWT 授权（企业特权应用）](https://docs.coze.cn/developer_guides/oauth_jwt_privilege)。 |
| **接口说明** | 查询组织成员列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息参考[OAuth JWT 授权（企业特权应用）](https://docs.coze.cn/developer_guides/oauth_jwt_privilege)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| organization_id | String | 必选 | 7559861372637*** | 需要查询组织成员列表的组织 ID。  <br> 你可以在**组织管理** > **组织设置**页面查看对应的组织 ID，或通过[查询组织列表](https://docs.coze.cn/developer_guides/list_organization) API 查询组织 ID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/b57d5e293ff944d2b83efe8217d35517~tplv-goo7wpa0wc-image.image) |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_num | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即返回第一页数据。 |
| page_size | Integer | 可选 | 20 | 每页返回的数据条数，用于分页查询。默认值为 `20`，最大支持 `50`条。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [ListOrganizationPeopleData](#listorganizationpeopledata) | {"items":[{"user_id":"41147914833****","is_valid":true,"avatar_url":"https://example.com/avatar/41147914833****.jpg","created_at":1715000000,"people_type":"employee","user_nickname":"John","user_unique_name":"John_123","organization_role_type":"organization_admin"}],"total_count":1} | 组织成员列表，包含成员详细信息和总数。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的日志信息，用于问题排查和技术支持。 |
### ListOrganizationPeopleData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [OrganizationPeople](#organizationpeople) | [{"user_id":"41147914833****","is_valid":true,"avatar_url":"https://example.com/avatar/41147914833****.jpg","created_at":1715000000,"people_type":"employee","user_nickname":"John","user_unique_name":"John_123","organization_role_type":"organization_admin"}] | 组织成员列表。 |
| total_count | Long | 100 | 组织中的成员数量。 |
### OrganizationPeople
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| user_id | String | 41147914833**** | 扣子用户的 UID。 |
| is_valid | Boolean | true | 标识该组织成员账号状态是否正常。 <br>  <br> * `true`：账号状态正常，可正常使用。 <br> * `false`：该成员的火山账号已被删除，不可使用。 |
| avatar_url | String | https://example.com/avatar/41147914833****.jpg | 用户头像的 URL 地址。 |
| created_at | Long | 1715000000 | 用户加入组织的时间戳，格式为 10 位的 Unix 时间戳，单位为秒。 |
| people_type | String | employee | 组织成员在企业中的身份类型，枚举值： <br>  <br> * `employee`：企业员工。 <br> * `guest`：访客。 |
| user_nickname | String | John | 用户昵称。 |
| user_unique_name | String | user5916927*** | 扣子用户名。 |
| organization_role_type | String | organization_admin | 组织成员在组织中的角色类型，枚举值： <br>  <br> * `organization_super_admin`：组织超级管理员。 <br> * `organization_admin`：组织管理员。 <br> * `organization_member`：组织成员。 <br> * `organization_guest`：访客。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/organizations/7559861372637***/members?page_num=1&page_size=20' \
--header 'Authorization: Bearer pat_O******' \
--header 'Content-Type: application/json' \
```

### 返回示例
```JSON
{
    "code": 0,
    "msg": "",
    "data": {
        "items": [
            {
                "user_id": "41147914833****",
                "is_valid": true,
                "avatar_url": "https://example.com/avatar/41147914833****.jpg",
                "created_at": 1715000000,
                "people_type": "employee",
                "user_nickname": "John",
                "user_unique_name": "John_123",
                "organization_role_type": "organization_admin"
            }
        ],
        "total_count": 1
    },
    "detail": {
        "logid": "20241210152726467C48D89D6DB2****"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。