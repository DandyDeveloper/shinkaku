const SETTINGS_KEY = 'shinkaku.settings.v1';

export const DEFAULT_SETTINGS = {
  includeConversationFurigana: false
};

export function getSettings() {
  if (typeof localStorage === 'undefined') {
    return { ...DEFAULT_SETTINGS };
  }

  try {
    const raw = localStorage.getItem(SETTINGS_KEY);
    if (!raw) return { ...DEFAULT_SETTINGS };
    const parsed = JSON.parse(raw);
    return { ...DEFAULT_SETTINGS, ...parsed };
  } catch {
    return { ...DEFAULT_SETTINGS };
  }
}

export function saveSettings(nextSettings) {
  if (typeof localStorage === 'undefined') return;
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(nextSettings));
}

export function updateSettings(patch) {
  const next = { ...getSettings(), ...patch };
  saveSettings(next);
  return next;
}