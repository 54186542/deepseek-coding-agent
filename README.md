# dca — DeepSeek Coding Agent

基于 DeepSeek API 的 AI 编程助手，支持 CLI 和 Wails GUI 两种交互模式。

## 快速开始 (CLI)

```bash
# 设置 API key
export DEEPSEEK_API_KEY=your_key_here

# 运行
go run . --root /path/to/your/project
```

### 命令行参数

| 参数 | 环境变量 | 说明 |
|---|---|---|
| `--api-key` | `DEEPSEEK_API_KEY` | DeepSeek API 密钥 |
| `--model` | — | 模型名称 (默认: deepseek-chat) |
| `--root` | — | 项目根目录 (默认: 当前目录) |

## Wails GUI

```bash
# 1. 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 2. 添加 Go 依赖
go get github.com/wailsapp/wails/v2@latest

# 3. 安装前端依赖
cd frontend && npm install

# 4. 开发模式运行
wails dev

# 5. 构建
wails build
```

## 项目结构

```
dca/
├── main.go                 # CLI 入口
├── main_wails.go           # Wails 入口 (build tag: wails)
├── internal/
│   ├── react/
│   │   ├── engine.go       # 纯状态驱动的 ReAct 引擎
│   │   └── loop.go         # CLI 交互循环
│   ├── llm/                # DeepSeek API 客户端
│   ├── tool/               # 文件编辑工具集
│   ├── platform/           # 平台适配层
│   ├── safety/             # 沙箱安全层
│   └── ui/bridge.go       # Wails 后端绑定
├── frontend/               # Vue 3 前端
│   ├── src/
│   │   ├── App.vue
│   │   ├── main.js
│   │   └── style.css
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
└── spec/v001.md            # 开发计划
```
