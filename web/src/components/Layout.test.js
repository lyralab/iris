import React, { act } from 'react';
import { createRoot } from 'react-dom/client';
import { MemoryRouter } from 'react-router-dom';
import Layout from './Layout';

window.IS_REACT_ACT_ENVIRONMENT = true;

let container;
let root;

beforeEach(() => {
    localStorage.clear();
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
const render = () => act(() => root.render(<MemoryRouter><Layout><p>page</p></Layout></MemoryRouter>));
const shell = () => container.querySelector('.app-shell');
const toggle = () => container.querySelector('.sidebar-toggle');
const press = (key, target = document.body) =>
    act(() => target.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true })));

test('toggle collapses and expands the sidebar and remembers the choice', () => {
    render();
    expect(shell().classList.contains('sidebar-collapsed')).toBe(false);
    expect(toggle().getAttribute('aria-label')).toBe('Collapse sidebar');

    act(() => toggle().click());
    expect(shell().classList.contains('sidebar-collapsed')).toBe(true);
    expect(toggle().getAttribute('aria-label')).toBe('Expand sidebar');
    expect(localStorage.getItem('iris.sidebar.collapsed')).toBe('1');

    act(() => toggle().click());
    expect(shell().classList.contains('sidebar-collapsed')).toBe(false);
    expect(localStorage.getItem('iris.sidebar.collapsed')).toBe('0');
});

test('starts collapsed when that was saved', () => {
    localStorage.setItem('iris.sidebar.collapsed', '1');
    render();
    expect(shell().classList.contains('sidebar-collapsed')).toBe(true);
});

test('"[" toggles, except while typing in a field', () => {
    render();
    press('[');
    expect(shell().classList.contains('sidebar-collapsed')).toBe(true);

    const input = document.createElement('input');
    container.appendChild(input);
    press('[', input);
    expect(shell().classList.contains('sidebar-collapsed')).toBe(true);
});
