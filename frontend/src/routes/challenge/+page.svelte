<script>
  import { onMount } from 'svelte';
  import { listGrammar, gradeChallenge, getConversationPrompt, gradeConversationReply } from '$lib/api.js';
  import LLMFeedback from '$lib/components/LLMFeedback.svelte';

  let grammarPoints = [];
  let selectedPoint = null;
  let mode = 'sentence';

  let userSentence = '';
  let feedback = null;
  let llmLoading = false;
  let llmError = '';

  let conversationPrompt = null;
  let conversationReply = '';
  let conversationFeedback = null;
  let conversationLoading = false;
  let conversationError = '';

  let loadError = '';

  onMount(async () => {
    try {
      grammarPoints = await listGrammar();
      if (grammarPoints.length > 0) {
        const params = new URLSearchParams(window.location.search);
        const requestedId = Number(params.get('grammar'));
        const requestedMode = params.get('mode');

        selectedPoint = grammarPoints.find((point) => point.id === requestedId)
          ?? grammarPoints[Math.floor(Math.random() * grammarPoints.length)];

        if (requestedMode === 'conversation') {
          mode = 'conversation';
          await startConversation();
        }
      }
    } catch (e) {
      loadError = e.message;
    }
  });

  async function pickRandom() {
    if (!grammarPoints.length) return;
    selectedPoint = grammarPoints[Math.floor(Math.random() * grammarPoints.length)];
    resetAll();
    if (mode === 'conversation') {
      await startConversation();
    }
  }

  function resetSentence() {
    userSentence = '';
    feedback = null;
    llmError = '';
  }

  function resetConversation() {
    conversationPrompt = null;
    conversationReply = '';
    conversationFeedback = null;
    conversationError = '';
  }

  function resetAll() {
    resetSentence();
    resetConversation();
  }

  function switchMode(nextMode) {
    if (mode === nextMode) return;
    mode = nextMode;
    if (nextMode === 'sentence') {
      resetSentence();
      return;
    }
    resetConversation();
    startConversation();
  }

  async function handleSentenceSubmit() {
    if (!selectedPoint || !userSentence.trim()) return;
    llmLoading = true;
    feedback = null;
    llmError = '';
    try {
      feedback = await gradeChallenge(selectedPoint.id, userSentence.trim());
    } catch (e) {
      llmError = e.message;
    } finally {
      llmLoading = false;
    }
  }

  async function startConversation() {
    if (!selectedPoint) return;
    conversationLoading = true;
    conversationPrompt = null;
    conversationFeedback = null;
    conversationReply = '';
    conversationError = '';
    try {
      conversationPrompt = await getConversationPrompt(selectedPoint.id);
    } catch (e) {
      conversationError = e.message;
    } finally {
      conversationLoading = false;
    }
  }

  async function handleConversationSubmit() {
    if (!selectedPoint || !conversationPrompt || !conversationReply.trim()) return;
    conversationLoading = true;
    conversationFeedback = null;
    conversationError = '';
    try {
      conversationFeedback = await gradeConversationReply(
        selectedPoint.id,
        conversationPrompt.scenario,
        conversationPrompt.assistant_message,
        conversationReply.trim()
      );
    } catch (e) {
      conversationError = e.message;
    } finally {
      conversationLoading = false;
    }
  }

  function handleSentenceKeydown(e) {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleSentenceSubmit();
    }
  }

  function handleConversationKeydown(e) {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleConversationSubmit();
    }
  }
</script>

<svelte:head><title>Challenge — 日本語先生</title></svelte:head>

