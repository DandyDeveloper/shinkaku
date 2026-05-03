<script>
  import { onMount } from 'svelte';
  import { getVocabReviewQueue, submitVocabGrade } from '$lib/api.js';
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
      const res = await getVocabReviewQueue();
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
      await submitVocabGrade(current.id, grade);
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

<svelte:head><title>Vocab Review — 日本語先生</title></svelte:head>

<main>
  <a href="/vocab" class="back">← Vocabulary</a>
  <h2>Vocab Review Session</h2>

  {#if loading}
    <p class="state-msg">Loading queue…</p>
  {:else if error}
    <div class="alert error">{error}</div>
  {:else if done}
    <div class="done-card">
      <span class="done-icon">🎉</span>
      <h3>All vocab done for today!</h3>
      <p>You reviewed {queue.length} card{queue.length !== 1 ? 's' : ''}.</p>
      <a href="/vocab" class="btn">Back to Vocabulary</a>
    </div>
  {:else if current}
    <div class="progress-bar">
      <div class="progress-fill" style="width: {progress}%"></div>
    </div>
    <p class="progress-label">{currentIndex} / {queue.length}</p>

    <article class="card">
      <p class="jlpt">{current.vocab_word.jlpt_level || 'Unspecified'}</p>
      <h3>{current.vocab_word.word}</h3>

      {#if !revealed}
        <p class="hint">Try to recall reading and meaning before revealing.</p>
      {:else}
        <div class="answer">
          <p><strong>Reading:</strong> {current.vocab_word.reading || '—'}</p>
          <p><strong>Meaning:</strong> {current.vocab_word.meaning}</p>
          {#if current.vocab_word.part_of_speech}<p><strong>POS:</strong> {current.vocab_word.part_of_speech}</p>{/if}
          {#if current.vocab_word.example_jp}<p><strong>JP:</strong> {current.vocab_word.example_jp}</p>{/if}
          {#if current.vocab_word.example_en}<p><strong>EN:</strong> {current.vocab_word.example_en}</p>{/if}
          {#if current.vocab_word.notes}<p><strong>Notes:</strong> {current.vocab_word.notes}</p>{/if}
        </div>
      {/if}
    </article>

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
  }
  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #48bb78, #38a169);
    transition: width 0.25s ease;
  }
  .progress-label {
    margin: 0 0 1rem;
    color: #718096;
    font-size: 0.9rem;
  }

  .card {
    border: 1px solid #e2e8f0;
    border-radius: 14px;
    padding: 1rem;
    background: #fff;
    margin-bottom: 1rem;
  }
  .jlpt {
    margin: 0;
    color: #38a169;
    font-weight: 700;
    font-size: 0.85rem;
  }
  h3 {
    margin: 0.1rem 0 0.8rem;
    font-size: 2rem;
    color: #1a202c;
  }
  .hint { margin: 0; color: #4a5568; }
  .answer p {
    margin: 0.3rem 0;
    color: #2d3748;
  }

  .actions { margin-top: 0.8rem; }
  .btn {
    display: inline-block;
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    padding: 0.55rem 0.9rem;
    background: #fff;
    color: #2d3748;
    text-decoration: none;
    cursor: pointer;
  }
  .btn.primary {
    border-color: #667eea;
    background: #667eea;
    color: #fff;
  }

  .done-card {
    border: 1px solid #bee3f8;
    background: linear-gradient(180deg, #ebf8ff, #ffffff);
    border-radius: 14px;
    padding: 1.4rem;
    text-align: center;
  }
  .done-icon { font-size: 1.5rem; }
  .done-card h3 { margin: 0.45rem 0; font-size: 1.25rem; }
  .done-card p { margin: 0 0 1rem; color: #4a5568; }

  .state-msg { color: #718096; }
  .alert.error {
    background: #fff5f5;
    border: 1px solid #fed7d7;
    color: #c53030;
    border-radius: 10px;
    padding: 0.7rem 0.85rem;
  }
</style>
