(function () {

    var DEFAULT_SCRIPT = EditorSetup.DEFAULT_SCRIPT;

    var addEditor, editEditor, settingEditor;
    var addModal, editModal;

    var TYPE_LABELS = {
        script: 'Go脚本',
        http: 'HTTP',
        shell: 'Shell',
        webhook: 'Webhook'
    };

    // ===== Utility Functions =====
    function escapeHtml(s) {
        if (!s) return '';
        return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
    }

    function trimContent(s) {
        if (!s) return '';
        s = s.trim();
        return s.length > 60 ? s.substring(0, 60) + '...' : s;
    }

    function clearTimer() {
        document.getElementById('timerTableBody').innerText = '';
    }

    function notice(msg, type) {
        var existing = document.querySelectorAll('.notification');
        var offset = 24 + existing.length * 60;
        var el = document.createElement('div');
        el.className = 'notification' + (type ? ' ' + type : '');
        el.style.top = offset + 'px';
        el.innerText = msg;
        document.body.appendChild(el);
        setTimeout(function () {
            if (el.parentNode) el.parentNode.removeChild(el);
        }, 3000);
    }

    // ===== Type Switching =====
    function switchType(modalPrefix, type) {
        var map = {
            script: modalPrefix + 'ContentScript',
            http: modalPrefix + 'ContentHttp',
            shell: modalPrefix + 'ContentShell',
            webhook: modalPrefix + 'ContentWebhook'
        };
        Object.keys(map).forEach(function (t) {
            var el = document.getElementById(map[t]);
            if (el) el.style.display = (t === type) ? 'block' : 'none';
        });
    }

    // ===== Editor Helper =====
    function getEditor(modalPrefix) {
        return modalPrefix === 'add' ? addEditor : editEditor;
    }

    // ===== Collect Content =====
    function collectContent(modalPrefix) {
        var type = document.getElementById(modalPrefix + 'Type').value;
        var content = '';
        if (type === 'script') {
            content = getEditor(modalPrefix).getValue();
        } else if (type === 'http') {
            var headersStr = document.getElementById(modalPrefix + 'HttpHeaders').value.trim();
            var headers = {};
            if (headersStr) {
                try {
                    headers = JSON.parse(headersStr);
                } catch (e) {
                    headers = {};
                }
            }
            content = JSON.stringify({
                method: document.getElementById(modalPrefix + 'HttpMethod').value,
                url: document.getElementById(modalPrefix + 'HttpUrl').value,
                headers: headers,
                body: document.getElementById(modalPrefix + 'HttpBody').value
            });
        } else if (type === 'shell') {
            content = document.getElementById(modalPrefix + 'ShellCommand').value;
        } else if (type === 'webhook') {
            content = JSON.stringify({
                url: document.getElementById(modalPrefix + 'WebhookUrl').value,
                payload: document.getElementById(modalPrefix + 'WebhookPayload').value
            });
        }
        return { type: type, content: content };
    }

    // ===== Populate Content =====
    function populateContent(modalPrefix, type, content) {
        type = type || 'script';
        document.getElementById(modalPrefix + 'Type').value = type;
        switchType(modalPrefix, type);

        if (type === 'script') {
            getEditor(modalPrefix).setValue(content || '');
        } else if (type === 'http') {
            var httpData = {};
            try {
                httpData = JSON.parse(content || '{}');
            } catch (e) {
                httpData = {};
            }
            document.getElementById(modalPrefix + 'HttpMethod').value = httpData.method || 'GET';
            document.getElementById(modalPrefix + 'HttpUrl').value = httpData.url || '';
            document.getElementById(modalPrefix + 'HttpHeaders').value = httpData.headers ? JSON.stringify(httpData.headers, null, 2) : '';
            document.getElementById(modalPrefix + 'HttpBody').value = httpData.body || '';
        } else if (type === 'shell') {
            document.getElementById(modalPrefix + 'ShellCommand').value = content || '';
        } else if (type === 'webhook') {
            var webhookData = {};
            try {
                webhookData = JSON.parse(content || '{}');
            } catch (e) {
                webhookData = {};
            }
            document.getElementById(modalPrefix + 'WebhookUrl').value = webhookData.url || '';
            document.getElementById(modalPrefix + 'WebhookPayload').value = webhookData.payload || '';
        }
    }

    // ===== Reset Add Form =====
    function resetAddForm() {
        document.getElementById('addType').value = 'script';
        switchType('add', 'script');
        addEditor.setValue(DEFAULT_SCRIPT);
        document.getElementById('addName').value = '';
        document.getElementById('addCronExpression').value = '';
        document.getElementById('addHttpMethod').value = 'GET';
        document.getElementById('addHttpUrl').value = '';
        document.getElementById('addHttpHeaders').value = '';
        document.getElementById('addHttpBody').value = '';
        document.getElementById('addShellCommand').value = '';
        document.getElementById('addWebhookUrl').value = '';
        document.getElementById('addWebhookPayload').value = '';
        var addResult = document.getElementById('addTestResult');
        addResult.style.display = 'none';
        addResult.innerHTML = '';
    }

    // ===== Table Rendering =====
    function loadingTimer(data) {
        var table = document.getElementById('timerTableBody');
        data.forEach(function (item) {
            var row = document.createElement('tr');
            var type = item.type || 'script';
            var typeLabel = TYPE_LABELS[type] || type;
            row.innerHTML =
                '<td>' + item.id + '</td>' +
                '<td title="' + escapeHtml(item.name) + '">' + escapeHtml(item.name) + '</td>' +
                '<td>' + typeLabel + '</td>' +
                '<td title="' + escapeHtml(item.cron) + '">' + escapeHtml(item.cron) + '</td>' +
                '<td title="' + escapeHtml(item.content) + '">' + escapeHtml(trimContent(item.content)) + '</td>' +
                '<td><label class="switch"><input type="checkbox"' + (item.enable ? ' checked' : '') + ' onchange="toggleStatus(this)"><span class="slider"></span></label></td>' +
                '<td>' + (item.next || '') + '</td>' +
                '<td class="actions">' +
                '<button class="exec" onclick="execTimer(this)">执行</button>' +
                '<button class="edit" onclick="openEditModal(this)">修改</button>' +
                '<button class="delete" onclick="deleteTimer(this)">删除</button>' +
                '</td>';
            row.dataset.content = item.content;
            row.dataset.type = type;
            table.appendChild(row);
        });
    }

    // ===== Toggle Status =====
    window.toggleStatus = async function (checkbox) {
        var row = checkbox.closest('tr');
        var enable = checkbox.checked;
        try {
            await API.enableTimer(row.children[0].innerText, enable);
        } catch (e) {
            notice(e.message || '操作失败', 'error');
            checkbox.checked = !enable;
            return;
        }
        notice(enable ? '已启用' : '已禁用');
        refresh();
    };

    // ===== Edit Modal =====
    window.openEditModal = function (button) {
        var row = button.closest('tr');
        var id = row.children[0].innerText;
        var name = row.children[1].getAttribute('title') || row.children[1].innerText;
        var type = row.dataset.type || 'script';
        var cron = row.children[3].getAttribute('title') || row.children[3].innerText;
        var content = row.dataset.content || '';

        document.getElementById('editName').value = name;
        document.getElementById('editCronExpression').value = cron;
        populateContent('edit', type, content);
        var editResult = document.getElementById('editTestResult');
        editResult.style.display = 'none';
        editResult.innerHTML = '';
        editModal.style.display = 'block';
        editModal.dataset.editId = id;
    };

    // ===== Delete Timer =====
    window.deleteTimer = async function (button) {
        var row = button.closest('tr');
        try {
            await API.deleteTimer(row.children[0].innerText);
        } catch (e) {
            notice(e.message || '删除失败', 'error');
            return;
        }
        notice('删除成功');
        row.remove();
    };

    // ===== Exec Timer (立即执行一次) =====
    window.execTimer = async function (button) {
        var row = button.closest('tr');
        var id = row.children[0].innerText;
        var name = row.children[1].getAttribute('title') || row.children[1].innerText;
        button.disabled = true;
        try {
            await API.execTimer(id);
            notice('[' + name + '] 已触发执行');
        } catch (e) {
            notice(e.message || '执行失败', 'error');
        } finally {
            button.disabled = false;
        }
    };

    // ===== Refresh =====
    async function refresh() {
        try {
            var data = await API.getTimers();
            clearTimer();
            if (data) {
                loadingTimer(data);
            }
        } catch (e) {
            notice(e.message || '获取数据失败', 'error');
        }
    }

    // ===== Log Rendering =====
    function loadingLogs(logs) {
        var tbody = document.getElementById('logTableBody');
        tbody.innerHTML = '';
        if (!logs || logs.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;padding:32px;color:var(--text-muted)">暂无日志</td></tr>';
            return;
        }
        logs.forEach(function (log) {
            var row = document.createElement('tr');
            var statusClass = log.status === 'success' ? 'log-status-success' : 'log-status-error';
            var statusText = log.status === 'success' ? '成功' : '失败';
            var resultText = log.status === 'success' ? (log.result || '') : (log.error || '');
            row.innerHTML =
                '<td>' + (log.createdAt || '') + '</td>' +
                '<td>' + (log.name || '') + '</td>' +
                '<td>' + (TYPE_LABELS[log.type] || (log.type ? log.type : 'Go脚本')) + '</td>' +
                '<td class="' + statusClass + '">' + statusText + '</td>' +
                '<td class="log-result" title="点击展开/收起">' + escapeHtml(resultText) + '</td>';
            tbody.appendChild(row);
        });
        // Click to expand/collapse result
        document.querySelectorAll('.log-result').forEach(function (el) {
            el.addEventListener('click', function () {
                this.classList.toggle('expanded');
            });
        });
    }

    // ===== Open Log Modal =====
    async function openLogModal() {
        try {
            // Load timers for filter
            var timers = await API.getTimers();
            var filter = document.getElementById('logFilter');
            filter.innerHTML = '<option value="0">全部任务</option>';
            if (timers) {
                timers.forEach(function (t) {
                    var opt = document.createElement('option');
                    opt.value = t.id;
                    opt.textContent = t.name;
                    filter.appendChild(opt);
                });
            }
            // Load all logs
            var logs = await API.getLogs(0, 200);
            loadingLogs(logs);
        } catch (e) {
            notice(e.message || '加载日志失败', 'error');
        }
        document.getElementById('logModal').style.display = 'block';
    }

    // ===== Refresh Logs =====
    async function refreshLogs() {
        try {
            var timerId = document.getElementById('logFilter').value;
            var logs = await API.getLogs(timerId && timerId !== '0' ? timerId : null, 200);
            loadingLogs(logs);
        } catch (e) {
            notice(e.message || '刷新日志失败', 'error');
        }
    }

    // ===== Clear Logs =====
    async function clearLogs() {
        try {
            var timerId = document.getElementById('logFilter').value;
            await API.clearLogs(timerId && timerId !== '0' ? timerId : null);
            notice('日志已清除');
            await refreshLogs();
        } catch (e) {
            notice(e.message || '清除日志失败', 'error');
        }
    }

    // ===== Open Setting Modal =====
    async function openSettingModal() {
        try {
            var script = await API.getErrorHandler();
            settingEditor.setValue(script || EditorSetup.ERROR_HANDLER_SCRIPT);
        } catch (e) {
            notice(e.message || '加载设置失败', 'error');
            settingEditor.setValue(EditorSetup.ERROR_HANDLER_SCRIPT);
        }
        document.getElementById('settingTestResult').style.display = 'none';
        document.getElementById('settingModal').style.display = 'block';
    }

    // ===== Save Setting =====
    async function saveSetting() {
        try {
            await API.saveErrorHandler(settingEditor.getValue());
            notice('保存成功,已生效');
            document.getElementById('settingModal').style.display = 'none';
        } catch (e) {
            notice(e.message || '保存失败', 'error');
        }
    }

    // ===== Test Setting Script =====
    async function testSettingScript() {
        var btn = document.getElementById('testSettingScript');
        var resultDiv = document.getElementById('settingTestResult');
        btn.disabled = true;
        btn.textContent = '执行中...';
        resultDiv.style.display = 'none';
        try {
            var result = await API.testErrorHandler(settingEditor.getValue());
            resultDiv.className = 'test-result success';
            resultDiv.textContent = result || '执行成功';
            resultDiv.style.display = 'block';
        } catch (e) {
            resultDiv.className = 'test-result error';
            resultDiv.textContent = e.message || '执行失败';
            resultDiv.style.display = 'block';
        } finally {
            btn.disabled = false;
            btn.textContent = '测试执行';
        }
    }

    // ===== DOM Ready =====
    document.addEventListener('DOMContentLoaded', function () {
        addModal = document.getElementById('addModal');
        editModal = document.getElementById('editModal');
        var logModal = document.getElementById('logModal');

        // Initialize Monaco editors
        EditorSetup.init(function (editors) {
            addEditor = editors.addEditor;
            editEditor = editors.editEditor;
            settingEditor = editors.settingEditor;

            // Load data after editors are ready
            refresh();
        });

        // ===== Modal Logic =====
        var addBtn = document.getElementById('openAddModal');
        var closeAdd = addModal.getElementsByClassName('close')[0];
        var closeEdit = editModal.getElementsByClassName('close')[0];
        var closeLog = logModal.getElementsByClassName('close')[0];

        addBtn.onclick = function () {
            resetAddForm();
            addModal.style.display = 'block';
        };

        closeAdd.onclick = function () {
            addModal.style.display = 'none';
        };

        closeEdit.onclick = function () {
            editModal.style.display = 'none';
        };

        closeLog.onclick = function () {
            logModal.style.display = 'none';
        };

        // ===== Setting Modal =====
        var settingModal = document.getElementById('settingModal');
        var closeSetting = settingModal.getElementsByClassName('close')[0];

        document.getElementById('openSettingModal').addEventListener('click', openSettingModal);
        document.getElementById('saveSetting').addEventListener('click', saveSetting);
        document.getElementById('testSettingScript').addEventListener('click', testSettingScript);

        closeSetting.onclick = function () {
            settingModal.style.display = 'none';
        };

        window.onclick = function (event) {
            if (event.target === addModal) {
                addModal.style.display = 'none';
            }
            if (event.target === editModal) {
                editModal.style.display = 'none';
            }
            if (event.target === logModal) {
                logModal.style.display = 'none';
            }
            if (event.target === settingModal) {
                settingModal.style.display = 'none';
            }
        };

        // ===== Type Select Bindings =====
        document.getElementById('addType').addEventListener('change', function () {
            switchType('add', this.value);
        });
        document.getElementById('editType').addEventListener('change', function () {
            switchType('edit', this.value);
        });

        // ===== Test Script (Add) =====
        document.getElementById('testAddScript').addEventListener('click', async function () {
            var name = document.getElementById('addName').value || '测试';
            var collected = collectContent('add');
            var btn = document.getElementById('testAddScript');
            var resultDiv = document.getElementById('addTestResult');
            btn.disabled = true;
            btn.textContent = '执行中...';
            resultDiv.style.display = 'none';
            try {
                var result = await API.testTimer(name, collected.content, collected.type);
                resultDiv.className = 'test-result success';
                resultDiv.textContent = result || '执行成功';
                resultDiv.style.display = 'block';
            } catch (e) {
                resultDiv.className = 'test-result error';
                resultDiv.textContent = e.message || '执行失败';
                resultDiv.style.display = 'block';
            } finally {
                btn.disabled = false;
                btn.textContent = '测试执行';
            }
        });

        // ===== Add Timer =====
        document.getElementById('addTimer').addEventListener('click', async function () {
            var name = document.getElementById('addName').value;
            var cron = document.getElementById('addCronExpression').value;
            var collected = collectContent('add');
            try {
                await API.addTimer(name, cron, collected.content, false, collected.type);
            } catch (e) {
                notice(e.message || '添加失败', 'error');
                return;
            }
            notice('添加成功');
            resetAddForm();
            addModal.style.display = 'none';
            refresh();
        });

        // ===== Test Script (Edit) =====
        document.getElementById('testEditScript').addEventListener('click', async function () {
            var name = document.getElementById('editName').value || '测试';
            var collected = collectContent('edit');
            var btn = document.getElementById('testEditScript');
            var resultDiv = document.getElementById('editTestResult');
            btn.disabled = true;
            btn.textContent = '执行中...';
            resultDiv.style.display = 'none';
            try {
                var result = await API.testTimer(name, collected.content, collected.type);
                resultDiv.className = 'test-result success';
                resultDiv.textContent = result || '执行成功';
                resultDiv.style.display = 'block';
            } catch (e) {
                resultDiv.className = 'test-result error';
                resultDiv.textContent = e.message || '执行失败';
                resultDiv.style.display = 'block';
            } finally {
                btn.disabled = false;
                btn.textContent = '测试执行';
            }
        });

        // ===== Save Changes =====
        document.getElementById('saveChanges').onclick = async function () {
            var collected = collectContent('edit');
            try {
                await API.updateTimer(
                    editModal.dataset.editId,
                    document.getElementById('editName').value,
                    document.getElementById('editCronExpression').value,
                    collected.content,
                    collected.type
                );
            } catch (e) {
                notice(e.message || '修改失败', 'error');
                return;
            }
            notice('修改成功');
            editModal.style.display = 'none';
            refresh();
        };

        // ===== Refresh Button =====
        document.getElementById('refreshAll').addEventListener('click', refresh);

        // ===== Log Modal =====
        document.getElementById('openLogModal').addEventListener('click', openLogModal);
        document.getElementById('refreshLogs').addEventListener('click', refreshLogs);
        document.getElementById('clearLogs').addEventListener('click', clearLogs);
        document.getElementById('logFilter').addEventListener('change', refreshLogs);

        // ===== Initialize WebSocket =====
        WSClient.onMessage = function (data) {
            try {
                var res = JSON.parse(data);
                notice(res.msg, res.success === false ? 'error' : 'success');
            } catch (e) {
                notice(data, 'info');
            }
        };
        WSClient.connect();
    });
})();
