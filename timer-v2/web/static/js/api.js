var API = (function () {

    async function request(url, options) {
        var res = await fetch(url, options);
        var data = await res.json();
        if (data.code !== 200) {
            throw new Error(data.msg || '请求失败');
        }
        return data.data;
    }

    function json(method, body) {
        return {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        };
    }

    return {
        getTimers: function () {
            return request('/api/timer/all');
        },

        addTimer: function (name, cron, content, enable, type) {
            return request('/api/timer', json('POST', {
                name: name,
                cron: cron,
                content: content,
                enable: enable,
                type: type
            }));
        },

        updateTimer: function (id, name, cron, content, type) {
            return request('/api/timer', json('PUT', {
                id: parseInt(id),
                name: name,
                cron: cron,
                content: content,
                type: type
            }));
        },

        enableTimer: function (id, enable) {
            return request('/api/timer/enable', json('PUT', {
                id: parseInt(id),
                enable: enable
            }));
        },

        deleteTimer: function (id) {
            return request('/api/timer', json('DELETE', {
                id: parseInt(id)
            }));
        },

        testTimer: function (name, content, type) {
            return request('/api/timer/test', json('POST', {
                name: name,
                content: content,
                type: type
            }));
        },

        execTimer: function (id) {
            return request('/api/timer/exec', json('POST', {
                id: parseInt(id)
            }));
        },

        getLogs: function (timerId, limit) {
            var params = [];
            if (timerId) params.push('timerId=' + timerId);
            if (limit) params.push('limit=' + limit);
            return request('/api/log/list' + (params.length ? '?' + params.join('&') : ''));
        },

        clearLogs: function (timerId) {
            var url = '/api/log';
            if (timerId) url += '?timerId=' + timerId;
            return request(url, json('DELETE', {}));
        }
    };
})();
