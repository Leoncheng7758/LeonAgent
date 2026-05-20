package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ChatRequest 前端发来的请求
type ChatRequest struct {
	Model    string `json:"model"`
	Message  string `json:"message"`
	History  []ChatMessage `json:"history"`
}

// ChatMessage 消息结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaResponse Ollama流式返回结构
type OllamaResponse struct {
	Model   string `json:"model"`
	Created int64  `json:"created_at"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// Chat 发送聊天请求（支持流式）
func (a *App) Chat(req ChatRequest) {
	go func() {
		payload := map[string]interface{}{
			"model":    req.Model,
			"messages": req.History,
			"stream":   true,
		}

		// 添加最新用户消息
		payload["messages"] = append(payload["messages"].([]ChatMessage), ChatMessage{
			Role:    "user",
			Content: req.Message,
		})

		jsonData, _ := json.Marshal(payload)

		resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			runtime.EventsEmit(a.ctx, "chat-error", err.Error())
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var ollamaResp OllamaResponse
			if err := json.Unmarshal([]byte(line), &ollamaResp); err != nil {
				continue
			}

			if ollamaResp.Message.Content != "" {
				runtime.EventsEmit(a.ctx, "chat-stream", ollamaResp.Message.Content)
			}

			if ollamaResp.Done {
				runtime.EventsEmit(a.ctx, "chat-done", nil)
				break
			}
		}
	}()
}

// GetModels 获取本地可用模型列表
func (a *App) GetModels() ([]string, error) {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range result.Models {
		models = append(models, m.Name)
	}
	return models, nil
}