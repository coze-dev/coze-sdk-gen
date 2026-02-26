# 双向流式语音对话
扣子编程提供流式语音对话 WebSocket OpenAPI，向指定的智能体发起语音对话。
双向流式语音对话场景下的各类事件详细信息可参考[双向流式对话上行事件](https://docs.coze.cn/api/open/docs/developer_guides/streaming_chat_event)。
## 接口信息
| **URL** | `wss://ws.coze.cn/v1/chat` |
| --- | --- |
| **Headers** | `Authorization Bearer `*`$Access_Token`* <br> 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息请参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| **权限** | `chat` |
| **接口说明** | 向指定的智能体发起语音对话。 |
### Query
| **参数** | **类型** | **是否必选** | **说明** |
| --- | --- | --- | --- |
| bot_id <br>  | String  <br>  | 必选 <br>  | 需要关联的智能体 ID。 <br> 进入智能体的开发页面，开发页面 URL 中 `bot` 参数后的数字就是智能体 ID。例如 `https://www.coze.com/space/341****/bot/73428668*****`，Bot ID 为 `73428668*****`。  <br> * 确保调用该接口使用的令牌开通了此智能体所在空间的权限。 <br> * 确保该智能体已发布为 API 服务。 <br>  |
| device_id | String | 可选 | 设备的唯一标识符，在建立 Websocket 连接时建议带上此参数，便于排查问题。 <br> `device_id` 的格式为 **int64 数字类型**的字符串。如果设备 ID 是纯数字，  可直接填写该数字作为 `device_id`。如果设备 ID 包含非数字字符，则需先将其转换为纯数字字符串，再填写到 `device_id` 中。 <br>  |
## 建连示例代码

<div type="doc-tabs">
<div type="tab-item" title="ws module(Node.js)" key="GRBLKftgkf">


```JavaScript
import WebSocket from 'ws';

const url = `wss://ws.coze.cn/v1/chat?bot_id=${BOT_ID}&authorization=Bearer ${ACCESS_TOKEN}`;
const ws = new WebSocket(url);

ws.on('open', function open() {
  console.log('Connected to server.');
});

ws.on('message', function incoming(message) {
  console.log(JSON.parse(message.toString()));
});
```



</div>
<div type="tab-item" title="websocket-client(Python)" key="Yha2n70XFV">

```Python
# example requires websocket-client library:
# pip install websocket-client

import os
import json
import websocket

ACCESS_TOKEN = os.environ.get("ACCESS_TOKEN")

url = "wss://ws.coze.cn/v1/chat?bot_id=73791654286875***"

headers = [
    "Authorization: Bearer " + ACCESS_TOKEN
]

def on_open(ws):
    print("Connected to server.")

def on_message(ws, message):
    data = json.loads(message)
    print("Received event:", json.dumps(data, indent=2))

ws = websocket.WebSocketApp(
    url,
    header=headers,
    on_open=on_open,
    on_message=on_message,
)

ws.run_forever()
```


</div>
<div type="tab-item" title="Websocket(browsers)" key="OWqBygoY0p">


```JavaScript
const url = `wss://ws.coze.cn/v1/chat?bot_id=${BOT_ID}&authorization=Bearer ${ACCESS_TOKEN}`;
const ws = new WebSocket(url);

ws.addEventListener('open', function () {
  console.log('Connected to server.');
});

ws.addEventListener('message', function (message) {
  console.log(JSON.parse(message.data.toString()));
});
```



</div>
</div>
## API 时序图
交互流程如下：
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/7f2c1f5256ad4a9aaffdc6dbcea1daa7~tplv-goo7wpa0wc-image.image)

