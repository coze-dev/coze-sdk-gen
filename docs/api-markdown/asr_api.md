# 双向流式语音识别
扣子编程提供流式语音识别 WebSocket OpenAPI，可以将指定的音频判断转为文字，支持识别中英文双语种。
双向流式语音识别场景下的各类事件详细信息可参考[双向流式语音识别事件](https://docs.coze.cn/api/open/docs/developer_guides/asr_event)。
## 接口信息
| **URL** | `wss://ws.coze.cn/v1/audio/transcriptions` |
| --- | --- |
| **Headers** | `Authorization Bearer `*`$Access_Token`* <br> 用于验证客户端身份的访问令牌。你可以在扣子编程中生成访问令牌，详细信息请参考[准备工作](https://www.coze.cn/docs/developer_guides/preparation)。 |
| **权限** | `createTranscription` |
| **接口说明** | 将指定的音频判断转为文字，支持识别中英文双语种。 |
## 建连示例代码

<div type="doc-tabs">
<div type="tab-item" title="ws module(Node.js)" key="H3NXw76qCz">


```JavaScript
import WebSocket from 'ws';

const url = `wss://ws.coze.cn/v1/audio/transcriptions?authorization=Bearer ${ACCESS_TOKEN}`;
const ws = new WebSocket(url);

ws.on('open', function open() {
  console.log('Connected to server.');
});

ws.on('message', function incoming(message) {
  console.log(JSON.parse(message.toString()));
});
```



</div>
<div type="tab-item" title="websocket-client(Python)" key="eiFzK04Gtw">

```Python
# example requires websocket-client library:
# pip install websocket-client

import os
import json
import websocket

ACCESS_TOKEN = os.environ.get("ACCESS_TOKEN")

url = "wss://ws.coze.cn/v1/audio/transcriptions"

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
<div type="tab-item" title="Websocket(browsers)" key="WvMzdhsWEn">


```JavaScript
const url = `wss://ws.coze.cn/v1/audio/transcriptions?authorization=Bearer ${ACCESS_TOKEN}`;
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
![Image](https://p9-arcosite.byteimg.com/tos-cn-i-goo7wpa0wc/3ab3d5338c8543a9abc9acf96ddf74e5~tplv-goo7wpa0wc-image.image)


