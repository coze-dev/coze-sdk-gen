# 修改组织成员角色
修改组织成员的角色。
组织成员角色包括组织超级管理员、组织管理员、组织成员和访客。你可以通过本 API 修改组织成员的角色。
## 接口限制

* 企业外部用户的角色只能是访客，不支持修改。
* 企业员工的角色不能设置为访客。
* 组织管理员不能将其他人设为组织超级管理员，只有组织超级管理员创建的访问令牌能设置其他人为组织超级管理员。

## 基础信息
| **请求方式** | PUT |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/organizations/:organization_id/members/:user_id <br> ``` <br>  |
| **权限** | `updateOrganizationPeople` <br> 确保调用该接口使用的访问令牌开通了**企业特权应用**中的 `updateOrganizationPeople` 权限，详细信息参考[OAuth JWT 授权（企业特权应用）](https://docs.coze.cn/developer_guides/oauth_jwt_privilege)。 |
| **接口说明** | 修改组织成员的角色。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息参考[OAuth JWT 授权（企业特权应用）](https://docs.coze.cn/developer_guides/oauth_jwt_privilege)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| organization_id | String | 必选 | 7490888144456*** | 待修改组织成员的组织 ID。  <br> 你可以在**组织管理** > **组织设置**页面查看对应的组织 ID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/b57d5e293ff944d2b83efe8217d35517~tplv-goo7wpa0wc-image.image) |
| user_id | String | 必选 | 41135614833**** | 待修改角色的组织成员的 UID。 <br> 你可以调用[查询组织成员列表](https://docs.coze.cn/developer_guides/list_organization_people) API，查看组织成员的扣子用户 UID。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| organization_role_type | String | 必选 | organization_member | 设置用户在组织中的角色，枚举值： <br>  <br> * `organization_super_admin`：组织超级管理员。 <br> * `organization_admin`：组织管理员。 <br> * `organization_member`：组织成员。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的详细信息，主要用于异常报错场景下的问题排查。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request PUT 'https://api.coze.cn/v1/organizations/74908881444562***/members/41135614833****' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type : application/json' \
--data-raw '{
    "organization_role_type": "organization_member"
}'
```

### 返回示例
```JSON
{
    "code": 0,
    "msg": "",
    "detail": {
        "logid": "20241210152726467C48D89D6DB2****"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。