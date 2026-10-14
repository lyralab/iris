const KEY = 'theme';

export const resolveTheme = (stored, prefersDark) =>
    stored === 'light' || stored === 'dark' ? stored : (prefersDark ? 'dark' : 'light');

const read = () => {
    try { return localStorage.getItem(KEY); } catch { return null; }
};

export const getTheme = () =>
    resolveTheme(read(), !!window.matchMedia?.('(prefers-color-scheme: dark)').matches);

export const setTheme = (theme) => {
    document.documentElement.setAttribute('data-theme', theme);
    try { localStorage.setItem(KEY, theme); } catch { /* private mode: applies for this session only */ }
};
