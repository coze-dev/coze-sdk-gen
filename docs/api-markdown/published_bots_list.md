# 【Deprecated】查看已发布智能体列表
查看指定空间发布到 Agent as API 渠道的智能体列表。
* 该 API 即将下线，建议替换为新版的 API [ 查看智能体列表](https://www.coze.cn/open/docs/developer_guides/bots_list_draft_published)。
* 此接口仅支持查看已发布为 API 服务的智能体列表。对于创建后从未发布到 API 渠道的 Bot，可以在[扣子平台](https://www.coze.cn/)中查看列表及配置。

## 基础信息
| **请求方式** | GET |
| --- | --- |
| **请求地址** | ```Plain Text <br> https://api.coze.cn/v1/space/published_bots_list <br> ``` <br>  |
| **权限** | `getPublishedBot` <br> 确保调用该接口使用的访问令牌开通了 `getPublishedBot` 权限，详细信息参考[鉴权方式](https://www.coze.cn/docs/developer_guides/authentication)。 |
| **接口说明** | 调用接口查看指定空间发布到 Agent as API 渠道的智能体列表。 |

## 请求参数
### Header
| **参数** | **取值** | **说明** |
| --- | --- | --- |
| Authorization | Bearer $AccessToken | 用于验证客户端身份的访问令牌。你可以在扣子平台中生成访问令牌，详细信息，参考[准备工作](https://docs.coze.cn/api/open/docs/docs/developer_guides/preparation)。 |
| Content-Type | application/json | 解释请求正文的方式。 |
### Query
| **参数** | **类型** | **是否必选** | **示例** | **说明** |
| --- | --- | --- | --- | --- |
| space_id | String | 必选 | 737620236629291**** | 智能体所在的空间的 Space ID。Space ID 是空间的唯一标识。 <br> 进入指定空间，空间页面 URL 中 `space` 参数后的数字就是 Space ID。例如`https://www.coze.cn/space/736163827687053****/bot`，Space ID `为736163827687053****`。 |
| page_index | Integer | 可选 | 1 | 分页查询时的页码。默认为 1，即从第一页数据开始返回。 |
| page_size | Integer | 可选 | 10 | 分页大小。默认为 20，即每页返回 20 条数据。 |

## 返回参数
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| code | Long | 0 | 状态码。 <br> `0`代表调用成功。 |
| msg | String | "" | 状态信息。API 调用失败时可通过此字段查看详细错误信息。 <br> 状态码为 0 时，msg 默认为空。 |
| data | Object of [SpacePublishedBotsInfo](#spacepublishedbotsinfo) | { "space_bots": [ { "bot_id": "737965697151719****", "bot_name": "图片生成文本", "description": "根据图片生成文本", "icon_url": "https://lf3-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_ICON/default_bot_icon3.png?lk3s=50ccb0c5&x-expires=1718343510&x-signature=Y2dTjqx6Oa1RtevCZPe2***", "publish_time": "1718212388" }, { "bot_id": "737946218936519****", "bot_name": "当代毕加索", "description": "根据用户描述自动生成毕加索油画风格的图片", "icon_url": "https://lf26-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_ICON/2667858400903179_17181861478***.jpeg?lk3s=50ccb0c5&x-expires=1718343510&x-signature=2FQ6%2FHmyswKDBdeTR***", "publish_time": "1718209964" } ], "total": 2 } | 智能体列表及总数信息。 |
| detail | Object of [ResponseDetail](#responsedetail) |  <br> { "logid": "20250106172024B5F607030EFFAD653960" } | 返回详情。 |
### SpacePublishedBotsInfo
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| total | Integer | 5 | 智能体列表中的智能体总数。 |
| space_bots | Array of [SpacePublishedBots](#spacepublishedbots) | [ { "bot_id": "737965697151719****", "bot_name": "图片生成文本", "description": "根据图片生成文本", "icon_url": "https://lf3-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_ICON/default_bot_icon3.png?lk3s=50ccb0c5&x-expires=1718343510&x-signature=Y2dTjqx6Oa1RtevCZP***", "publish_time": "1718212388" }, { "bot_id": "737946218936519****", "bot_name": "当代毕加索", "description": "根据用户描述自动生成毕加索油画风格的图片", "icon_url": "https://lf26-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_ICON/2667858400903179_17181861***.jpeg?lk3s=50ccb0c5&x-expires=1718343510&x-signature=2FQ6%2FHmyswKDBdeTROTO***", "publish_time": "1718209964" } ] | 指定空间发布到 Agent as API 渠道的智能体列表。 |
### SpacePublishedBots
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| bot_id | String | 737965697151719**** | 智能体的唯一标识。 |
| bot_name | String | 图片生成文本 | 智能体的名称。 |
| description | String | 根据图片生成文本 | 智能体的描述。 |
| icon_url | String | https://lf3-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_ICON/default_bot_icon3.png?lk3s=50ccb0c5&x-expires=1718343510&x-signature=Y2dTjqx6Oa1RtevCZPe27k4RKbs%3D | 智能体的头像。 |
| publish_time | String | 1718212388 | 智能体的最近一次发布时间，格式为 10 位的 Unixtime 时间戳。此 API 返回的智能体列表按照此字段降序排列。 |
### ResponseDetail
| **参数** | **类型** | **示例** | **说明** |
| --- | --- | --- | --- |
| logid | String | 20241210152726467C48D89D6DB2**** | 本次请求的日志 ID。如果遇到异常报错场景，且反复重试仍然报错，可以根据此 logid 及错误码联系扣子团队获取帮助。详细说明可参考[获取帮助和技术支持](https://www.coze.cn/docs/guides/help_and_support)。 |

## 示例
### 请求示例
```JSON
GET https://api.coze.cn/v1/space/published_bots_list?space_id=737620236629291****
```

### 返回示例
```JSON
{
  "code": 0,
  "msg": "",
  "data": {
    "space_bots": [
      {
        "bot_id": "737965697151719****",
        "bot_name": "图片生成文本",
        "description": "根据图片生成文本",
        "icon_url": "https://lf3-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_ICON/default_bot_icon3.png?lk3s=50ccb0c5&x-expires=1718343510&x-signature=Y2dTjqx6Oa1RtevCZPe27k4RKbs%3D",
        "publish_time": "1718212388"
      },
      {
        "bot_id": "737946218936519****",
        "bot_name": "当代毕加索",
        "description": "根据用户描述自动生成毕加索油画风格的图片",
        "icon_url": "https://lf26-appstore-sign.oceancloudapi.com/ocean-cloud-tos/FileBizType.BIZ_BOT_ICON/2667858400903179_1718186147899963557.jpeg?lk3s=50ccb0c5&x-expires=1718343510&x-signature=2FQ6%2FHmyswKDBdeTROTOn0IYyrA%3D",
        "publish_time": "1718209964"
      }
    ],
    "total": 2
  }
}
```

## 错误码
如果成功调用扣子的 API，返回信息中 code 字段为 0。如果状态码为其他值，则表示接口调用失败。此时 msg 字段中包含详细错误信息，你可以参考[错误码](https://www.coze.cn/docs/developer_guides/coze_error_codes)文档查看对应的解决方法。

