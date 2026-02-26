# 添加组织成员
添加组织成员。
## 接口描述
用户加入企业时将自动加入默认组织。[创建组织](https://docs.coze.cn/developer_guides/add_organization_people)后，你可以调用本 API 将员工和访客加入对应组织。
## 接口限制

* **套餐限制**：扣子企业版（企业标准版、企业旗舰版）。
* 用户加入组织前，需先将其添加至对应企业，具体请参见[添加企业成员](https://docs.coze.cn/developer_guides/add_enterprise_member)。
* 访客加入组织后，组织角色只能是访客。

* 每次请求只能添加一位成员。如需添加多位，请依次发送请求。
* 该 API 不支持并发请求。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/organizations/:organization_id/members <br> ``` <br>  |
| **权限** | `batchAddOrganizationPeople` <br> 确保调用该接口使用的访问令牌开通了**企业特权应用**中的 `batchAddOrganizationPeople` 权限，详细信息参考[OAuth JWT 授权（企业特权应用）](https://docs.coze.cn/developer_guides/oauth_jwt_privilege)。 |
| **接口说明** | 添加组织成员。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息参考[OAuth JWT 授权（企业特权应用）](https://docs.coze.cn/developer_guides/oauth_jwt_privilege)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| organization_id | String | 必选 | 7490888144456*** | 需要添加组织成员的组织 ID。  <br> 你可以在**组织管理** > **组织设置**页面查看对应的组织 ID，或通过[查询组织列表](https://docs.coze.cn/developer_guides/list_organization) API 查询组织 ID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/b57d5e293ff944d2b83efe8217d35517~tplv-goo7wpa0wc-image.image) |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| organization_people | Array of [OrganizationPeopleAddData](#organizationpeopleadddata) | 必选 | [{"user_id":"41135614833****","organization_role_type":"organization_admin"}] | 待添加至组织的成员列表。 |
### OrganizationPeopleAddData
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| user_id | String | 必选 | 41135614833**** | 需要添加至组织的扣子用户的 UID。 <br> 你可以调用火山引擎的 [ListCozeUser-成员列表](https://api.volcengine.com/api-docs/view?serviceCode=coze&version=2025-06-01&action=ListCozeUser) API，其中 `CozeUserId` 的值即为扣子用户的 UID。 |
| organization_role_type | String | 必选 | organization_admin | 设置用户在组织中的角色，枚举值： <br>  <br> * `organization_super_admin`：组织超级管理员。 <br> * `organization_admin`：组织管理员。 <br> * `organization_member`：组织成员。 <br> * `organization_guest`：访客。 |
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
curl --location --request POST 'https://api.coze.cn/v1/organizations/7490888144456***/members' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type : application/json' \
--data-raw '{
    "organization_people": [
        {
            "user_id": "41135614833****",
            "organization_role_type": "organization_admin"
        }
    ]
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