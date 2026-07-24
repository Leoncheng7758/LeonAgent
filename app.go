package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	httpClient *http.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model   string        `json:"model"`
	Message string        `json:"message"`
	History []ChatMessage `json:"history"`
}

type FileEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

type GitStatus struct {
	Branch  string   `json:"branch"`
	Changes []string `json:"changes"`
}

type ollamaChunk struct {
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
	Error   string      `json:"error"`
}

func NewApp() *App {
	return &App{httpClient: &http.Client{Timeout: 10 * time.Minute}}
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

func (a *App) GetHomeDir() (string, error) { return os.UserHomeDir() }

func (a *App) ListDir(path string) ([]FileEntry, error) {
	if path == "" {
		var err error
		path, err = os.UserHomeDir()
		if err != nil { return nil, err }
	}
	entries, err := os.ReadDir(path)
	if err != nil { return nil, err }
	result := make([]FileEntry, 0, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") { continue }
		info, err := entry.Info()
		if err != nil { continue }
		result = append(result, FileEntry{Name: entry.Name(), Path: filepath.Join(path, entry.Name()), IsDir: entry.IsDir(), Size: info.Size(), ModTime: info.ModTime().Format("2006-01-02 15:04")})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir { return result[i].IsDir }
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	return result, nil
}

func (a *App) ReadFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil { return "", err }
	if info.IsDir() { return "", errors.New("cannot read a directory") }
	if info.Size() > 2*1024*1024 { return "", errors.New("file exceeds 2 MB safety limit") }
	content, err := os.ReadFile(path)
	return string(content), err
}

func (a *App) WriteFile(path, content string) error {
	if strings.TrimSpace(path) == "" { return errors.New("path is required") }
	return os.WriteFile(path, []byte(content), 0o644)
}

func (a *App) GetModels() ([]string, error) {
	resp, err := a.httpClient.Get("http://127.0.0.1:11434/api/tags")
	if err != nil { return nil, fmt.Errorf("Ollama unavailable: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("Ollama returned %s", resp.Status) }
	var data struct { Models []struct { Name string `json:"name"` } `json:"models"` }
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil { return nil, err }
	models := make([]string, 0, len(data.Models))
	for _, model := range data.Models { models = append(models, model.Name) }
	return models, nil
}

func (a *App) Chat(req ChatRequest) {
	go func() {
		messages := append([]ChatMessage{}, req.History...)
		messages = append(messages, ChatMessage{Role: "user", Content: req.Message})
		payload, _ := json.Marshal(map[string]any{"model": req.Model, "messages": messages, "stream": true})
		resp, err := a.httpClient.Post("http://127.0.0.1:11434/api/chat", "application/json", bytes.NewReader(payload))
		if err != nil { runtime.EventsEmit(a.ctx, "chat-error", err.Error()); return }
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK { runtime.EventsEmit(a.ctx, "chat-error", resp.Status); return }
		scanner := bufio.NewScanner(resp.Body)
		buffer := make([]byte, 0, 64*1024)
		scanner.Buffer(buffer, 1024*1024)
		for scanner.Scan() {
			var chunk ollamaChunk
			if json.Unmarshal(scanner.Bytes(), &chunk) != nil { continue }
			if chunk.Error != "" { runtime.EventsEmit(a.ctx, "chat-error", chunk.Error); return }
			if chunk.Message.Content != "" { runtime.EventsEmit(a.ctx, "chat-stream", chunk.Message.Content) }
			if chunk.Done { runtime.EventsEmit(a.ctx, "chat-done"); return }
		}
		if err := scanner.Err(); err != nil { runtime.EventsEmit(a.ctx, "chat-error", err.Error()) }
	}()
}

func (a *App) GetGitStatus(projectPath string) (GitStatus, error) {
	branch, err := runGit(projectPath, "branch", "--show-current")
	if err != nil { return GitStatus{}, err }
	status, err := runGit(projectPath, "status", "--short")
	if err != nil { return GitStatus{}, err }
	changes := []string{}
	if strings.TrimSpace(status) != "" { changes = strings.Split(strings.TrimSpace(status), "\n") }
	return GitStatus{Branch: strings.TrimSpace(branch), Changes: changes}, nil
}

func (a *App) GetGitDiff(projectPath string) (string, error) { return runGit(projectPath, "diff", "--", ".") }

func runGit(projectPath string, args ...string) (string, error) {
	if strings.TrimSpace(projectPath) == "" { return "", errors.New("project path is required") }
	cmd := exec.Command("git", append([]string{"-C", projectPath}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil { return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(output))) }
	return string(output), nil
}
