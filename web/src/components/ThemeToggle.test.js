import React, { act } from 'react';
import { createRoot } from 'react-dom/client';
import ThemeToggle from './ThemeToggle';

window.IS_REACT_ACT_ENVIRONMENT = true;

let container;
let root;

beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
    window.matchMedia = jest.fn().mockReturnValue({ matches: false });
    container = document.createElement('div');
    document.body.appendChild(container);
    root = createRoot(container);
});

afterEach(() => {
    act(() => root.unmount());
    container.remove();
});

// eslint-disable-next-line testing-library/no-unnecessary-act -- raw react-dom, no RTL installed
const render = () => act(() => root.render(<ThemeToggle />));
const button = () => container.querySelector('button');

test('clicking switches light -> dark -> light and persists each choice', () => {
    render();
    expect(button().getAttribute('aria-label')).toBe('Switch to dark theme');

    act(() => button().click());
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    expect(localStorage.getItem('theme')).toBe('dark');
    expect(button().getAttribute('aria-label')).toBe('Switch to light theme');

    act(() => button().click());
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
    expect(localStorage.getItem('theme')).toBe('light');
});

test('starts in dark when the OS prefers dark', () => {
    window.matchMedia = jest.fn().mockReturnValue({ matches: true });
    render();
    expect(button().getAttribute('aria-label')).toBe('Switch to light theme');
});
