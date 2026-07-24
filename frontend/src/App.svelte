<script>
  import { onMount } from 'svelte';

  let projectPath = '';
  let files = [];
  let activeFile = '';
  let content = '';
  let messages = [{ role: 'assistant', content: 'LeonAgent 已就绪。选择项目目录后，可以浏览代码并与本地模型对话。' }];
  let input = '';
  let models = [];
  let selectedModel = '';
  let busy = false;
  let branch = '';
  let changes = [];
  let stream = '';

  onMount(async () => {
    window.runtime.EventsOn('chat-stream', chunk => {
      stream += chunk;
      const last = messages[messages.length - 1];
      if (last?.role === 'assistant' && last.streaming) {
        last.content = stream;
        messages = [...messages];
      } else {
        messages = [...messages, { role: 'assistant', content: stream, streaming: true }];
      }
    });
    window.runtime.EventsOn('chat-done', () => { busy = false; stream = ''; messages = messages.map(m => ({...m, streaming: false})); });
    window.runtime.EventsOn('chat-error', error => { busy = false; messages = [...messages, { role: 'assistant', content: `错误：${error}` }]; });
    try { models = await window.go.main.App.GetModels(); selectedModel = models[0] || ''; } catch (_) {}
    try { projectPath = await window.go.main.App.GetHomeDir(); await openFolder(projectPath); } catch (_) {}
  });

  async function openFolder(path) {
    projectPath = path;
    files = await window.go.main.App.ListDir(path);
    await refreshGit();
  }

  async function openEntry(entry) {
    if (entry.isDir) return openFolder(entry.path);
    activeFile = entry.path;
    content = await window.go.main.App.ReadFile(entry.path);
  }

  async function saveFile() {
    if (!activeFile) return;
    await window.go.main.App.WriteFile(activeFile, content);
    await refreshGit();
  }

  async function refreshGit() {
    try {
      const status = await window.go.main.App.GetGitStatus(projectPath);
      branch = status.branch;
      changes = status.changes || [];
    } catch (_) { branch = ''; changes = []; }
  }

  async function send() {
    const text = input.trim();
    if (!text || busy || !selectedModel) return;
    const history = messages.filter(m => !m.streaming).map(({role, content}) => ({role, content}));
    messages = [...messages, { role: 'user', content: text }];
    input = '';
    busy = true;
    stream = '';
    await window.go.main.App.Chat({ model: selectedModel, message: text, history });
  }
</script>

<div class="app-shell">
  <header>
    <div><strong>LeonAgent</strong><span>Local coding workspace</span></div>
    <select bind:value={selectedModel} aria-label="模型">
      {#if models.length === 0}<option value="">Ollama 未连接</option>{/if}
      {#each models as model}<option value={model}>{model}</option>{/each}
    </select>
  </header>

  <main>
    <aside class="explorer">
      <div class="panel-title"><span>项目</span><button on:click={() => openFolder(projectPath)}>刷新</button></div>
      <div class="path">{projectPath || '未选择目录'}</div>
      <div class="file-list">
        {#each files as file}
          <button class:active={activeFile === file.path} on:click={() => openEntry(file)}>
            <span>{file.isDir ? '▸' : '·'}</span><span>{file.name}</span>
          </button>
        {/each}
      </div>
      <div class="git-box">
        <div><b>Git</b><span>{branch || '非仓库'}</span></div>
        <small>{changes.length} 个变更</small>
        {#each changes.slice(0, 6) as change}<code>{change}</code>{/each}
      </div>
    </aside>

    <section class="editor">
      <div class="panel-title"><span>{activeFile ? activeFile.split('/').pop() : '编辑器'}</span><button disabled={!activeFile} on:click={saveFile}>保存 ⌘S</button></div>
      <textarea bind:value={content} spellcheck="false" placeholder="从左侧选择文件"></textarea>
    </section>

    <section class="chat">
      <div class="panel-title"><span>Agent</span><span class:online={selectedModel}>{selectedModel ? '本地模型在线' : '等待 Ollama'}</span></div>
      <div class="messages">
        {#each messages as message}
          <article class:user={message.role === 'user'}><b>{message.role === 'user' ? '你' : 'LeonAgent'}</b><p>{message.content}</p></article>
        {/each}
      </div>
      <div class="composer">
        <textarea bind:value={input} on:keydown={(e) => e.key === 'Enter' && !e.shiftKey && (e.preventDefault(), send())} placeholder="描述要分析或实现的任务…"></textarea>
        <button on:click={send} disabled={busy || !selectedModel || !input.trim()}>{busy ? '生成中' : '发送'}</button>
      </div>
    </section>
  </main>
</div>
