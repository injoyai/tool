# Timer V2 项目指南

## 项目简介
一个 Go 定时任务管理工具，支持通过 Web 界面管理定时任务。支持 4 种任务类型：Go 脚本(yaegi 解释器)、HTTP 请求、Shell 命令、Webhook。带执行日志记录与查看功能。

## 项目结构
```
timer-v2/
├── api.go              # HTTP handler + 路由(含 embed 静态资源)
├── auth.go             # 登录认证(HMAC签名Cookie,无密码免登录)
├── model.go            # Timer 结构体 + xorm 模型
├── executor.go         # 多类型任务执行器(script/http/shell/webhook)
├── script.go           # yaegi Go 脚本引擎(i 包) + stdout 捕获
├── log.go              # 执行日志模型 + ExecWithLog + WebSocket 通知
├── error_handler.go    # 全局错误处理脚本(onError + Setting 读写缓存)
├── error_handler_test.go # 错误处理测试
├── script_test.go      # 脚本引擎测试
├── auth_test.go        # token 派生测试
├── web/                # 前端项目目录(embed 到二进制)
│   ├── index.html      # 前端入口
│   ├── login.html      # 登录页(内联CSS,暗色主题)
│   └── static/
│       ├── css/main.css
│       └── js/         # api.js, editor.js, ws.js, app.js
├── build.sh            # 发布构建(带压缩,上传 minio)
├── Dockerfile          # 多阶段构建(golang:1.25-alpine -> alpine)
├── docker-push.sh      # 构建并推送镜像(可配多架构)
├── .dockerignore       # 排除 data/、*.exe、docs/
├── .gitignore          # 忽略 config/config.yaml(含敏感密钥,不入库)
├── cmd/server/         # 入口 main.go
├── cmd/tray/           # 系统托盘入口
├── lib/                # yaegi extract 生成的 injoyai 包符号
└── data/database/      # SQLite 数据库
```

## 快速开始
```bash
go build ./cmd/server && server.exe    # 访问 http://localhost:8078/
```

## 技术栈

### 后端
- **HTTP 框架**: gofiber/fiber/v3 -> injoyai/frame/fiber 封装
- **数据库**: SQLite (xorm)
- **定时任务**: robfig/cron (via injoyai/goutil/task)
- **脚本引擎**: yaegi v0.16.1 (Go 解释器)
- **stdout 捕获**: `interp.Options{Stdout: &buf, Stderr: &buf}` 直接捕获脚本 `fmt.Println` 输出

