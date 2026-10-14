import fs from 'fs';
import path from 'path';
import { getTheme } from './theme';

// public/index.html carries a copy of getTheme() so the theme is set before first paint.
// This keeps the two copies in agreement.
const html = fs.readFileSync(path.join(__dirname, '../../public/index.html'), 'utf8');
const inline = html.match(/<script>([\s\S]*?)<\/script>/)[1];

describe.each([null, 'light', 'dark', 'purple'])('stored=%s', (stored) => {
    test.each([true, false])('osDark=%s', (osDark) => {
        localStorage.clear();
        if (stored) localStorage.setItem('theme', stored);
        window.matchMedia = jest.fn().mockReturnValue({ matches: osDark });
        document.documentElement.removeAttribute('data-theme');

        // eslint-disable-next-line no-new-func
        new Function(inline)();

        expect(document.documentElement.getAttribute('data-theme')).toBe(getTheme());
    });
});
