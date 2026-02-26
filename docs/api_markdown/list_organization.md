# 查询组织列表
查询指定企业中的组织列表。
## 接口限制
**套餐限制**：扣子企业版（企业标准版、企业旗舰版）。
## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/enterprises/:enterprise_id/organizations <br> ``` <br>  |
| **权限** | `Enterprise.listOrganization` <br> 确保调用该接口使用的访问令牌开通了**企业特权应用**中的 `Enterprise.listOrganization` 权限，详细信息参考[OAuth JWT 授权（企业特权应用）](https://docs.coze.cn/developer_guides/oauth_jwt_privilege)。 |
| **接口说明** | 查询指定企业中的组织列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息参考[OAuth JWT 授权（企业特权应用）](https://docs.coze.cn/developer_guides/oauth_jwt_privilege)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| enterprise_id | String | 必选 | volcano_210195*** | 企业 ID，查询组织列表所属的企业。 <br> 你可以在**组织管理** > **组织设置**页面查看企业 ID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/02db2078f0c84bc2aa189f5cca93d49d~tplv-goo7wpa0wc-image.image) |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_num | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即返回第一页数据。 |
| page_size | Integer | 可选 | 20 | 每页返回的数据条数，用于分页查询。默认值为 20，最大值为 20。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [ListOrganizationData](#listorganizationdata) | {"items":[{"organization_id":"7490888144456***","organization_name":"默认组织","organization_icon_url":"https://example.com/organization_icons/7490888144456***.png","is_default_organization":true},{"organization_id":"7490888144457***","organization_name":"市场部","organization_icon_url":"https://example.com/organization_icons/7490888144457***.png","is_default_organization":false}],"total_count":2} | 包含符合查询条件的企业组织列表及其数量。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含本次请求的详细信息，主要用于异常报错场景下的问题排查。 |
### ListOrganizationData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| items | Array of [OrganizationBasicInfo](#organizationbasicinfo) | [{"organization_id":"7490888144456***","organization_name":"默认组织","organization_icon_url":"https://example.com/organization_icons/7490888144456***.png","is_default_organization":true},{"organization_id":"7490888144457***","organization_name":"市场部","organization_icon_url":"https://example.com/organization_icons/7490888144457***.png","is_default_organization":false}] | 企业下的组织列表，包含每个组织的详细信息。 |
| total_count | Long | 2 | 企业中的组织数量。 |
### OrganizationBasicInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| organization_id | String | 7490888144456*** | 组织 ID。 |
| organization_name | String | 市场部 | 组织名称。 |
| organization_icon_url | String | https://example.com/organization_icons/7490888144456***.png | 组织的头像 URL 地址。 |
| is_default_organization | Boolean | true | 标识当前组织是否为默认组织。 <br>  <br> * `true` ：默认组织。 <br> * `false`：自定义组织。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request GET 'https://api.coze.cn/v1/enterprises/volcano_210195***/organizations?page_num=1&page_size=20' \
--header 'Authorization : Bearer pat_O******' \
--header 'Content-Type : application/json' \
```

### 返回示例
```JSON
{
    "code": 0,
    "msg": "",
    "data": {
        "items": [
            {
                "organization_id": "7490888144456***",
                "organization_name": "默认组织",
                "organization_icon_url": "https://example.com/organization_icons/7490888144456***.png",
                "is_default_organization": true
            },
            {
                "organization_id": "7490888144457***",
                "organization_name": "市场部",
                "organization_icon_url": "https://example.com/organization_icons/7490888144457***.png",
                "is_default_organization": false
            }
        ],
        "total_count": 2
    },
    "detail": {
        "logid": "20241210152726467C48D89D6DB2****"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。