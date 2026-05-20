package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ==================== Ollama Chat (保留) ====================
type ChatRequest struct {
	Model   string        `json:"model"`
	Message string        `json:"message"`
	History []ChatMessage `json:"history"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (a *App) Chat(req ChatRequest) {
	// ... (保持你上一个版本的 Chat 实现)
	go func() {
		// 流式实现代码（与之前相同）
		fmt.Println("Chat request received:", req.Message)
		runtime.EventsEmit(a.ctx, "chat-stream", "正在思考中...")
		// 实际 Ollama 调用代码保持之前版本
	}()
}

// ==================== 文件浏览器 & 操作 ====================

type FileEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	Size     int64  `json:"size"`
	ModTime  string `json:"modTime"`
}

// ListDir 列出目录内容
func (a *App) ListDir(path string) ([]FileEntry, error) {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = home
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileEntry
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, FileEntry{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
	}

	// 文件夹排前面
	for i := range files {
		for j := i + 1; j < len(files); j++ {
			if files[i].IsDir != files[j].IsDir {
				if files[i].IsDir {
					continue
				}
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	return files, nil
}

// ReadFile 读取文件内容
func (a *App) ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile 写入文件
func (a *App) WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// CreateFolder 创建文件夹
func (a *App) CreateFolder(path string) error {
	return os.MkdirAll(path, 0755)
}

// GetHomeDir 获取用户主目录
func (a *App) GetHomeDir() (string, error) {
	return os.UserHomeDir()
}