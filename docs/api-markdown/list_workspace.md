# 查看空间列表
查看当前扣子用户或指定扣子用户创建或加入的空间列表。
## 接口描述

* 你可以查看当前账号创建和加入的工作空间，包括个人账号的工作空间和企业账号的工作空间。
* 在扣子企业版中，你可以在 `Query` 中指定 `enterprise_id`，查看该企业当前账号创建或加入的所有工作空间。
* 你可以在 `Query` 中同时指定 `user_id` 和 `coze_account_id`，查看指定用户在指定组织下加入的工作空间。这里的组织可以是企业中的某个组织，也可以是个人版账号。

不同场景的 Query 参数配置说明如下：
| **场景** | **Query 参数** |  |  |
| --- | --- | --- | --- |
|  | **enterprise_id** | **user_id** | **coze_account_id** |
| 查看当前账号的所有工作空间（含个人空间） | 不传 | 不传 | 不传 |
| 查看当前账号在指定企业中的工作空间（不含个人空间） | 必选 | 不传 | 不传 |
| 查看指定用户在指定组织的工作空间（不含个人空间） | 不传 | 必选 | 必选 |
暂不支持[第三方渠道应用](https://docs.coze.cn/dev_how_to_guides/channel_integration_overview)调用此 API。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/workspaces <br> ``` <br>  |
| **权限** | `listWorkspace` <br> 确保调用该接口使用的访问令牌开通了`listWorkspace`权限，详细信息参考[鉴权](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 查看当前扣子用户或指定扣子用户创建或加入的空间列表。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| page_num | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即返回第一页数据。 |
| page_size | Integer | 可选 | 30 | 分页大小，即每页返回多少个工作空间。默认为 20，最大为 50。 |
| enterprise_id | String | 可选 | volcano_2105850*** | 企业 ID，你可以在**组织管理** > **组织设置**页面查看企业 ID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/02db2078f0c84bc2aa189f5cca93d49d~tplv-goo7wpa0wc-image.image) <br>  <br> * 如果指定 `enterprise_id`，则查询对应 `EnterpriseId` 中，当前账号创建或加入的所有工作空间。 <br> * 如果不指定 `enterprise_id`，则查询当前账号创建和加入的所有工作空间，包括个人账号的工作空间和企业/团队账号的工作空间。 <br>  <br> 仅扣子企业版套餐中的用户支持该参数。个人免费版和个人进阶版不支持该参数。 <br>  |
| user_id | String | 可选 | 38509118307*** | 扣子用户 ID，用于查询特定用户的工作空间信息。 <br> 你可以单击扣子编程左下角的头像，选择**账号设置**，在页面底部查看扣子用户的 UID。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/c77171908fad4b69a328dfd3210b3576~tplv-goo7wpa0wc-image.image) <br> 如果同时指定 `user_id` 和 `coze_account_id`，则查询该用户在该组织下加入的工作空间。暂时不支持仅指定`user_id` 或仅指定`coze_account_id`。 <br>  |
| coze_account_id | String | 可选 | 38509118307*** | 扣子编程的组织 ID，用于查询特定组织的工作空间信息。 <br>  <br> * 扣子个人版中，`coze_account_id`  即为 `user_id`。 <br> * 扣子企业版中，`coze_account_id` 为组织 ID。你可以在**组织管理** > **组织设置**页面查看对应的组织 ID，或通过[查询组织列表](https://docs.coze.cn/developer_guides/list_organization) API 查询组织 ID。 <br>  <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/0aa86f3e0957493d82b3f534a2800fcc~tplv-goo7wpa0wc-image.image) <br> 如果同时指定 `user_id` 和 `coze_account_id`，则查询该用户在该组织下加入的工作空间。暂时不支持仅指定`user_id` 或仅指定`coze_account_id`。 <br>  |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。  <br>  <br> * 0 表示调用成功。  <br> * 其他值表示调用失败。你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [OpenSpaceData](#openspacedata) | \ | 接口响应的业务信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"1234567890abcdef"} | 包含请求的详细日志信息，用于问题排查和调试。 |
### OpenSpaceData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| workspaces | Array of [OpenSpace](#openspace) | \ | 用户创建或加入的空间列表。 |
| total_count | Long | 2 | 用户加入的空间总数。 |
### OpenSpace
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 74876004423701**** | 空间 ID。 |
| name | String | test | 空间名称。 |
| icon_url | String | https://***/obj/ocean-cloud-tos/FileBizType.BIZ_BOT_SPACE/team.png | 空间图标的 url 地址。 |
| owner_uid | String | 24787743932*** | 空间所有者的扣子用户 UID。 |
| role_type | String | member | 用户在空间中的角色。枚举值包括： <br>  <br> * owner：所有者 <br> * admin：管理员 <br> * member：成员 |
| admin_uids | Array of String | ["24787743932***","24787743932***"] | 空间管理员的用户 UID 列表，用于标识当前空间的所有管理员。 |
| description | String | 这是一个用于测试的空间。 | 空间的描述信息。 |
| enterprise_id | String | volcano_2105850*** | 企业 ID。 <br> 个人账号下的工作空间，返回的企业 ID 为空。 <br>  |
| joined_status | String | joined | 用户在空间中的加入状态。枚举值： <br>  <br> * `joined`：已加入 |
| workspace_type | String | team | 空间类型。枚举值包括： <br>  <br> * `personal`：个人空间 <br> * `team`：工作空间 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例

<div type="doc-tabs">
<div type="tab-item" title="查询当前账号加入的所有空间" key="B1_s7znamw3PS7GpgY8ef">

```JSON
GET 'https://api.coze.cn/v1/workspaces?&page_num=1&page_size=20' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
```


</div>
<div type="tab-item" title="查询指定用户在指定组织下加入的空间" key="wu72ntdUh79eVOOSFKs0n">

```JSON
GET 'https://api.coze.cn/v1/workspaces?user_id=38509118307***&coze_account_id=74867411766917***' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
```


</div>
<div type="tab-item" title="查看当前账号在指定企业中的工作空间" key="el2xxt6eK5etYbGnT3XTZ">

```JSON
GET 'https://api.coze.cn/v1/workspaces?enterprise_id=volcano_2101955**' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json' \
```


</div>
</div>
### 返回示例
```JSON
{
    "data": {
        "workspaces": [
            {
                "id": "74876004423701****",
                "name": "test",
                "icon_url": "https://***/obj/ocean-cloud-tos/FileBizType.BIZ_BOT_SPACE/team.png",
                "role_type": "member",
                "enterprise_id": "volcano_2105850***",
                "workspace_type": "team"
            }
            {
                "id": "74879061161065***",
                "name": "个人空间",
                "icon_url": "https://***/obj/ocean-cloud-tos/FileBizType.BIZ_BOT_SPACE/team.png",
                "role_type": "owner",
                "enterprise_id": "",
                "workspace_type": "personal"
            }
        ],
        "total_count": 2
    },
    "code": 0,
    "msg": "",
    "detail": {
        "logid": "1234567890abcdef****"
    }
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
