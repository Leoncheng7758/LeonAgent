<script>
	import { onMount } from 'svelte';
	let messages = [];
	let input = '';
	let isLoading = false;
	let selectedModel = 'qwen2.5-coder:7b';
	let models = [];

	let currentResponse = '';

	async function loadModels() {
		try {
			const res = await window.go.main.App.GetModels();
			models = res;
			if (models.length > 0) selectedModel = models[0];
		} catch(e) {
			console.error(e);
		}
	}

	async function sendMessage() {
		if (!input.trim() || isLoading) return;

		messages = [...messages, { role: 'user', content: input }];
		const userMsg = input;
		input = '';
		isLoading = true;
		currentResponse = '';

		try {
			await window.go.main.App.Chat({
				model: selectedModel,
				message: userMsg,
				history: messages.slice(0, -1)
			});
		} catch(e) {
			console.error(e);
		}
	}

	// 监听流式返回
	onMount(() => {
		loadModels();

		window.runtime.EventsOn("chat-stream", (chunk) => {
			currentResponse += chunk;
			// 更新最后一条 assistant 消息
			if (messages.length > 0 && messages[messages.length-1].role === 'assistant') {
				messages[messages.length-1].content = currentResponse;
			} else {
				messages = [...messages, { role: 'assistant', content: currentResponse }];
			}
		});

		window.runtime.EventsOn("chat-done", () => {
			isLoading = false;
			currentResponse = '';
		});

		window.runtime.EventsOn("chat-error", (err) => {
			isLoading = false;
			messages = [...messages, { role: 'assistant', content: `❌ 错误: ${err}` }];
		});
	});
</script>

<div class="h-screen flex flex-col bg-gray-950 text-white">
	<!-- Header -->
	<div class="p-4 border-b border-gray-800 flex items-center justify-between bg-gray-900">
		<h1 class="text-xl font-bold flex items-center gap-2">
			<span>🧠</span> LeonAgent
		</h1>
		<select bind:value={selectedModel} class="bg-gray-800 px-4 py-2 rounded-lg border border-gray-700">
			{#each models as model}
				<option value={model}>{model}</option>
			{/each}
		</select>
	</div>

	<!-- Messages -->
	<div class="flex-1 overflow-y-auto p-6 space-y-6">
		{#each messages as msg}
			<div class={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
				<div class={`max-w-[80%] px-5 py-3 rounded-2xl ${
					msg.role === 'user' ? 'bg-blue-600' : 'bg-gray-800'
				}`}>
					{msg.content}
				</div>
			</div>
		{/each}
		{#if isLoading && currentResponse}
			<div class="flex justify-start">
				<div class="bg-gray-800 px-5 py-3 rounded-2xl">
					{currentResponse}
				</div>
			</div>
		{/if}
	</div>

	<!-- Input -->
	<div class="p-4 border-t border-gray-800 bg-gray-900">
		<div class="flex gap-3">
			<input
			bind:value={input}
			on:keydown={(e) => e.key === 'Enter' && sendMessage()}
			placeholder="输入编程需求，例如：帮我写一个 Go HTTP API..."
			class="flex-1 bg-gray-800 border border-gray-700 rounded-xl px-5 py-4 focus:outline-none focus:border-blue-500"
			disabled={isLoading}
			/>
			<button
			on:click={sendMessage}
			disabled={isLoading || !input.trim()}
			class="bg-emerald-600 hover:bg-emerald-500 px-8 rounded-xl disabled:bg-gray-700"
			>
				发送
			</button>
		</div>
	</div>
</div>