/**
 * @fileoverview JSDoc type definitions mirroring the Go backend models.
 */

/**
 * @typedef {Object} GrammarPoint
 * @property {number} id
 * @property {string} jlpt_level - e.g. "N5", "N4", "N3", "N2", "N1"
 * @property {string} pattern - e.g. "〜てもいい"
 * @property {string} meaning - English meaning
 * @property {string} example_jp - Japanese example sentence
 * @property {string} example_en - English translation
 * @property {string} notes
 * @property {string} source - e.g. "hanabira", "manual"
 * @property {string} created_at - ISO date string
 */

/**
 * @typedef {Object} ReviewCard
 * @property {number} id
 * @property {number} grammar_point_id
 * @property {number} interval - days until next review
 * @property {number} repetitions - consecutive correct reviews
 * @property {number} e_factor - ease factor, min 1.3
 * @property {string} due_date - ISO date string
 * @property {string|null} last_reviewed - ISO date string or null
 */

/**
 * @typedef {ReviewCard & { grammar_point: GrammarPoint }} QueueItem
 */

/**
 * @typedef {Object} LLMGrade
 * @property {boolean} correct
 * @property {string} explanation
 * @property {string} [correction]
 * @property {string} [natural_alternative]
 * @property {string} [raw_response]
 */

export {};
