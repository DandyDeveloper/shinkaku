<script>
  /**
   * Emits a 'grade' event with value 0–5 when a button is pressed.
   * @type {import('svelte').EventDispatcher}
   */
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  export let disabled = false;

  const grades = [
    { value: 0, label: 'Blackout',  sublabel: 'Complete blank',        color: '#fc8181' },
    { value: 1, label: 'Wrong',     sublabel: 'Recalled after seeing', color: '#f6ad55' },
    { value: 2, label: 'Hard',      sublabel: 'Wrong but close',       color: '#fbd38d' },
    { value: 3, label: 'OK',        sublabel: 'Correct w/ difficulty', color: '#68d391' },
    { value: 4, label: 'Good',      sublabel: 'Correct w/ hesitation', color: '#4fd1c5' },
    { value: 5, label: 'Perfect',   sublabel: 'Instant recall',        color: '#76e4f7' },
  ];

  function handleGrade(value) {
    if (!disabled) dispatch('grade', value);
  }
</script>

<div class="grade-grid">
  {#each grades as g}
    <button
      class="grade-btn"
      style="--accent: {g.color}"
      on:click={() => handleGrade(g.value)}
      {disabled}
      title="Grade {g.value}: {g.sublabel}"
    >
      <span class="grade-num">{g.value}</span>
      <span class="grade-label">{g.label}</span>
      <span class="grade-sub">{g.sublabel}</span>
    </button>
  {/each}
</div>

<style>
  .grade-grid {
    display: grid;
    grid-template-columns: repeat(6, 1fr);
    gap: 0.5rem;
    max-width: 640px;
    margin: 0 auto;
  }
  @media (max-width: 600px) {
    .grade-grid { grid-template-columns: repeat(3, 1fr); }
  }
  .grade-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0.75rem 0.25rem;
    border: 2px solid var(--accent);
    border-radius: 8px;
    background: white;
    cursor: pointer;
    transition: background 0.15s, transform 0.1s;
    gap: 0.2rem;
  }
  .grade-btn:hover:not(:disabled) {
    background: var(--accent);
    transform: translateY(-2px);
  }
  .grade-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .grade-num {
    font-size: 1.5rem;
    font-weight: 800;
    color: #2d3748;
  }
  .grade-label {
    font-size: 0.8rem;
    font-weight: 600;
    color: #4a5568;
  }
  .grade-sub {
    font-size: 0.65rem;
    color: #718096;
    text-align: center;
    line-height: 1.2;
  }
</style>
