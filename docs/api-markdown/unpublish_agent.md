# 下架智能体
智能体发布后，你可以调用本 API 从扣子编程官方渠道及自定义渠道下架智能体。
## 接口限制

* 仅智能体所有者和自定义渠道的所有者可以下架智能体。
* 暂不支持调用本 API 下架豆包渠道的智能体。

## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/bots/:bot_id/unpublish <br> ``` <br>  |
| **权限** | `disconnectBot`、`Connector.disconnectBot` <br> 确保调用该接口使用的访问令牌开通了 `disconnectBot` 、`Connector.disconnectBot`权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 从扣子编程官方渠道及自定义渠道下架智能体。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Path
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| bot_id | String | 必选 | 73428668***** | 待下架的智能体的 ID。你可通过智能体开发页面 URL 中的 `bot` 参数获取智能体 ID 。例如 URL 为 `https://www.coze.com/space/341****/bot/73428668*****` 时，智能体 ID 为 `73428668*****`。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| connector_id | String | 必选 | 1024 | 待下架的渠道 ID。支持下架如下发布渠道中的智能体： <br>  <br> * 1024：API 渠道。 <br> * 999：Chat SDK。 <br> * 10000122：扣子商店。 <br> * 10000113：微信客服。 <br> * 10000120：微信服务号。 <br> * 10000121：微信订阅号。 <br> * 10000126：抖音小程序。 <br> * 10000127：微信小程序。 <br> * 10000011：飞书。 <br> * 10000128：飞书多维表格。 <br> * 10000117：掘金。 <br> * 自定义渠道 ID。自定义渠道 ID 的获取方式如下：在扣子编程左下角单击头像，在**账号设置** > **发布渠道** > **企业自定义渠道管理**页面查看渠道 ID。 |
| unpublish_reason | String | 可选 | 终止服务 | 下架渠道的原因说明，用于记录或说明为何执行下架操作。最大支持 `1024`个字符。 <br> 该原因会在扣子编程**工作空间** > **发布管理**页面对应智能体的渠道右侧展示，建议提供清晰易懂的下架原因。 <br> ![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/a6f846fc44184e02a2e2c579ec27e10c~tplv-goo7wpa0wc-image.image) |
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
curl --location --request POST 'https://api.coze.cn/v1/bots/73428668*****/unpublish' \
--header 'Authorization:  Bearer pat_xitq9LWlowpX3qGCih1lwpAdzvXNqgmpfhpV28HLWFN****' \
--header 'Content-Type: application/json' \
--data-raw '{
    "connector_id": "1024",
    "unpublish_reason": "终止服务"
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
