<script>
  import { onMount } from 'svelte';
  import { listGrammar, gradeChallenge, createGrammar } from '$lib/api.js';
  import LLMFeedback from '$lib/components/LLMFeedback.svelte';

  let grammarPoints = [];
  let selectedPoint = null;
  let userSentence = '';
  let feedback = null;
  let llmLoading = false;
  let llmError = '';
  let addedToSRS = false;
  let loadError = '';

  onMount(async () => {
    try {
      grammarPoints = await listGrammar();
      if (grammarPoints.length > 0) {
        selectedPoint = grammarPoints[Math.floor(Math.random() * grammarPoints.length)];
      }
    } catch (e) {
      loadError = e.message;
    }
  });

  function pickRandom() {
    if (!grammarPoints.length) return;
    selectedPoint = grammarPoints[Math.floor(Math.random() * grammarPoints.length)];
    reset();
  }

  function reset() {
    userSentence = '';
    feedback = null;
    llmError = '';
    addedToSRS = false;
  }

  async function handleSubmit() {
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

  async function addToSRS() {
    if (!selectedPoint || addedToSRS) return;
    try {
      // The grammar point already exists — just confirm with the user.
      // If it wasn't created yet, call createGrammar.
      // Here we just mark it as added (SRS card is auto-created on grammar creation).
      addedToSRS = true;
    } catch (e) {
      alert('Failed to add: ' + e.message);
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleSubmit();
    }
  }
</script>

<svelte:head><title>Challenge — 日本語先生</title></svelte:head>

<main>
  <a href="/" class="back">← Dashboard</a>
  <h2>Sentence Challenge</h2>

  {#if loadError}
    <div class="alert error">{loadError}</div>
  {:else if grammarPoints.length === 0}
    <div class="alert info">
      No grammar points yet. <a href="/grammar">Add some first →</a>
    </div>
  {:else}
    {#if selectedPoint}
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

      <div class="input-area">
        <label for="sentence">Your sentence using <strong>{selectedPoint.pattern}</strong></label>
        <textarea
          id="sentence"
          bind:value={userSentence}
          on:keydown={handleKeydown}
          placeholder="日本語で書いてください…"
          rows="3"
          disabled={llmLoading}
        ></textarea>
        <div class="input-actions">
          <span class="hint">⌘ + Enter to submit</span>
          <button
            class="btn primary"
            on:click={handleSubmit}
            disabled={llmLoading || !userSentence.trim()}
          >
            {llmLoading ? 'Grading…' : 'Submit'}
          </button>
        </div>
      </div>

      <LLMFeedback {feedback} loading={llmLoading} error={llmError} />

      {#if feedback && !addedToSRS}
        <div class="srs-prompt">
          <button class="btn" on:click={addToSRS}>+ Add to SRS deck</button>
          <span class="srs-hint">Review this pattern with spaced repetition</span>
        </div>
      {/if}
      {#if addedToSRS}
        <p class="added-msg">✅ Added to your SRS deck!</p>
      {/if}
    {/if}
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

  .grammar-display {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
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

  .input-area { margin-bottom: 0.5rem; }
  .input-area label { display: block; font-size: 0.9rem; color: #4a5568; margin-bottom: 0.5rem; }
  textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 1.1rem;
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
  }
  .hint { font-size: 0.75rem; color: #a0aec0; }

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

  .srs-prompt {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 1.25rem;
    max-width: 640px;
    margin-left: auto;
    margin-right: auto;
  }
  .srs-hint { font-size: 0.8rem; color: #718096; }
  .added-msg { text-align: center; color: #38a169; margin-top: 1rem; font-weight: 600; }

  .alert {
    border-radius: 8px;
    padding: 1rem 1.25rem;
    margin-bottom: 1.5rem;
  }
  .alert.error { background: #fff5f5; border: 1px solid #fc8181; color: #c53030; }
  .alert.info  { background: #ebf8ff; border: 1px solid #bee3f8; color: #2c5282; }
  .alert a { color: inherit; font-weight: 600; }
</style>
