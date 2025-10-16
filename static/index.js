
function init() {
    const form = document.querySelector('#login-form');
    form.addEventListener('submit', async (event) => {
        event.preventDefault();

        const formData = new FormData(form);
        const url = "/v1/user/login";

        try {
            const response = await fetch(url, {
                method: 'POST',
                body: formData,
            });

            if (!response.ok) {
                throw new Error(`HTTP 错误! 状态码: ${response.status}`);
            }

            const result = await response.json();
            if (result["code"] != 0) {
                alert("登录失败：", result["msg"])
                return;
            }
            
            window.login_data = {
                "token": result["token"],
                "userId": result["userId"],
                "name": result["name"],
            }

            document.querySelector("#login-view").style.display = "none";
            document.querySelector("#online-user-view").style.display = "";

            initWebSocket();

        } catch (error) {
            let msg = 'Fetch 错误:' + error;
            alert("登录失败", msg)
        }
    });

    var coinList = document.getElementById("online-user-view");
    coinList.onclick = (e) => {
        // console.log("onclick coin list");
    
        var ele = document.elementFromPoint(e.x, e.y);
        var name = ele.textContent;
        var userId = ele.getAttribute("user-id");
        if (!userId) {
            return;
        }

        window.other_user = {
            "name": name,
            "user_id": Number(userId),
        }
        console.log(name, userId);

        document.querySelector(".lite-chatmaster").style.display = "";
    };

    initTestUser();

}

init();

async function initTestUser() {
    const url = "/v1/user/test_user";

    try {
        const response = await fetch(url);

        if (!response.ok) {
            throw new Error(`HTTP 错误! 状态码: ${response.status}`);
        }

        const result = await response.json();
        if (result["code"] != 0) {
            alert("获取测试账号失败：" + result["msg"])
            return;
        }
        
        var name = result["name"];
        var password = result["password"];

        document.querySelector("#login-name").value = name;
        document.querySelector("#login-password").value = password;

    } catch (error) {
        let msg = 'Fetch 错误:' + error;
        alert(msg)
    }
}

function initWebSocket() {
    const WS_URL = "ws://" + window.location.host + "/ws?token=" + window.login_data.token + "&userId=" + window.login_data.userId;
    var socket = window.mysocket;

    function connectWebSocket() {
        // 检查是否已经连接
        if (socket && socket.readyState === WebSocket.OPEN) {
            console.log("WebSocket 已经连接");
            return;
        }

        try {
            socket = new WebSocket(WS_URL);
            
            // ------------------------------------
            // 1. 处理连接打开事件 (Connection Opened)
            // ------------------------------------
            socket.onopen = (event) => {
                // statusElement.textContent = '状态: 已连接 ✅';
                console.log('连接已建立', event);

                window.mysocket = socket;
                // 连接成功后可以发送一个初始消息
                socket.send(JSON.stringify({ cmd: 2 }));
            };

            // ------------------------------------
            // 2. 处理接收消息事件 (Message Received)
            // ------------------------------------
            socket.onmessage = (event) => {
                const data = event.data;
                console.log('收到消息:', data);
                processServerMsg(data);
                
                // 在页面上显示消息
                // const li = document.createElement('li');
                // li.textContent = `服务器: ${data}`;
                // messagesElement.appendChild(li);
                
                // // 滚动到底部以显示最新消息
                // messagesElement.scrollTop = messagesElement.scrollHeight;
            };

            // ------------------------------------
            // 3. 处理连接关闭事件 (Connection Closed)
            // ------------------------------------
            socket.onclose = (event) => {
                // statusElement.textContent = '状态: 已断开 ❌';
                console.log('连接已关闭:', event);
                
                // 尝试重连 (可选功能)
                // setTimeout(connectWebSocket, 5000); 
            };

            // ------------------------------------
            // 4. 处理错误事件 (Error)
            // ------------------------------------
            socket.onerror = (error) => {
                // statusElement.textContent = '状态: 发生错误 ⚠️';
                console.error('WebSocket 发生错误:', error);
            };

        } catch (e) {
            // statusElement.textContent = `状态: 连接失败 (${e.message})`;
            console.error('创建 WebSocket 失败:', e);
        }
    }

    connectWebSocket();
}

function processServerMsg(data) {
    msg = JSON.parse(data)
    switch (msg.cmd) {
        case 1:
            processTextMsg(msg);
            break;
        case 2:
            break;
        case 3:
            processOnlineUserMsg(msg);
            break;
        default:
            console.log("unknown msg cmd: ", msg.cmd)
    }
}

function processOnlineUserMsg(msg) {
    var itemTmpl = "<div class='user-info'><span tooltip-right='点击与他聊天' user-id='{{userId}}'>{{userName}}</span></div>";
    var node = document.querySelector("#online-user-view");
    node.innerHTML = "";
    node.insertAdjacentHTML("beforeend", "<div class='tip-text'>当前在线用户：</div>");

    var myId = window.login_data.userId;
    for (let i = 0; i < msg["user_info"].length; i++) {
        var userName = msg["user_info"][i].name;
        var userId = msg["user_info"][i].user_id;
        if (userId == myId) {
            continue;
        }
        var item = itemTmpl.replace("{{userName}}", userName);
        item = item.replace("{{userId}}", userId);
        node.insertAdjacentHTML("beforeend", item);
    }
}

function processTextMsg(msg) {
    console.log(msg);
    var htmls = [];
    htmls.push({
        messageType: 'raw',
        headIcon: '/static/images/B.jpg',
        name: msg["from_id"].toString(),
        position: 'left',
        html: msg["text"]
    })
    beforeRenderingHTML(htmls, '.lite-chatbox');    
}

document.querySelector('.send').onclick = function () {

    var text = document.querySelector('.chatinput').innerHTML;
    console.log(text);

    var socket = window.mysocket;
    if (!socket) {
        console.log("socket not init");
        return;
    }

    socket.send(JSON.stringify({
        cmd: 1,
        from_id: window.login_data["userId"],
        to_id: Number(window.other_user.user_id),
        text: text,
    }));

    htmls.push({
        messageType: 'raw',
        headIcon: '/static/images/B.jpg',
        name: window.login_data.name,
        position: 'right',
        html: text
    })

    document.querySelector('.chatinput').innerHTML = '';
    beforeRenderingHTML(htmls, '.lite-chatbox');
};