/**
 * API client — thin fetch wrapper pointing at the Go backend.
 * During development, Vite proxies /api/* to http://localhost:8080.
 * In production (static build), the Go server should serve the built frontend
 * and handle /api/* directly.
 */

const BASE = '/api';

async function request(method, path, body) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' }
  };
  if (body !== undefined) {
    opts.body = JSON.stringify(body);
  }

  const res = await fetch(BASE + path, opts);

  if (!res.ok) {
    if (res.status === 401) {
      window.location.href = '/auth/login';
      return;
    }
    let errMsg = `HTTP ${res.status}`;
    try {
      const errBody = await res.json();
      errMsg = errBody.error ?? errMsg;
    } catch (_) {}
    throw new Error(errMsg);
  }

  // 204 No Content
  if (res.status === 204) return null;
  return res.json();
}

// --- Grammar Points ---

/** @returns {Promise<import('./types').GrammarPoint[]>} */
export function listGrammar(jlptLevel = '') {
  const qs = jlptLevel ? `?jlpt=${encodeURIComponent(jlptLevel)}` : '';
  return request('GET', `/grammar${qs}`);
}

/** @returns {Promise<import('./types').GrammarPoint>} */
export function getGrammar(id) {
  return request('GET', `/grammar/${id}`);
}

/** @returns {Promise<import('./types').GrammarPoint>} */
export function createGrammar(data) {
  return request('POST', '/grammar', data);
}

// --- SRS Review ---

/** @returns {Promise<{cards: import('./types').QueueItem[], total: number}>} */
export function getReviewQueue() {
  return request('GET', '/review/queue');
}

/**
 * Submit a grade (0–5) for a card.
 * @param {number} cardId
 * @param {number} grade
 */
export function submitGrade(cardId, grade) {
  return request('POST', `/review/${cardId}/grade`, { grade });
}

// --- LLM Challenge ---

/**
 * Ask Ollama to grade the user's sentence for a given grammar point.
 * @param {number} grammarPointId
 * @param {string} userSentence
 * @returns {Promise<import('./types').LLMGrade>}
 */
export function gradeChallenge(grammarPointId, userSentence) {
  return request('POST', '/challenge/grade', {
    grammar_point_id: grammarPointId,
    user_sentence: userSentence
  });
}
