# 添加发布平台
为指定空间添加一个发布渠道。
## 接口说明
在扣子编程中发布智能体或应用时，可发布的渠道范围与工作空间有关。每个工作空间均配置了官方默认发布渠道，例如扣子商店、豆包、API 、SDK 等。除此之外，工作空间中还可以手动添加公共渠道和企业自定义渠道，按需拓展 AI 项目的分发渠道。添加渠道后，空间中的每个开发者都可以将自己的 AI 项目发布到这些渠道中。
* 添加发布渠道之前，需要获取发布渠道的渠道 ID。
   * 企业自定义渠道：在**我的**>**设置**>**发布渠道**>**我的渠道管理**页面查看当前登录用户已创建的渠道列表，列表中可查看渠道 ID。企业自定义渠道入驻扣子编程的方式可参考[渠道入驻](https://docs.coze.cn/dev_how_to_guides/channel_integration_overview)文档。
   * 公开渠道：联系公开平台的管理员获取渠道 ID。
* 扣子企业版（企业标准版、企业旗舰版）中，仅**超级管理员和管理员**能添加企业的公共渠道和自定义渠道，成员只能给个人空间添加公共渠道和自定义渠道。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/connectors/:connector_id/install <br> ``` <br>  |
| **权限** | `installConnector` <br> 确保调用该接口使用的访问令牌开通了`installConnector`权限，详细信息参考[鉴权](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 为指定空间添加一个发布渠道。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| connector_id | String | 必选 | 745264392345****** | 渠道 ID。 <br>  <br> * 企业自定义渠道：在**我的**>**设置**>**发布渠道**>**我的渠道管理**页面查看当前登录用户已创建的渠道列表，列表中可查看渠道 ID。企业自定义渠道入驻扣子编程的方式可参考[渠道入驻](https://docs.coze.cn/dev_how_to_guides/channel_integration_overview)文档。 <br> * 公开渠道：联系公开平台的管理员获取渠道 ID。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| workspace_id | String | 必选 | 736163827687053**** | 需要添加新渠道的工作空间 ID。空间 ID 是空间的唯一标识。 <br> 进入指定空间，空间页面 URL 中 `w` 参数后的数字就是 工作空间 ID。例如`https://code.coze.cn/w/75814654762959***/projects`，工作空间 ID 为 `75814654762959***`。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。  <br>  <br> * 0 表示调用成功。  <br> * 其他值表示调用失败。你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "202410242028595CCF353CEC86A8*****" } | 本次请求的详细信息。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
安装渠道
```JSON
curl --location --request POST 'https://api.coze.cn/v1/connectors/745264392345******/install' \
--header 'Authorization : Bearer pat_Osa******' \
--header 'Content-Type : application/json' \
--data-raw '{
    "workspace_id": "736163827687053****"
}'
```

### 返回示例
```JSON
{
  "detail": {
    "logid": "20241029152003BC531DC784F1897B****"
  },
  "code": 0,
  "msg": ""
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。
