# 浮窗浏览器 (Floating Browser)

基于 Electron 的桌面浮窗浏览器应用。

## Electron 版本

- **Electron**: ^28.1.0

## 文件说明

- `main.js` - Electron 主进程，负责创建窗口和处理 IPC 通信
- `preload.js` - 预加载脚本，安全地暴露 API 给渲染进程
- `index.html` - 渲染进程页面，包含控制栏和 iframe
- `package.json` - 项目配置

## 功能

- 顶部 40px 控制栏，包含：
  - 透明度滑块（0.1 ~ 1.0，步长 0.1）
  - 窗口置顶复选框
- 剩余区域使用 iframe 全屏嵌入目标网页
- 窗口默认大小 900×650，可调整大小
- 标题为"浮窗浏览器"

## 启动方式

### 1. 安装依赖

```bash
npm install
```

### 2. 启动应用

```bash
# 默认加载 https://www.example.com
npm start

# 自定义加载网址
npm start -- --url=https://www.baidu.com
```

### 3. 开发模式

```bash
npm run dev
```

## 架构说明

```
┌─────────────────────────────────────┐
│         控制栏 (40px)                │
│  [透明度滑块] [1.0]  [☐ 置顶窗口]    │
├─────────────────────────────────────┤
│                                     │
│         iframe (剩余区域)            │
│    嵌入目标网页 (如 example.com)      │
│                                     │
│                                     │
└─────────────────────────────────────┘
```

## IPC 通信

- `set-opacity` - 设置窗口透明度
- `set-always-on-top` - 设置窗口是否置顶
