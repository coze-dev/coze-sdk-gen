# 创建声纹组
创建声纹组。
## 接口描述
声纹组是声纹的集合单元，例如，你可以为每个设备分别创建一个声纹组。声纹识别时，扣子编程会在指定的声纹组中匹配声纹。
默认最多可创建 1000 个声纹组。如需提高配额，需要升级至扣子企业旗舰版，并联系对应销售申请配额扩容。
## 基础信息
| **请求方式** | POST |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/audio/voiceprint_groups <br> ``` <br>  |
| **权限** | `createVoiceprintGroup` <br> 确保调用该接口使用的个人令牌开通了 `createVoiceprintGroup` 权限，详细信息参考[鉴权方式](https://docs.coze.cn/developer_guides/authentication)。 |
| **接口说明** | 创建声纹组。 |
## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $Access_Token | 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Body
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| name | String | 必选 | 音箱2的声纹组 | 声纹组的名称，用于标识和管理不同的声纹组，长度不超过 `20`个字符。 |
| desc | String | 可选 | 该声纹组用于客厅智能音箱的用户语音识别，支持多用户语音指令区分。 | 声纹组的描述信息，用于详细说明该声纹组的用途或特点，长度不超过 `500`个字符。 |
## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 调用状态码。0 表示调用成功，其他值表示调用失败，你可以通过 msg 字段判断详细的错误原因。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [CreateVoicePrintGroupData](#createvoiceprintgroupdata) | {"id":"426614174000"} | 声纹组的信息。 |
| detail | Object of [ResponseDetail](#responsedetail) | { "logid": "202410242028595CCF353CEC86A8*****" } | 响应详情信息。 |
### CreateVoicePrintGroupData
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| id | String | 426614174000 | 声纹组 ID。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://docs.coze.cn/guides/help_and_support)。 |
## 示例
### 请求示例
```JSON
curl --location --request POST 'https://api.coze.cn/v1/audio/voiceprint_groups' \
--header 'Authorization: Bearer pat_OYDacMzM3WyOWV3Dtj2bHRMymzxP****' \
--header 'Content-Type: application/json'
{
  "name": "音箱1的声纹组",
  "desc": "音箱1的声纹组"
}
```

### 返回示例
```JSON
{
  "msg": "",
  "detail": {
    "logid": "202506091635425D49D29729AD9****"
  },
  "data": {
    "id": "75138649771889***"
  },
  "code": 0
}
```

## 错误码
如果成功调用扣子编程的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://docs.coze.cn/developer_guides/coze_error_codes)文档查看对应的解决方法。