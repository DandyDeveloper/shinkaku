<script>
  import { onMount } from 'svelte';
  import { getReviewQueue, getVocabReviewQueue, listGrammar, listVocab } from '$lib/api.js';

  let queue = [];
  let vocabQueue = [];
  let totalGrammar = 0;
  let totalVocab = 0;
  let loading = true;
  let error = '';

  onMount(async () => {
    try {
      const [queueRes, vocabQueueRes, grammarRes, vocabRes] = await Promise.all([
        getReviewQueue(),
        getVocabReviewQueue(),
        listGrammar(),
        listVocab()
      ]);
      queue = queueRes.cards ?? [];
      vocabQueue = vocabQueueRes.cards ?? [];
      totalGrammar = grammarRes.length ?? 0;
      totalVocab = vocabRes.length ?? 0;
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head><title>日本語先生 — Dashboard</title></svelte:head>

<main>
  <header>
    <h1>日本語先生</h1>
    <p class="tagline">Japanese Grammar SRS · Powered by Ollama</p>
    <a href="/auth/logout" class="logout-btn">Sign out</a>
  </header>

  {#if loading}
    <p class="state-msg">Loading…</p>
  {:else if error}
    <div class="alert error">
      <strong>Could not connect to backend.</strong><br/>
      Make sure the Go server is running on <code>localhost:8080</code>.<br/>
      <small>{error}</small>
    </div>
  {:else}
    <div class="stats-grid">
      <div class="stat-card">
        <span class="stat-num">{queue.length}</span>
        <span class="stat-label">Grammar due</span>
      </div>
      <div class="stat-card">
        <span class="stat-num">{vocabQueue.length}</span>
        <span class="stat-label">Vocab due</span>
      </div>
      <div class="stat-card">
        <span class="stat-num">{totalGrammar}</span>
        <span class="stat-label">Grammar points</span>
      </div>
      <div class="stat-card">
        <span class="stat-num">{totalVocab}</span>
        <span class="stat-label">Vocab words</span>
      </div>
    </div>

    <div class="nav-grid">
      <a href="/review" class="nav-card" class:disabled={queue.length === 0}>
        <span class="nav-icon">📚</span>
        <strong>Review</strong>
        <span>{queue.length} card{queue.length !== 1 ? 's' : ''} due</span>
      </a>
      <a href="/challenge" class="nav-card">
        <span class="nav-icon">✍️</span>
        <strong>Challenge</strong>
        <span>Write a sentence</span>
      </a>
      <a href="/grammar" class="nav-card">
        <span class="nav-icon">🔍</span>
        <strong>Grammar</strong>
        <span>Browse & add points</span>
      </a>
      <a href="/vocab" class="nav-card">
        <span class="nav-icon">🈶</span>
        <strong>Vocabulary</strong>
        <span>Browse & add words</span>
      </a>
      <a href="/vocab/review" class="nav-card" class:disabled={vocabQueue.length === 0}>
        <span class="nav-icon">🧠</span>
        <strong>Vocab Review</strong>
        <span>{vocabQueue.length} card{vocabQueue.length !== 1 ? 's' : ''} due</span>
      </a>
      <a href="/settings" class="nav-card">
        <span class="nav-icon">⚙️</span>
        <strong>Settings</strong>
        <span>Customize conversation options</span>
      </a>
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
  header { text-align: center; margin-bottom: 2.5rem; position: relative; }
  h1 { font-size: 2.5rem; margin-bottom: 0.25rem; color: #2d3748; }
  .tagline { color: #718096; font-size: 0.95rem; }
  .logout-btn {
    position: absolute;
    top: 0;
    right: 0;
    font-size: 0.8rem;
    color: #718096;
    text-decoration: none;
    padding: 0.25rem 0.5rem;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
  }
  .logout-btn:hover { color: #2d3748; border-color: #cbd5e0; }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
    margin-bottom: 2rem;
  }
  .stat-card {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 1.5rem;
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  .stat-num { font-size: 2.5rem; font-weight: 800; color: #667eea; }
  .stat-label { color: #718096; font-size: 0.9rem; }

  .nav-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
  }
  @media (max-width: 500px) { .nav-grid { grid-template-columns: 1fr; } }

  .nav-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 1.5rem 1rem;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    text-decoration: none;
    color: #2d3748;
    gap: 0.35rem;
    transition: box-shadow 0.15s, transform 0.1s;
  }
  .nav-card:hover:not(.disabled) {
    box-shadow: 0 4px 16px rgba(102,126,234,0.2);
    transform: translateY(-2px);
  }
  .nav-card.disabled { opacity: 0.5; pointer-events: none; }
  .nav-icon { font-size: 2rem; }
  .nav-card strong { font-size: 1.1rem; }
  .nav-card span { font-size: 0.8rem; color: #718096; }

  .alert {
    border-radius: 8px;
    padding: 1rem 1.25rem;
    line-height: 1.6;
  }
  .alert.error { background: #fff5f5; border: 1px solid #fc8181; color: #c53030; }

  .state-msg { text-align: center; color: #a0aec0; padding: 3rem 0; }
</style>
