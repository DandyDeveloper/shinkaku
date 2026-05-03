<script>
  import { onMount } from 'svelte';
  import { getOnboardingStatus, completeOnboarding } from '$lib/api.js';

  const LEVELS = ['N5', 'N4', 'N3', 'N2', 'N1'];

  let startingJLPT = 'N5';
  let loading = true;
  let submitting = false;
  let error = '';

  onMount(async () => {
    try {
      const status = await getOnboardingStatus();
      if (!status.needs_onboarding) {
        window.location.href = '/';
        return;
      }
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  async function handleSubmit() {
    submitting = true;
    error = '';
    try {
      await completeOnboarding(startingJLPT);
      window.location.href = '/';
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head><title>Welcome — 日本語先生</title></svelte:head>

<main>
  <section class="onboarding-card">
    <h1>Welcome to 日本語先生</h1>
    <p class="intro">Choose your starting JLPT level. We will pre-populate your starter grammar deck with this level.</p>

    {#if loading}
      <p class="state">Checking setup…</p>
    {:else}
      <div class="levels">
        {#each LEVELS as level}
          <label class="level-option">
            <input
              type="radio"
              name="jlpt-level"
              value={level}
              checked={startingJLPT === level}
              on:change={() => startingJLPT = level}
            />
            <span>{level}</span>
          </label>
        {/each}
      </div>

      {#if error}
        <div class="alert error">{error}</div>
      {/if}

      <button class="btn primary" on:click={handleSubmit} disabled={submitting}>
        {submitting ? 'Setting up…' : 'Start Learning'}
      </button>
    {/if}
  </section>
</main>

<style>
  main {
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: 1.5rem;
    background: linear-gradient(135deg, #f7fafc, #edf2f7);
    font-family: system-ui, -apple-system, sans-serif;
  }
  .onboarding-card {
    width: min(680px, 100%);
    border: 1px solid #e2e8f0;
    border-radius: 14px;
    background: #fff;
    padding: 1.5rem;
  }
  h1 {
    margin: 0 0 0.5rem;
    color: #1a202c;
    font-size: 1.7rem;
  }
  .intro {
    margin: 0 0 1rem;
    color: #4a5568;
    line-height: 1.5;
  }
  .levels {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 0.6rem;
    margin-bottom: 1rem;
  }
  .level-option {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 0.55rem 0.6rem;
    color: #2d3748;
    font-weight: 600;
  }
  .state {
    color: #718096;
    margin: 0.6rem 0;
  }
  .alert.error {
    background: #fff5f5;
    border: 1px solid #fed7d7;
    color: #c53030;
    border-radius: 8px;
    padding: 0.65rem 0.75rem;
    margin-bottom: 0.75rem;
  }
  .btn {
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    padding: 0.65rem 1rem;
    cursor: pointer;
    font-weight: 700;
  }
  .btn.primary {
    border-color: #667eea;
    background: #667eea;
    color: #fff;
  }
  .btn:disabled {
    opacity: 0.65;
    cursor: not-allowed;
  }
  @media (max-width: 700px) {
    .levels {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
</style>