<main>
  <a href="/" class="back">← Dashboard</a>
  <h2>Grammar Challenge</h2>

  {#if loadError}
    <div class="alert error">{loadError}</div>
  {:else if grammarPoints.length === 0}
    <div class="alert info">
      No grammar points yet. <a href="/grammar">Add some first →</a>
    </div>
  {:else if selectedPoint}
    <div class="grammar-display">
      <div class="grammar-header">
        {#if selectedPoint.jlpt_level}
          <span class="badge">{selectedPoint.jlpt_level}</span>
        {/if}
        <h3 class="pattern">{selectedPoint.pattern}</h3>
        <button class="icon-btn" on:click={pickRandom} title="Pick random grammar point">🔀</button>
      </div>
      <p class="meaning">{selectedPoint.meaning}</p>
      {#if selectedPoint.example_jp}
        <div class="example">
          <span class="example-jp">{selectedPoint.example_jp}</span>
          {#if selectedPoint.example_en}
            <span class="example-en"> — {selectedPoint.example_en}</span>
          {/if}
        </div>
      {/if}
    </div>

    <div class="mode-switcher">
      <button class="mode-tab" class:active={mode === 'sentence'} on:click={() => switchMode('sentence')}>Write a sentence</button>
      <button class="mode-tab" class:active={mode === 'conversation'} on:click={() => switchMode('conversation')}>Have a conversation</button>
    </div>

    {#if mode === 'sentence'}
      <div class="input-area">
        <label for="sentence">Your sentence using <strong>{selectedPoint.pattern}</strong></label>
        <textarea
          id="sentence"
          bind:value={userSentence}
          on:keydown={handleSentenceKeydown}
          placeholder="日本語で書いてください…"
          rows="3"
          disabled={llmLoading}
        ></textarea>
        <div class="input-actions">
          <span class="hint">Ctrl/Cmd + Enter to submit</span>
          <button class="btn primary" on:click={handleSentenceSubmit} disabled={llmLoading || !userSentence.trim()}>
            {llmLoading ? 'Grading…' : 'Submit'}
          </button>
        </div>
      </div>

      <LLMFeedback {feedback} loading={llmLoading} error={llmError} />
    {:else}
      <div class="conversation-card">
        <div class="conversation-meta">
          <span class="eyebrow">Roleplay</span>
          <button class="btn" on:click={startConversation} disabled={conversationLoading}>New prompt</button>
        </div>

        {#if conversationPrompt}
          <p class="scenario">{conversationPrompt.scenario}</p>
          <div class="message-row assistant">
            <div class="bubble assistant-bubble">{conversationPrompt.assistant_message}</div>
          </div>

          <div class="input-area conversation-input">
            <label for="conversation-reply">Reply in Japanese using <strong>{selectedPoint.pattern}</strong></label>
            <textarea
              id="conversation-reply"
              bind:value={conversationReply}
              on:keydown={handleConversationKeydown}
              placeholder="自然な返事を書いてください…"
              rows="4"
              disabled={conversationLoading}
            ></textarea>
            <div class="input-actions">
              <span class="hint">Ctrl/Cmd + Enter to submit</span>
              <button class="btn primary" on:click={handleConversationSubmit} disabled={conversationLoading || !conversationReply.trim()}>
                {conversationLoading ? 'Grading…' : 'Send reply'}
              </button>
            </div>
          </div>
        {/if}

        <LLMFeedback feedback={conversationFeedback} loading={conversationLoading && !conversationPrompt} error={conversationError} />
        {#if conversationPrompt && conversationLoading}
          <div class="loading-inline">Asking sensei to review your reply…</div>
        {/if}
        {#if conversationFeedback}
          <div class="message-row learner">
            <div class="bubble learner-bubble">{conversationReply}</div>
          </div>
        {/if}
      </div>
    {/if}
  {/if}
</main>

<style>
  main {
    max-width: 760px;
    margin: 0 auto;
    padding: 2rem 1rem;
    font-family: system-ui, -apple-system, sans-serif;
  }
  .back { color: #667eea; text-decoration: none; font-size: 0.9rem; }
  h2 { margin: 0.5rem 0 1.5rem; color: #2d3748; }

  .grammar-display {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.5rem;
    margin-bottom: 1rem;
  }
  .grammar-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.5rem;
  }
  .pattern { font-size: 1.8rem; font-weight: 700; color: #2d3748; margin: 0; flex: 1; }
  .badge {
    background: #ebf8ff;
    color: #2b6cb0;
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
  }
  .icon-btn {
    background: none;
    border: none;
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 4px;
  }
  .icon-btn:hover { background: #f7fafc; }
  .meaning { color: #4a5568; margin-bottom: 0.75rem; }
  .example {
    font-size: 0.9rem;
    background: #f7fafc;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    border-left: 3px solid #667eea;
  }
  .example-jp { color: #2d3748; }
  .example-en { color: #718096; }

  .mode-switcher {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }
  .mode-tab {
    padding: 0.6rem 0.9rem;
    border: 1px solid #d7deeb;
    border-radius: 999px;
    background: #fff;
    color: #4a5568;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
  }
  .mode-tab.active {
    background: #667eea;
    color: #fff;
    border-color: #667eea;
  }

  .input-area { margin-bottom: 0.5rem; }
  .input-area label { display: block; font-size: 0.9rem; color: #4a5568; margin-bottom: 0.5rem; }
  textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 1.05rem;
    resize: vertical;
    box-sizing: border-box;
    font-family: inherit;
    transition: border-color 0.15s;
  }
  textarea:focus { outline: none; border-color: #667eea; }
  .input-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 0.5rem;
    gap: 0.75rem;
  }
  .hint { font-size: 0.75rem; color: #a0aec0; }

  .conversation-card {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1rem;
  }
  .conversation-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    margin-bottom: 0.75rem;
  }
  .eyebrow {
    font-size: 0.75rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #718096;
  }
  .scenario {
    margin: 0 0 1rem;
    color: #4a5568;
    line-height: 1.5;
  }
  .message-row {
    display: flex;
    margin-bottom: 0.75rem;
  }
  .message-row.learner {
    justify-content: flex-end;
    margin-top: 1rem;
  }
  .bubble {
    max-width: 85%;
    padding: 0.75rem 0.9rem;
    border-radius: 14px;
    line-height: 1.5;
    font-size: 1rem;
  }
  .assistant-bubble {
    background: #f7fafc;
    color: #2d3748;
    border: 1px solid #e2e8f0;
  }
  .learner-bubble {
    background: #667eea;
    color: #fff;
  }
  .conversation-input {
    margin-top: 1rem;
  }
  .loading-inline {
    color: #718096;
    text-align: center;
    margin-top: 0.75rem;
  }

  .btn {
    padding: 0.65rem 1.25rem;
    border-radius: 8px;
    font-size: 0.95rem;
    font-weight: 600;
    cursor: pointer;
    border: 1px solid #e2e8f0;
    background: #f7fafc;
    color: #2d3748;
    transition: opacity 0.15s;
  }
  .btn.primary { background: #667eea; color: white; border-color: #667eea; }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .btn:hover:not(:disabled) { opacity: 0.85; }

  .alert {
    border-radius: 8px;
    padding: 1rem 1.25rem;
    margin-bottom: 1.5rem;
  }
  .alert.error { background: #fff5f5; border: 1px solid #fc8181; color: #c53030; }
  .alert.info  { background: #ebf8ff; border: 1px solid #bee3f8; color: #2c5282; }
  .alert a { color: inherit; font-weight: 600; }
</style>
