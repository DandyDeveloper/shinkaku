<script>
  /** @type {{ correct: boolean, explanation: string, correction?: string, natural_alternative?: string, assistant_reply?: string } | null} */
  export let feedback = null;
  export let loading = false;
  export let error = '';
</script>

{#if loading}
  <div class="feedback loading">
    <span class="spinner" aria-label="Loading"></span>
    <p>Asking sensei…</p>
  </div>
{:else if error}
  <div class="feedback error">
    <p class="error-msg">⚠️ {error}</p>
  </div>
{:else if feedback}
  <div class="feedback" class:correct={feedback.correct} class:incorrect={!feedback.correct}>
    <div class="verdict">
      {#if feedback.correct}
        <span class="verdict-icon">✅</span>
        <strong>正解！ Correct!</strong>
      {:else}
        <span class="verdict-icon">❌</span>
        <strong>Not quite — let's look at this…</strong>
      {/if}
    </div>

    <p class="explanation">{feedback.explanation}</p>

    {#if !feedback.correct && feedback.correction}
      <div class="section">
        <h4>Correction</h4>
        <p class="jp-text">{feedback.correction}</p>
      </div>
    {/if}

    {#if feedback.natural_alternative}
      <div class="section">
        <h4>Natural alternative</h4>
        <p class="jp-text">{feedback.natural_alternative}</p>
      </div>
    {/if}

    {#if feedback.assistant_reply}
      <div class="section">
        <h4>Assistant follow-up</h4>
        <p class="jp-text">{feedback.assistant_reply}</p>
      </div>
    {/if}
  </div>
{/if}

<style>
  .feedback {
    border-radius: 10px;
    padding: 1.25rem 1.5rem;
    margin-top: 1.5rem;
    max-width: 640px;
    margin-left: auto;
    margin-right: auto;
    border: 1px solid #e2e8f0;
    background: #f7fafc;
  }
  .feedback.correct  { border-color: #68d391; background: #f0fff4; }
  .feedback.incorrect { border-color: #fc8181; background: #fff5f5; }
  .feedback.loading  { text-align: center; color: #718096; }
  .feedback.error    { border-color: #f6ad55; background: #fffaf0; }

  .verdict {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 1.1rem;
    margin-bottom: 0.75rem;
  }
  .verdict-icon { font-size: 1.4rem; }

  .explanation { color: #4a5568; line-height: 1.6; }

  .section { margin-top: 1rem; }
  .section h4 {
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #718096;
    margin-bottom: 0.25rem;
  }
  .jp-text {
    font-size: 1.1rem;
    color: #2d3748;
    background: white;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    border: 1px solid #e2e8f0;
  }
  .error-msg { color: #c05621; }

  .spinner {
    display: inline-block;
    width: 20px; height: 20px;
    border: 3px solid #e2e8f0;
    border-top-color: #667eea;
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
    margin-right: 0.5rem;
    vertical-align: middle;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
