<script>
	import { onMount, onDestroy } from 'svelte';
	import * as monaco from 'monaco-editor';

	let editorContainer;
	let editor;
	let currentFilePath = '';
	let currentContent = '';

	export let onSave = () => {};

	// 监听文件打开事件
	onMount(() => {
		// 初始化 Monaco Editor
		editor = monaco.editor.create(editorContainer, {
			value: currentContent || '// 欢迎使用 LeonAgent\n// 打开左侧文件开始编辑',
			language: 'typescript', // 默认语言，后续自动检测
			theme: 'vs-dark',
			fontSize: 15,
			automaticLayout: true,
			minimap: { enabled: true },
			scrollBeyondLastLine: false,
			wordWrap: 'on',
			tabSize: 2,
		});

		// 监听文件打开事件
		window.runtime.EventsOn("file-opened", (data) => {
			currentFilePath = data.path;
			currentContent = data.content;
			
			if (editor) {
				editor.setValue(data.content);
				
				// 自动检测语言
				const ext = currentFilePath.split('.').pop().toLowerCase();
				const languageMap = {
					'go': 'go',
					'py': 'python',
					'js': 'javascript',
					'ts': 'typescript',
					'svelte': 'html',
					'html': 'html',
					'css': 'css',
					'md': 'markdown'
				};
				const lang = languageMap[ext] || 'plaintext';
				monaco.editor.setModelLanguage(editor.getModel(), lang);
			}
		});

		// 保存快捷键 Cmd/Ctrl + S
		editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
			saveFile();
		});
	});

	async function saveFile() {
		if (!currentFilePath || !editor) return;
		
		const newContent = editor.getValue();
		try {
			await window.go.main.App.WriteFile(currentFilePath, newContent);
			alert('✅ 文件保存成功！');
		} catch (e) {
			alert('保存失败: ' + e);
		}
	}

	onDestroy(() => {
		if (editor) editor.dispose();
	});
</script>

<div class="flex flex-col h-full">
	<!-- 编辑器头部 -->
	<div class="bg-gray-900 border-b border-gray-800 px-4 py-2 flex items-center justify-between">
		<div class="flex items-center gap-3">
			<span class="text-sm text-gray-400">当前文件：</span>
			<span class="text-sm font-medium text-white truncate max-w-[400px]">
				{currentFilePath ? currentFilePath.split('/').pop() : '未打开文件'}
			</span>
		</div>
		<button 
			on:click={saveFile}
			class="bg-emerald-600 hover:bg-emerald-500 px-5 py-1.5 rounded-lg text-sm font-medium flex items-center gap-2"
		>
			保存 (⌘S)
		</button>
	</div>

	<!-- Monaco 编辑器容器 -->
	<div bind:this={editorContainer} class="flex-1 overflow-hidden"></div>
</div>