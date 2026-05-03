<script>
  import { onMount } from 'svelte';
  import { getReviewQueue, submitGrade, getConversationPrompt, gradeConversationReply } from '$lib/api.js';
  import ReviewCard from '$lib/components/ReviewCard.svelte';
  import GradeButtons from '$lib/components/GradeButtons.svelte';
  import LLMFeedback from '$lib/components/LLMFeedback.svelte';
  import { getSettings } from '$lib/settings.js';

  let queue = [];
  let currentIndex = 0;
  let revealed = false;
  let loading = true;
  let submitting = false;
  let error = '';
  let done = false;

  let reviewMode = 'self';
  let conversationPrompt = null;
  let conversationReply = '';
  let conversationFeedback = null;
  let conversationLoading = false;
  let conversationError = '';
  let suggestedGrade = null;
  let includeConversationFurigana = false;

  $: current = queue[currentIndex] ?? null;
  $: progress = queue.length ? Math.round((currentIndex / queue.length) * 100) : 0;

  onMount(async () => {
    try {
      const res = await getReviewQueue();
      queue = res.cards ?? [];
      includeConversationFurigana = getSettings().includeConversationFurigana;
      if (queue.length === 0) done = true;
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  function reveal() { revealed = true; }

  function resetConversationState() {
    reviewMode = 'self';
    conversationPrompt = null;
    conversationReply = '';
    conversationFeedback = null;
    conversationLoading = false;
    conversationError = '';
    suggestedGrade = null;
  }

  function switchReviewMode(mode) {
    reviewMode = mode;
    if (mode === 'conversation' && !conversationPrompt && !conversationLoading) {
      startConversation();
    }
  }

  async function startConversation() {
    if (!current) return;
    conversationLoading = true;
    conversationError = '';
    conversationFeedback = null;
    conversationReply = '';
    suggestedGrade = null;
    try {
      conversationPrompt = await getConversationPrompt(current.grammar_point.id, includeConversationFurigana);
    } catch (e) {
      conversationError = e.message;
    } finally {
      conversationLoading = false;
    }
  }

  async function handleConversationSubmit() {
    if (!current || !conversationPrompt || !conversationReply.trim()) return;
    conversationLoading = true;
    conversationError = '';
    conversationFeedback = null;
    try {
      conversationFeedback = await gradeConversationReply(
        current.grammar_point.id,
        conversationPrompt.scenario,
        conversationPrompt.assistant_message,
        conversationReply.trim(),
        includeConversationFurigana
      );
      // Keep this conservative: "correct" should advance, "incorrect" should fail.
      suggestedGrade = conversationFeedback.correct ? 4 : 2;
    } catch (e) {
      conversationError = e.message;
    } finally {
      conversationLoading = false;
    }
  }

  async function submitCardGrade(grade) {
    if (!current || submitting) return;
    submitting = true;
    try {
      await submitGrade(current.id, grade);
      currentIndex += 1;
      revealed = false;
      resetConversationState();
      if (currentIndex >= queue.length) done = true;
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }

  async function handleGrade(event) {
    await submitCardGrade(event.detail);
  }

  async function applySuggestedGrade() {
    if (suggestedGrade === null) return;
    await submitCardGrade(suggestedGrade);
  }

  function handleConversationKeydown(e) {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleConversationSubmit();
    }
  }
</script>

<svelte:head><title>Review — 日本語先生</title></svelte:head>

<main>
  <a href="/" class="back">← Dashboard</a>
  <h2>Review Session</h2>

  {#if loading}
    <p class="state-msg">Loading queue…</p>
  {:else if error}
    <div class="alert error">{error}</div>
  {:else if done}
    <div class="done-card">
      <span class="done-icon">🎉</span>
      <h3>All done for today!</h3>
      <p>You reviewed {queue.length} card{queue.length !== 1 ? 's' : ''}.</p>
      <a href="/" class="btn">Back to Dashboard</a>
    </div>
  {:else if current}
    <div class="progress-bar">
      <div class="progress-fill" style="width: {progress}%"></div>
    </div>
    <p class="progress-label">{currentIndex} / {queue.length}</p>

    <ReviewCard grammarPoint={current.grammar_point} {revealed} />

    <div class="actions">
      {#if !revealed}
        <button class="btn primary" on:click={reveal}>Show Answer</button>
      {:else}
        <div class="mode-switcher">
          <button class="mode-tab" class:active={reviewMode === 'self'} on:click={() => switchReviewMode('self')}>
            Self grade
          </button>
          <button class="mode-tab" class:active={reviewMode === 'conversation'} on:click={() => switchReviewMode('conversation')}>
            Conversation check
          </button>
        </div>

        {#if reviewMode === 'conversation'}
          <div class="conversation-card">
            <div class="conversation-head">
              <span>Practice this grammar in context</span>
              <button class="btn" on:click={startConversation} disabled={conversationLoading}>New prompt</button>
            </div>

            {#if conversationPrompt}
              <p class="scenario">{conversationPrompt.scenario}</p>
              <div class="assistant-msg">{conversationPrompt.assistant_message}</div>

              <label class="reply-label" for="conversation-reply">Reply in Japanese using <strong>{current.grammar_point.pattern}</strong></label>
              <textarea
                id="conversation-reply"
                bind:value={conversationReply}
                rows="4"
                on:keydown={handleConversationKeydown}
                placeholder="自然な返事を書いてください…"
                disabled={conversationLoading}
              ></textarea>

              <div class="conversation-actions">
                <span class="hint">Ctrl/Cmd + Enter to submit</span>
                <button class="btn primary" on:click={handleConversationSubmit} disabled={conversationLoading || !conversationReply.trim()}>
                  {conversationLoading ? 'Checking…' : 'Check reply'}
                </button>
              </div>
            {/if}

            <LLMFeedback feedback={conversationFeedback} loading={conversationLoading && !conversationPrompt} error={conversationError} />

            {#if suggestedGrade !== null}
              <div class="suggestion-row">
                <span>Suggested grade: <strong>{suggestedGrade}</strong></span>
                <button class="btn primary" on:click={applySuggestedGrade} disabled={submitting}>Use suggested grade</button>
              </div>
            {/if}
          </div>
        {/if}

        <GradeButtons disabled={submitting} on:grade={handleGrade} />
      {/if}
    </div>
  {/if}
</main>

<style>
  main {
    max-width: 720px;
    margin: 0 auto;
    padding: 2rem 1rem;
    font-family: system-ui, -apple-system, sans-serif;
  }
  .back { color: #667eea; text-decoration: none; font-size: 0.9rem; }
  h2 { margin: 0.5rem 0 1.5rem; color: #2d3748; }

  .progress-bar {
    height: 6px;
    background: #e2e8f0;
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 0.4rem;
    max-width: 640px;
    margin-left: auto;
    margin-right: auto;
  }
  .progress-fill {
    height: 100%;
    background: #667eea;
    transition: width 0.3s ease;
  }
  .progress-label {
    text-align: center;
    font-size: 0.8rem;
    color: #a0aec0;
    margin-bottom: 1.5rem;
  }

  .actions { margin-top: 1.5rem; text-align: center; }

  .mode-switcher {
    display: flex;
    justify-content: center;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }
  .mode-tab {
    padding: 0.5rem 0.85rem;
    border: 1px solid #d7deeb;
    background: #fff;
    border-radius: 999px;
    font-size: 0.85rem;
    font-weight: 700;
    color: #4a5568;
    cursor: pointer;
  }
  .mode-tab.active {
    background: #667eea;
    color: #fff;
    border-color: #667eea;
  }

  .conversation-card {
    max-width: 640px;
    margin: 0 auto 1rem;
    padding: 1rem;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    background: #fff;
    text-align: left;
  }
  .conversation-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
    font-size: 0.9rem;
    color: #4a5568;
  }
  .scenario {
    margin: 0 0 0.5rem;
    color: #4a5568;
    font-size: 0.92rem;
  }
  .assistant-msg {
    padding: 0.7rem 0.9rem;
    background: #f7fafc;
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    margin-bottom: 0.75rem;
    color: #2d3748;
  }
  .reply-label {
    display: block;
    margin-bottom: 0.4rem;
    color: #4a5568;
    font-size: 0.9rem;
  }
  textarea {
    width: 100%;
    box-sizing: border-box;
    border: 1px solid #d7deeb;
    border-radius: 8px;
    padding: 0.7rem;
    font-size: 1rem;
    resize: vertical;
    font-family: inherit;
  }
  textarea:focus {
    outline: none;
    border-color: #667eea;
  }
  .conversation-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.55rem;
  }
  .hint {
    font-size: 0.75rem;
    color: #a0aec0;
  }
  .suggestion-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.75rem;
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid #edf2f7;
    color: #4a5568;
    font-size: 0.9rem;
  }

  .btn {
    display: inline-block;
    padding: 0.65rem 1.5rem;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    text-decoration: none;
    border: none;
    background: #e2e8f0;
    color: #2d3748;
  }
  .btn.primary { background: #667eea; color: white; }
  .btn:hover { opacity: 0.9; }

  .done-card {
    text-align: center;
    padding: 3rem 1rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
  }
  .done-icon { font-size: 3rem; }
  .done-card h3 { font-size: 1.5rem; color: #2d3748; margin: 0; }
  .done-card p { color: #718096; margin: 0; }

  .state-msg { text-align: center; color: #a0aec0; padding: 3rem 0; }
  .alert.error { background: #fff5f5; border: 1px solid #fc8181; color: #c53030;
    border-radius: 8px; padding: 1rem; }
</style>
