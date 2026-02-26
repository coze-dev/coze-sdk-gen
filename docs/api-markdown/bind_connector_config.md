# 绑定设备
将设备与自定义渠道绑定。
## 接口描述
硬件厂商可以调用本 API 将企业内的硬件设备与企业的自定义渠道绑定，当开发者发布智能体到该自定义渠道时，在发布配置页面的设备列表中选择对应的设备，即可快速发布到对应设备。
支持批量绑定多台设备。
创建自定义渠道后，默认未开启 API 绑定设备的能力，如果需要调用本 API，你需要将自定义渠道的渠道 ID 提供给扣子商务经理，申请开通渠道设备绑定的能力，并由商务经理配置设备绑定的 key。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/connectors/:connector_id/user_configs <br> ``` <br>  |
| **权限** | `bindConnectorConfig` <br> 确保调用该接口使用的访问令牌开通了 `bindConnectorConfig` 权限，详细信息参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| **接口说明** | 将设备与自定义渠道绑定。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer *$Access_Token* | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| connector_id | String | 必选 | 745264392345****** | 企业自定义渠道的渠道 ID。 <br> 自定义渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。自定义渠道入驻扣子编程的方式可参考[渠道入驻](https://docs.coze.cn/dev_how_to_guides/channel_integration_overview)文档。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| configs | Array of [UserConfig](#userconfig) | 必选 | [{"key":"device_id","enums":[{"label":"设备123","value":"1237824***"}]}] | 设备配置信息列表。 |
| user_id | String | 可选 | 4497462571*** | 智能体开发者的扣子用户 UID，用于标识设备绑定的扣子用户。 <br> 你可以单击扣子编程左下角的头像，选择**账号设置**，在页面底部查看扣子用户的 UID。 |
### UserConfig
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| key | String | 必选 | device_id | 参数的键值，需要与提供给商务经理的 `key` 保持一致，用于标识设备配置信息的具体类型。 |
| enums | Array of [UserConfigEnum](#userconfigenum) | 必选 | [{"label":"设备123","value":"1237824***"}] | 设备信息的数组，可以批量添加该渠道中的所有设备。 |
### UserConfigEnum
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| label | String | 必选 | 智能音箱-客厅 | 设备的显示名称，用于在用户界面中标识设备，方便用户识别和管理。 |
| value | String | 必选 | 51237824*** | 设备的 ID。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| detail | Object of [ResponseDetail](#responsedetail) | {"logid":"20241210152726467C48D89D6DB2****"} | 包含请求的详细信息的对象，主要用于记录请求的日志 ID 以便于排查问题。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/connectors/745264392345******/user_configs' \
--header 'Authorization: Bearer pat_xitq****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "configs": [
        {
            "key": "device_id",
            "enums": [
                {
                    "label": "智能音箱-客厅",
                    "value": "51237824***"
                },
                {
                    "label": "智能音箱-主卧",
                    "value": "51237825***"
                }
            ]
        }
    ]
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