### 前端
- 入口 `index.html` + 分离的 CSS/JS 文件，通过 `//go:embed web` 打包
- **编辑器**: Monaco Editor (v0.52.2)，通过 CDN (jsDelivr) 加载，vs-dark 主题
- 静态资源(`/static/*`) 通过 embed.FS + fiber 路由提供，自动按后缀设置 Content-Type
- **前端模块结构**:
  - `static/css/main.css` - 暗色主题(背景 #0f0f1a, 卡片 #1a1a2e, 主色琥珀 #f59e0b, 辅色青 #14b8a6)
  - `static/js/api.js` - API 客户端(IIFE 全局 `API` 对象)
  - `static/js/editor.js` - Monaco 编辑器初始化 + 字典补全(GO_HINTS/GO_SIGS/GO_IMPORTS)
  - `static/js/ws.js` - WebSocket 客户端(自动重连)
  - `static/js/app.js` - 主应用逻辑(类型切换/日志/测试/CRUD)
- **多类型 UI**: 类型下拉框切换 4 种内容编辑区，`switchType()` 控制显隐，`collectContent()`/`populateContent()` 按 type 序列化/反序列化
- 字体: Google Fonts - Outfit (UI) + JetBrains Mono (代码)

### 代码补全系统
基于 Monaco `registerCompletionItemProvider` + `registerSignatureHelpProvider` 的字典补全：
- **GO_HINTS**: 包成员提示，覆盖标准库 + injoyai 库全部 32 个包(ios/v2 子包、conv、logs、crypt 等)
- **GO_SIGS**: 函数签名，覆盖全部 32 个包的函数参数
- **GO_IMPORTS**: 导入路径补全，标准库短路径 + injoyai 全路径
- **GO_BUILTIN_SIGS**: 内置函数签名(make/len/append 等)
- 字典补全纯前端实现，无 gopls 依赖

## 核心约定

### 配置文件 (config/config.yaml)
- **已 gitignore，不入库**：含敏感信息(ServerChan key、夸克签到 token 等)，仅本地保留
- 模板文件 `config/config.yaml.exmplate` 可入库供参考
- password 配置见下方「登录认证」

### 脚本编写 (Go 语法)
脚本必须是完整 Go 程序(`package main` + `func main`)，可用的包:
- **`i` 包** (内置函数): `i.Start`, `i.Ping`, `i.Notice`, `i.Dial`, `i.DialTCP`, `i.Print`, `i.Println`, `i.GetPrices`, `i.Set`, `i.Get`, `i.Del`
- **标准库**: fmt, time, net, encoding/json, strings 等
- **injoyai 库**: conv, base/crypt/*, base/maps/*, base/coding, logs 等 (需 `import "github.com/injoyai/..."`)
- `fmt.Println` 输出会被 `interp.Options.Stdout` 捕获，返回给前端展示

### 任务类型 (executor.go)
`Timer.Type` 字段决定执行方式，`Exec()` 根据类型分发：
- `script` (默认): yaegi Go 脚本引擎，content 为 Go 源码，**捕获 stdout 返回输出**
- `http`: HTTP 请求，content 为 JSON `{method, url, headers, body}`，返回响应体
- `shell`: 系统命令，content 为命令字符串(Windows `cmd /c`，Linux `sh -c`)，返回命令输出
- `webhook`: Webhook 通知，content 为 JSON `{url, payload}`，POST application/json
- `PutTimer` 需更新 `Type` 列: `Cols("Name", "Cron", "Content", "Type")`

### 执行日志 (log.go)
- **自动记录**: 每次 cron 触发或手动测试都会记录日志(Log 模型: 时间/任务名/类型/状态/结果/错误)
- **ExecWithLog**: 封装 `Exec()` + 写日志 + WebSocket 推送通知
- **TestTimer**: 也记录日志，并返回执行结果给前端
- **API 端点**: `GET /api/log/list?timerId=&limit=` 查询，`DELETE /api/log?timerId=` 清除
- **前端**: 日志弹窗(900px 宽)，支持按任务筛选、刷新、清除，结果列点击展开/收起
- **输出截断**: 结果超过 2000 字符自动截断
- **自动清理**: `_init()` 启动时 `cleanupOldLogs()` 清一次(适配开机自启) + `Corn.SetTask("cleanup_logs","0 0 3 * * *",...)` 每天 3 点定时清理,删除 7 天前的日志(`WHERE CreatedAt < cutoff`)
- **WebSocket 并发写坑(已修复)**: `noticeWS` 在 `WS.Range` 中调 `conn.WriteJSON`，多个 cron 任务并发执行时多 goroutine 同时写同一连接，触发 fasthttp/websocket 的 `concurrent write to websocket connection` panic。cron 在独立 goroutine 运行，`WithRecover` 不覆盖，会直接崩溃整个进程。修复：`WS` 值类型改为 `*wsClient`(封装 `conn + sync.Mutex`)，`noticeWS` 写前加锁序列化。注意 `maps.Generic` 本身线程安全(RWMutex)，问题仅在 ws 写操作
- **时区坑(已修复)**: xorm 默认 `DatabaseTZ=time.UTC`，`created` 字段(如 `Log.CreatedAt`，string->Varchar)按 `time.Now().In(UTC)` 格式化存储，导致比本地少 8 小时。修复：`_init()` 中 `DB.SetTZDatabase(time.Local)`。注意 `xorms.Engine` 是嵌入 `*xorm.Engine` 的薄包装，`SetTZDatabase` 直接作用于底层引擎。仅影响新写入，历史记录仍是 UTC。`TZLocation`(app 时区)默认已是 `time.Local`，无需改
- **列名坑(重要)**: `sqlite.NewXorm` -> `xorms.NewSqlite` -> `WithSyncField` 用 `core.SameMapper`，列名 = 字段名(PascalCase)。Log 表实际列名为 `ID/TimerID/Name/Type/Status/Result/Error/CreatedAt`，**不是** snake_case。写 `Where` 时必须用 PascalCase(如 `Where("CreatedAt < ?", x)`、`Where("TimerID = ?", id)`)。已修复历史遗留的 `timer_id` 误用(原 `GetLogs`/`ClearLogs` 按 timerId 过滤/清除会报 `no such column`)

### 全局错误处理 (error_handler.go)
- 任务执行失败(ExecWithLog, cron/立即执行路径; 手动"测试执行"不触发)时异步调用 `onError(taskID, taskName, errMsg)`(log.go 失败分支, `go onError(...)` 在 noticeWS 之前)
- 错误处理脚本存 Setting 表(key=error_handler_script)，进程内双检锁缓存，Web 设置页(工具栏 ⚙ 按钮)编辑，保存即生效；传空脚本即停用
- 脚本为完整 Go 程序，必须定义 `OnError(taskID int64, taskName, errMsg string)`；taskID 稳定(改名不影响)，可配合 i.Set/i.Get 做失败计数
- `scriptEngine.CallFunc(code, fn, args...)`: Eval 完整脚本声明后 Eval `OnError(int64(123), "name", "msg")` 带参调用；**args[0] 必须是数字字符串**(内部拼 `int64(...)`)，其余 strconv.Quote 转义；void 函数返回值已归一化为 nil(避免 *interface{} 指针地址进日志)；**注意: 若脚本含 func main，首次 Eval 会执行它**——错误处理脚本只应包含函数声明
- 处理脚本自身失败仅写 Log 表(Type=error_handler) + logs.Errorf，**不递归触发**；脚本为空时 onError 静默跳过；CallFunc 无超时控制，脚本死循环会泄漏 goroutine(设计上列为非目标)
- API: GET/PUT /api/setting/error_handler, POST /api/setting/error_handler/test(模拟参数 999/测试任务/测试错误信息；请求 script 为空时回退测试已保存脚本)
- Setting 表为通用 KV(xorm pk=Key, PascalCase 列名)，后续全局配置可复用

### 路由参数 (关键易错)
**gofiber 中 `c.Get(key)` 取的是请求头，取路由参数必须用 `c.Params("*")`！**

### 登录认证 (auth.go)
- 密码来源: `cfg.GetString("password","")`(自动读环境变量 `password`)，回退 `os.Getenv("PASSWORD")`；配置文件亦可
- 无密码: 不启用认证，所有路由开放(本地裸跑友好)
- 有密码: `GET /` 未认证跳转 `/login`；`POST /api/login` 校验密码后下发 `timer_token` Cookie(HttpOnly, SameSite=Lax, 7天)
- token = `HMAC-SHA256(password, "timer-v2-session")` 的 hex，无状态确定性，重启不掉线
- `authMiddleware`(签名 `func(c fbr.Ctx) error`) 挂在 `/api/*` 组首行，login/logout 放行，其余校验 Cookie；WebSocket `/api/notice/ws` 同样受保护
- 密码比对用 `subtle.ConstantTimeCompare` 防时序攻击
- **关键坑**: `fbr.Ctx` 的 `next()` 是非导出方法，包外中间件不能调用！必须用 `fiber.Ctx` 的 `c.Next()`(返回 error)；中间件写成 `func(c fbr.Ctx) error` 并 `return c.Next()`，由 fbr transfer 的 `dealErr` 处理下游错误
- 新增文件: `web/login.html`(独立登录页，内联CSS，含 autofill 修复)

### Handler 响应标准化
所有 handler 必须调用 `c.CheckErr(err)` 处理错误 + `c.Succ(nil)` 返回成功响应。
`TestTimer` 例外：直接 `c.JSON(map[string]interface{}{"code": 200, "data": resultStr})` 返回执行结果。

### 前端逻辑模式
- `api.js` 的 `request()` 统一处理 code!=200 -> throw Error(msg)，调用方用 try/catch 捕获
- `notice(msg, type)` 弹出通知，多条通知垂直堆叠(每条偏移 60px)
- 测试执行：按钮变为"执行中..."并禁用，结果展示在 `div.test-result` 面板(成功绿色/失败红色)
- WebSocket 消息通过 `WSClient.onMessage` 回调 -> `notice()` 显示

### CSS 注意
- 主表列宽规则必须限定到 `#timerTable`，否则会影响日志表 `#logTable` 的列宽和颜色
- `.modal` 设 `overflow-y: auto` 滚动，`.modal-content` 不设 `max-height`(避免双重滚动条)
- 日志表 `#logTable` 使用 `table-layout: fixed` + 独立列宽规则

## 构建与运行
```bash
# 开发运行
go build ./cmd/server && server.exe

# 发布构建(压缩,无控制台窗口)
./build.sh

# 测试(需 -vet=off 绕过 Go 1.25 vet 兼容问题)
go test -vet=off ./...

# Docker
docker build -t injoyai/timer:v2 .
docker run -d -p 8078:8078 -e password=xxx -v timer-data:/app/data injoyai/timer:v2
# docker-push.sh: 构建并推送(REGISTRY 留空则仅本地构建,PLATFORMS 可设多架构)
```

## 维护 lib 符号
lib/ 目录是 `yaegi extract` 生成的 injoyai 包符号定义，供脚本使用。
如需添加新包:
```bash
cd lib
yaegi extract github.com/injoyai/<pkg>
```
然后在 `lib/lib.go` 中加入 `_ "github.com/injoyai/<pkg>"` blank import，
并执行 `go mod tidy`。

**注意**: 添加新包后需同步更新 `editor.js` 中的 GO_HINTS、GO_SIGS、GO_IMPORTS 字典。
