# LeonAgent

LeonAgent 是基于 **Go + Wails + Svelte** 的 macOS 本地编程工作台。它将项目浏览、代码编辑、Git 状态和 Ollama 流式对话放在同一个原生桌面界面中。

## 当前能力

- 本地目录与文件浏览
- 文本文件读取、编辑和保存
- Ollama 模型发现与流式聊天
- Git 分支、工作区变更和 Diff 后端接口
- 文件大小限制、错误处理和请求超时
- 三栏式原生桌面工作区

## 技术栈

- Backend: Go 1.22, Wails v2
- Frontend: Svelte 4, Vite 5
- LLM runtime: Ollama
- Platform: macOS（Wails 也可扩展至 Windows/Linux）

## 本地开发

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 安装并启动 Ollama
ollama serve
ollama pull qwen2.5-coder:7b

# 安装前端依赖
cd frontend && npm install && cd ..

# 开发模式
wails dev

# 构建 macOS 应用
wails build
```

## 架构

```text
Svelte Workspace
  ├─ Project Explorer
  ├─ Source Editor
  ├─ Git Status
  └─ Agent Chat
          │ Wails bindings/events
Go Application Service
  ├─ Safe filesystem API
  ├─ Ollama streaming client
  └─ Git command adapter
```

## 下一阶段

1. Monaco 多标签编辑器与语法高亮
2. Agent 任务规划、Diff 预览和修改审批
3. Git commit、分支和回滚工作流
4. 项目级记忆与代码索引
5. OpenAI、Claude 和兼容 API Provider
6. MCP 工具生态与受控终端执行

## 安全原则

LeonAgent 默认只提供显式文件操作，不自动执行模型生成的 Shell 命令。未来加入工具执行时，应采用工作区边界、命令白名单、Diff 审批和可回滚提交。

## License

建议发布前添加 MIT 或 Apache-2.0 License。
