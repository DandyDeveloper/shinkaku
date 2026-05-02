<script>
  import { onMount } from 'svelte';
  import { getReviewQueue, submitGrade } from '$lib/api.js';
  import ReviewCard from '$lib/components/ReviewCard.svelte';
  import GradeButtons from '$lib/components/GradeButtons.svelte';

  let queue = [];
  let currentIndex = 0;
  let revealed = false;
  let loading = true;
  let submitting = false;
  let error = '';
  let done = false;

  $: current = queue[currentIndex] ?? null;
  $: progress = queue.length ? Math.round((currentIndex / queue.length) * 100) : 0;

  onMount(async () => {
    try {
      const res = await getReviewQueue();
      queue = res.cards ?? [];
      if (queue.length === 0) done = true;
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  function reveal() { revealed = true; }

  async function handleGrade(event) {
    const grade = event.detail;
    if (!current || submitting) return;
    submitting = true;
    try {
      await submitGrade(current.id, grade);
      currentIndex += 1;
      revealed = false;
      if (currentIndex >= queue.length) done = true;
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
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
