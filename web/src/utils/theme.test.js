import { resolveTheme, getTheme, setTheme } from './theme';

const mockPrefersDark = (dark) => {
    window.matchMedia = jest.fn().mockReturnValue({ matches: dark });
};

beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
    mockPrefersDark(false);
});

test('a stored choice wins over the OS preference', () => {
    expect(resolveTheme('light', true)).toBe('light');
    expect(resolveTheme('dark', false)).toBe('dark');
});

test('without a stored choice the OS preference decides', () => {
    expect(resolveTheme(null, true)).toBe('dark');
    expect(resolveTheme(null, false)).toBe('light');
});

test('garbage in storage is ignored', () => {
    expect(resolveTheme('purple', true)).toBe('dark');
});

test('getTheme follows the OS when nothing is stored', () => {
    mockPrefersDark(true);
    expect(getTheme()).toBe('dark');
});

test('setTheme persists the choice and tags <html>', () => {
    setTheme('dark');
    expect(localStorage.getItem('theme')).toBe('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    expect(getTheme()).toBe('dark');
});

test('setTheme still applies when storage throws', () => {
    jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw new Error('denied'); });
    setTheme('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    jest.restoreAllMocks();
});
