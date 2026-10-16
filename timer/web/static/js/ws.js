var WSClient = (function () {

    var ws = null;
    var url = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/api/notice/ws';
    var onMessage = null;
    var reconnectTimer = null;
    var manualClose = false;

    function connect() {
        if (ws && (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN)) {
            return;
        }
        manualClose = false;
        try {
            ws = new WebSocket(url);
        } catch (e) {
            scheduleReconnect();
            return;
        }

        ws.onopen = function () {
            // 连接建立
        };

        ws.onmessage = function (event) {
            if (typeof onMessage === 'function') {
                onMessage(event.data);
            }
        };

        ws.onclose = function () {
            if (!manualClose) {
                scheduleReconnect();
            }
        };

        ws.onerror = function () {
            // 错误时关闭会触发 onclose
        };
    }

    function scheduleReconnect() {
        if (reconnectTimer) return;
        reconnectTimer = setTimeout(function () {
            reconnectTimer = null;
            connect();
        }, 3000);
    }

    function send(msg) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(msg);
        }
    }

    function close() {
        manualClose = true;
        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }
        if (ws) {
            ws.close();
        }
    }

    return {
        connect: connect,
        send: send,
        close: close,
        set onMessage(fn) { onMessage = fn; },
        get onMessage() { return onMessage; }
    };
})();
