# MEMORY.md — 项目记忆

## 项目概况

监听股票价格（A 股），突破阈值时通过 Windows 托盘通知提醒的桌面小工具。

## 技术栈

- Go 1.23，Windows 桌面应用
- `github.com/injoyai/lorca`：内嵌 `index.html`（`//go:embed`）作为配置界面，通过 `app.Bind` 绑定 `getConfig` / `setConfig` 前后端交互
- `github.com/injoyai/tdx`：连接通达信行情服务获取收盘价（`getPrice` 取日 K 最新收盘）
- `github.com/injoyai/goutil`：托盘（`tray.Run`）、通知（`notice.DefaultWindows`）、配置存储（`oss`）
- 配置/数据库目录：`~/.injoyai/monitor-price/`（config/config.json 与 database/codes.db）

## 架构与数据流

1. 托盘启动 → 连接 tdx → 读取配置文件 → `mon.Run` 定时轮询（间隔为配置的 interval 秒）。
2. 定时内跳过非交易时段（9:30 前 / 15:00 后 / 11:30~13:00），托盘"刷新"手动触发不受限。
3. 每条规则：code(如 sh600519) + price(元) + greater(≥/≤) + enable(开关)。涨/跌穿越阈值时各触发一次通知，`limit` 字段记录相对阈值状态防止重复通知。
4. 配置界面：lorca 窗口 800×600，`app.Eval("initialize()")` 回调 `getConfig` 回填表单，提交时 `setConfig` 写回 JSON。

## 前端约定（index.html）

- 纯原生 HTML/CSS/JS，单文件内嵌，无外部依赖（离线可用，无法加载外网字体，只用系统中文字体栈）。
- JS 依赖的类名/ID 结构：`#interval`、`#rules`、`.rule`、`.code/.price/.enable/.greater`、`.deleteRule`、`#addRule`、`#submit`、`#toast`、`showToast()`、`addRule()`、`initialize()`。改版时不得破坏这些契约。
- 样式已重做为深色行情终端风格，主色调金色 `#daa949`（与托盘图标同色系），CSS 变量集中在 `:root`（--gold、--bg-*、--border-* 等）。
- 窗口尺寸 800×600 固定，注意内容密度与滚动条样式。

## 决策记录

- **设计方向**（2026-08）：沿用图标琥珀金色作为品牌色，深色终端的克制风格，避免与股票软件常用的红涨绿跌配色冲突（界面本身不用红绿做主视觉，仅删除按钮 hover 用红色）。
- 加规则动画用 `ruleIn` keyframes，容器 `backdrop-filter` 在旧版 WebView 可能不生效，视觉上有纯色渐变兜底。

## 踩坑

- lorca 内嵌 HTML 无法加载外网字体/CDN，避免引入任何外部资源。
- 桌面窗口固定 800×600，预留滚动场景；`::-webkit-scrollbar` 需单独美化。
- 改版后必须重新编译（`build_win.sh`，含 UPX 压缩与 minio 上传），`index.html` 是 `//go:embed` 嵌入的，热改不生效。